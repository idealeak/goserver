// protocol
package netlib

import (
	"encoding/binary"
	"fmt"
)

var (
	LenOfPacketHeader int
	LenOfProtoHeader  int
	MaxPacketSize     int = 64 * 1024
)

type ProtoHeader struct {
	Len     uint16 //包长度
	Seq     uint16 //包序号
	LogicNo uint32 //逻辑号
}

type PacketHeader struct {
	EncodeType int16
	PacketId   int16
}

type RWBuffer struct {
	pheader ProtoHeader
	seq     uint16
	buf     []byte
}

func (rwb *RWBuffer) Init() {
	rwb.seq = 0
}

func init() {
	LenOfPacketHeader = binary.Size(&PacketHeader{})
	LenOfProtoHeader = binary.Size(&ProtoHeader{})
	fmt.Println("sizeof(PacketHeader)=", LenOfPacketHeader, " sizeof(ProtoHeader)=", LenOfProtoHeader)
}
