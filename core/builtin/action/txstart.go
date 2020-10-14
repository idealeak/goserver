package action

import (
	"errors"
	"strconv"
	"time"

	"code.google.com/p/goprotobuf/proto"
	"github.com/idealeak/goserver/core/builtin/protocol"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/transact"
)

type TxStartPacketFactory struct {
}

type TxStartHandler struct {
}

func (this *TxStartPacketFactory) CreatePacket() interface{} {
	pack := &protocol.TransactStart{}
	return pack
}

func (this *TxStartHandler) Process(session *netlib.Session, packetid int, data interface{}) error {
	//logger.Logger.Trace("TxStartHandler.Process")
	if ts, ok := data.(*protocol.TransactStart); ok {
		netptnp := ts.GetParenTNP()
		if netptnp == nil {
			return nil
		}
		netmtnp := ts.GetMyTNP()
		if netmtnp == nil {
			return nil
		}

		ptnp := &transact.TransNodeParam{
			TId:        transact.TransNodeID(netptnp.GetTransNodeID()),
			Tt:         transact.TransType(netptnp.GetTransType()),
			Ot:         transact.TransOwnerType(netptnp.GetOwnerType()),
			Tct:        transact.TransactCommitPolicy(netptnp.GetTransCommitType()),
			Oid:        int(netptnp.GetOwnerID()),
			SkeletonID: int(netptnp.GetSkeletonID()),
			LevelNo:    int(netptnp.GetLevelNo()),
			AreaID:     int(netptnp.GetAreaID()),
			TimeOut:    time.Duration(netptnp.GetTimeOut()),
		}
		mtnp := &transact.TransNodeParam{
			TId:        transact.TransNodeID(netmtnp.GetTransNodeID()),
			Tt:         transact.TransType(netmtnp.GetTransType()),
			Ot:         transact.TransOwnerType(netmtnp.GetOwnerType()),
			Tct:        transact.TransactCommitPolicy(netmtnp.GetTransCommitType()),
			Oid:        int(netmtnp.GetOwnerID()),
			SkeletonID: int(netmtnp.GetSkeletonID()),
			LevelNo:    int(netmtnp.GetLevelNo()),
			AreaID:     int(netmtnp.GetAreaID()),
			TimeOut:    time.Duration(netmtnp.GetTimeOut()),
		}

		if !transact.ProcessTransStart(ptnp, mtnp, ts.GetCustomData(), mtnp.TimeOut) {
			return errors.New("TxStartHandler error, tid=" + strconv.FormatInt(netmtnp.GetTransNodeID(), 16))
		}
	}
	return nil
}

func init() {
	netlib.RegisterHandler(int(protocol.CoreBuiltinPacketID_PACKET_SS_TX_START), &TxStartHandler{})
	netlib.RegisterFactory(int(protocol.CoreBuiltinPacketID_PACKET_SS_TX_START), &TxStartPacketFactory{})
}

func ContructTxStartPacket(parent, me *transact.TransNodeParam, ud interface{}) proto.Message {
	packet := &protocol.TransactStart{
		MyTNP: &protocol.TransactParam{
			TransNodeID:     proto.Int64(int64(me.TId)),
			TransType:       proto.Int32(int32(me.Tt)),
			OwnerType:       proto.Int32(int32(me.Ot)),
			TransCommitType: proto.Int32(int32(me.Tct)),
			OwnerID:         proto.Int32(int32(me.Oid)),
			SkeletonID:      proto.Int32(int32(me.SkeletonID)),
			LevelNo:         proto.Int32(int32(me.LevelNo)),
			AreaID:          proto.Int32(int32(me.AreaID)),
			TimeOut:         proto.Int64(int64(me.TimeOut)),
		},
		ParenTNP: &protocol.TransactParam{
			TransNodeID:     proto.Int64(int64(parent.TId)),
			TransType:       proto.Int32(int32(parent.Tt)),
			OwnerType:       proto.Int32(int32(parent.Ot)),
			TransCommitType: proto.Int32(int32(parent.Tct)),
			OwnerID:         proto.Int32(int32(parent.Oid)),
			SkeletonID:      proto.Int32(int32(parent.SkeletonID)),
			LevelNo:         proto.Int32(int32(parent.LevelNo)),
			AreaID:          proto.Int32(int32(parent.AreaID)),
			TimeOut:         proto.Int64(int64(parent.TimeOut)),
		},
	}

	if ud != nil {
		b, err := netlib.MarshalPacketNoPackId(ud)
		if err != nil {
			logger.Logger.Warn("ContructTxStartPacket Marshal UserData error:", err)
		} else {
			packet.CustomData = b
		}
	}
	proto.SetDefaults(packet)
	return packet
}
