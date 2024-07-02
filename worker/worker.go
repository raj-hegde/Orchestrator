package worker

import (
	"fmt"
	"orchestrator/task"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Worker struct {
	Name      string
	Queue     queue.Queue
	Db        map[uuid.UUID]task.Task
	TaskCount int
}

func (W *Worker) CollectStats() {
	fmt.Println("I will collect stats")
}

func (W *Worker) RunTask() {
	fmt.Println("I will start or stop a task")
}

func (W *Worker) StartTask() {
	fmt.Println("I will start a task")
}

func (W *Worker) StopTask() {
	fmt.Println("I will stop a task")
}
