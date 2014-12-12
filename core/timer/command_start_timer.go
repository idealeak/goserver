package timer

import (
	"container/heap"
	"time"

	"github.com/idealeak/goserver/core/basic"
)

type startTimerCommand struct {
	ta       TimerAction
	ud       interface{}
	interval time.Duration
	times    int
	h        TimerHandle
}

func (stc *startTimerCommand) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()

	te := &TimerEntity{
		sink:     o,
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

	return StartTimerByObject(ta, ud, interval, times)
}

func StartTimerByObject(ta TimerAction, ud interface{}, interval time.Duration, times int) (TimerHandle, bool) {
	h := generateTimerHandle()
	ret := TimerModule.SendCommand(
		&startTimerCommand{
			ta:       ta,
			ud:       ud,
			interval: interval,
			times:    times,
			h:        h,
		},
		true)
	return h, ret
}
