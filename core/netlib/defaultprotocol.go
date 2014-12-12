// protocol
package netlib

import (
	"encoding/binary"
)

var (
	LenOfPacketHeader int
	LenOfProtoHeader  int
	MaxPacketSize     int = 1024
)

type ProtoHeader struct {
	Len uint16
	Seq uint16
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
}
