// serversessionfilter
package main

import (
	"time"

	"code.google.com/p/goprotobuf/proto"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/examples/protocol"
)

var (
	ServerSessionFilterName = "serversessionfilter"
)

type ServerSessionFilter struct {
	netlib.BasicSessionFilter
}

func (ssf ServerSessionFilter) GetName() string {
	return ServerSessionFilterName
}

func (ssf *ServerSessionFilter) GetInterestOps() uint {
	return 1 << netlib.InterestOps_Opened
}

func (ssf *ServerSessionFilter) OnSessionOpened(s *netlib.Session) bool {
	logger.Logger.Trace("(ssf *ServerSessionFilter) OnSessionOpened")
	packet := &protocol.CSPacketPing{
		TimeStamb: proto.Int64(time.Now().Unix()),
		Message:   []byte("=1234567890abcderghijklmnopqrstuvwxyz="),
	}
	//for i := 0; i < 1024*32; i++ {
	//	packet.Message = append(packet.Message, byte('x'))
	//}
	proto.SetDefaults(packet)
	s.Send(int(protocol.PacketID_PACKET_CS_PING), packet)
	return true
}

func init() {
	netlib.RegisteSessionFilterCreator(ServerSessionFilterName, func() netlib.SessionFilter {
		return &ServerSessionFilter{}
	})
}
