package task

import (
	"fmt"

	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/stathat/consistent"
)

var (
	WorkerIdGenerator int       = 0
	WorkerInitialCnt            = 8
	WorkerVirtualNum            = 8
	TaskExecutor      *Executor = NewExecutor()
)

type Executor struct {
	*basic.Object
	c       *consistent.Consistent
	workers map[string]*Worker
}

func NewExecutor() *Executor {
	e := &Executor{
		c:       consistent.New(),
		workers: make(map[string]*Worker),
	}

	return e
}

func (e *Executor) Start() {
	logger.Logger.Trace("Executor Start")
	defer logger.Logger.Trace("Executor Start [ok]")

	e.Object = basic.NewObject(core.ObjId_ExecutorId,
		"executor",
		Config.Options,
		nil)
	e.c.NumberOfReplicas = WorkerVirtualNum
	e.UserData = e
	e.AddWorker(Config.Worker.WorkerCnt)

	core.LaunchChild(TaskExecutor.Object)
}

func (e *Executor) AddWorker(workerCnt int) {
	for i := 0; i < workerCnt; i++ {
		w := &Worker{
			Object: basic.NewObject(WorkerIdGenerator,
				fmt.Sprintf("worker_%d", WorkerIdGenerator),
				Config.Worker.Options,
				nil),
		}
		WorkerIdGenerator++

		w.UserData = w
		e.LaunchChild(w.Object)
		e.c.Add(w.Name)
		e.workers[w.Name] = w
	}
}

func (e *Executor) GetWorker(name string) *Worker {
	if w, exist := e.workers[name]; exist {
		return w
	}
	return nil
}
