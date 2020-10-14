package basic

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"

	"fmt"
	"github.com/idealeak/goserver/core/container"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/utils"
)

const (
	DefaultQueueBacklog int = 4
)

var (
//Waitor = utils.NewWaitor()
)

//  Base class for need alone goroutine objects
//  that easy to start and when to exit the unified management
//	Feature.
//		establish a tree structure between objects
//		asynchronous message queue
type Object struct {
	*utils.Waitor
	sync.RWMutex
	//  Identify
	Id int

	//  Name
	Name string

	//  True if termination was already initiated. If so, we can destroy
	//  the object if there are no more child objects or pending term acks.
	terminating bool

	//  True if termination was already finished.
	terminated bool

	//  Sequence number of the last command sent to this object.
	sentSeqnum uint32

	//  Sequence number of the last command processed by this object.
	processedSeqnum uint32

	//  Number of events we have to get before we can destroy the object.
	termAcks int

	//  List of all objects owned by this object. We are responsible
	//  for deallocating them before we quit.
	childs *container.SynchronizedMap

	//  Socket owning this object. It's responsible for shutting down
	//  this object.
	owner *Object

	//	Command queue
	que *list.List

	//	Configuration Options
	opt Options

	//	Currently resides goroutine id. I do not know how get it.
	gid int

	//
	waitActive chan struct{}
	//
	waitEnlarge chan struct{}

	//	UserData
	UserData interface{}
	//
	sinker Sinker
	//
	timer *time.Ticker
	//object local storage
	ols [OLS_MAX_SLOT]interface{}
	//
	recvCmdCnt int64
	//
	sendCmdCnt int64
	//
	cond *Cond
}

func NewObject(id int, name string, opt Options, sinker Sinker) *Object {
	o := &Object{
		Id:          id,
		Name:        name,
		opt:         opt,
		sinker:      sinker,
		waitActive:  make(chan struct{}, 1),
		waitEnlarge: make(chan struct{}, 1),
		childs:      container.NewSynchronizedMap(),
		cond:        NewCond(1),
	}

	o.init()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Logger.Error(o, "panic, o.ProcessCommand error=", err)
			}
		}()
		o.ProcessCommand()
	}()

	return o
}

func (o *Object) GetTreeName() string {
	name := o.Name
	parent := o.owner
	for parent != nil {
		name = parent.Name + "/" + name
		parent = parent.owner
	}
	return "/" + name
}

func (o *Object) init() {
	o.que = list.New()
}

//	Active inner goroutine
func (o *Object) Active() {
	o.waitActive <- struct{}{}
}

//  Launch the supplied object and become its owner.
func (o *Object) LaunchChild(c *Object) {
	if c == nil {
		return
	}

	if c.owner != nil {
		panic("An object can have only one parent node")
	}

	c.owner = o
	c.Waitor = o.Waitor
	c.Active()
	c.safeStart()

	//  Take ownership of the object.
	SendOwn(o, c)
}

//thread safe
func (o *Object) GetChildById(id int) *Object {
	c := o.childs.Get(id)
	if cc, ok := c.(*Object); ok {
		return cc
	}
	return nil
}

//  When another owned object wants to send command to this object
//  it calls this function to let it know it should not shut down
//  before the command is delivered.
func (o *Object) incSeqnum() {
	atomic.AddUint32(&(o.sentSeqnum), 1)
}

//  Special handler called after a command that requires a seqnum
//  was processed. The implementation should catch up with its counter
//  of processed commands here.
func (o *Object) ProcessSeqnum() {
	//  Catch up with counter of processed commands.
	o.processedSeqnum++

	//  We may have catched up and still have pending terms acks.
	o.checkTermAcks()
}

//  Check whether all the peding term acks were delivered.
//  If so, deallocate this object.
func (o *Object) checkTermAcks() {
	name := o.GetTreeName()
	logger.Logger.Debugf("(%v) object checkTermAcks terminating=%v processedSeqnum=%v sentSeqnum=%v termAcks=%v ", name, o.terminating, o.processedSeqnum, o.sentSeqnum, o.termAcks)
	if o.terminating && o.processedSeqnum == o.sentSeqnum && o.termAcks == 0 {

		//  Sanity check. There should be no active children at this point.

		//  The root object has nobody to confirm the termination to.
		//  Other nodes will confirm the termination to the owner.
		if o.owner != nil {
			logger.Logger.Debugf("(%v)->(%v) Object SendTermAck ", o.Name, o.owner.Name)
			SendTermAck(o.owner)
		}

		//  Deallocate the resources.
		o.processDestroy()
	}
}

//  Ask owner object to terminate this object. It may take a while
//  while actual termination is started. This function should not be
//  called more than once.
func (o *Object) Terminate(s *Object) {
	//  If termination is already underway, there's no point
	//  in starting it anew.
	if o.terminating {
		return
	}

	name := o.GetTreeName()
	logger.Logger.Debugf("(%v) object Terminate ", name)
	//  As for the root of the ownership tree, there's noone to terminate it,
	//  so it has to terminate itself.
	if o.owner == nil {
		o.processTerm()
		return
	}

	//  If I am an owned object, I'll ask my owner to terminate me.
	SendTermReq(o.owner, o)
}

