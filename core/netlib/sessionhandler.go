package netlib

import (
	"container/list"
)

var (
	sessionHandlerCreatorPool = make(map[string]SessionHandlerCreator)
)

type SessionHandlerCreator func() SessionHandler

type SessionHandler interface {
	GetName() string
	GetInterestOps() uint
	OnSessionOpened(s *Session)                                                    //run in main goroutine
	OnSessionClosed(s *Session)                                                    //run in main goroutine
	OnSessionIdle(s *Session)                                                      //run in main goroutine
	OnPacketReceived(s *Session, packetid int, logicNo uint32, packet interface{}) //run in session receive goroutine
	OnPacketSent(s *Session, packetid int, logicNo uint32, data []byte)            //run in session send goroutine
}

type BasicSessionHandler struct {
}

func (bsh *BasicSessionHandler) GetName() string            { return "BasicSessionHandler" }
func (bsh *BasicSessionHandler) GetInterestOps() uint       { return 0 }
func (bsh *BasicSessionHandler) OnSessionOpened(s *Session) {}
func (bsh *BasicSessionHandler) OnSessionClosed(s *Session) {}
func (bsh *BasicSessionHandler) OnSessionIdle(s *Session)   {}
func (bsh *BasicSessionHandler) OnPacketReceived(s *Session, packetid int, logicNo uint32, packet interface{}) {
}
func (bsh *BasicSessionHandler) OnPacketSent(s *Session, packetid int, logicNo uint32, data []byte) {}

type SessionHandlerChain struct {
	handlers            *list.List
	handlersInterestOps [InterestOps_Max]*list.List
}

func NewSessionHandlerChain() *SessionHandlerChain {
	shc := &SessionHandlerChain{
		handlers: list.New(),
	}
	for i := uint(0); i < InterestOps_Max; i++ {
		shc.handlersInterestOps[i] = list.New()
	}
	return shc
}

func (shc *SessionHandlerChain) AddFirst(sh SessionHandler) {
	shc.handlers.PushFront(sh)
	ops := sh.GetInterestOps()
	for i := uint(0); i < InterestOps_Max; i++ {
		if ops&(1<<i) != 0 {
			shc.handlersInterestOps[i].PushFront(sh)
		}
	}
}

func (shc *SessionHandlerChain) AddLast(sh SessionHandler) {
	shc.handlers.PushBack(sh)
	ops := sh.GetInterestOps()
	for i := uint(0); i < InterestOps_Max; i++ {
		if ops&(1<<i) != 0 {
			shc.handlersInterestOps[i].PushBack(sh)
		}
	}
}

func (shc *SessionHandlerChain) GetHandler(name string) SessionHandler {
	for e := shc.handlers.Front(); e != nil; e = e.Next() {
		sh := e.Value.(SessionHandler)
		if sh != nil && sh.GetName() == name {
			return sh
		}
	}
	return nil
}

func (shc *SessionHandlerChain) OnSessionOpened(s *Session) {
	for e := shc.handlersInterestOps[InterestOps_Opened].Front(); e != nil; e = e.Next() {
		sh := e.Value.(SessionHandler)
		if sh != nil {
			sh.OnSessionOpened(s)
		}
	}
}

func (shc *SessionHandlerChain) OnSessionClosed(s *Session) {
	for e := shc.handlersInterestOps[InterestOps_Closed].Front(); e != nil; e = e.Next() {
		sh := e.Value.(SessionHandler)
		if sh != nil {
			sh.OnSessionClosed(s)
		}
	}
}

func (shc *SessionHandlerChain) OnSessionIdle(s *Session) {
	for e := shc.handlersInterestOps[InterestOps_Idle].Front(); e != nil; e = e.Next() {
		sh := e.Value.(SessionHandler)
		if sh != nil {
			sh.OnSessionIdle(s)
		}
	}
}

func (shc *SessionHandlerChain) OnPacketReceived(s *Session, packetid int, logicNo uint32, packet interface{}) {
	for e := shc.handlersInterestOps[InterestOps_Received].Front(); e != nil; e = e.Next() {
		sh := e.Value.(SessionHandler)
		if sh != nil {
			sh.OnPacketReceived(s, packetid, logicNo, packet)
		}
	}
}

func (shc *SessionHandlerChain) OnPacketSent(s *Session, packetid int, logicNo uint32, data []byte) {
	for e := shc.handlersInterestOps[InterestOps_Sent].Front(); e != nil; e = e.Next() {
		sh := e.Value.(SessionHandler)
		if sh != nil {
			sh.OnPacketSent(s, packetid, logicNo, data)
		}
	}
}

func RegisteSessionHandlerCreator(name string, shc SessionHandlerCreator) {
	if shc == nil {
		return
	}
	if _, exist := sessionHandlerCreatorPool[name]; exist {
		panic("repeate registe SessionHandler:" + name)
	}

	sessionHandlerCreatorPool[name] = shc
}

func GetSessionHandlerCreator(name string) SessionHandlerCreator {
	if shc, exist := sessionHandlerCreatorPool[name]; exist {
		return shc
	}
	return nil
}
