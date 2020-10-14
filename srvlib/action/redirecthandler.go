package action

import (
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/srvlib"
	"github.com/idealeak/goserver/srvlib/protocol"
)

type PacketRedirectPacketFactory struct {
}

type PacketRedirectHandler struct {
}

func init() {
	netlib.RegisterHandler(int(protocol.SrvlibPacketID_PACKET_SS_REDIRECT), &PacketRedirectHandler{})
	netlib.RegisterFactory(int(protocol.SrvlibPacketID_PACKET_SS_REDIRECT), &PacketRedirectPacketFactory{})
}

func (this *PacketRedirectPacketFactory) CreatePacket() interface{} {
	pack := &protocol.SSPacketRedirect{}
	return pack
}

func (this *PacketRedirectHandler) Process(s *netlib.Session, packetid int, data interface{}) error {
	logger.Logger.Trace("PacketRedirectHandler.Process")
	if pr, ok := data.(*protocol.SSPacketRedirect); ok {
		packid, pack, err := netlib.UnmarshalPacket(pr.GetData())
		if err != nil {
			return err
		}
		h := srvlib.GetHandler(packid)
		if h != nil {
			return h.Process(s, packid, pack, pr.GetClientSid(), pr.GetSrvRoutes())
		} else {
			nh := netlib.GetHandler(packid)
			if nh != nil {
				return nh.Process(s, packid, pack)
			}
		}
	}
	return nil
}
