package queue

import (
	"container/list"
	"sync"
	"time"
)

type queueS struct {
	fifo *list.List
	lock *sync.Mutex
}

func NewQueueS() Queue {
	q := &queueS{
		fifo: list.New(),
		lock: new(sync.Mutex),
	}
	return q
}

func (q *queueS) Enqueue(i interface{}, timeout time.Duration) bool {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.fifo.PushBack(i)
	return true
}

func (q *queueS) Dequeue(timeout time.Duration) (interface{}, bool) {
	if q.fifo.Len() == 0 {
		return nil, false
	}

	q.lock.Lock()
	defer q.lock.Unlock()

	e := q.fifo.Front()
	if e != nil {
		q.fifo.Remove(e)
		return e.Value, true
	}

	return nil, false
}

func (q *queueS) Len() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.fifo.Len()
}
