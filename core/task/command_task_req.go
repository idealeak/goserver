package task

import (
	"errors"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
)

var (
	TaskErr_CannotFindWorker  = errors.New("Cannot find fit worker.")
	TaskErr_TaskExecuteObject = errors.New("Task can only be executed executor")
)

type taskReqCommand struct {
	t *Task
	n string
}

func (trc *taskReqCommand) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()

	workerName, err := TaskExecutor.c.Get(trc.n)
	if err != nil {
		return err
	}
	worker := TaskExecutor.GetWorker(workerName)
	if worker != nil {
		SendTaskExe(worker.Object, trc.t)
	} else {
		return TaskErr_CannotFindWorker
	}

	return nil
}

func sendTaskReqToExecutor(t *Task, name string) bool {
	if t == nil {
		return false
	}
	if t.n != nil && t.s == nil {
		logger.Logger.Error(name, " You must specify the source object task.")
		return false
	}
	return TaskExecutor.SendCommand(t.s, &taskReqCommand{t: t, n: name}, true)
}
