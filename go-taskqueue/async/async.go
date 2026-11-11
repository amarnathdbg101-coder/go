package async

import (
	"github.com/amarnathdbg101-coder/go/go-taskqueue/async/taskqueue"
)

const Version = "1.0.0"

type Task = taskqueue.Task

func NewTaskQueue(workers int) (*taskqueue.Queue, error) {
	return taskqueue.New(taskqueue.Config{Workers: workers})
}
