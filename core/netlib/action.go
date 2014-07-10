package netlib

import (
	"fmt"
	"reflect"

	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/profile"
	"github.com/idealeak/goserver/core/utils"
)

type action struct {
	s      *Session
	p      interface{}
	n      string
	packid int
}

func (this *action) do() {
	defer FreeAction(this)
	defer utils.DumpStackIfPanic(fmt.Sprintf("netlib.session.task.do exe error, packet type:%v", reflect.TypeOf(this.p)))

	watch := profile.TimeStatisticMgr.WatchStart(this.n)
	defer watch.Stop()

	h := GetHandler(this.packid)
	if h != nil {
		err := h.Process(this.s, this.p)
		if err != nil {
			logger.Infof("%v process error %v", this.n, err)
		}
	} else {
		logger.Infof("%v not registe handler", this.n)
	}
}
