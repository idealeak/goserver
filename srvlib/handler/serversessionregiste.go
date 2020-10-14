// SessionHandlerSrvRegiste
package handler

import (
	"code.google.com/p/goprotobuf/proto"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/srvlib"
	"github.com/idealeak/goserver/srvlib/protocol"
)

var (
	SessionHandlerSrvRegisteName = "session-srv-registe"
)

type SessionHandlerSrvRegiste struct {
}

func (sfl SessionHandlerSrvRegiste) GetName() string {
	return SessionHandlerSrvRegisteName
}

func (sfl *SessionHandlerSrvRegiste) GetInterestOps() uint {
	return 1<<netlib.InterestOps_Opened | 1<<netlib.InterestOps_Closed
}

func (sfl *SessionHandlerSrvRegiste) OnSessionOpened(s *netlib.Session) {
	registePacket := &protocol.SSSrvRegiste{
		Id:     proto.Int(netlib.Config.SrvInfo.Id),
		Type:   proto.Int(netlib.Config.SrvInfo.Type),
		AreaId: proto.Int(netlib.Config.SrvInfo.AreaID),
		Name:   proto.String(netlib.Config.SrvInfo.Name),
		Data:   proto.String(netlib.Config.SrvInfo.Data),
	}
	proto.SetDefaults(registePacket)
	s.Send(int(protocol.SrvlibPacketID_PACKET_SS_REGISTE), registePacket)
}

func (sfl *SessionHandlerSrvRegiste) OnSessionClosed(s *netlib.Session) {
	srvlib.ServerSessionMgrSington.UnregisteSession(s)
}

func (sfl *SessionHandlerSrvRegiste) OnSessionIdle(s *netlib.Session) {
}

func (sfl *SessionHandlerSrvRegiste) OnPacketReceived(s *netlib.Session, packetid int, logicNo uint32, packet interface{}) {
}

func (sfl *SessionHandlerSrvRegiste) OnPacketSent(s *netlib.Session, packetid int, logicNo uint32, data []byte) {
}

func init() {
	netlib.RegisteSessionHandlerCreator(SessionHandlerSrvRegisteName, func() netlib.SessionHandler {
		return &SessionHandlerSrvRegiste{}
	})

	netlib.RegisterFactory(int(protocol.SrvlibPacketID_PACKET_SS_REGISTE), netlib.PacketFactoryWrapper(func() interface{} {
		return &protocol.SSSrvRegiste{}
	}))

	netlib.RegisterHandler(int(protocol.SrvlibPacketID_PACKET_SS_REGISTE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, data interface{}) error {
		if registePacket, ok := data.(*protocol.SSSrvRegiste); ok {
			s.SetAttribute(srvlib.SessionAttributeServerInfo, registePacket)
			srvlib.ServerSessionMgrSington.RegisteSession(s)
		}
		return nil
	}))
}
