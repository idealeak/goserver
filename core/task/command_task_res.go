package task

import (
	"github.com/idealeak/goserver/core/basic"
)

type taskResCommand struct {
	t *Task
}

func (trc *taskResCommand) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()
	trc.t.n.Done(<-trc.t.r)
	return nil
}

func SendTaskRes(o *basic.Object, t *Task) bool {
	if o == nil {
		return false
	}
	return o.SendCommand(TaskExecutor.Object, &taskResCommand{t: t}, true)
}
