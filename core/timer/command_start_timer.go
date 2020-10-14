package timer

import (
	"container/heap"
	"time"

	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
)

type startTimerCommand struct {
	src      *basic.Object
	ta       TimerAction
	ud       interface{}
	interval time.Duration
	times    int
	h        TimerHandle
}

func (stc *startTimerCommand) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()

	te := &TimerEntity{
		sink:     stc.src,
		ud:       stc.ud,
		ta:       stc.ta,
		interval: stc.interval,
		times:    stc.times,
		h:        stc.h,
		next:     time.Now().Add(stc.interval),
	}

	heap.Push(TimerModule.tq, te)

	return nil
}

// StartTimer only can be called in main module
func StartTimer(ta TimerAction, ud interface{}, interval time.Duration, times int) (TimerHandle, bool) {
	return StartTimerByObject(core.CoreObject(), ta, ud, interval, times)
}
func AfterTimer(taw TimerActionWrapper, ud interface{}, interval time.Duration) (TimerHandle, bool) {
	var tac = &TimerActionCommon{
		Taw: taw,
	}
	return StartTimerByObject(core.CoreObject(), tac, ud, interval, 1)
}

func StartTimerByObject(src *basic.Object, ta TimerAction, ud interface{}, interval time.Duration, times int) (TimerHandle, bool) {
	h := generateTimerHandle()
	ret := TimerModule.SendCommand(
		&startTimerCommand{
			src:      src,
			ta:       ta,
			ud:       ud,
			interval: interval,
			times:    times,
			h:        h,
		},
		true)
	return h, ret
}
