package utils

import (
	"sync/atomic"
)

type Waitor struct {
	counter int32
	waiters int32
	c       chan bool
}

func NewWaitor() *Waitor {
	w := &Waitor{c: make(chan bool, 16)}
	return w
}

func (w *Waitor) Add(delta int) {
	v := atomic.AddInt32(&w.counter, int32(delta))
	if v < 0 {
		panic("negative Waitor counter")
	}
}

func (w *Waitor) Wait() {
	v := atomic.AddInt32(&w.waiters, 1)
	if v > 1 {
		panic("only support one waitor")
	}
	for w.counter > 0 {
		<-w.c
		atomic.AddInt32(&w.counter, -1)
	}
}

func (w *Waitor) Done() {
	w.c <- true
}
