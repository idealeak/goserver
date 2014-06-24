package srvlib

import (
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/srvlib/protocol"
)

var (
	SessionAttributeServerInfo = &ServerSessionMgr{}
	ServerSessionMgrSington    = &ServerSessionMgr{sessions: make(map[int]map[int]map[int]*netlib.Session)}
)

type ServerSessionRegisteListener interface {
	OnRegiste(*netlib.Session)
	OnUnregiste(*netlib.Session)
}

type ServerSessionMgr struct {
	sessions map[int]map[int]map[int]*netlib.Session //keys=>areaid:type:id
	listener ServerSessionRegisteListener
}

func (ssm *ServerSessionMgr) SetListener(l ServerSessionRegisteListener) ServerSessionRegisteListener {
	ol := ssm.listener
	ssm.listener = l
	return ol
}

func (ssm *ServerSessionMgr) RegisteSession(s *netlib.Session) bool {
	attr := s.GetAttribute(SessionAttributeServerInfo)
	if attr != nil {
		logger.Logger.Trace("ServerSessionMgr.RegisteSession")
		if srvInfo, ok := attr.(*protocol.SSSrvRegiste); ok && srvInfo != nil {
			areaId := int(srvInfo.GetAreaId())
			srvType := int(srvInfo.GetType())
			srvId := int(srvInfo.GetId())
			if a, exist := ssm.sessions[areaId]; !exist {
				ssm.sessions[areaId] = make(map[int]map[int]*netlib.Session)
				a = ssm.sessions[areaId]
				a[srvType] = make(map[int]*netlib.Session)
			} else {
				if _, exist := a[srvType]; !exist {
					a[srvType] = make(map[int]*netlib.Session)
				}
			}

			ssm.sessions[areaId][srvType][srvId] = s
			if ssm.listener != nil {
				ssm.listener.OnRegiste(s)
			}
		}
	}
	return true
}

func (ssm *ServerSessionMgr) UnregisteSession(s *netlib.Session) bool {
	attr := s.GetAttribute(SessionAttributeServerInfo)
	if attr != nil {
		logger.Logger.Trace("ServerSessionMgr.UnregisteSession")
		if srvInfo, ok := attr.(*protocol.SSSrvRegiste); ok && srvInfo != nil {
			areaId := int(srvInfo.GetAreaId())
			srvType := int(srvInfo.GetType())
			srvId := int(srvInfo.GetId())
			if a, exist := ssm.sessions[areaId]; exist {
				if b, exist := a[srvType]; exist {
					delete(b, srvId)
					if ssm.listener != nil {
						ssm.listener.OnUnregiste(s)
					}
				}
			}
		}
	}
	return true
}

func (ssm *ServerSessionMgr) GetSession(areaId, srvType, srvId int) *netlib.Session {
	if a, exist := ssm.sessions[areaId]; exist {
		if b, exist := a[srvType]; exist {
			if c, exist := b[srvId]; exist {
				return c
			}
		}
	}
	return nil
}

func (ssm *ServerSessionMgr) GetSessions(areaId, srvType int) (sessions []*netlib.Session) {
	if a, exist := ssm.sessions[areaId]; exist {
		if b, exist := a[srvType]; exist {
			for _, s := range b {
				sessions = append(sessions, s)
			}
		}
	}
	return
}

func (ssm *ServerSessionMgr) Broadcast(pack interface{}, areaId, srvType int) {
	if areaId >= 0 {
		if srvType >= 0 {
			if a, exist := ssm.sessions[areaId]; exist {
				if b, exist := a[srvType]; exist {
					for _, s := range b {
						s.Send(pack)
					}
				}
			}
		} else {
			if a, exist := ssm.sessions[areaId]; exist {
				for _, b := range a {
					for _, s := range b {
						s.Send(pack)
					}
				}
			}
		}
	} else {
		if srvType >= 0 {
			for _, a := range ssm.sessions {
				if b, exist := a[srvType]; exist {
					for _, s := range b {
						s.Send(pack)
					}
				}
			}
		} else {
			for _, a := range ssm.sessions {
				for _, b := range a {
					for _, s := range b {
						s.Send(pack)
					}
				}
			}
		}
	}
}
