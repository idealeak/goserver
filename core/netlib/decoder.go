package netlib

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var (
	DefaultProtocoDecoderName = "default-protocol-decoder"
	SessionAttributeRcvBuf    = &RWBuffer{}
	protocolDecoders          = make(map[string]ProtocolDecoder)
)

type ProtocolDecoder interface {
	Decode(s *Session, r io.Reader) (packetid int, packet interface{}, err error)
	FinishDecode(s *Session)
}

type DefaultProtocolDecoder struct {
}

func (pdi *DefaultProtocolDecoder) Decode(s *Session, r io.Reader) (packetid int, packet interface{}, err error) {
	attr := s.GetAttribute(SessionAttributeRcvBuf)
	if attr == nil {
		attr = AllocRWBuf()
		s.SetAttribute(SessionAttributeRcvBuf, attr)
	}
	if attr == nil {
		err = errors.New("Session rdbuf set failed")
		return
	}

	rdbuf := attr.(*RWBuffer)
	_, err = io.ReadFull(r, rdbuf.buf[:LenOfProtoHeader])
	if err != nil {
		return
	}
	err = binary.Read(bytes.NewBuffer(rdbuf.buf[:LenOfProtoHeader]), binary.LittleEndian, &rdbuf.pheader)
	if err != nil {
		return
	}
	if int(rdbuf.pheader.Len) > MaxPacketSize {
		err = fmt.Errorf("PacketHeader len exceed MaxPacket. get %v limit %v", rdbuf.pheader.Len, MaxPacketSize)
		return
	}
	if rdbuf.pheader.Seq != rdbuf.seq+1 {
		err = fmt.Errorf("PacketHeader sno not matched. get %v want %v", rdbuf.pheader.Seq, rdbuf.seq+1)
		return
	}
	rdbuf.seq++
	_, err = io.ReadFull(r, rdbuf.buf[0:rdbuf.pheader.Len])
	if err != nil {
		return
	}
	packetid, packet, err = UnmarshalPacket(rdbuf.buf[0:rdbuf.pheader.Len])
	if err != nil {
		return
	}
	return
}

func (pdi *DefaultProtocolDecoder) FinishDecode(s *Session) {
	attr := s.GetAttribute(SessionAttributeRcvBuf)
	if attr != nil {
		FreeRWBuf(attr.(*RWBuffer))
	}
}

func RegisteProtocolDecoder(name string, dec ProtocolDecoder) {
	if _, exist := protocolDecoders[name]; exist {
		panic("repeated registe protocol decoder:" + name)
	}
	protocolDecoders[name] = dec
}

func GetProtocolDecoder(name string) ProtocolDecoder {
	if dec, exist := protocolDecoders[name]; exist {
		return dec
	}

	return nil
}

func init() {
	RegisteProtocolDecoder(DefaultProtocoDecoderName, &DefaultProtocolDecoder{})
}
