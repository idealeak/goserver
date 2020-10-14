package utils

import (
	"sync/atomic"

	"github.com/idealeak/goserver/core/logger"
)

type Waitor struct {
	name    string
	counter int32
	waiters int32
	c       chan string
}

func NewWaitor(name string) *Waitor {
	w := &Waitor{name: name, c: make(chan string, 16)}
	return w
}

func (w *Waitor) Add(name string, delta int) {
	v := atomic.AddInt32(&w.counter, int32(delta))
	if v < 0 {
		panic("negative Waitor counter")
	}
	cnt := atomic.LoadInt32(&w.counter)
	logger.Logger.Debugf("(w *Waitor)(%v:%p) Add(%v,%v) counter(%v)", w.name, w, name, delta, cnt)
}

func (w *Waitor) Wait(name string) {
	v := atomic.AddInt32(&w.waiters, 1)
	if v > 1 {
		panic("only support one waitor")
	}
	cnt := atomic.LoadInt32(&w.waiters)
	logger.Logger.Debugf("(w *Waitor)(%v:%p) Waiter(%v) waiters(%v)", w.name, w, name, cnt)
	for w.counter > 0 {
		dname := <-w.c
		v = atomic.AddInt32(&w.counter, -1)
		cnt = atomic.LoadInt32(&w.counter)
		logger.Logger.Debugf("(w *Waitor)(%v:%p) Waiter(%v) after(%v)done! counter(%v)", w.name, w, name, dname, cnt)
	}
}

func (w *Waitor) Done(name string) {
	w.c <- name
	logger.Logger.Debugf("(w *Waitor)(%v:%p) Done(%v)!!!", w.name, w, name)
}
