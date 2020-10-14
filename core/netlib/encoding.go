// encoding
package netlib

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"code.google.com/p/goprotobuf/proto"
)

const (
	EncodingTypeNil = iota
	EncodingTypeGPB
	EncodingTypeBinary
	EncodingTypeGob
	EncodingTypeMax
)

var (
	encodingArray [EncodingTypeMax]EncDecoder
	typeTesters   [EncodingTypeMax]TypeTester
)

type EncDecoder interface {
	Unmarshal(buf []byte, pack interface{}) error
	Marshal(pack interface{}) ([]byte, error)
}

type UnparsePacketTypeErr struct {
	EncodeType int16
	PacketId   int16
	Err        error
}

type TypeTester func(pack interface{}) int

func (this *UnparsePacketTypeErr) Error() string {
	return fmt.Sprintf("cannot parse proto type:%v packetid:%v err:%v", this.EncodeType, this.PacketId, this.Err)
}

func NewUnparsePacketTypeErr(et, packid int16, err error) *UnparsePacketTypeErr {
	return &UnparsePacketTypeErr{EncodeType: et, PacketId: packid, Err: err}
}

func UnmarshalPacket(data []byte) (int, interface{}, error) {
	var ph PacketHeader
	err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &ph)
	if err != nil {
		return int(ph.PacketId), nil, err
	}

	if ph.EncodeType >= EncodingTypeMax {
		return int(ph.PacketId), nil, NewUnparsePacketTypeErr(ph.EncodeType, ph.PacketId, fmt.Errorf("EncodeType:%d unregiste", ph.EncodeType))
	}

	pck := CreatePacket(int(ph.PacketId))
	if pck == nil {
		return int(ph.PacketId), nil, NewUnparsePacketTypeErr(ph.EncodeType, ph.PacketId, fmt.Errorf("packetId:%d unregiste", ph.PacketId))
	} else {
		err = encodingArray[ph.EncodeType].Unmarshal(data[LenOfPacketHeader:], pck)
		return int(ph.PacketId), pck, err
	}

	return 0, nil, nil
}

func MarshalPacket(packetid int, pack interface{}) ([]byte, error) {
	et := typetest(pack)
	if et < EncodingTypeNil || et > EncodingTypeMax {
		return nil, fmt.Errorf("MarshalPacket unkown data type:%v", et)
	}

	if encodingArray[et] == nil {
		return nil, fmt.Errorf("MarshalPacket unkown data type:%v", et)
	}

	data, err := encodingArray[et].Marshal(pack)
	if err != nil {
		return nil, fmt.Errorf("%v %v", pack, err.Error())
	}

	ph := PacketHeader{
		EncodeType: int16(et),
		PacketId:   int16(packetid),
	}

	w := bytes.NewBuffer(nil)
	binary.Write(w, binary.LittleEndian, &ph)
	binary.Write(w, binary.LittleEndian, data)
	return w.Bytes(), nil
}

func MarshalPacketNoPackId(pack interface{}) (data []byte, err error) {
	et := typetest(pack)
	if et < EncodingTypeNil || et > EncodingTypeMax {
		return nil, fmt.Errorf("MarshalPacket unkown data type:%v", et)
	}

	if encodingArray[et] == nil {
		return nil, fmt.Errorf("MarshalPacket unkown data type:%v", et)
	}

	data, err = encodingArray[et].Marshal(pack)
	if err != nil {
		return nil, fmt.Errorf("%v %v", pack, err.Error())
	}

	ph := PacketHeader{
		EncodeType: int16(et),
		PacketId:   int16(0),
	}

	w := bytes.NewBuffer(nil)
	binary.Write(w, binary.LittleEndian, &ph)
	binary.Write(w, binary.LittleEndian, data)
	return w.Bytes(), nil
}

func UnmarshalPacketNoPackId(data []byte, pck interface{}) error {
	var ph PacketHeader
	err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &ph)
	if err != nil {
		return err
	}

	if ph.EncodeType >= EncodingTypeMax {
		return NewUnparsePacketTypeErr(ph.EncodeType, ph.PacketId, fmt.Errorf("EncodeType:%d unregiste", ph.EncodeType))
	}

	err = encodingArray[ph.EncodeType].Unmarshal(data[LenOfPacketHeader:], pck)
	if err != nil {
		return NewUnparsePacketTypeErr(ph.EncodeType, ph.PacketId, err)
	}
	return err
}

func typetest(pack interface{}) int {
	switch pack.(type) {
	case proto.Message:
		return EncodingTypeGPB
	case []byte:
		return EncodingTypeBinary
	default:
		return EncodingTypeGob
	}
	return -1
}

func RegisteEncoding(edtype int, ed EncDecoder, tt TypeTester) {
	if encodingArray[edtype] != nil {
		panic(fmt.Sprintf("repeated registe EncDecoder %d", edtype))
	}
	encodingArray[edtype] = ed
	typeTesters[edtype] = tt
}
