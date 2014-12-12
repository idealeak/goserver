package task

import (
	"github.com/idealeak/goserver/core/basic"
)

type taskExeCommand struct {
	t *Task
}

func (ttc *taskExeCommand) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()
	return ttc.t.run()
}

func SendTaskExe(o *basic.Object, t *Task) bool {
	return o.SendCommand(&taskExeCommand{t: t}, true)
}
