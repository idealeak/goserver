package srvlib

import (
	"fmt"
	"reflect"

	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/srvlib/protocol"
)

var handlers = make(map[int]Handler)

type Handler interface {
	Process(s *netlib.Session, packetid int, data interface{}, sid int64, routes []*protocol.SrvInfo) error
}

type HandlerWrapper func(s *netlib.Session, packetid int, data interface{}, sid int64, routes []*protocol.SrvInfo) error

func (hw HandlerWrapper) Process(s *netlib.Session, packetid int, data interface{}, sid int64, routes []*protocol.SrvInfo) error {
	return hw(s, packetid, data, sid, routes)
}

func RegisterHandler(packetId int, h Handler) {
	if _, ok := handlers[packetId]; ok {
		panic(fmt.Sprintf("repeate register handler: %v Handler type=%v", packetId, reflect.TypeOf(h)))
	}

	handlers[packetId] = h
}

func Register1ToMHandler(h Handler, packetIds ...int) {
	for _, packetId := range packetIds {
		RegisterHandler(packetId, h)
	}
}

func RegisterRangeHandler(start, end int, h Handler) {
	for ; start <= end; start++ {
		RegisterHandler(start, h)
	}
}

func GetHandler(packetId int) Handler {
	if h, ok := handlers[packetId]; ok {
		return h
	}

	return nil
}
