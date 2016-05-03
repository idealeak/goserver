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
	sessions  map[int]map[int]map[int]*netlib.Session //keys=>areaid:type:id
	listeners []ServerSessionRegisteListener
}

func (ssm *ServerSessionMgr) AddListener(l ServerSessionRegisteListener) ServerSessionRegisteListener {
	ssm.listeners = append(ssm.listeners, l)
	return l
}

func (ssm *ServerSessionMgr) RegisteSession(s *netlib.Session) bool {
	attr := s.GetAttribute(SessionAttributeServerInfo)
	if attr != nil {
		if srvInfo, ok := attr.(*protocol.SSSrvRegiste); ok && srvInfo != nil {
			logger.Tracef("ServerSessionMgr.RegisteSession %v", srvInfo)
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
			if len(ssm.listeners) != 0 {
				for _, l := range ssm.listeners {
					l.OnRegiste(s)
				}
			}
		}
	} else {
		logger.Tracef("ServerSessionMgr.RegisteSession SessionAttributeServerInfo=nil")
	}
	return true
}

func (ssm *ServerSessionMgr) UnregisteSession(s *netlib.Session) bool {
	attr := s.GetAttribute(SessionAttributeServerInfo)
	if attr != nil {
		if srvInfo, ok := attr.(*protocol.SSSrvRegiste); ok && srvInfo != nil {
			logger.Tracef("ServerSessionMgr.UnregisteSession %v", srvInfo)
			areaId := int(srvInfo.GetAreaId())
			srvType := int(srvInfo.GetType())
			srvId := int(srvInfo.GetId())
			if a, exist := ssm.sessions[areaId]; exist {
				if b, exist := a[srvType]; exist {
					delete(b, srvId)
					if len(ssm.listeners) != 0 {
						for _, l := range ssm.listeners {
							l.OnUnregiste(s)
						}
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

func (ssm *ServerSessionMgr) GetServerId(areaId, srvType int) int {
	if a, exist := ssm.sessions[areaId]; exist {
		if b, exist := a[srvType]; exist {
			for sid, _ := range b {
				return sid
			}
		}
	}
	return -1
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
