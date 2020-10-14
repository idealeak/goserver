package filter

import (
	"container/list"
	"net"
	"sync"

	"github.com/idealeak/goserver/core/netlib"
)

var (
	BlackListFilterName = "session-filter-blacklist"
)

type BlackListFilter struct {
	lock      sync.RWMutex // locker
	blacklist *list.List
}

func (blf *BlackListFilter) GetName() string {
	return BlackListFilterName
}

func (blf *BlackListFilter) GetInterestOps() uint {
	return 1<<netlib.InterestOps_Opened | 1<<netlib.InterestOps_Received | 1<<netlib.InterestOps_Sent
}

func (blf *BlackListFilter) OnSessionOpened(s *netlib.Session) bool {
	if blf.isBlock(s) {
		blf.blockSession(s)
		return false
	}
	return true
}

func (blf *BlackListFilter) OnSessionClosed(s *netlib.Session) bool {
	return true
}

func (blf *BlackListFilter) OnSessionIdle(s *netlib.Session) bool {
	return true
}

func (blf *BlackListFilter) OnPacketReceived(s *netlib.Session, packetid int, logicNo uint32, packet interface{}) bool {
	if blf.isBlock(s) {
		blf.blockSession(s)
		return false
	}
	return true
}

func (blf *BlackListFilter) OnPacketSent(s *netlib.Session, packetid int, logicNo uint32, data []byte) bool {
	if blf.isBlock(s) {
		blf.blockSession(s)
		return false
	}
	return true
}

func (blf *BlackListFilter) isBlock(s *netlib.Session) bool {
	host, _, err := net.SplitHostPort(s.RemoteAddr())
	if err != nil {
		return true
	}

	ip := net.ParseIP(host)
	blf.lock.RLock()
	defer blf.lock.RUnlock()
	for e := blf.blacklist.Front(); e != nil; e = e.Next() {
		if e.Value.(*net.IPNet).Contains(ip) {
			return true
		}
	}
	return false
}

func (blf *BlackListFilter) blockSession(s *netlib.Session) {
	s.Close()
}

func (blf *BlackListFilter) Block(ipnet *net.IPNet) {
	blf.lock.Lock()
	defer blf.lock.Unlock()
	blf.blacklist.PushBack(ipnet)
}

func (blf *BlackListFilter) UnBlock(ipnet *net.IPNet) {
	blf.lock.Lock()
	defer blf.lock.Unlock()
	for e := blf.blacklist.Front(); e != nil; e = e.Next() {
		if e.Value.(*net.IPNet).String() == ipnet.String() {
			blf.blacklist.Remove(e)
			return
		}
	}
}

func init() {
	netlib.RegisteSessionFilterCreator(BlackListFilterName, func() netlib.SessionFilter {
		return &BlackListFilter{blacklist: list.New()}
	})
}
