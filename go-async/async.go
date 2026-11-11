package async

import (
	"github.com/shopsilo/go-async/taskqueue"
)

// Version of the unified go-async suite.
const Version = "1.0.0"

// Re-exported types for top-level convenience.
type Task = taskqueue.Task

// NewTaskQueue initializes a task queue engine.
func NewTaskQueue(workers int) (*taskqueue.Queue, error) {
	return taskqueue.New(taskqueue.Config{Workers: workers})
}
