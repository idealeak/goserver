package action

import (
	"errors"
	"strconv"

	"code.google.com/p/goprotobuf/proto"
	"github.com/idealeak/goserver/core/builtin/protocol"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/transact"
)

type TxCtrlCmdPacketFactory struct {
}

type TxCtrlCmdHandler struct {
}

func (this *TxCtrlCmdPacketFactory) CreatePacket() interface{} {
	pack := &protocol.TransactCtrlCmd{}
	return pack
}

func (this *TxCtrlCmdHandler) Process(session *netlib.Session, packetid int, data interface{}) error {
	//logger.Logger.Trace("TxCtrlCmdHandler.Process")
	if txcmd, ok := data.(*protocol.TransactCtrlCmd); ok {
		if !transact.ProcessTransCmd(transact.TransNodeID(txcmd.GetTId()), transact.TransCmd(txcmd.GetCmd())) {
			return errors.New("TxCtrlCmdHandler error, tid=" + strconv.FormatInt(txcmd.GetTId(), 16) + " cmd=" + strconv.Itoa(int(txcmd.GetCmd())))
		}
	}
	return nil
}

func init() {
	netlib.RegisterHandler(int(protocol.CoreBuiltinPacketID_PACKET_SS_TX_CMD), &TxCtrlCmdHandler{})
	netlib.RegisterFactory(int(protocol.CoreBuiltinPacketID_PACKET_SS_TX_CMD), &TxCtrlCmdPacketFactory{})
}

func ConstructTxCmdPacket(tnp *transact.TransNodeParam, cmd transact.TransCmd) proto.Message {
	packet := &protocol.TransactCtrlCmd{
		TId: proto.Int64(int64(tnp.TId)),
		Cmd: proto.Int32(int32(cmd)),
	}
	proto.SetDefaults(packet)
	return packet
}
