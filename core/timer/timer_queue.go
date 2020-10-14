package timer

import (
	"container/heap"
	"sync/atomic"
	"time"

	"github.com/idealeak/goserver/core/basic"
)

type TimerEntity struct {
	sink     *basic.Object
	ud       interface{}
	interval time.Duration
	next     time.Time
	times    int
	ta       TimerAction
	h        TimerHandle
	stoped   bool
}

type TimerQueue struct {
	queue []*TimerEntity
	ref   map[TimerHandle]int
}

func generateTimerHandle() TimerHandle {
	return TimerHandle(atomic.AddUint32(&TimerHandleGenerator, 1))
}

func NewTimerQueue() *TimerQueue {
	tq := &TimerQueue{
		ref: make(map[TimerHandle]int),
	}
	heap.Init(tq)
	return tq
}
func (tq TimerQueue) Len() int {
	return len(tq.queue)
}

func (tq TimerQueue) Less(i, j int) bool {
	return tq.queue[i].next.Before(tq.queue[j].next)
}

func (tq *TimerQueue) Swap(i, j int) {
	tq.queue[i], tq.queue[j] = tq.queue[j], tq.queue[i]
	tq.ref[tq.queue[i].h] = i
	tq.ref[tq.queue[j].h] = j
}

func (tq *TimerQueue) Push(x interface{}) {
	n := len(tq.queue)
	te := x.(*TimerEntity)
	tq.ref[te.h] = n
	tq.queue = append(tq.queue, te)
}

func (tq *TimerQueue) Pop() interface{} {
	old := tq.queue
	n := len(old)
	te := old[n-1]
	delete(tq.ref, te.h)
	tq.queue = old[0 : n-1]
	return te
}
