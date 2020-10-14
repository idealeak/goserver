package action

import (
	"bytes"
	"errors"
	"strconv"

	"code.google.com/p/goprotobuf/proto"
	"github.com/idealeak/goserver/core/builtin/protocol"
	"github.com/idealeak/goserver/core/netlib"
)

var (
	SessionAttributeBigBuf = &PacketSlicesHandler{}
)

type PacketSlicesPacketFactory struct {
}

type PacketSlicesHandler struct {
}

func (this *PacketSlicesPacketFactory) CreatePacket() interface{} {
	pack := &protocol.SSPacketSlices{}
	return pack
}

func (this *PacketSlicesHandler) Process(s *netlib.Session, packetid int, data interface{}) error {
	if packetslices, ok := data.(*protocol.SSPacketSlices); ok {
		seqNo := int(packetslices.GetSeqNo())
		if seqNo < 1 {
			return errors.New("PacketSlicesHandler unexpect packet seq:" + strconv.Itoa(seqNo))
		}
		totalSize := int(packetslices.GetTotalSize())
		if totalSize > s.GetSessionConfig().MaxPacket {
			return errors.New("PacketSlicesHandler exceed MaxPacket size:" + strconv.Itoa(s.GetSessionConfig().MaxPacket) + " size=" + strconv.Itoa(totalSize))
		}
		attr := s.GetAttribute(SessionAttributeBigBuf)
		if seqNo == 1 {
			if attr == nil {
				attr = bytes.NewBuffer(make([]byte, 0, packetslices.GetTotalSize()))
				s.SetAttribute(SessionAttributeBigBuf, attr)
			}
		}
		if seqNo > 1 {
			if attr == nil {
				return errors.New("PacketSlicesHandler Incorrect packet seq, expect seq=1")
			}
		} else if attr == nil {
			return errors.New("PacketSlicesHandler get bytesbuf failed")
		}

		buf := attr.(*bytes.Buffer)
		if seqNo == 1 {
			buf.Reset()
		}
		if buf.Len() != int(packetslices.GetOffset()) {
			return errors.New("PacketSlicesHandler get next packet offset error")
		}
		buf.Write(packetslices.GetPacketData())
		if buf.Len() == totalSize {
			packetid, pck, err := netlib.UnmarshalPacket(buf.Bytes())
			if err != nil {
				return err
			}
			h := netlib.GetHandler(packetid)
			if h != nil {
				h.Process(s, packetid, pck)
			}
		}
	}
	return nil
}

func init() {
	netlib.RegisterHandler(int(protocol.CoreBuiltinPacketID_PACKET_SS_SLICES), &PacketSlicesHandler{})
	netlib.RegisterFactory(int(protocol.CoreBuiltinPacketID_PACKET_SS_SLICES), &PacketSlicesPacketFactory{})

	netlib.DefaultBuiltinProtocolEncoder.PacketCutor = func(data []byte) (packid int, packs []interface{}) {

		var (
			offset    = 0
			sendSize  = 0
			seqNo     = 1
			totalSize = len(data)
			restSize  = len(data)
		)
		packid = int(protocol.CoreBuiltinPacketID_PACKET_SS_SLICES)
		for restSize > 0 {
			sendSize = restSize
			if sendSize > netlib.MaxPacketSize-128 {
				sendSize = netlib.MaxPacketSize - 128
			}
			pack := &protocol.SSPacketSlices{
				SeqNo:      proto.Int32(int32(seqNo)),
				TotalSize:  proto.Int32(int32(totalSize)),
				Offset:     proto.Int32(int32(offset)),
				PacketData: data[offset : offset+sendSize],
			}
			proto.SetDefaults(pack)
			seqNo++
			restSize -= sendSize
			offset += sendSize
			packs = append(packs, pack)
		}
		return
	}
}
