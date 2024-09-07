package accrual

import (
	"fmt"
	"sync"
)

type Queue struct {
	mu     sync.RWMutex
	buffer []Task
}

func NewQueue() *Queue {
	return &Queue{
		mu:     sync.RWMutex{},
		buffer: make([]Task, 0, 10),
	}
}

func (q *Queue) PopWait() (*Task, bool) {
	// получаем задачу
	q.mu.RLock()
	l := q.buffer
	q.mu.RUnlock()
	if len(l) > 0 {
		fmt.Println("берем из очереди")
		return &l[0], true
	}
	return nil, false
}

func (q *Queue) Push(t *Task) {
	// добавляем задачу
	fmt.Println("кладём 1")
	q.mu.Lock()
	q.buffer = append(q.buffer, *t)
	q.mu.Unlock()
	fmt.Println("кладём в очередь")
}

func (q *Queue) RemoveLastCompleted() {
	q.mu.Lock()
	if len(q.buffer) > 1 {
		q.buffer = q.buffer[1:]
	} else {
		q.buffer = make([]Task, 0, 10)
	}
	q.mu.Unlock()
}
