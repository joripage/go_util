package taskmanager

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

type TaskManager struct {
	tasks sync.Map // key: string, value: context.CancelFunc
	wg    sync.WaitGroup
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
	s.wg.Add(1)

	go func() {
		defer func() {
			s.tasks.Delete(id)
			s.wg.Done()
		}()

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

func (s *TaskManager) GracefulShutdown(wait bool, timeout time.Duration) {
	// Cancel all tasks
	s.tasks.Range(func(key, value interface{}) bool {
		value.(context.CancelFunc)()
		return true
	})

	if wait {
		done := make(chan struct{})
		go func() {
			s.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			log.Println("All tasks completed gracefully")
		case <-time.After(timeout):
			log.Println("Graceful shutdown timed out")
		}
	} else {
		log.Println("Graceful shutdown triggered without waiting")
	}
}