//  Term handler is protocted rather than private so that it can
//  be intercepted by the derived class. This is useful to add custom
//  steps to the beginning of the termination process.
func (o *Object) processTerm() {
	//  Double termination should never happen.
	if o.terminating {
		return
	}

	//  Send termination request to all owned objects.
	cnt := 0
	childs := o.childs.Items()
	for _, c := range childs {
		if cc, ok := c.(*Object); ok && cc != nil {
			SendTerm(cc)
			cnt++
		}
	}
	o.termAcks += cnt

	name := o.GetTreeName()
	logger.Logger.Debugf("(%v) object processTerm, termAcks=%v", name, o.termAcks)

	o.safeStop()
	//  Start termination process and check whether by chance we cannot
	//  terminate immediately.
	o.terminating = true
	o.checkTermAcks()
}

//  A place to hook in when phyicallal destruction of the object
//  is to be delayed.
func (o *Object) processDestroy() {
	name := o.GetTreeName()
	logger.Logger.Debugf("(%v) object processDestroy ", name)
	o.terminated = true
	//clear ols
	o.OlsClrValue()
}

func (o *Object) GetPendingCommandCnt() int {
	o.RLock()
	cnt := o.que.Len()
	o.RUnlock()
	return cnt
}

//	Enqueue command
func (o *Object) SendCommand(c Command, incseq bool) bool {
	if incseq {
		o.incSeqnum()
	}

	o.Lock()
	o.que.PushBack(c)
	o.Unlock()

	atomic.AddInt64(&o.sendCmdCnt, 1)

	//notify
	o.cond.Signal()
	return true
}

//	Dequeue command and process it.
func (o *Object) ProcessCommand() {

	//wait for active
	<-o.waitActive

	//deamon or no
	if o.Waitor != nil {
		o.Waitor.Add(o.Name, 1)
		defer o.Waitor.Done(o.Name)
	}

	var tickMode bool
	if o.opt.Interval > 0 && o.sinker != nil && o.timer == nil {
		o.timer = time.NewTicker(o.opt.Interval)
		defer o.timer.Stop()
		tickMode = true
	}

	name := o.GetTreeName()
	logger.Logger.Debug("(", name, ") object active!!!")
	doneCnt := 0
	for !o.terminated {
		cnt := o.GetPendingCommandCnt()
		if cnt == 0 {
			if tickMode {
				if o.cond.WaitForTick(o.timer) {
					//logger.Logger.Debug("(", name, ") object safeTick 1 ", time.Now())
					o.safeTick()
					doneCnt = 0
					continue
				}
			} else {
				o.cond.Wait()
			}
		}

		o.Lock()
		e := o.que.Front()
		if e != nil {
			o.que.Remove(e)
		}
		o.Unlock()

		if e != nil {
			if cmd, ok := e.Value.(Command); ok {
				o.safeDone(cmd)
				doneCnt++
			}
		}

		if tickMode {
			select {
			case <-o.timer.C:
				//logger.Logger.Debug("(", name, ") object safeTick 2 ", time.Now())
				o.safeTick()
				doneCnt = 0
			default:
			}

			if doneCnt > o.opt.MaxDone || cnt > o.opt.MaxDone {
				logger.Logger.Warn("(", name, ") object queue cmd count(", cnt, ") maxdone(", o.opt.MaxDone, ")", " this tick process cnt(", doneCnt, ")")
			}
		}
	}

	cnt := o.GetPendingCommandCnt()
	logger.Logger.Debug("(", name, ") object ProcessCommand done!!! queue rest cmd count(", cnt, ") ")
}

func (o *Object) safeDone(cmd Command) {
	defer utils.DumpStackIfPanic("Object::Command::Done")
	if StatsWatchMgr != nil {
		watch := StatsWatchMgr.WatchStart(fmt.Sprintf("/object/%v/cmdone", o.Name), 4)
		if watch != nil {
			defer watch.Stop()
		}
	}

	err := cmd.Done(o)
	atomic.AddInt64(&o.recvCmdCnt, 1)
	if err != nil {
		panic(err)
	}
}

func (o *Object) safeStart() {
	defer utils.DumpStackIfPanic("Object::OnStart")

	if o.sinker != nil {
		o.sinker.OnStart()
	}
}

func (o *Object) safeTick() {
	defer utils.DumpStackIfPanic("Object::OnTick")

	if o.sinker != nil {
		o.sinker.OnTick()
	}
}

func (o *Object) safeStop() {
	defer utils.DumpStackIfPanic("Object::OnStop")

	if o.sinker != nil {
		o.sinker.OnStop()
	}
}

func (o *Object) IsTermiated() bool {
	return o.terminated
}

func (o *Object) StatsSelf() (stats CmdStats) {
	stats.PendingCnt = int64(o.GetPendingCommandCnt())
	stats.SendCmdCnt = atomic.LoadInt64(&o.sendCmdCnt)
	stats.RecvCmdCnt = atomic.LoadInt64(&o.recvCmdCnt)
	return
}

func (o *Object) GetStats() map[string]CmdStats {
	if o.childs == nil {
		return nil
	}
	stats := make(map[string]CmdStats)
	stats[o.GetTreeName()] = o.StatsSelf()
	childs := o.childs.Items()
	for _, c := range childs {
		if cc, ok := c.(*Object); ok && cc != nil {
			stats[cc.GetTreeName()] = cc.StatsSelf()
			subStats := cc.GetStats()
			if subStats != nil && len(subStats) > 0 {
				for k, v := range subStats {
					stats[k] = v
				}
			}
		}
	}
	return stats
}
