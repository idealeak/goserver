package main

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

func (this *MulticastPacketFactory) CreateMulticastPacket(packetid int, data interface{}, sis ...*protocol.MCSessionUnion) (proto.Message, error) {
	pack := &protocol.SSPacketMulticast{
		Sessions: sis,
		PacketId: proto.Int(packetid),
	}
	if byteData, ok := data.([]byte); ok {
		pack.Data = byteData
	} else {
		byteData, err := netlib.MarshalPacket(packetid, data)
		if err == nil {
			pack.Data = byteData
		} else {
			logger.Logger.Info("MulticastPacketFactory.CreateMulticastPacket err:", err)
			return nil, err
		}
	}
	proto.SetDefaults(pack)
	return pack, nil
}

func (this *MulticastHandler) Process(s *netlib.Session, packetid int, data interface{}) error {
	if mp, ok := data.(*protocol.SSPacketMulticast); ok {
		pd := mp.GetData()
		sis := mp.GetSessions()
		for _, si := range sis {
			cs := si.GetMccs()
			if cs != nil {
				sid := srvlib.SessionId(cs.GetSId())
				bundleId := sid.SeqId()
				bs := BundleMgrSington.GetBundleSession(uint16(bundleId))
				if bs != nil {
					bs.Send(int(mp.GetPacketId()), pd)
				}
			} else {
				ss := si.GetMcss()
				if ss != nil {
					ns := srvlib.ServerSessionMgrSington.GetSession(int(ss.GetSArea()), int(ss.GetSType()), int(ss.GetSId()))
					if ns != nil {
						ns.Send(int(mp.GetPacketId()), pd /*, s.GetSessionConfig().IsInnerLink*/)
					}
				}
			}
		}
	}
	return nil
}
