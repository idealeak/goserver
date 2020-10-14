package srvlib

import (
	"github.com/idealeak/goserver/core/builtin/action"
	"github.com/idealeak/goserver/core/builtin/protocol"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/transact"
)

type TxCommSkeleton struct {
}

func (tcs *TxCommSkeleton) SendTransResult(parent, me *transact.TransNodeParam, tr *transact.TransResult) bool {
	//logger.Logger.Trace("TxCommSkeleton.SendTransResult")
	p := action.ContructTxResultPacket(parent, me, tr)
	if p == nil {
		return false
	}
	s := ServerSessionMgrSington.GetSession(parent.AreaID, int(parent.Ot), parent.Oid)
	if s == nil {
		logger.Logger.Trace("TxCommSkeleton.SendTransResult s=nil")
		return false
	}

	s.Send(int(protocol.CoreBuiltinPacketID_PACKET_SS_TX_RESULT), p)
	//logger.Logger.Trace("TxCommSkeleton.SendTransResult success")
	return true
}

func (tcs *TxCommSkeleton) SendTransStart(parent, me *transact.TransNodeParam, ud interface{}) bool {
	//logger.Logger.Trace("TxCommSkeleton.SendTransStart")
	p := action.ContructTxStartPacket(parent, me, ud)
	if p == nil {
		return false
	}
	s := ServerSessionMgrSington.GetSession(me.AreaID, int(me.Ot), me.Oid)
	if s == nil {
		logger.Logger.Trace("TxCommSkeleton.SendTransStart s=nil")
		return false
	}

	s.Send(int(protocol.CoreBuiltinPacketID_PACKET_SS_TX_START), p)
	return true
}

func (tcs *TxCommSkeleton) SendCmdToTransNode(tnp *transact.TransNodeParam, c transact.TransCmd) bool {
	//logger.Logger.Trace("TxCommSkeleton.SendCmdToTransNode")
	p := action.ConstructTxCmdPacket(tnp, c)
	if p == nil {
		return false
	}
	s := ServerSessionMgrSington.GetSession(tnp.AreaID, int(tnp.Ot), tnp.Oid)
	if s == nil {
		logger.Logger.Trace("TxCommSkeleton.SendCmdToTransNode s=nil")
		return false
	}

	s.Send(int(protocol.CoreBuiltinPacketID_PACKET_SS_TX_CMD), p)
	return true
}

func (tcs *TxCommSkeleton) GetSkeletonID() int {
	return netlib.Config.SrvInfo.Id
}

func (tcs *TxCommSkeleton) GetAreaID() int {
	return netlib.Config.SrvInfo.AreaID
}

func init() {
	transact.RegisteTxCommSkeleton("github.com/idealeak/goserver/srvlib/txcommskeleton", &TxCommSkeleton{})
}
