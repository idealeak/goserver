package srvlib

import (
	"github.com/idealeak/goserver/core/netlib"
)

const (
	SessionIdSeqIdBits     uint32 = 32
	SessionIdSrvIdBits            = 16
	SessionIdSrvTypeBits          = 8
	SessionIdSrvAreaIdBits        = 8
	SessionIdSrvIdOffset          = SessionIdSeqIdBits
	SessionIdSrvTypeOffset        = SessionIdSrvIdOffset + SessionIdSrvIdBits
	SessionIdSrvAreaOffset        = SessionIdSrvTypeOffset + SessionIdSrvTypeBits
	SessionIdSeqIdMask            = 1<<SessionIdSeqIdBits - 1
	SessionIdSrvIdMask            = 1<<SessionIdSrvIdBits - 1
	SessionIdSrvTypeMask          = 1<<SessionIdSrvTypeBits - 1
	SessionIdSrvAreaIdMask        = 1<<SessionIdSrvAreaIdBits - 1
)

type SessionId int64

func NewSessionId(s *netlib.Session) SessionId {
	sc := s.GetSessionConfig()
	id := int64(sc.AreaId&SessionIdSrvAreaIdMask)<<SessionIdSrvAreaOffset | int64(sc.Type&SessionIdSrvTypeMask)<<SessionIdSrvTypeOffset | int64(sc.Id&SessionIdSrvIdMask)<<SessionIdSrvIdOffset | int64(s.Id)&SessionIdSeqIdMask
	return SessionId(id)
}

func NewSessionIdEx(areaId, srvType, srvId, seq int32) SessionId {
	id := int64(areaId&SessionIdSrvAreaIdMask)<<SessionIdSrvAreaOffset | int64(srvType&SessionIdSrvTypeMask)<<SessionIdSrvTypeOffset | int64(srvId&SessionIdSrvIdMask)<<SessionIdSrvIdOffset | int64(seq)&SessionIdSeqIdMask
	return SessionId(id)
}

func (id SessionId) IsNil() bool {
	return int64(id) == int64(0)
}

func (id SessionId) Get() int64 {
	return int64(id)
}

func (id *SessionId) Set(sid int64) {
	*id = SessionId(sid)
}

func (id SessionId) AreaId() uint32 {
	return uint32(id>>SessionIdSrvAreaOffset) & SessionIdSrvAreaIdMask
}

func (id SessionId) SrvType() uint32 {
	return uint32(id>>SessionIdSrvTypeOffset) & SessionIdSrvTypeMask
}

func (id SessionId) SrvId() uint32 {
	return uint32(id>>SessionIdSrvIdOffset) & SessionIdSrvIdMask
}

func (id SessionId) SeqId() uint32 {
	return uint32(id) & SessionIdSeqIdMask
}
