package task

import (
	"fmt"
	"sync/atomic"

	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/stathat/consistent"
)

var (
	WorkerIdGenerator int32     = 0
	WorkerInitialCnt            = 8
	WorkerVirtualNum            = 8
	TaskExecutor      *Executor = NewExecutor()
)

type WorkerGroup struct {
	name       string
	c          *consistent.Consistent
	e          *Executor
	workers    map[string]*Worker
	fixWorkers map[string]*Worker
}

type Executor struct {
	*basic.Object
	c          *consistent.Consistent
	workers    map[string]*Worker
	fixWorkers map[string]*Worker
	group      map[string]*WorkerGroup
}

func NewExecutor() *Executor {
	e := &Executor{
		c:          consistent.New(),
		workers:    make(map[string]*Worker),
		fixWorkers: make(map[string]*Worker),
		group:      make(map[string]*WorkerGroup),
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
	e.addWorker(Config.Worker.WorkerCnt)

	core.LaunchChild(TaskExecutor.Object)
}

func (e *Executor) addWorker(workerCnt int) {
	for i := 0; i < workerCnt; i++ {
		id := atomic.AddInt32(&WorkerIdGenerator, 1)
		w := &Worker{
			Object: basic.NewObject(int(id),
				fmt.Sprintf("worker_%d", id),
				Config.Worker.Options,
				nil),
		}

		w.UserData = w
		e.LaunchChild(w.Object)
		e.c.Add(w.Name)
		e.workers[w.Name] = w
	}
}

func (e *Executor) getWorker(name string) *Worker {
	if w, exist := e.workers[name]; exist {
		return w
	}
	return nil
}

func (e *Executor) getFixWorker(name string) *Worker {
	if w, exist := e.fixWorkers[name]; exist {
		return w
	}
	return nil
}

func (e *Executor) addFixWorker(name string) *Worker {
	logger.Logger.Infof("Executor.AddFixWorker(%v)", name)
	id := atomic.AddInt32(&WorkerIdGenerator, 1)
	w := &Worker{
		Object: basic.NewObject(int(id),
			name,
			Config.Worker.Options,
			nil),
	}

	w.UserData = w
	e.LaunchChild(w.Object)
	e.fixWorkers[name] = w
	return w
}

func (e *Executor) getGroup(gname string) (*WorkerGroup, bool) {
	wg, ok := e.group[gname]
	return wg, ok
}

func (e *Executor) AddGroup(gname string) *WorkerGroup {
	wg := &WorkerGroup{
		e:          e,
		c:          consistent.New(),
		name:       gname,
		workers:    make(map[string]*Worker),
		fixWorkers: make(map[string]*Worker),
	}

	for i := 0; i < Config.Worker.WorkerCnt; i++ {
		id := atomic.AddInt32(&WorkerIdGenerator, 1)
		w := &Worker{
			Object: basic.NewObject(int(id),
				fmt.Sprintf("g_%v_worker_%d", gname, id),
				Config.Worker.Options,
				nil),
		}

		w.UserData = w
		e.LaunchChild(w.Object)
		wg.c.Add(w.Name)
		wg.workers[w.Name] = w
	}

	e.group[gname] = wg
	return wg
}

func (wg *WorkerGroup) getWorker(name string) *Worker {
	if w, exist := wg.workers[name]; exist {
		return w
	}
	return nil
}

func (wg *WorkerGroup) getFixWorker(name string) *Worker {
	if w, exist := wg.fixWorkers[name]; exist {
		return w
	}
	return nil
}

func (wg *WorkerGroup) addFixWorker(name string) *Worker {
	logger.Logger.Infof("WorkerGroup(%v).AddFixWorker(%v)", wg.name, name)
	id := atomic.AddInt32(&WorkerIdGenerator, 1)
	w := &Worker{
		Object: basic.NewObject(int(id),
			fmt.Sprintf("%s_%s", wg.name, name),
			Config.Worker.Options,
			nil),
	}

	w.UserData = w
	wg.e.LaunchChild(w.Object)
	wg.fixWorkers[name] = w
	return w
}
