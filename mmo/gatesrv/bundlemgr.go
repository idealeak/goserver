package main

import (
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/srvlib"
)

var (
	SessionHandlerBundleName = "handler-client-bundle"
	BundleMgrSington         = &BundleMgr{}
)

type SendPendItem struct {
	sid      int //当前pending包发给的session
	packetid int
	logicNo  uint32
	data     interface{}
	ts       int64
}

type BundleSession struct {
	bundleKey       string
	bundleSession   []*netlib.Session
	currentSession  *netlib.Session
	fastestSession  *netlib.Session
	worldsrvSession *netlib.Session
	gamesrvSession  *netlib.Session
	waitAckChain    []*SendPendItem
	rcvLogicNo      uint32
	sndLogicNo      uint32
	lastSndTs       int64
	lastAckTs       int64
	RTT             int64 //Round Trip Time
	RTO             int64 //Retransmission TimeOut
}

func (bs *BundleSession) OnSessionClose(s *netlib.Session) bool {
	if bs.fastestSession == s {
		bs.fastestSession = nil
	}
	if bs.currentSession == s {
		bs.currentSession = nil
	}
	cnt := len(bs.bundleSession)
	for i := 0; i < cnt; i++ {
		if bs.bundleSession[i] == s {
			if i == 0 {
				bs.bundleSession = bs.bundleSession[1:]
			} else if i == cnt-1 {
				bs.bundleSession = bs.bundleSession[:cnt-1]
			} else {
				temp := bs.bundleSession[:i]
				temp = append(temp, bs.bundleSession[i+1:]...)
				bs.bundleSession = temp
			}
			break
		}
	}
	//优先最快的一个连接
	if bs.currentSession == nil {
		bs.currentSession = bs.fastestSession
	}
	//从池里面挑选一个
	if bs.currentSession == nil {
		if len(bs.bundleSession) > 0 {
			bs.currentSession = bs.bundleSession[0]
		}
	}

	//重发没有回应的包
	if bs.currentSession != nil {
		idx := -1
		for i, pack := range bs.waitAckChain {
			if pack.sid == s.Id {
				bs.currentSession.SendEx(pack.packetid, pack.logicNo, pack.data, false)
				idx = i
			} else {
				break
			}
		}
		if idx != -1 {
			if idx < len(bs.waitAckChain) {
				bs.waitAckChain = bs.waitAckChain[idx+1:]
			}
		}
	} else { //所有连接都已关闭
		return true
	}
	return false
}

func (bs *BundleSession) CacheSendItem(sid, packetid int, logicNo uint32, data interface{}) {
	ts := module.AppModule.GetCurrTimeNano()
	bs.waitAckChain = append(bs.waitAckChain, &SendPendItem{
		sid:      sid,
		packetid: packetid,
		logicNo:  logicNo,
		data:     data,
		ts:       ts,
	})
	bs.lastSndTs = ts
}

func (bs *BundleSession) Send(packetid int, pack interface{}) {
	if bs.currentSession != nil {
		logicNo := atomic.AddUint32(&bs.sndLogicNo, 1)
		if bs.currentSession.SendEx(packetid, logicNo, pack, false) {
			bs.CacheSendItem(bs.currentSession.Id, packetid, logicNo, pack)
		}
	}
}

type BundleMgr struct {
	freeBundle []uint16
	bundles    [math.MaxUint16]*BundleSession
	bundlesMap map[string]uint16
	Debug      bool
}

//主线程中调用
func (bm *BundleMgr) AllocBundleId() uint16 {
	last := len(bm.freeBundle)
	if last > 0 {
		id := bm.freeBundle[last-1]
		bm.freeBundle = bm.freeBundle[:last-1]
		return id
	}
	return 0
}

//主线程中调用
func (bm *BundleMgr) FreeBundleId(id uint16) {
	if bm.Debug {
		for i := 0; i < len(bm.freeBundle); i++ {
			if bm.freeBundle[i] == id {
				panic(fmt.Sprintf("BundleMgr.FreeBundleId found repeat id:%v", id))
			}
		}
	}
	bm.freeBundle = append(bm.freeBundle, id)
	delete(bm.bundlesMap, bm.bundles[id].bundleKey)
	bm.bundles[id] = nil
}

