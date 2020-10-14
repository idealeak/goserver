package netlib

import (
	"container/list"
)

var (
	sessionFilterCreatorPool = make(map[string]SessionFilterCreator)
)

const (
	InterestOps_Opened uint = iota
	InterestOps_Closed
	InterestOps_Idle
	InterestOps_Received
	InterestOps_Sent
	InterestOps_Max
)

type SessionFilterCreator func() SessionFilter

type SessionFilter interface {
	GetName() string
	GetInterestOps() uint
	OnSessionOpened(s *Session) bool                                                    //run in main goroutine
	OnSessionClosed(s *Session) bool                                                    //run in main goroutine
	OnSessionIdle(s *Session) bool                                                      //run in main goroutine
	OnPacketReceived(s *Session, packetid int, logicNo uint32, packet interface{}) bool //run in session receive goroutine
	OnPacketSent(s *Session, packetid int, logicNo uint32, data []byte) bool            //run in session send goroutine
}

type BasicSessionFilter struct {
}

func (bsf *BasicSessionFilter) GetName() string                 { return "BasicSessionFilter" }
func (bsf *BasicSessionFilter) GetInterestOps() uint            { return 0 }
func (bsf *BasicSessionFilter) OnSessionOpened(s *Session) bool { return true }
func (bsf *BasicSessionFilter) OnSessionClosed(s *Session) bool { return true }
func (bsf *BasicSessionFilter) OnSessionIdle(s *Session) bool   { return true }
func (bsf *BasicSessionFilter) OnPacketReceived(s *Session, packetid int, logicNo uint32, packet interface{}) bool {
	return true
}
func (bsf *BasicSessionFilter) OnPacketSent(s *Session, packetid int, logicNo uint32, data []byte) bool {
	return true
}

type SessionFilterChain struct {
	filters            *list.List
	filtersInterestOps [InterestOps_Max]*list.List
}

func NewSessionFilterChain() *SessionFilterChain {
	sfc := &SessionFilterChain{
		filters: list.New(),
	}
	for i := uint(0); i < InterestOps_Max; i++ {
		sfc.filtersInterestOps[i] = list.New()
	}
	return sfc
}

func (sfc *SessionFilterChain) AddFirst(sf SessionFilter) {
	sfc.filters.PushFront(sf)
	ops := sf.GetInterestOps()
	for i := uint(0); i < InterestOps_Max; i++ {
		if ops&(1<<i) != 0 {
			sfc.filtersInterestOps[i].PushFront(sf)
		}
	}
}

func (sfc *SessionFilterChain) AddLast(sf SessionFilter) {
	sfc.filters.PushBack(sf)
	ops := sf.GetInterestOps()
	for i := uint(0); i < InterestOps_Max; i++ {
		if ops&(1<<i) != 0 {
			sfc.filtersInterestOps[i].PushBack(sf)
		}
	}
}

func (sfc *SessionFilterChain) GetFilter(name string) SessionFilter {
	for e := sfc.filters.Front(); e != nil; e = e.Next() {
		sf := e.Value.(SessionFilter)
		if sf != nil && sf.GetName() == name {
			return sf
		}
	}
	return nil
}

func (sfc *SessionFilterChain) OnSessionOpened(s *Session) bool {
	for e := sfc.filtersInterestOps[InterestOps_Opened].Front(); e != nil; e = e.Next() {
		sf := e.Value.(SessionFilter)
		if sf != nil {
			if !sf.OnSessionOpened(s) {
				return false
			}
		}
	}
	return true
}

func (sfc *SessionFilterChain) OnSessionClosed(s *Session) bool {
	for e := sfc.filtersInterestOps[InterestOps_Closed].Front(); e != nil; e = e.Next() {
		sf := e.Value.(SessionFilter)
		if sf != nil {
			if !sf.OnSessionClosed(s) {
				return false
			}
		}
	}
	return true
}

func (sfc *SessionFilterChain) OnSessionIdle(s *Session) bool {
	for e := sfc.filtersInterestOps[InterestOps_Idle].Front(); e != nil; e = e.Next() {
		sf := e.Value.(SessionFilter)
		if sf != nil {
			if !sf.OnSessionIdle(s) {
				return false
			}
		}
	}
	return true
}

func (sfc *SessionFilterChain) OnPacketReceived(s *Session, packetid int, logicNo uint32, packet interface{}) bool {
	for e := sfc.filtersInterestOps[InterestOps_Received].Front(); e != nil; e = e.Next() {
		sf := e.Value.(SessionFilter)
		if sf != nil {
			if !sf.OnPacketReceived(s, packetid, logicNo, packet) {
				return false
			}
		}
	}
	return true
}

func (sfc *SessionFilterChain) OnPacketSent(s *Session, packetid int, logicNo uint32, data []byte) bool {
	for e := sfc.filtersInterestOps[InterestOps_Sent].Front(); e != nil; e = e.Next() {
		sf := e.Value.(SessionFilter)
		if sf != nil {
			if !sf.OnPacketSent(s, packetid, logicNo, data) {
				return false
			}
		}
	}
	return true
}

func RegisteSessionFilterCreator(filterName string, sfc SessionFilterCreator) {
	if sfc == nil {
		return
	}
	if _, exist := sessionFilterCreatorPool[filterName]; exist {
		panic("repeate registe SessionFilter:" + filterName)
	}

	sessionFilterCreatorPool[filterName] = sfc
}

func GetSessionFilterCreator(name string) SessionFilterCreator {
	if sfc, exist := sessionFilterCreatorPool[name]; exist {
		return sfc
	}
	return nil
}
