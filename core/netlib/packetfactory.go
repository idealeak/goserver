package netlib

import (
	"fmt"
	"reflect"
)

var factories = make(map[int]PacketFactory)
var packetQuickMap = make(map[reflect.Type]packetInfo)

type packetInfo struct {
	ptype int
	pid   int
}

type PacketFactory interface {
	CreatePacket() interface{}
}

type PacketFactoryWrapper func() interface{}

func (pfw PacketFactoryWrapper) CreatePacket() interface{} {
	return pfw()
}

func RegisterFactory(packetId int, factory PacketFactory) {
	if _, ok := factories[packetId]; ok {
		panic(fmt.Sprintf("repeate register packet factory: %v", packetId))
	}

	factories[packetId] = factory
	tp := factory.CreatePacket()
	if tp != nil {
		pt := typetest(tp)
		packetQuickMap[reflect.TypeOf(tp)] = packetInfo{ptype: pt, pid: packetId}
	}
}

func CreatePacket(packetId int) interface{} {
	if v, ok := factories[packetId]; ok {
		return v.CreatePacket()
	}
	return nil
}

func GetPacketTypeAndId(pack interface{}) (int, int) {
	t := reflect.TypeOf(pack)
	if tp, exist := packetQuickMap[t]; exist {
		return tp.ptype, tp.pid
	}
	return 0, 0
}
