// transnode
package transact

import (
	"time"

	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/timer"
	"sync"
	"sync/atomic"
)

const (
	///transact execute result
	TransResult_Success int = iota
	TransResult_Failed
	TransResult_TimeOut
	TransResult_Max
)
const (
	///transact result
	TransExeResult_Success TransExeResult = iota
	TransExeResult_Failed
	TransExeResult_Yield
	TransExeResult_NullPointer
	TransExeResult_NoStart
	TransExeResult_NoSetHandler
	TransExeResult_ChildNodeNotExist
	TransExeResult_ChildNodeRepeateRet
	TransExeResult_AsynFailed
	TransExeResult_HadDone
	TransExeResult_StartChildFailed
	TransExeResult_UnsafeExecuteEnv
)
const (
	///transact owner type
	TransOwnerType_Invalid TransOwnerType = iota
	TransOwnerType_Max
)
const (
	///transact command type
	TransCmd_Invalid TransCmd = iota
	TransCmd_Commit
	TransCmd_RollBack
)
const (
	///transact
	TransRootNodeLevel     int           = 0
	DefaultTransactTimeout time.Duration = 30 * time.Second
)

var (
	TransNodeIDNil = TransNodeID(0)
	transStats     = new(sync.Map)
)

type TransExeResult int
type TransOwnerType int
type TransCmd int
type TransNodeID int64
type TransNodeParam struct {
	TId        TransNodeID
	Tt         TransType
	Ot         TransOwnerType
	Tct        TransactCommitPolicy
	Oid        int
	SkeletonID int
	LevelNo    int
	AreaID     int
	TimeOut    time.Duration
}

type TransResult struct {
	RetCode  int
	RetFiels interface{}
}

const (
	TransStatsOp_Exe = iota
	TransStatsOp_Rollback
	TransStatsOp_Commit
	TransStatsOp_Yiled
	TransStatsOp_Resume
	TransStatsOp_Timeout
)

type TransStats struct {
	ExecuteTimes    int64
	RollbackTimes   int64
	CommitTimes     int64
	TimeoutTimes    int64
	YieldTimes      int64
	ResumeTimes     int64
	TotalRuningTime int64
	MaxRuningTime   int64
}

func (stats *TransStats) incStats(op int) {
	switch op {
	case TransStatsOp_Exe:
		atomic.AddInt64(&stats.ExecuteTimes, 1)
	case TransStatsOp_Rollback:
		atomic.AddInt64(&stats.RollbackTimes, 1)
	case TransStatsOp_Commit:
		atomic.AddInt64(&stats.CommitTimes, 1)
	case TransStatsOp_Yiled:
		atomic.AddInt64(&stats.YieldTimes, 1)
	case TransStatsOp_Resume:
		atomic.AddInt64(&stats.ResumeTimes, 1)
	case TransStatsOp_Timeout:
		atomic.AddInt64(&stats.TimeoutTimes, 1)
	}
}

type TransCallback func(*TransNode)
type TransBrotherNotify func(*TransNode, TransExeResult)
type TransNode struct {
	TransEnv     *TransCtx
	TransRep     *TransResult
	MyTnp        *TransNodeParam
	ParentTnp    *TransNodeParam
	ownerObj     *basic.Object
	Childs       map[TransNodeID]*TransNodeParam
	finChild     map[TransNodeID]interface{}
	timeHandle   timer.TimerHandle
	handler      TransHandler
	AsynCallback TransCallback
	brothers     map[*TransNode]TransBrotherNotify
	createTime   time.Time
	start        bool
	yield        bool
	resume       bool
	done         bool
	owner        *transactCoordinater
	ud           interface{}
}

func (this *TransNode) incStats(op int) {
	if s, exist := transStats.Load(this.MyTnp.Tt); exist {
		if stats, ok := s.(*TransStats); ok {
			stats.incStats(op)
		}
	} else {
		stats := &TransStats{}
		transStats.Store(this.MyTnp.Tt, stats)
		stats.incStats(op)
	}
}

func (this *TransNode) statsRuningTime() {
	if s, exist := transStats.Load(this.MyTnp.Tt); exist {
		if stats, ok := s.(*TransStats); ok {
			runingTime := int64(time.Now().Sub(this.createTime) / time.Millisecond)
			if runingTime > stats.MaxRuningTime {
				stats.MaxRuningTime = runingTime
			}
			stats.TotalRuningTime += runingTime
		}
	}
}

