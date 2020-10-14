package main

import (
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/transact"
	"github.com/idealeak/goserver/examples/protocol"
	"github.com/idealeak/goserver/srvlib"
)

type traceTransHandler struct {
}

func init() {
	transact.RegisteHandler(protocol.TxTrace, &traceTransHandler{})
	srvlib.ServerSessionMgrSington.AddListener(&MyServerSessionRegisteListener{})
}

func (this *traceTransHandler) OnExcute(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("traceTransHandler.OnExcute ")
	userData := &protocol.StructA{}
	err := netlib.UnmarshalPacketNoPackId(ud.([]byte), userData)
	if err == nil {
		logger.Logger.Tracef("==========%#v", userData)
	}
	return transact.TransExeResult_Success
}

func (this *traceTransHandler) OnCommit(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("traceTransHandler.OnCommit ")
	return transact.TransExeResult_Success
}

func (this *traceTransHandler) OnRollBack(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("traceTransHandler.OnRollBack ")
	return transact.TransExeResult_Success
}

func (this *traceTransHandler) OnChildTransRep(tNode *transact.TransNode, hChild transact.TransNodeID, retCode int, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("traceTransHandler.OnChildTransRep ")
	return transact.TransExeResult_Success
}

type MyServerSessionRegisteListener struct {
}

func (mssrl *MyServerSessionRegisteListener) OnRegiste(*netlib.Session) {
	logger.Logger.Trace("MyServerSessionRegisteListener.OnRegiste")
}

func (mssrl *MyServerSessionRegisteListener) OnUnregiste(*netlib.Session) {
	logger.Logger.Trace("MyServerSessionRegisteListener.OnUnregiste")
}
