package taskmanager

import (
	"context"
	"errors"
	"log"
	"sync"
)

type TaskManager struct {
	tasks sync.Map // key: string, value: context.CancelFunc
}

func NewTaskManager() *TaskManager {
	return &TaskManager{}
}

func (s *TaskManager) HasTask(id string) bool {
	_, ok := s.tasks.Load(id)
	return ok
}

func (s *TaskManager) StartTask(ctx context.Context, id string, fn func(ctx context.Context) error) error {
	if id == "" {
		return ErrInvalidTaskID
	}

	if fn == nil {
		return ErrNilTaskFunc
	}

	if ctx.Err() != nil {
		log.Printf("Context already canceled, task %s not started", id)
		return ctx.Err()
	}

	if cancelFn, ok := s.tasks.Load(id); ok {
		cancelFn.(context.CancelFunc)()
	}

	ctxTask, cancel := context.WithCancel(ctx)
	s.tasks.Store(id, cancel)

	go func() {
		defer s.tasks.Delete(id)

		err := fn(ctxTask)
		if errors.Is(err, context.Canceled) {
			log.Printf("Task %s was canceled", id)
		} else if err != nil {
			log.Printf("Task %s failed: %v", id, err)
		} else {
			log.Printf("Task %s completed successfully", id)
		}
	}()

	return nil
}

func (s *TaskManager) StopTask(id string) bool {
	if cancelFn, ok := s.tasks.Load(id); ok {
		cancelFn.(context.CancelFunc)()
		s.tasks.Delete(id)
		return true
	}
	return false
}
