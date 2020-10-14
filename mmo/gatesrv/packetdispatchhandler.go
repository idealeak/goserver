package main

import (
	"bytes"
	"sync/atomic"

	"code.google.com/p/goprotobuf/proto"
	"games.jiexunjiayin.com/jxjyqp/protocol"
	"github.com/idealeak/goserver/core/builtin/filter"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

func init() {
	netlib.RegisteErrorPacketHandlerCreator("packetdispatchhandler", func() netlib.ErrorPacketHandler {
		return netlib.ErrorPacketHandlerWrapper(func(s *netlib.Session, packetid int, logicNo uint32, data []byte) bool {
			if s.GetAttribute(filter.SessionAttributeAuth) == nil {
				logger.Logger.Trace("packetdispatchhandler session not auth! ")
				return false
			}

			bs := BundleMgrSington.GetBundleSession(uint16(s.GroupId))
			if bs == nil {
				logger.Logger.Trace("packetdispatchhandler BundleSession is nil! ")
				return false
			}

			if atomic.CompareAndSwapUint32(&bs.rcvLogicNo, logicNo-1, logicNo) {
				var ss *netlib.Session
				if packetid >= 2000 && packetid < 3000 {
					ss = bs.worldsrvSession
				} else {
					ss = bs.gamesrvSession
				}
				if ss == nil {
					logger.Logger.Trace("packetdispatchhandler redirect server session is nil ", packetid)
					return true
				}
				//must copy
				buf := bytes.NewBuffer(nil)
				buf.Write(data)
				pack := &protocol.SSTransmit{
					SessionId:  proto.Int64(s.Sid),
					PacketData: buf.Bytes(),
				}
				proto.SetDefaults(pack)
				ss.Send(int(protocol.MmoPacketID_PACKET_SS_PACKET_TRANSMIT), pack)
				return true
			}

			//丢掉
			return false
		})
	})
}
