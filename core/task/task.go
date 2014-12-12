package task

import (
	"time"

	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/container/recycler"
	"github.com/idealeak/goserver/core/utils"
)

type Callable interface {
	Call() interface{}
}

type CompleteNotify interface {
	Done(interface{})
}

type CallableWrapper func() interface{}

func (cw CallableWrapper) Call() interface{} {
	return cw()
}

type CompleteNotifyWrapper func(interface{})

func (cnw CompleteNotifyWrapper) Done(i interface{}) {
	cnw(i)
}

type Task struct {
	s   *basic.Object
	c   Callable
	n   CompleteNotify
	r   chan interface{}
	env map[interface{}]interface{}
}

func New(s *basic.Object, c Callable, n CompleteNotify) *Task {
	t := &Task{
		s: s,
		c: c,
		n: n,
		r: make(chan interface{}, 1),
	}

	if s == nil {
		t.s = core.CoreObject()
	}

	return t
}

func (t *Task) Get() interface{} {
	if t.n != nil {
		panic("Task result by CompleteNotify return")
	}

	return <-t.r
}

func (t *Task) GetWithTimeout(timeout time.Duration) interface{} {
	if timeout == 0 {
		return t.Get()
	} else {
		timer := recycler.GetTimer(timeout)
		defer recycler.GiveTimer(timer)
		select {
		case r, ok := <-t.r:
			if ok {
				return r
			} else {
				return nil
			}
		case <-timer.C:
			return nil
		}
	}
	return nil
}

func (t *Task) GetEnv(k interface{}) interface{} {
	if t.env == nil {
		return nil
	}

	if v, exist := t.env[k]; exist {
		return v
	}
	return nil
}

func (t *Task) PutEnv(k, v interface{}) bool {
	if t.env == nil {
		t.env = make(map[interface{}]interface{})
	}
	if t.env != nil {
		t.env[k] = v
	}

	return true
}

func (t *Task) run() (e error) {
	defer utils.DumpStackIfPanic("Task::run")

	ret := t.c.Call()

	if t.r != nil {
		t.r <- ret
	}

	if t.n != nil {
		SendTaskRes(t.s, t)
	}

	return nil
}

func (t *Task) Start() {
	go t.run()
}

func (t *Task) StartByExecutor(name string) bool {
	return sendTaskReqToExecutor(t, name)
}
