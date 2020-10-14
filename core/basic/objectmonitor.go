package basic

var StatsWatchMgr IStatsWatchMgr

type ObjectMonitor struct {
}

func (om *ObjectMonitor) OnStart(o *Object) {
}

func (om *ObjectMonitor) OnTick(o *Object) {
}

func (om *ObjectMonitor) OnStop(o *Object) {
}

type IStatsWatchMgr interface {
	WatchStart(name string, elementype int) IStatsWatch
}

type IStatsWatch interface {
	Stop()
}

type CmdStats struct {
	PendingCnt int64
	SendCmdCnt int64
	RecvCmdCnt int64
}
