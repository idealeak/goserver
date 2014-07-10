package action

import (
	"code.google.com/p/goprotobuf/proto"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/srvlib"
	"github.com/idealeak/goserver/srvlib/protocol"
)

var (
	MulticastMaker = &MulticastPacketFactory{}
)

type MulticastPacketFactory struct {
}

type MulticastHandler struct {
}

func init() {
	netlib.RegisterHandler(int(protocol.SrvlibPacketID_PACKET_SS_MULTICAST), &MulticastHandler{})
	netlib.RegisterFactory(int(protocol.SrvlibPacketID_PACKET_SS_MULTICAST), MulticastMaker)
}

func (this *MulticastPacketFactory) CreatePacket() interface{} {
	pack := &protocol.SSPacketMulticast{}
	return pack
}

func (this *MulticastPacketFactory) CreateMulticastPacket(data interface{}, sis ...*protocol.MCSessionUnion) proto.Message {
	pack := &protocol.SSPacketMulticast{Sessions: sis}
	if byteData, ok := data.([]byte); ok {
		pack.Data = byteData
	} else {
		byteData, err := netlib.MarshalPacket(data)
		if err == nil {
			pack.Data = byteData
		} else {
			logger.Warn("MulticastPacketFactory.CreateMulticastPacket err:", err)
		}
	}
	proto.SetDefaults(pack)
	return pack
}

func (this *MulticastHandler) Process(s *netlib.Session, data interface{}) error {
	if mp, ok := data.(*protocol.SSPacketMulticast); ok {
		pd := mp.GetData()
		sis := mp.GetSessions()
		for _, si := range sis {
			ns := this.getSession(si)
			if ns != nil {
				ns.Send(pd, true)
			}
		}
	}
	return nil
}

func (this *MulticastHandler) getSession(su *protocol.MCSessionUnion) *netlib.Session {
	cs := su.GetMccs()
	if cs != nil {
		return srvlib.ClientSessionMgrSington.GetSession(cs.GetSId())
	}

	ss := su.GetMcss()
	if ss != nil {
		return srvlib.ServerSessionMgrSington.GetSession(int(ss.GetSArea()), int(ss.GetSType()), int(ss.GetSId()))
	}

	return nil
}
