package netlib

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
)

var (
	DefaultProtocoDecoderName = "default-protocol-decoder"
	protocolDecoders          = make(map[string]ProtocolDecoder)
	ErrRcvBufCannotGet        = errors.New("Session rcvbuf get failed")
)

type ProtocolDecoder interface {
	Decode(s *Session, r io.Reader) (packetid int, logicNo uint32, packet interface{}, err error, raw []byte)
	FinishDecode(s *Session)
}

type DefaultProtocolDecoder struct {
}

func (pdi *DefaultProtocolDecoder) Decode(s *Session, r io.Reader) (packetid int, logicNo uint32, packet interface{}, err error, raw []byte) {
	if s.rcvbuf == nil {
		s.rcvbuf = AllocRWBuf()
	}
	rdbuf := s.rcvbuf
	if rdbuf == nil {
		err = ErrRcvBufCannotGet
		return
	}
	err = binary.Read(r, binary.LittleEndian, &rdbuf.pheader)
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
	logicNo = rdbuf.pheader.LogicNo
	_, err = io.ReadFull(r, rdbuf.buf[0:rdbuf.pheader.Len])
	if err != nil {
		return
	}
	raw = rdbuf.buf[0:rdbuf.pheader.Len]
	packetid, packet, err = UnmarshalPacket(rdbuf.buf[0:rdbuf.pheader.Len])
	if err != nil {
		return
	}

	atomic.AddInt64(&s.recvedBytes, int64(int(rdbuf.pheader.Len)+LenOfProtoHeader))
	atomic.AddInt64(&s.recvedPack, 1)
	return
}

func (pdi *DefaultProtocolDecoder) FinishDecode(s *Session) {
	if s.rcvbuf != nil {
		FreeRWBuf(s.rcvbuf)
		s.rcvbuf = nil
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
