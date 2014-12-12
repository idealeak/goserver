package netlib

import (
	"fmt"
)

var factories = make(map[int]PacketFactory)

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
}

func CreatePacket(packetId int) interface{} {
	if v, ok := factories[packetId]; ok {
		return v.CreatePacket()
	}
	return nil
}
