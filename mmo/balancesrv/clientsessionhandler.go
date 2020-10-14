package main

import (
	"code.google.com/p/goprotobuf/proto"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/mmo/protocol"
	"github.com/idealeak/goserver/srvlib"
)

var (
	SessionHandlerClientBalanceName = "handler-client-balance"
	SessionHandlerClientBalanceMgr  = &SessionHandlerClientBalance{gates: make(map[int32]*gateService)}
)

type gateService struct {
	load   int
	active bool
}

type SessionHandlerClientBalance struct {
	netlib.BasicSessionHandler
	gates map[int32]*gateService
}

func (sfcb SessionHandlerClientBalance) GetName() string {
	return SessionHandlerClientBalanceName
}

func (sfcb *SessionHandlerClientBalance) GetInterestOps() uint {
	return 1 << netlib.InterestOps_Opened
}

func (sfcb *SessionHandlerClientBalance) OnSessionOpened(s *netlib.Session) {
	logger.Logger.Trace("SessionHandlerClientBalance.OnSessionOpened")
	services := srvlib.ServiceMgr.GetServices(srvlib.ClientServiceType)
	if services != nil {
		/*清理掉线的gate*/
		for k, _ := range sfcb.gates {
			if _, has := services[k]; !has {
				logger.Logger.Trace("gate leave: ", k)
				delete(sfcb.gates, k)
			}
		}
		/*补充新上线的gate*/
		for k, v := range services {
			if _, has := sfcb.gates[k]; !has {
				sfcb.gates[k] = &gateService{active: true}
				logger.Logger.Trace("new gate come in: ", k, v)
			}
		}
	}

	/*查找最小负载的gate*/
	var minsrvid int32
	var min = 100000
	for k, v := range sfcb.gates {
		if v.active && v.load < min {
			minsrvid = k
			min = v.load
		}
	}

	pack := &protocol.SCGateInfo{}
	if mls, has := services[minsrvid]; has {
		pack.SrvType = proto.Int32(mls.GetSrvType())
		pack.SrvId = proto.Int32(mls.GetSrvId())
		pack.AuthKey = proto.String(mls.GetAuthKey())
		pack.Ip = proto.String(mls.GetOuterIp())
		pack.Port = proto.Int32(mls.GetPort())
	}
	proto.SetDefaults(pack)
	s.Send(int(protocol.MmoPacketID_PACKET_SC_GATEINFO), pack)
	logger.Logger.Trace(pack)
	s.Close()
}


func init() {
	netlib.RegisteSessionHandlerCreator(SessionHandlerClientBalanceName, func() netlib.SessionHandler {
		return SessionHandlerClientBalanceMgr
	})

	netlib.RegisterFactory(int(protocol.MmoPacketID_PACKET_GB_CUR_LOAD), netlib.PacketFactoryWrapper(func() interface{} {
		return &protocol.ServerLoad{}
	}))

	netlib.RegisterHandler(int(protocol.MmoPacketID_PACKET_GB_CUR_LOAD), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		if sr, ok := pack.(*protocol.ServerLoad); ok {
			srvid := sr.GetSrvId()
			if v, has := SessionHandlerClientBalanceMgr.gates[srvid]; has {
				v.load = int(sr.GetCurLoad())
				logger.Logger.Trace("receive gate load info 1, sid=", srvid, " load=", v.load)
			} else {
				services := srvlib.ServiceMgr.GetServices(srvlib.ClientServiceType)
				if _, has := services[srvid]; has {
					SessionHandlerClientBalanceMgr.gates[srvid] = &gateService{active: true, load: int(sr.GetCurLoad())}
					logger.Logger.Trace("receive gate load info 2, sid=", srvid, " load=", sr.GetCurLoad())
				}
			}
		}
		return nil
	}))
}
