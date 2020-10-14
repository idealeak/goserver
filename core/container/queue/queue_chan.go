// queue
package queue

import "time"

type queueC struct {
	c chan interface{}
}

func NewQueueC(backlog int) Queue {
	return &queueC{
		c: make(chan interface{}, backlog),
	}
}

func (q *queueC) Enqueue(i interface{}, timeout time.Duration) bool {
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		select {
		case q.c <- i:
		case <-timer.C:
			return false
		}
	} else {
		q.c <- i
	}

	return true
}

func (q *queueC) Dequeue(timeout time.Duration) (i interface{}, ok bool) {
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		select {
		case i, ok = <-q.c:
			return i, ok
		case <-timer.C:
			return nil, false
		}
	} else {
		select {
		case i, ok = <-q.c:
			return i, ok
		}
	}
	return nil, false
}

func (q *queueC) Len() int {
	return len(q.c)
}
