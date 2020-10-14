// distributed transcation coordinater
package transact

import (
	"sync"
	"time"

	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/timer"
	"github.com/idealeak/goserver/core/utils"
)

var (
	DTCModule = &transactCoordinater{transPool: make(map[TransNodeID]*TransNode)}
	tta       = &transactTimerAction{}
)

type transactCoordinater struct {
	idGen     utils.AtomicIdGen
	lock      sync.Mutex
	transPool map[TransNodeID]*TransNode
	quit      bool
	reaped    bool
}

func (this *transactCoordinater) ModuleName() string {
	return module.ModuleName_Transact
}

func (this *transactCoordinater) Init() {
}

func (this *transactCoordinater) Update() {
}

func (this *transactCoordinater) Shutdown() {
	if this.quit {
		return
	}

	this.quit = true

	if len(this.transPool) > 0 {
		go this.reapRoutine()
		return
	} else {
		this.destroy()
	}
}

func (this *transactCoordinater) reapRoutine() {
	if this.reaped {
		return
	}

	this.reaped = true

	for len(this.transPool) > 0 {
		time.Sleep(time.Second)
	}

	this.destroy()
}

func (this *transactCoordinater) destroy() {
	module.UnregisteModule(this)
}

func (this *transactCoordinater) releaseTrans(tnode *TransNode) {
	if this == nil || tnode == nil {
		return
	}
	timer.StopTimer(tnode.timeHandle)
	this.delTransNode(tnode)
}

func (this *transactCoordinater) spawnTransNodeID() TransNodeID {
	tid := int64(this.idGen.NextId())
	if Config.tcs != nil {
		tid = int64(Config.tcs.GetAreaID())<<48 | int64(Config.tcs.GetSkeletonID())<<32 | tid
	}
	return TransNodeID(tid)
}

func (this *transactCoordinater) createTransNode(tnp *TransNodeParam, ud interface{}, timeout time.Duration) *TransNode {
	if this == nil || tnp == nil {
		logger.Logger.Warn("transactCoordinater.createTransNode failed, Null Pointer")
		return nil
	}
	if this.quit {
		logger.Logger.Warn("transactCoordinater.createTransNode failed, module shutdowning")
		return nil
	}
	transHandler := GetHandler(tnp.Tt)
	if transHandler == nil {
		logger.Logger.Warnf("transactCoordinater.createTransNode failed, TransNodeParam=%v", *tnp)
		return nil
	}

	if tnp.TId == TransNodeIDNil {
		tnp.TId = this.spawnTransNodeID()
	}

	if Config.tcs != nil {
		tnp.SkeletonID = Config.tcs.GetSkeletonID()
		tnp.AreaID = Config.tcs.GetAreaID()
	}
	tnp.TimeOut = timeout * 3 / 2
	tnode := &TransNode{
		MyTnp:      tnp,
		handler:    transHandler,
		owner:      this,
		TransRep:   &TransResult{},
		TransEnv:   NewTransCtx(),
		ud:         ud,
		createTime: time.Now(),
	}

	this.addTransNode(tnode)

	if h, ok := timer.StartTimer(tta, tnode, tnp.TimeOut, 1); ok {
		tnode.timeHandle = h
	} else {
		return nil
	}
	return tnode
}

func (this *transactCoordinater) StartTrans(tnp *TransNodeParam, ud interface{}, timeout time.Duration) *TransNode {
	if this.quit {
		return nil
	}
	tnode := this.createTransNode(tnp, ud, timeout)
	if tnode == nil {
		return nil
	}
	return tnode
}

func (this *transactCoordinater) ProcessTransResult(tid, childtid TransNodeID, retCode int, ud interface{}) bool {
	tnode := this.getTransNode(tid)
	if tnode == nil {
		return false
	}
	ret := tnode.childTransRep(childtid, retCode, ud)
	if ret != TransExeResult_Success {
		return false
	}
	return true
}

func (this *transactCoordinater) ProcessTransStart(parentTnp, myTnp *TransNodeParam, ud interface{}, timeout time.Duration) bool {
	if this.quit {
		logger.Logger.Warn("transactCoordinater.processTransStart find shutdowning, parent=", parentTnp, " selfparam=", myTnp)
		return false
	}
	tnode := this.createTransNode(myTnp, ud, timeout)
	if tnode == nil {
		return false
	}

	tnode.ParentTnp = parentTnp
	tnode.ownerObj = core.CoreObject()
	ret := tnode.execute(ud)
	if ret != TransExeResult_Success {
		return false
	}
	return true
}

func (this *transactCoordinater) ProcessTransCmd(tid TransNodeID, cmd TransCmd) bool {
	tnode := this.getTransNode(tid)
	if tnode == nil {
		return false
	}

	switch cmd {
	case TransCmd_Commit:
		tnode.commit()
	case TransCmd_RollBack:
		tnode.rollback(TransNodeIDNil)
	}
	return true
}

func (this *transactCoordinater) getTransNode(tid TransNodeID) *TransNode {
	this.lock.Lock()
	defer this.lock.Unlock()
	if v, exist := this.transPool[tid]; exist {
		return v
	}
	return nil
}

func (this *transactCoordinater) addTransNode(tnode *TransNode) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.transPool[tnode.MyTnp.TId] = tnode
}

func (this *transactCoordinater) delTransNode(tnode *TransNode) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.transPool, tnode.MyTnp.TId)
}

func init() {
	module.RegisteModule(DTCModule, time.Hour, 1)
}

func ProcessTransResult(tid, childtid TransNodeID, retCode int, ud interface{}) bool {
	return DTCModule.ProcessTransResult(tid, childtid, retCode, ud)
}

func ProcessTransStart(parentTnp, myTnp *TransNodeParam, ud interface{}, timeout time.Duration) bool {
	return DTCModule.ProcessTransStart(parentTnp, myTnp, ud, timeout)
}

func ProcessTransCmd(tid TransNodeID, cmd TransCmd) bool {
	return DTCModule.ProcessTransCmd(tid, cmd)
}
