package netlib

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"sync"
	"sync/atomic"
)

var (
	DefaultProtocolEncoderName    = "default-protocol-encoder"
	DefaultBuiltinProtocolEncoder = &DefaultProtocolEncoder{}
	protocolEncoders              = make(map[string]ProtocolEncoder)
	ErrSndBufCannotGet            = errors.New("Session sndbuf get failed")
	ErrExceedMaxPacketSize        = errors.New("exceed max packet size")
)

type PacketCutSlicesFunc func(data []byte) (int, []interface{})

var bytesBufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 512)
	},
}

type ProtocolEncoder interface {
	Encode(s *Session, packetid int, logicNo uint32, packet interface{}, w io.Writer) (data []byte, err error)
	FinishEncode(s *Session)
}

type DefaultProtocolEncoder struct {
	PacketCutor PacketCutSlicesFunc
}

func (dec *DefaultProtocolEncoder) Encode(s *Session, packetid int, logicNo uint32, packet interface{}, w io.Writer) (data []byte, err error) {
	if s.sndbuf == nil {
		s.sndbuf = AllocRWBuf()
	}
	sndbuf := s.sndbuf
	if sndbuf == nil {
		err = ErrSndBufCannotGet
		return
	}

	var (
		ok bool
	)
	if data, ok = packet.([]byte); !ok {
		data, err = MarshalPacket(packetid, packet)
		if err != nil {
			return
		}
	}

	var size int = len(data)
	if size > MaxPacketSize-LenOfProtoHeader {
		if s.sc.SupportFragment {
			err = dec.CutAndSendPacket(s, logicNo, data, w)
			return
		} else {
			err = ErrExceedMaxPacketSize
			return
		}
	}

	//fill packerHeader
	sndbuf.seq++
	sndbuf.pheader.Len = uint16(size)
	sndbuf.pheader.Seq = sndbuf.seq
	sndbuf.pheader.LogicNo = logicNo

	buf := bytesBufferPool.Get().([]byte)
	defer func() {
		bytesBufferPool.Put(buf[:0])
	}()
	ioBuf := bytes.NewBuffer(buf)

	//err = binary.Write(w, binary.LittleEndian, &sndbuf.pheader)
	err = binary.Write(ioBuf, binary.LittleEndian, sndbuf.pheader.Len)
	err = binary.Write(ioBuf, binary.LittleEndian, sndbuf.pheader.Seq)
	err = binary.Write(ioBuf, binary.LittleEndian, sndbuf.pheader.LogicNo)
	if err != nil {
		return
	}

	lenPack := len(data)
	_, err = ioBuf.Write(data[:])
	if err != nil {
		return
	}
	_, err = w.Write(ioBuf.Bytes())
	//_, err = w.Write(data[:])
	//_, err = io.Copy(w, bytes.NewBuffer(data))
	if err != nil {
		return
	}

	atomic.AddInt64(&s.sendedBytes, int64(lenPack+LenOfProtoHeader))
	atomic.AddInt64(&s.sendedPack, 1)
	return
}

func (dec *DefaultProtocolEncoder) CutAndSendPacket(s *Session, logicNo uint32, data []byte, w io.Writer) (err error) {
	if dec.PacketCutor != nil {
		packid, slices := dec.PacketCutor(data)
		for i := 0; i < len(slices); i++ {
			_, err = dec.Encode(s, packid, logicNo, slices[i], w)
			if err != nil {
				return
			}
			if s.scpl != nil {
				err = s.scpl.onCutPacket(w)
				if err != nil {
					return
				}
			}
		}
	}
	return
}

func (dec *DefaultProtocolEncoder) FinishEncode(s *Session) {
	if s.sndbuf != nil {
		FreeRWBuf(s.sndbuf)
		s.sndbuf = nil
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
