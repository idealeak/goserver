package basic

import (
	"sync/atomic"
	"time"
)

type Cond struct {
	notify  chan struct{}
	countor int32
}

func NewCond(waitor int) *Cond {
	return &Cond{notify: make(chan struct{}, waitor)}
}

func (c *Cond) Wait() {
	atomic.AddInt32(&c.countor, 1)
	defer atomic.AddInt32(&c.countor, -1)

	select {
	case <-c.notify:
	}
}

func (c *Cond) WaitForTimeout(dura time.Duration) bool {
	atomic.AddInt32(&c.countor, 1)
	defer atomic.AddInt32(&c.countor, -1)

	select {
	case <-c.notify:
	case <-time.Tick(dura):
		return true
	}
	return false
}

func (c *Cond) WaitForTick(ticker *time.Ticker) bool {
	atomic.AddInt32(&c.countor, 1)
	defer atomic.AddInt32(&c.countor, -1)

	select {
	case <-c.notify:
	case <-ticker.C:
		return true
	}
	return false
}

func (c *Cond) Signal() {
	select {
	case c.notify <- struct{}{}:
	default:
		return
	}
}

func (c *Cond) Drain() {
	for {
		select {
		case <-c.notify:
		default:
			return
		}
	}
}

func (c *Cond) Broadcast() {
	for atomic.LoadInt32(&c.countor) > 0 {
		c.notify <- struct{}{}
	}
}