//主线程中调用
func (bm *BundleMgr) BindSession(bundleId uint16, s *netlib.Session) {
	if bm.bundles[bundleId] == nil {
		bm.bundles[bundleId] = &BundleSession{
			RTO: int64(time.Second),
		}
	}
	bs := bm.bundles[bundleId]
	if bs != nil {
		s.GroupId = int(bundleId)
		param := s.GetAttribute(srvlib.SessionAttributeClientSession)
		if param != nil {
			if sid, ok := param.(srvlib.SessionId); ok {
				s.Sid = int64(srvlib.NewSessionIdEx(int32(sid.AreaId()), int32(sid.SrvType()), int32(sid.SrvId()), int32(bundleId)))
			}
		}

		bs.bundleSession = append(bs.bundleSession, s)
		if bs.fastestSession == nil {
			bs.fastestSession = s
		}
		if bs.currentSession == nil {
			bs.currentSession = s
		}
	}
}

func (bm *BundleMgr) GetBundleSession(bundleId uint16) *BundleSession {
	return bm.bundles[bundleId]
}

//主线程中调用
func (bm *BundleMgr) OnSessionAck(s *netlib.Session, logicNo uint32) {
	if s != nil {
		bs := bm.GetBundleSession(uint16(s.GroupId))
		if bs != nil {
			if len(bs.waitAckChain) > 0 && bs.waitAckChain[0].logicNo == logicNo {
				pendingItem := bs.waitAckChain[0]
				bs.waitAckChain = bs.waitAckChain[1:]
				bs.fastestSession = s
				if len(bs.waitAckChain) == 0 { //切换响应更快的链接
					bs.currentSession = bs.fastestSession
				}
				bs.lastAckTs = module.AppModule.GetCurrTimeNano()
				bs.RTT = bs.lastAckTs - pendingItem.ts
				bs.RTO = bs.RTT * 5
				if bs.RTO > int64(time.Second) { //最长1秒
					bs.RTO = int64(time.Second)
				}
			}
		}
	}
}

func (bm *BundleMgr) Broadcast(packetid int, pack interface{}) {
	for _, bid := range bm.bundlesMap {
		bs := bm.GetBundleSession(bid)
		if bs != nil {
			bs.Send(packetid, pack)
		}
	}
}

func (bm *BundleMgr) ModuleName() string {
	return "BundleMgr"
}

func (bm *BundleMgr) Init() {
	bm.freeBundle = make([]uint16, 0, math.MaxUint16)
	for i := uint16(math.MaxUint16); i > 0; i-- {
		bm.freeBundle = append(bm.freeBundle, i)
	}
	bm.bundlesMap = make(map[string]uint16)
}

func (bm *BundleMgr) Update() {
	ts := module.AppModule.GetCurrTimeNano()
	for _, id := range bm.bundlesMap {
		bs := bm.GetBundleSession(id)
		if bs != nil {
			if len(bs.waitAckChain) > 0 && ts-bs.waitAckChain[0].ts > bs.RTO { //重发
				s := bs.fastestSession
				if s == bs.currentSession {
					for _, ss := range bs.bundleSession {
						if s != ss {
							s = ss
							break
						}
					}
				}
				if s != bs.currentSession {
					bs.currentSession = s
					chain := bs.waitAckChain
					bs.waitAckChain = nil
					for _, item := range chain {
						if ts-item.ts > bs.RTO {
							if bs.currentSession.SendEx(item.packetid, item.logicNo, item.data, false) {
								bs.CacheSendItem(bs.currentSession.Id, item.packetid, item.logicNo, item.data)
							}
						}
					}
				}
			}
		}
	}
}

func (bm *BundleMgr) Shutdown() {
	module.UnregisteModule(bm)
}

func init() {
	module.RegisteModule(BundleMgrSington, time.Millisecond*100, 0)
}
