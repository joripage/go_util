package main

import (
	"context"
	"fmt"
	"time"

	taskmanager "github.com/joripage/go_util/pkg/task_manager"
)

var orders = []string{"order1", "order2", "order3"}

func process(order string) {
	fmt.Println("Processing", order)
	time.Sleep(1 * time.Second) // simulate work
}

func processAllOrders(ctx context.Context) error {
	for _, order := range orders {
		select {
		case <-ctx.Done():
			fmt.Println("Task canceled early")
			return ctx.Err()
		default:
			process(order)
		}
	}
	fmt.Println("Task completed")
	return nil
}

func main() {
	tm := taskmanager.NewTaskManager()

	// --- Pattern 1: Cancel multiple tasks at once ---
	fmt.Println("Pattern 1: Cancel multiple tasks individually")
	_ = tm.StartTask(context.Background(), "task1", processAllOrders)
	_ = tm.StartTask(context.Background(), "task2", processAllOrders)
	time.Sleep(1500 * time.Millisecond)
	tm.StopTask("task1")
	tm.StopTask("task2")
	time.Sleep(2000 * time.Millisecond)

	// --- Pattern 2: Cancel tasks via shared parent context ---
	fmt.Println("\nPattern 2: Cancel all tasks via shared parent context")
	parentCtx, cancelAll := context.WithCancel(context.Background())
	_ = tm.StartTask(parentCtx, "task3", processAllOrders)
	_ = tm.StartTask(parentCtx, "task4", processAllOrders)
	time.Sleep(1500 * time.Millisecond)
	cancelAll() // stops task3 and task4
	time.Sleep(2000 * time.Millisecond)

	// --- Pattern 3: Cancel tasks by tag ---
	fmt.Println("\nPattern 3: Cancel tasks by tag")
	taskTags := map[string]string{}
	startTagged := func(id, tag string) {
		taskTags[id] = tag
		_ = tm.StartTask(context.Background(), id, processAllOrders)
	}
	startTagged("sync1", "sync")
	startTagged("sync2", "sync")
	startTagged("report1", "report")
	time.Sleep(1500 * time.Millisecond)
	// stop all tasks with tag "sync"
	for id, tag := range taskTags {
		if tag == "sync" {
			tm.StopTask(id)
		}
	}
	time.Sleep(2000 * time.Millisecond)

	// --- Pattern 4: Timeout for automatic cancellation ---
	fmt.Println("\nPattern 4: Timeout for automatic cancellation")
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	_ = tm.StartTask(ctxTimeout, "task_with_timeout", processAllOrders)
	time.Sleep(3 * time.Second) // wait to see timeout

	// --- Pattern 5: Graceful shutdown of all tasks ---
	fmt.Println("\nPattern 5: Graceful shutdown of all tasks")
	_ = tm.StartTask(context.Background(), "task5", processAllOrders)
	_ = tm.StartTask(context.Background(), "task6", processAllOrders)
	time.Sleep(1500 * time.Millisecond)
	fmt.Println("Shutting down...")
	tm.GracefulShutdown(true, 3*time.Second)
	time.Sleep(500 * time.Millisecond)
}
