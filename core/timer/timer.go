package timer

import (
	"container/heap"
	"time"

	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
)

var (
	TimerHandleGenerator uint32      = 1
	InvalidTimerHandle   TimerHandle = 0
	TimerModule          *TimerMgr   = NewTimerMgr()
)

type TimerMgr struct {
	*basic.Object
	tq *TimerQueue
}

func NewTimerMgr() *TimerMgr {
	tm := &TimerMgr{
		tq: NewTimerQueue(),
	}

	return tm
}

func (tm *TimerMgr) Start() {
	logger.Logger.Trace("Timer Start")
	defer logger.Logger.Trace("Timer Start [ok]")

	tm.Object = basic.NewObject(core.ObjId_TimerId,
		"timer",
		Config.Options,
		tm)
	tm.UserData = tm

	core.LaunchChild(TimerModule.Object)
}

func (tm *TimerMgr) TimerCount() int {
	return tm.tq.Len()
}

func (tm *TimerMgr) OnTick() {
	nowTime := time.Now()
	for {
		if tm.tq.Len() > 0 {
			t := heap.Pop(tm.tq)
			if te, ok := t.(*TimerEntity); ok {
				if !te.stoped && te.next.Before(nowTime) {
					if te.times > 0 {
						te.times--
					}
					//Avoid async stop timer failed
					if te.times != 0 {
						te.next = te.next.Add(te.interval)
						heap.Push(tm.tq, te)
					}
					if !SendTimeout(te) {
						if v, ok := tm.tq.ref[te.h]; ok {
							heap.Remove(tm.tq, v)
						}
					}
				} else {
					if !te.stoped {
						heap.Push(tm.tq, te)
					}
					return
				}
			}
		} else {
			return
		}
	}
}

func (tm *TimerMgr) OnStart() {}

func (tm *TimerMgr) OnStop() {}
