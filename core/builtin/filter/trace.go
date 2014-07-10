// sessionfiltertrace
package filter

import (
	"reflect"

	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

var (
	SessionFilterTraceName = "session-filter-trace"
)

type SessionFilterTrace struct {
}

func (sft SessionFilterTrace) GetName() string {
	return SessionFilterTraceName
}

func (sft *SessionFilterTrace) GetInterestOps() uint {
	return 1<<netlib.InterestOps_Max - 1
}

func (sft *SessionFilterTrace) OnSessionOpened(s *netlib.Session, bAccept bool) bool {
	logger.Tracef("SessionFilterTrace.OnSessionOpened sid=%v accept=%v ", s.Id, bAccept)
	return true
}

func (sft *SessionFilterTrace) OnSessionClosed(s *netlib.Session) bool {
	logger.Tracef("SessionFilterTrace.OnSessionClosed sid=%v", s.Id)
	return true
}

func (sft *SessionFilterTrace) OnSessionIdle(s *netlib.Session) bool {
	logger.Tracef("SessionFilterTrace.OnSessionIdle sid=%v", s.Id)
	return true
}

func (sft *SessionFilterTrace) OnPacketReceived(s *netlib.Session, packetid int, packet interface{}) bool {
	logger.Tracef("SessionFilterTrace.OnPacketReceived sid=%v packetid=%v packet=%v", s.Id, packetid, reflect.TypeOf(packet))
	return true
}

func (sft *SessionFilterTrace) OnPacketSent(s *netlib.Session, data []byte) bool {
	logger.Tracef("SessionFilterTrace.OnPacketSent sid=%v size=%d", s.Id, len(data))
	return true
}

func init() {
	netlib.RegisteSessionFilterCreator(SessionFilterTraceName, func() netlib.SessionFilter {
		return &SessionFilterTrace{}
	})
}