func (this *TransNode) execute(ud interface{}) TransExeResult {
	if this == nil {
		return TransExeResult_NullPointer
	}
	if this.handler == nil {
		return TransExeResult_NoSetHandler
	}
	this.start = true
	ret := this.handler.OnExcute(this, ud)
	this.incStats(TransStatsOp_Exe)
	if ret == TransExeResult_Yield {
		return this.Yield()
	}

	return this.doneExecRet(ret)
}

func (this *TransNode) doneExecRet(ter TransExeResult) TransExeResult {
	if this.done {
		return TransExeResult_HadDone
	}
	if ter == TransExeResult_Success {
		if len(this.Childs) == len(this.finChild) {
			if this.MyTnp.LevelNo <= TransRootNodeLevel {
				return this.commit()
			} else {
				if Config.tcs != nil {
					this.TransRep.RetCode = TransResult_Success
					Config.tcs.SendTransResult(this.ParentTnp, this.MyTnp, this.TransRep)
				}
				if this.MyTnp.Tct == TransactCommitPolicy_SelfDecide {
					return this.commit()
				}
			}
		}
	} else {
		if this.MyTnp.LevelNo == TransRootNodeLevel {
			return this.rollback(TransNodeIDNil)
		} else {
			if Config.tcs != nil {
				this.TransRep.RetCode = TransResult_Failed
				Config.tcs.SendTransResult(this.ParentTnp, this.MyTnp, this.TransRep)
			}
			return this.rollback(TransNodeIDNil)
		}
	}
	return TransExeResult_Success
}

func (this *TransNode) commit() TransExeResult {
	defer this.owner.releaseTrans(this)

	if this == nil {
		return TransExeResult_NullPointer
	}
	if !this.start {
		return TransExeResult_NoStart
	}
	if this.handler == nil {
		return TransExeResult_NoSetHandler
	}
	if this.done {
		return TransExeResult_HadDone
	}

	defer this.notifyBrother(TransExeResult_Success)

	this.done = true
	this.handler.OnCommit(this)
	this.incStats(TransStatsOp_Commit)
	this.statsRuningTime()
	if len(this.Childs) > 0 && Config.tcs != nil {
		for _, v := range this.Childs {
			if v.Tct == TransactCommitPolicy_TwoPhase {
				Config.tcs.SendCmdToTransNode(v, TransCmd_Commit)
			}
		}
	}

	return TransExeResult_Success
}

func (this *TransNode) rollback(exclude TransNodeID) TransExeResult {
	defer this.owner.releaseTrans(this)

	if this == nil {
		return TransExeResult_NullPointer
	}
	if !this.start {
		return TransExeResult_NoStart
	}
	if this.handler == nil {
		return TransExeResult_NoSetHandler
	}
	if this.done {
		return TransExeResult_HadDone
	}

	defer this.notifyBrother(TransExeResult_Failed)

	this.done = true
	this.handler.OnRollBack(this)
	this.incStats(TransStatsOp_Rollback)
	this.statsRuningTime()
	if len(this.Childs) > 0 && Config.tcs != nil {
		for k, v := range this.Childs {
			if k != exclude && v.Tct == TransactCommitPolicy_TwoPhase {
				Config.tcs.SendCmdToTransNode(v, TransCmd_RollBack)
			}
		}
	}

	return TransExeResult_Success
}

func (this *TransNode) timeout() TransExeResult {
	if this == nil {
		return TransExeResult_NullPointer
	}
	if !this.start {
		return TransExeResult_NoStart
	}
	if this.handler == nil {
		return TransExeResult_NoSetHandler
	}
	if this.done {
		return TransExeResult_HadDone
	}
	if this.MyTnp.LevelNo > TransRootNodeLevel {
		if Config.tcs != nil {
			this.TransRep.RetCode = TransResult_TimeOut
			Config.tcs.SendTransResult(this.ParentTnp, this.MyTnp, this.TransRep)
		}
	}
	this.incStats(TransStatsOp_Timeout)
	this.rollback(TransNodeIDNil)
	return TransExeResult_Success
}

