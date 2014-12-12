package recycler

import (
	"time"
)

const (
	NewTimerDefaultDuration time.Duration = time.Minute
	TimerRecyclerBacklog    int           = 128
)

var TimerRecycler = NewRecycler(
	TimerRecyclerBacklog,
	func() interface{} {
		return time.NewTimer(NewTimerDefaultDuration)
	},
	"timer_recycler",
)

func GetTimer(timeout time.Duration) *time.Timer {
	t := TimerRecycler.Get()
	timer := t.(*time.Timer)
	timer.Reset(timeout)
	return timer
}

func GiveTimer(t *time.Timer) {
	TimerRecycler.Give(t)
}
