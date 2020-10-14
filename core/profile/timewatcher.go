package profile

import "time"

type TimeWatcher struct {
	name       string    //模块名称
	elementype int       //类型
	tStart     time.Time //开始时间
	next       *TimeWatcher
}

func newTimeWatcher(name string, elementype int) *TimeWatcher {
	w := AllocWatcher()
	w.name = name
	w.elementype = elementype
	w.tStart = time.Now()
	return w
}

func (this *TimeWatcher) Stop() {
	defer FreeWatcher(this)
	d := time.Now().Sub(this.tStart)
	TimeStatisticMgr.addStatistic(this.name, this.elementype, int64(d))
}
