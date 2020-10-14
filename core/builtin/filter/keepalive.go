package filter

import (
	"code.google.com/p/goprotobuf/proto"
	"github.com/idealeak/goserver/core/builtin/protocol"
	"github.com/idealeak/goserver/core/netlib"
)

var (
	KeepAliveFilterName = "session-filter-keepalive"
)

type KeepAliveFilter struct {
}

func (kf *KeepAliveFilter) GetName() string {
	return KeepAliveFilterName
}

func (kf *KeepAliveFilter) GetInterestOps() uint {
	return 1 << netlib.InterestOps_Idle
}

func (kf *KeepAliveFilter) OnSessionOpened(s *netlib.Session) bool {
	return true
}

func (kf *KeepAliveFilter) OnSessionClosed(s *netlib.Session) bool {
	return true
}

func (kf *KeepAliveFilter) OnSessionIdle(s *netlib.Session) bool {
	p := &protocol.SSPacketKeepAlive{Flag: proto.Int32(0)}
	proto.SetDefaults(p)
	s.Send(int(protocol.CoreBuiltinPacketID_PACKET_SS_KEEPALIVE), p)
	return true
}

func (kf *KeepAliveFilter) OnPacketReceived(s *netlib.Session, packetid int, logicNo uint32, packet interface{}) bool {
	return true
}

func (kf *KeepAliveFilter) OnPacketSent(s *netlib.Session, packetid int, logicNo uint32, data []byte) bool {
	return true
}

func init() {
	netlib.RegisterFactory(int(protocol.CoreBuiltinPacketID_PACKET_SS_KEEPALIVE), netlib.PacketFactoryWrapper(func() interface{} {
		return &protocol.SSPacketKeepAlive{}
	}))
	netlib.RegisteSessionFilterCreator(KeepAliveFilterName, func() netlib.SessionFilter {
		return &KeepAliveFilter{}
	})
	netlib.RegisterHandler(int(protocol.CoreBuiltinPacketID_PACKET_SS_KEEPALIVE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, data interface{}) error {
		if p, ok := data.(*protocol.SSPacketKeepAlive); ok {
			if p.GetFlag() == 0 {
				p.Flag = proto.Int32(1)
				s.Send(int(protocol.CoreBuiltinPacketID_PACKET_SS_KEEPALIVE), p)
			}
		}
		return nil
	}))
}
