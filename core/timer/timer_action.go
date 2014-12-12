package timer

type TimerHandle uint32

type TimerAction interface {
	OnTimer(h TimerHandle, ud interface{}) bool
}

type TimerActionWrapper func(h TimerHandle, ud interface{}) bool

func (taw TimerActionWrapper) OnTimer(h TimerHandle, ud interface{}) bool {
	return taw(h, ud)
}
