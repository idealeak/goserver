package netlib

var (
	errorPacketHandlerCreatorPool = make(map[string]ErrorPacketHandlerCreator)
)

type ErrorPacketHandlerCreator func() ErrorPacketHandler

type ErrorPacketHandler interface {
	OnErrorPacket(s *Session, packetid int, logicNo uint32, data []byte) bool //run in session receive goroutine
}

type ErrorPacketHandlerWrapper func(session *Session, packetid int, logicNo uint32, data []byte) bool

func (hw ErrorPacketHandlerWrapper) OnErrorPacket(session *Session, packetid int, logicNo uint32, data []byte) bool {
	return hw(session, packetid, logicNo, data)
}

func RegisteErrorPacketHandlerCreator(name string, ephc ErrorPacketHandlerCreator) {
	if ephc == nil {
		return
	}
	if _, exist := errorPacketHandlerCreatorPool[name]; exist {
		panic("repeate registe ErrorPacketHandler:" + name)
	}

	errorPacketHandlerCreatorPool[name] = ephc
}

func GetErrorPacketHandlerCreator(name string) ErrorPacketHandlerCreator {
	if ephc, exist := errorPacketHandlerCreatorPool[name]; exist {
		return ephc
	}
	return nil
}
