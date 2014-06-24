package netlib

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

var (
	DefaultProtocolEncoderName    = "default-protocol-encoder"
	DefaultBuiltinProtocolEncoder = &DefaultProtocolEncoder{}
	SessionAttributeSndBuf        = &RWBuffer{}
	protocolEncoders              = make(map[string]ProtocolEncoder)
	ErrSndBufCannotGet            = errors.New("Session sndbuf get failed")
	ErrExceedMaxPacketSize        = errors.New("exceed max packet size")
)

type PacketCutSlicesFunc func(data []byte) []interface{}

type ProtocolEncoder interface {
	Encode(s *Session, packet interface{}, w io.Writer) (data []byte, err error)
	FinishEncode(s *Session)
}

type DefaultProtocolEncoder struct {
	PacketCutor PacketCutSlicesFunc
}

func (dec *DefaultProtocolEncoder) Encode(s *Session, packet interface{}, w io.Writer) (data []byte, err error) {
	attr := s.GetAttribute(SessionAttributeSndBuf)
	if attr == nil {
		attr = AllocRWBuf()
		s.SetAttribute(SessionAttributeSndBuf, attr)
	}
	if attr == nil {
		err = ErrSndBufCannotGet
		return
	}

	var (
		ok bool
	)
	if data, ok = packet.([]byte); !ok {
		data, err = MarshalPacket(packet)
		if err != nil {
			return
		}
	}

	var size int = len(data)
	if size > MaxPacketSize-LenOfProtoHeader {
		if s.sc.SupportFragment {
			err = dec.CutAndSendPacket(s, data, w)
			return
		} else {
			err = ErrExceedMaxPacketSize
			return
		}
	}

	sndbuf := attr.(*RWBuffer)
	//fill packerHeader
	sndbuf.seq++
	sndbuf.pheader.Len = uint16(size)
	sndbuf.pheader.Seq = sndbuf.seq

	err = binary.Write(w, binary.LittleEndian, &sndbuf.pheader)
	if err != nil {
		return
	}
	_, err = io.Copy(w, bytes.NewBuffer(data))
	if err != nil {
		return
	}
	return
}

func (dec *DefaultProtocolEncoder) CutAndSendPacket(s *Session, data []byte, w io.Writer) (err error) {
	if dec.PacketCutor != nil {
		slices := dec.PacketCutor(data)
		for i := 0; i < len(slices); i++ {
			_, err = dec.Encode(s, slices[i], w)
			if err != nil {
				return
			}
		}
	}
	return
}

func (dec *DefaultProtocolEncoder) FinishEncode(s *Session) {
	attr := s.GetAttribute(SessionAttributeSndBuf)
	if attr != nil {
		FreeRWBuf(attr.(*RWBuffer))
	}
}

func RegisteProtocolEncoder(name string, enc ProtocolEncoder) {
	if _, exist := protocolEncoders[name]; exist {
		panic("repeated registe protocol encoder:" + name)
	}
	protocolEncoders[name] = enc
}

func GetProtocolEncoder(name string) ProtocolEncoder {
	if enc, exist := protocolEncoders[name]; exist {
		return enc
	}

	return nil
}

func init() {
	RegisteProtocolEncoder(DefaultProtocolEncoderName, DefaultBuiltinProtocolEncoder)
}
