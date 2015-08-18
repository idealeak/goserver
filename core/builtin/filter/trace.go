// sessionfiltertrace
package filter

import (
	//"reflect"
	"sync"
	"time"

	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

var (
	SessionFilterTraceName = "session-filter-trace"
)

type SessionFilterTrace struct {
	sync.Mutex
	recvCntPerSec int
	maxRecvPerSec int
	recvTime      time.Time
	sendCntPerSec int
	maxSendPerSec int
	sendTime      time.Time
	dumpTime      time.Time
}

func (sft SessionFilterTrace) GetName() string {
	return SessionFilterTraceName
}

func (sft *SessionFilterTrace) GetInterestOps() uint {
	return 1<<netlib.InterestOps_Max - 1
}

func (sft *SessionFilterTrace) OnSessionOpened(s *netlib.Session) bool {
	//logger.Tracef("SessionFilterTrace.OnSessionOpened sid=%v", s.Id)
	return true
}

func (sft *SessionFilterTrace) OnSessionClosed(s *netlib.Session) bool {
	//logger.Tracef("SessionFilterTrace.OnSessionClosed sid=%v", s.Id)
	return true
}

func (sft *SessionFilterTrace) OnSessionIdle(s *netlib.Session) bool {
	//logger.Tracef("SessionFilterTrace.OnSessionIdle sid=%v", s.Id)
	return true
}

func (sft *SessionFilterTrace) OnPacketReceived(s *netlib.Session, packetid int, packet interface{}) bool {
	//logger.Tracef("SessionFilterTrace.OnPacketReceived sid=%v packetid=%v packet=%v", s.Id, packetid, reflect.TypeOf(packet))
	sft.Lock()
	sft.recvCntPerSec++
	if time.Now().Sub(sft.recvTime) > time.Second {
		if sft.recvCntPerSec > sft.maxRecvPerSec {
			sft.maxRecvPerSec = sft.recvCntPerSec
			sft.recvCntPerSec = 0
			sft.recvTime = time.Now()
		}
	}
	sft.dump()
	sft.Unlock()
	return true
}

func (sft *SessionFilterTrace) OnPacketSent(s *netlib.Session, data []byte) bool {
	//logger.Tracef("SessionFilterTrace.OnPacketSent sid=%v size=%d", s.Id, len(data))
	sft.Lock()
	sft.sendCntPerSec++
	if time.Now().Sub(sft.sendTime) > time.Second {
		if sft.sendCntPerSec > sft.maxSendPerSec {
			sft.maxSendPerSec = sft.sendCntPerSec
			sft.sendCntPerSec = 0
			sft.sendTime = time.Now()
		}
	}
	sft.dump()
	sft.Unlock()
	return true
}

func (sft *SessionFilterTrace) dump() {
	if time.Now().Sub(sft.dumpTime) >= time.Minute*5 {
		logger.Info("Session per five minuts: recvCntPerSec=", sft.recvCntPerSec, " sendCntPerSec=", sft.sendCntPerSec)
		sft.dumpTime = time.Now()
	}
}
func init() {
	netlib.RegisteSessionFilterCreator(SessionFilterTraceName, func() netlib.SessionFilter {
		tNow := time.Now()
		return &SessionFilterTrace{dumpTime: tNow, recvTime: tNow, sendTime: tNow}
	})
}
