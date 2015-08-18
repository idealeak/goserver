package main

import (
	"time"

	"code.google.com/p/goprotobuf/proto"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/mmo/protocol"
	"github.com/idealeak/goserver/srvlib"
	libproto "github.com/idealeak/goserver/srvlib/protocol"
)

var (
	SessionHandlerClientBalanceName = "handler-client-balance"
	SessionHandlerClientBalanceMgr  = &SessionHandlerClientBalance{gates: make(map[int32]*gateService)}
)

type gateService struct {
	*libproto.ServiceInfo
	load   int
	active bool
}

type SessionHandlerClientBalance struct {
	gates map[int32]*gateService
}

func (sfcb SessionHandlerClientBalance) GetName() string {
	return SessionHandlerClientBalanceName
}

func (sfcb *SessionHandlerClientBalance) GetInterestOps() uint {
	return 1 << netlib.InterestOps_Opened
}

func (sfcb *SessionHandlerClientBalance) OnSessionOpened(s *netlib.Session) {
	logger.Trace("SessionHandlerClientBalance.OnSessionOpened")
	services := srvlib.ServiceMgr.GetServices(srvlib.ClientServiceType)
	if services != nil {
		/*清理掉线的gate*/
		for k, _ := range sfcb.gates {
			if _, has := services[k]; !has {
				delete(sfcb.gates, k)
			}
		}
		/*补充新上线的gate*/
		for k, v := range services {
			if _, has := sfcb.gates[k]; !has {
				sfcb.gates[k] = &gateService{ServiceInfo: v, active: true}
			}
		}
	}

	/*查找最小负载的gate*/
	var mls *libproto.ServiceInfo
	var min = 100000
	for _, v := range sfcb.gates {
		if v.active && v.load < min {
			mls = v.ServiceInfo
		}
	}
	pack := &protocol.SCGateInfo{}
	if mls != nil {
		pack.SrvType = proto.Int32(mls.GetSrvType())
		pack.SrvId = proto.Int32(mls.GetSrvId())
		pack.AuthKey = proto.String(mls.GetAuthKey())
		pack.Ip = proto.String(mls.GetIp())
		pack.Port = proto.Int32(mls.GetPort())
	}
	proto.SetDefaults(pack)
	s.Send(pack)
	time.AfterFunc(time.Second*5, func() { s.Close() })
}

func (sfcb *SessionHandlerClientBalance) OnSessionClosed(s *netlib.Session) {
}

func (sfcb *SessionHandlerClientBalance) OnSessionIdle(s *netlib.Session) {
}

func (sfcb *SessionHandlerClientBalance) OnPacketReceived(s *netlib.Session, packetid int, packet interface{}) {
}

func (sfcb *SessionHandlerClientBalance) OnPacketSent(s *netlib.Session, data []byte) {
}

func init() {
	netlib.RegisteSessionHandlerCreator(SessionHandlerClientBalanceName, func() netlib.SessionHandler {
		return SessionHandlerClientBalanceMgr
	})

	netlib.RegisterFactory(int(protocol.MmoPacketID_PACKET_GB_CUR_LOAD), netlib.PacketFactoryWrapper(func() interface{} {
		return &protocol.ServerLoad{}
	}))

	netlib.RegisterHandler(int(protocol.MmoPacketID_PACKET_GB_CUR_LOAD), netlib.HandlerWrapper(func(s *netlib.Session, pack interface{}) error {
		logger.Trace("receive gate load info==", pack)
		if sr, ok := pack.(*protocol.ServerLoad); ok {
			srvid := sr.GetSrvId()
			if v, has := SessionHandlerClientBalanceMgr.gates[srvid]; has {
				v.load = int(sr.GetCurLoad())
			}
		}
		return nil
	}))
}
