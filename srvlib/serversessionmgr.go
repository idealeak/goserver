package srvlib

import (
	"strings"

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

			if _, exist := ssm.sessions[areaId][srvType][srvId]; !exist {
				logger.Logger.Infof("(ssm *ServerSessionMgr) RegisteSession %v", srvInfo)
				ssm.sessions[areaId][srvType][srvId] = s
				if len(ssm.listeners) != 0 {
					for _, l := range ssm.listeners {
						l.OnRegiste(s)
					}
				}
			} else {
				logger.Logger.Warnf("###(ssm *ServerSessionMgr) RegisteSession repeated areaid:%v srvType:%v srvId:%v", areaId, srvType, srvId)
			}
		}
	} else {
		logger.Logger.Warnf("ServerSessionMgr.RegisteSession SessionAttributeServerInfo=nil")
	}
	return true
}

func (ssm *ServerSessionMgr) UnregisteSession(s *netlib.Session) bool {
	attr := s.GetAttribute(SessionAttributeServerInfo)
	if attr != nil {
		if srvInfo, ok := attr.(*protocol.SSSrvRegiste); ok && srvInfo != nil {
			logger.Logger.Infof("ServerSessionMgr.UnregisteSession try %v", srvInfo)
			areaId := int(srvInfo.GetAreaId())
			srvType := int(srvInfo.GetType())
			srvId := int(srvInfo.GetId())
			if a, exist := ssm.sessions[areaId]; exist {
				if b, exist := a[srvType]; exist {
					if ss, exist := b[srvId]; exist && ss == s {
						logger.Logger.Infof("ServerSessionMgr.UnregisteSession %v success", srvInfo)
						delete(b, srvId)
						if len(ssm.listeners) != 0 {
							for _, l := range ssm.listeners {
								l.OnUnregiste(s)
							}
						}
					} else {
						logger.Logger.Warnf("(ssm *ServerSessionMgr) UnregisteSession found not fit session, area:%v type:%v id:%v", areaId, srvType, srvId)
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

func (ssm *ServerSessionMgr) GetServerIdByMaxData(areaId, srvType int) int {
	var bestSid int = -1
	var data string
	if a, exist := ssm.sessions[areaId]; exist {
		if b, exist := a[srvType]; exist {
			for sid, s := range b {
				if srvInfo, ok := s.GetAttribute(SessionAttributeServerInfo).(*protocol.SSSrvRegiste); ok && srvInfo != nil {
					if strings.Compare(data, srvInfo.GetData()) <= 0 {
						data = srvInfo.GetData()
						bestSid = sid
					}
				}
			}
		}
	}
	return bestSid
}

func (ssm *ServerSessionMgr) GetServerIds(areaId, srvType int) (ids []int) {
	if a, exist := ssm.sessions[areaId]; exist {
		if b, exist := a[srvType]; exist {
			for sid, _ := range b {
				ids = append(ids, sid)
			}
		}
	}
	return
}

func (ssm *ServerSessionMgr) Broadcast(packetid int, pack interface{}, areaId, srvType int) {
	if areaId >= 0 {
		if srvType >= 0 {
			if a, exist := ssm.sessions[areaId]; exist {
				if b, exist := a[srvType]; exist {
					for _, s := range b {
						s.Send(packetid, pack)
					}
				}
			}
		} else {
			if a, exist := ssm.sessions[areaId]; exist {
				for _, b := range a {
					for _, s := range b {
						s.Send(packetid, pack)
					}
				}
			}
		}
	} else {
		if srvType >= 0 {
			for _, a := range ssm.sessions {
				if b, exist := a[srvType]; exist {
					for _, s := range b {
						s.Send(packetid, pack)
					}
				}
			}
		} else {
			for _, a := range ssm.sessions {
				for _, b := range a {
					for _, s := range b {
						s.Send(packetid, pack)
					}
				}
			}
		}
	}
}
