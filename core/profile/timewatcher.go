package profile

import (
	"strings"
	"time"

	"github.com/idealeak/goserver/core/logger"
)

type TimeWatcher struct {
	name   string    //模块名称
	tStart time.Time //开始时间
}

func newTimeWatcher(name string) *TimeWatcher {
	w := AllocWatcher()
	w.name = name
	w.tStart = time.Now()
	return w
}

func (this *TimeWatcher) Stop() {
	defer FreeWatcher(this)
	d := time.Now().Sub(this.tStart)
	if Config.SlowMS > 0 && d >= time.Duration(Config.SlowMS)*time.Millisecond {
		logger.Logger.Warnf("###slow timespan name: %s  take:%s", strings.ToLower(this.name), toS(d))
	}
	TimeStatisticMgr.addStatistic(this.name, d)
}
