package main

import (
	"code.google.com/p/goprotobuf/proto"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/mmo/protocol"
	"github.com/idealeak/goserver/srvlib"
)

var (
	SessionHandlerClientLoadName = "handler-client-load"
)

type SessionHandlerClientLoad struct {
	netlib.BasicSessionHandler
}

func (sfcl SessionHandlerClientLoad) GetName() string {
	return SessionHandlerClientLoadName
}

func (sfcl *SessionHandlerClientLoad) GetInterestOps() uint {
	return 1<<netlib.InterestOps_Opened | 1<<netlib.InterestOps_Closed
}

func (sfcl *SessionHandlerClientLoad) OnSessionOpened(s *netlib.Session) {
	sfcl.reportLoad(s)
}

func (sfcl *SessionHandlerClientLoad) OnSessionClosed(s *netlib.Session) {
	sfcl.reportLoad(s)

}

func (sfcl *SessionHandlerClientLoad) reportLoad(s *netlib.Session) {
	sc := s.GetSessionConfig()
	pack := &protocol.ServerLoad{
		SrvType: proto.Int32(int32(sc.Type)),
		SrvId:   proto.Int32(int32(sc.Id)),
		CurLoad: proto.Int32(int32(srvlib.ClientSessionMgrSington.Count())),
	}
	proto.SetDefaults(pack)
	srvlib.ServerSessionMgrSington.Broadcast(int(protocol.MmoPacketID_PACKET_SC_GATEINFO), pack, netlib.Config.SrvInfo.AreaID, srvlib.BalanceServerType)
	logger.Logger.Tracef("SessionHandlerClientLoad.reportLoad %v", pack)
}

func init() {
	netlib.RegisteSessionHandlerCreator(SessionHandlerClientLoadName, func() netlib.SessionHandler {
		return &SessionHandlerClientLoad{}
	})
}
