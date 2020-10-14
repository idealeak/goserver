package queue

import (
	"container/list"
	"sync"
	"time"
)

type queueS struct {
	fifo *list.List
	lock *sync.RWMutex
}

func NewQueueS() Queue {
	q := &queueS{
		fifo: list.New(),
		lock: new(sync.RWMutex),
	}
	return q
}

func (q *queueS) Enqueue(i interface{}, timeout time.Duration) bool {
	q.lock.Lock()
	q.fifo.PushBack(i)
	q.lock.Unlock()
	return true
}

func (q *queueS) Dequeue(timeout time.Duration) (interface{}, bool) {
	if q.fifo.Len() == 0 {
		return nil, false
	}

	q.lock.Lock()
	e := q.fifo.Front()
	if e != nil {
		q.fifo.Remove(e)
		q.lock.Unlock()
		return e.Value, true
	}
	q.lock.Unlock()
	return nil, false
}

func (q *queueS) Len() int {
	q.lock.RLock()
	l := q.fifo.Len()
	q.lock.RUnlock()
	return l
}
