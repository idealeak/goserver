package module

import (
	"container/list"
	"time"

	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/utils"
)

const (
	///module state
	ModuleStateInvalid int = iota
	ModuleStateInit
	ModuleStateRun
	ModuleStateShutdown
	ModuleStateWaitShutdown
	ModuleStateFini
	///other
	ModuleMaxCount = 1000
)

var (
	AppModule = newModuleMgr()
)

type ModuleEntity struct {
	lastTick     time.Time
	tickInterval time.Duration
	priority     int
	module       Module
	quited       bool
}

type PreloadModuleEntity struct {
	priority int
	module   PreloadModule
}

type ModuleMgr struct {
	*basic.Object
	modules       *list.List
	modulesByName map[string]*ModuleEntity
	preloadModule *list.List
	state         int
	waitShutAct   chan interface{}
	waitShutCnt   int
	waitShut      bool
}

func newModuleMgr() *ModuleMgr {
	mm := &ModuleMgr{
		modules:       list.New(),
		preloadModule: list.New(),
		modulesByName: make(map[string]*ModuleEntity),
		waitShutAct:   make(chan interface{}, ModuleMaxCount),
		state:         ModuleStateInvalid,
	}

	return mm
}

func (this *ModuleMgr) RegisteModule(m Module, tickInterval time.Duration, priority int) {
	mentiry := &ModuleEntity{
		lastTick:     time.Now(),
		tickInterval: tickInterval,
		priority:     priority,
		module:       m,
	}

	this.modulesByName[m.ModuleName()] = mentiry

	for e := this.modules.Front(); e != nil; e = e.Next() {
		if me, ok := e.Value.(*ModuleEntity); ok {
			if priority < me.priority {
				this.modules.InsertBefore(mentiry, e)
				return
			}
		}
	}
	this.modules.PushBack(mentiry)
}

func (this *ModuleMgr) UnregisteModule(m Module) {
	for e := this.modules.Front(); e != nil; e = e.Next() {
		if me, ok := e.Value.(*ModuleEntity); ok {
			if me.module == m {
				delete(this.modulesByName, m.ModuleName())
				this.modules.Remove(e)
				return
			}
		}
	}
}

func (this *ModuleMgr) RegistePreloadModule(m PreloadModule, priority int) {
	mentiry := &PreloadModuleEntity{
		priority: priority,
		module:   m,
	}

	for e := this.preloadModule.Front(); e != nil; e = e.Next() {
		if me, ok := e.Value.(*PreloadModuleEntity); ok {
			if priority < me.priority {
				this.preloadModule.InsertBefore(mentiry, e)
				return
			}
		}
	}
	this.preloadModule.PushBack(mentiry)
}

func (this *ModuleMgr) UnregistePreloadModule(m PreloadModule) {
	for e := this.preloadModule.Front(); e != nil; e = e.Next() {
		if me, ok := e.Value.(*PreloadModuleEntity); ok {
			if me.module == m {
				this.preloadModule.Remove(e)
				return
			}
		}
	}
}

func (this *ModuleMgr) Start() *utils.Waitor {
	logger.Logger.Trace("Startup PreloadModules")
	for e := this.preloadModule.Front(); e != nil; e = e.Next() {
		if me, ok := e.Value.(*PreloadModuleEntity); ok {
			me.module.Start()
		}
	}
	logger.Logger.Trace("Startup PreloadModules [ok]")

	this.Object = basic.NewObject(core.ObjId_CoreId,
		"core",
		Config.Options,
		this)
	this.UserData = this
	core.LaunchChild(this.Object)

	this.state = ModuleStateInit

	return basic.Waitor
}

func (this *ModuleMgr) Close() {
	this.state = ModuleStateShutdown
}

func (this *ModuleMgr) init() {
	logger.Logger.Trace("Start Initialize Modules")
	defer logger.Logger.Trace("Start Initialize Modules [ok]")
	for e := this.modules.Front(); e != nil; e = e.Next() {
		if me, ok := e.Value.(*ModuleEntity); ok && !me.quited {
			logger.Logger.Trace(me.module.ModuleName(), " Init...")
			me.module.Init()
			logger.Logger.Trace(me.module.ModuleName(), " Init [ok]")
		}
	}
	this.state = ModuleStateRun
}

func (this *ModuleMgr) update() {
	nowTime := time.Now()
	for e := this.modules.Front(); e != nil; e = e.Next() {
		if me, ok := e.Value.(*ModuleEntity); ok && !me.quited {
			me.safeUpt(nowTime)
		}
	}
}

func (this *ModuleMgr) shutdown() {
	if this.waitShut {
		return
	}
	this.waitShut = true

	for e := this.modules.Front(); e != nil; e = e.Next() {
		if me, ok := e.Value.(*ModuleEntity); ok {
			me.safeShutdown(this.waitShutAct)
			this.waitShutCnt++
		}
	}

	this.state = ModuleStateWaitShutdown
}

func (this *ModuleMgr) checkShutdown() bool {
	select {
	case param := <-this.waitShutAct:
		logger.Logger.Trace(param, " shutdowned")
		if name, ok := param.(string); ok {
			me := this.getModuleEntityByName(name)
			if me != nil {
				me.quited = true
			}
		}
		this.waitShutCnt--
	default:
	}
	if this.waitShutCnt == 0 {
		this.state = ModuleStateFini
		return true
	}
	this.update()
	return false
}

func (this *ModuleMgr) tick() {

	switch this.state {
	case ModuleStateInit:
		this.init()
	case ModuleStateRun:
		this.update()
	case ModuleStateShutdown:
		this.shutdown()
	case ModuleStateWaitShutdown:
		this.checkShutdown()
	case ModuleStateFini:
		this.fini()
	}
}

func (this *ModuleMgr) fini() {
	core.Terminate(this.Object)
	this.state = ModuleStateInvalid
}

func (this *ModuleMgr) getModuleEntityByName(name string) *ModuleEntity {
	if me, exist := this.modulesByName[name]; exist {
		return me
	}
	return nil
}

func (this *ModuleMgr) GetModuleByName(name string) Module {
	if me, exist := this.modulesByName[name]; exist {
		return me.module
	}
	return nil
}

func (this *ModuleEntity) safeUpt(nowTime time.Time) {
	defer utils.DumpStackIfPanic("ModuleEntity.safeTick")

	if nowTime.Sub(this.lastTick) > this.tickInterval {
		this.module.Update()
		this.lastTick = nowTime
	}
}

func (this *ModuleEntity) safeShutdown(shutWaitAck chan<- interface{}) {
	defer utils.DumpStackIfPanic("ModuleEntity.safeShutdown")
	this.module.Shutdown(shutWaitAck)
}

func (this *ModuleMgr) OnStart() {}
func (this *ModuleMgr) OnStop()  {}
func (this *ModuleMgr) OnTick() {
	this.tick()
}

func RegistePreloadModule(m PreloadModule, priority int) {
	AppModule.RegistePreloadModule(m, priority)
}

func RegisteModule(m Module, tickInterval time.Duration, priority int) {
	AppModule.RegisteModule(m, tickInterval, priority)
}

func Start() *utils.Waitor {
	return AppModule.Start()
}

func Stop() {
	AppModule.Close()
}
