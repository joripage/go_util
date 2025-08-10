# Task manager

`taskmanager` is a lightweight Go package for running and managing multiple concurrent tasks with:

- Start tasks with a `context.Context`.
- Automatic cancellation of an existing task if a new one with the same ID is started.
- Automatic cleanup of tasks after completion.
- Start a new task via `StartTask`.
- Task status tracking via `HasTask`.
- Stop a running task via `StopTask`.

This implementation uses `sync.Map` for thread-safe storage without manual locking.

---

## Installation

```bash
go get github.com/joripage/go_util/pkg/task_manager
```

## How to use

```go
    tm := taskmanager.NewTaskManager()

    // start a task
    err := tm.StartTask(ctx, "task1", func(ctx context.Context) error {
        return nil
    })

    // check if a task exist
    exist := tm.HasTask("task1")

    // stop a task
    success := tm.StopTask("task1")
```

## Cancel Tasks Gracefully with StartTask

- When running long-running or loop-based tasks, you may want to stop them before completion — for example:
  - User manually stops the task.
  - System shutdown.
  - Timeout or deadline.
- By default, if your task ignores `context.Context`, calling `StopTask` will signal cancellation but **your function will keep running until it finishes.**
- To allow graceful early exit, your task function should check `ctx.Done()` periodically.

### Key points

- Always check ctx.Done() inside loops or between heavy operations.
- Return ctx.Err() when canceled to help the caller know why the task ended.
- StopTask(id) will call the cancel function for that task, triggering your cancellation checks.
- This approach prevents wasted work and frees resources earlier.

### Simple Example

**Before – function without cancellation support:**

```go
// Before – function without cancellation support:
func processAllOrders() error {
    for _, order := range orders {
        process(order)
    }
    return nil
}
```

**After – function with cancellation support:**

```go
// After – function with cancellation support:
func processAllOrders(ctx context.Context) error {
    for _, order := range orders {
        select {
        case <-ctx.Done():
            // Stop early and return cancellation reason
            return ctx.Err()
        default:
            process(order)
        }
    }
    return nil
}
```

**Using with `StartTask`**

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/joripage/go_util/pkg/task_manager"
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
            return ctx.Err()
        default:
            process(order)
        }
    }
    return nil
}

func main() {
    tm := taskmanager.NewTaskManager()
    ctx := context.Background()

    // Start task
    _ = tm.StartTask(ctx, "orders", processAllOrders)

    // Stop after 2.5 seconds
    time.Sleep(2500 * time.Millisecond)
    tm.StopTask("orders")

    // Wait for graceful shutdown
    time.Sleep(500 * time.Millisecond)
}
```

**Expected output:**

```bash
Processing order1
Processing order2
Task orders was canceled
```

(order3 is skipped because the task was canceled before reaching it.)

### Advance example

```golang
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/joripage/go_util/pkg/task_manager"
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
    tm.StartTask(context.Background(), "task1", processAllOrders)
    tm.StartTask(context.Background(), "task2", processAllOrders)
    time.Sleep(1500 * time.Millisecond)
    tm.StopTask("task1")
    tm.StopTask("task2")
    time.Sleep(500 * time.Millisecond)

    // --- Pattern 2: Cancel tasks via shared parent context ---
    fmt.Println("\nPattern 2: Cancel all tasks via shared parent context")
    parentCtx, cancelAll := context.WithCancel(context.Background())
    tm.StartTask(parentCtx, "task3", processAllOrders)
    tm.StartTask(parentCtx, "task4", processAllOrders)
    time.Sleep(1500 * time.Millisecond)
    cancelAll() // stops task3 and task4
    time.Sleep(500 * time.Millisecond)

    // --- Pattern 3: Cancel tasks by tag ---
    fmt.Println("\nPattern 3: Cancel tasks by tag")
    taskTags := map[string]string{}
    startTagged := func(id, tag string) {
    taskTags[id] = tag
        tm.StartTask(context.Background(), id, processAllOrders)
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
    time.Sleep(500 * time.Millisecond)

    // --- Pattern 4: Timeout for automatic cancellation ---
    fmt.Println("\nPattern 4: Timeout for automatic cancellation")
    ctxTimeout, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
    defer cancel()
    tm.StartTask(ctxTimeout, "task_with_timeout", processAllOrders)
    time.Sleep(3 * time.Second) // wait to see timeout

    // --- Pattern 5: Graceful shutdown of all tasks ---
    fmt.Println("\nPattern 5: Graceful shutdown of all tasks")
    tm.StartTask(context.Background(), "task5", processAllOrders)
    tm.StartTask(context.Background(), "task6", processAllOrders)
    time.Sleep(1500 * time.Millisecond)
    fmt.Println("Shutting down...")
    tm.Tasks.Range(func(key, value interface{}) bool {
        value.(context.CancelFunc)()
        return true
    })
    time.Sleep(500 * time.Millisecond)
}

```
