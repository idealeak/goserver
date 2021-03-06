// Code generated by protoc-gen-go.
// source: protocol/pingpong.proto
// DO NOT EDIT!

package protocol

import proto "code.google.com/p/goprotobuf/proto"
import json "encoding/json"
import math "math"

// Reference proto, json, and math imports to suppress error if they are not otherwise used.
var _ = proto.Marshal
var _ = &json.SyntaxError{}
var _ = math.Inf

type CSPacketPing struct {
	TimeStamb        *int64 `protobuf:"varint,1,req" json:"TimeStamb,omitempty"`
	Message          []byte `protobuf:"bytes,2,req" json:"Message,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *CSPacketPing) Reset()         { *m = CSPacketPing{} }
func (m *CSPacketPing) String() string { return proto.CompactTextString(m) }
func (*CSPacketPing) ProtoMessage()    {}

func (m *CSPacketPing) GetTimeStamb() int64 {
	if m != nil && m.TimeStamb != nil {
		return *m.TimeStamb
	}
	return 0
}

func (m *CSPacketPing) GetMessage() []byte {
	if m != nil {
		return m.Message
	}
	return nil
}

type SCPacketPong struct {
	TimeStamb        *int64 `protobuf:"varint,1,req" json:"TimeStamb,omitempty"`
	Message          []byte `protobuf:"bytes,2,req" json:"Message,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *SCPacketPong) Reset()         { *m = SCPacketPong{} }
func (m *SCPacketPong) String() string { return proto.CompactTextString(m) }
func (*SCPacketPong) ProtoMessage()    {}

func (m *SCPacketPong) GetTimeStamb() int64 {
	if m != nil && m.TimeStamb != nil {
		return *m.TimeStamb
	}
	return 0
}

func (m *SCPacketPong) GetMessage() []byte {
	if m != nil {
		return m.Message
	}
	return nil
}

func init() {
}
