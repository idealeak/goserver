package main

import (
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/mmo/protocol"
)

func init() {
	netlib.RegisterFactory(int(protocol.MmoPacketID_PACKET_SC_GATEINFO), netlib.PacketFactoryWrapper(func() interface{} {
		return &protocol.SCGateInfo{}
	}))
	netlib.RegisterHandler(int(protocol.MmoPacketID_PACKET_SC_GATEINFO), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive gateinfo==", pack)
		if sr, ok := pack.(*protocol.SCGateInfo); ok {
			sc := &netlib.SessionConfig{
				Id:           int(sr.GetSrvId()),
				Type:         int(sr.GetSrvType()),
				Ip:           sr.GetIp(),
				Port:         int(sr.GetPort()),
				AuthKey:      sr.GetAuthKey(),
				WriteTimeout: 30,
				ReadTimeout:  30,
				IdleTimeout:  30,
				MaxDone:      20,
				MaxPend:      20,
				MaxPacket:    1024,
				RcvBuff:      1024,
				SndBuff:      1024,
				IsClient:     true,
				NoDelay:      true,
				FilterChain:  []string{"session-filter-trace", "session-filter-auth"},
			}
			sc.Init()
			err := netlib.Connect(sc)
			if err != nil {
				logger.Logger.Warn("connect server failed err:", err)
			}
		}
		return nil
	}))
}
