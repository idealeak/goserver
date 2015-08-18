package srvlib

import (
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

var (
	SessionAttributeClientSession = &ClientSessionMgr{}
	ClientSessionMgrSington       = &ClientSessionMgr{sessions: make(map[int64]*netlib.Session)}
)

type ClientSessionMgr struct {
	sessions map[int64]*netlib.Session //keys=>sessionid
}

func (csm *ClientSessionMgr) RegisteSession(s *netlib.Session) bool {
	attr := s.GetAttribute(SessionAttributeClientSession)
	if attr == nil {
		sid := NewSessionId(s)
		s.SetAttribute(SessionAttributeClientSession, sid)
		csm.sessions[sid.Get()] = s
		logger.Tracef("client session %v registe", sid.Get())
	}
	return true
}

func (csm *ClientSessionMgr) UnregisteSession(s *netlib.Session) bool {
	attr := s.GetAttribute(SessionAttributeClientSession)
	if attr != nil {
		if sid, ok := attr.(SessionId); ok {
			delete(csm.sessions, sid.Get())
			logger.Tracef("client session %v unregiste", sid.Get())
		}
	}
	return true
}

func (csm *ClientSessionMgr) GetSession(srvId int64) *netlib.Session {
	if s, exist := csm.sessions[srvId]; exist {
		return s
	}
	return nil
}

func (csm *ClientSessionMgr) Broadcast(pack interface{}) {
	for _, s := range csm.sessions {
		s.Send(pack)
	}
}

func (csm *ClientSessionMgr) Count() int {
	return len(csm.sessions)
}

func (csm *ClientSessionMgr) CloseAll() {
	for _, s := range csm.sessions {
		s.Close()
	}
}