func (this *TransNode) childTransRep(child TransNodeID, retCode int, ud interface{}) TransExeResult {
	if this == nil {
		return TransExeResult_NullPointer
	}
	if this.handler == nil {
		return TransExeResult_NoSetHandler
	}
	if !this.start {
		return TransExeResult_NoStart
	}
	if this.done {
		return TransExeResult_HadDone
	}
	if _, exist := this.Childs[child]; !exist {
		return TransExeResult_ChildNodeNotExist
	}
	if this.finChild == nil {
		this.finChild = make(map[TransNodeID]interface{})
	}
	if _, exist := this.finChild[child]; exist {
		return TransExeResult_ChildNodeRepeateRet
	}
	this.finChild[child] = ud
	ret := this.handler.OnChildTransRep(this, child, retCode, ud)
	if retCode == TransResult_Success && ret == TransExeResult_Success {
		// the child nodes are returned and also run their own end (note: they may be executed asynchronously)
		if len(this.Childs) == len(this.finChild) && this.yield == this.resume {
			if this.MyTnp.LevelNo == TransRootNodeLevel {
				this.commit()
			} else {
				if Config.tcs != nil {
					this.TransRep.RetCode = retCode
					Config.tcs.SendTransResult(this.ParentTnp, this.MyTnp, this.TransRep)
				}
				if this.MyTnp.Tct == TransactCommitPolicy_SelfDecide {
					this.commit()
				}
			}
		}
	} else {
		// They are not the root, then the parent would like to report fails
		if this.MyTnp.LevelNo > TransRootNodeLevel {
			if Config.tcs != nil {
				this.TransRep.RetCode = retCode
				Config.tcs.SendTransResult(this.ParentTnp, this.MyTnp, this.TransRep)
			}
		}
		var exclude TransNodeID
		if retCode != TransResult_Success {
			exclude = child
		}
		// Sub-transaction fails or times out or the results were not satisfactory, timing optimization, advance RollBack
		this.rollback(exclude)
	}

	return TransExeResult_Success
}

func (this *TransNode) StartChildTrans(tnp *TransNodeParam, ud interface{}, timeout time.Duration) TransExeResult {
	if this.done {
		return TransExeResult_HadDone
	}

	tnp.TId = this.owner.spawnTransNodeID()
	tnp.TimeOut = timeout
	tnp.LevelNo = this.MyTnp.LevelNo + 1

	if this.Childs == nil {
		this.Childs = make(map[TransNodeID]*TransNodeParam)
	}
	this.Childs[tnp.TId] = tnp
	if Config.tcs != nil {
		Config.tcs.SendTransStart(this.MyTnp, tnp, ud)
	}
	return TransExeResult_Success
}

func (this *TransNode) GetChildTransParam(childid TransNodeID) *TransNodeParam {
	if v, exist := this.Childs[childid]; exist {
		return v
	}
	return nil
}

func (this *TransNode) Yield() TransExeResult {
	this.yield = true
	SendTranscatYield(this)
	this.incStats(TransStatsOp_Yiled)
	return TransExeResult_Success
}

func (this *TransNode) Resume() TransExeResult {
	this.resume = true
	SendTranscatResume(this)
	this.incStats(TransStatsOp_Resume)
	return TransExeResult_Success
}

func (this *TransNode) Go(obj *basic.Object) TransExeResult {
	this.ownerObj = obj
	return this.execute(this.ud)
}

func (this *TransNode) checkExeOver() {
	if this.resume == this.yield {
		if this.AsynCallback != nil {
			this.AsynCallback(this)
		}
		if this.done == false {
			var ter TransExeResult
			if this.TransRep.RetCode == TransResult_Success {
				ter = TransExeResult_Success
			} else {
				ter = TransExeResult_AsynFailed
			}
			this.doneExecRet(ter)
		}
	}
}

func (this *TransNode) MakeBrotherWith(brother *TransNode, tbn TransBrotherNotify) {
	if this.brothers == nil {
		this.brothers = make(map[*TransNode]TransBrotherNotify)
	}
	this.brothers[brother] = tbn
}

func (this *TransNode) notifyBrother(ter TransExeResult) {
	for k, v := range this.brothers {
		v(k, ter)
	}
}

func Stats() map[int]TransStats {
	stats := make(map[int]TransStats)
	transStats.Range(func(k, v interface{}) bool {
		if s, ok := v.(*TransStats); ok {
			d := *s
			stats[int(k.(TransType))] = d
		}
		return true
	})
	return stats
}
