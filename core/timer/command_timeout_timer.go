package timer

import (
	"fmt"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/profile"
	"reflect"
)

type timeoutCommand struct {
	te *TimerEntity
}

func (tc *timeoutCommand) Done(o *basic.Object) error {
	tta := reflect.TypeOf(tc.te.ta)
	watch := profile.TimeStatisticMgr.WatchStart(fmt.Sprintf("/timer/%v/ontimer", tta.Name()), profile.TIME_ELEMENT_TIMER)
	defer func() {
		o.ProcessSeqnum()
		if watch != nil {
			watch.Stop()
		}
	}()
	if tc.te.stoped {
		return nil
	}
	if tc.te.ta.OnTimer(tc.te.h, tc.te.ud) == false {
		tc.te.stoped = true
		if tc.te.times < 0 {
			StopTimer(tc.te.h)
		}
	}
	return nil
}

func SendTimeout(te *TimerEntity) bool {
	if te.sink == nil {
		return false
	}

	return te.sink.SendCommand(&timeoutCommand{te: te}, true)
}
