// sessionfiltertrace
package main

import (
	"github.com/idealeak/goserver/core/netlib"
)

var (
	SessionFilterName = "my-session-filter-trace"
)

type SessionFilterTrace struct {
}

func (sft SessionFilterTrace) GetName() string {
	return SessionFilterName
}

func (sft *SessionFilterTrace) GetInterestOps() uint {
	return 0
}

func (sft *SessionFilterTrace) OnSessionOpened(s *netlib.Session) bool {
	return true
}

func (sft *SessionFilterTrace) OnSessionClosed(s *netlib.Session) bool {
	return true
}

func (sft *SessionFilterTrace) OnSessionIdle(s *netlib.Session) bool {
	return true
}

func (sft *SessionFilterTrace) OnPacketReceived(s *netlib.Session, packetid int, packet interface{}) bool {
	return true
}

func (sft *SessionFilterTrace) OnPacketSent(s *netlib.Session, data []byte) bool {
	return true
}

func init() {
	netlib.RegisteSessionFilterCreator(SessionFilterName, func() netlib.SessionFilter {
		return &SessionFilterTrace{}
	})
}
