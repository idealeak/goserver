// serversessionfilter
package main

import (
	"time"

	"code.google.com/p/goprotobuf/proto"
	"github.com/idealeak/goserver/examples/protocol"
	"github.com/idealeak/goserver/core/netlib"
)

var (
	ServerSessionFilterName = "serversessionfilter"
)

type ServerSessionFilter struct {
}

func (ssf ServerSessionFilter) GetName() string {
	return ServerSessionFilterName
}

func (ssf *ServerSessionFilter) GetInterestOps() uint {
	return 1 << netlib.InterestOps_Opened
}

func (ssf *ServerSessionFilter) OnSessionOpened(s *netlib.Session) bool {
	packet := &protocol.CSPacketPing{
		TimeStamb: proto.Int64(time.Now().Unix()),
		Message:   []byte("=1234567890abcderghijklmnopqrstuvwxyz="),
	}
	//for i := 0; i < 1024*32; i++ {
	//	packet.Message = append(packet.Message, byte('x'))
	//}
	proto.SetDefaults(packet)
	s.Send(packet)
	return true
}

func (ssf *ServerSessionFilter) OnSessionClosed(s *netlib.Session) bool {
	return true
}

func (ssf *ServerSessionFilter) OnSessionIdle(s *netlib.Session) bool {
	return true
}

func (ssf *ServerSessionFilter) OnPacketReceived(s *netlib.Session, packetid int, packet interface{}) bool {
	return true
}

func (ssf *ServerSessionFilter) OnPacketSent(s *netlib.Session, data []byte) bool {
	return true
}

func init() {
	netlib.RegisteSessionFilterCreator(ServerSessionFilterName, func() netlib.SessionFilter {
		return &ServerSessionFilter{}
	})
}
