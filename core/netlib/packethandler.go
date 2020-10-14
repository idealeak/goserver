// handler
package netlib

import (
	"fmt"
	"reflect"
)

var handlers = make(map[int]Handler)

type Handler interface {
	Process(session *Session, packetid int, data interface{}) error
}

type HandlerWrapper func(session *Session, packetid int, data interface{}) error

func (hw HandlerWrapper) Process(session *Session, packetid int, data interface{}) error {
	return hw(session, packetid, data)
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
