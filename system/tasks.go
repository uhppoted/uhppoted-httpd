package system

import ()

type TaskQ interface {
	Add(task Task)
}

type Task struct {
	f func()
}

type taskQ struct {
	queue chan Task
}

func NewTaskQ() TaskQ {
	q := taskQ{
		queue: make(chan Task, 32),
	}

	q.run()

	return &q
}

func (q taskQ) Add(task Task) {
	q.queue <- task
}

func (q taskQ) run() {
	go func() {
		for task := range q.queue {
			task.f()
		}
	}()
}
