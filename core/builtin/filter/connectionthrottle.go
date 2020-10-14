package filter

import (
	"net"
	"time"

	"github.com/idealeak/goserver/core/container"
	"github.com/idealeak/goserver/core/netlib"
)

var (
	DefaultAllowedInterval       = 500 //ms
	ConnectionThrottleFilterName = "session-filter-connectionthrottle"
)

type ConnectionThrottleFilter struct {
	clients         *container.SynchronizedMap //need synchronize
	AllowedInterval int                        //ms
}

func (ctf *ConnectionThrottleFilter) GetName() string {
	return ConnectionThrottleFilterName
}

func (ctf *ConnectionThrottleFilter) GetInterestOps() uint {
	return 1 << netlib.InterestOps_Opened
}

func (ctf *ConnectionThrottleFilter) OnSessionOpened(s *netlib.Session) bool {
	if !ctf.isConnectionOk(s) {
		s.Close()
		return false
	}
	return true
}

func (ctf *ConnectionThrottleFilter) OnSessionClosed(s *netlib.Session) bool {
	return true
}

func (ctf *ConnectionThrottleFilter) OnSessionIdle(s *netlib.Session) bool {
	return true
}

func (ctf *ConnectionThrottleFilter) OnPacketReceived(s *netlib.Session, packetid int, logicNo uint32, packet interface{}) bool {
	return true
}

func (ctf *ConnectionThrottleFilter) OnPacketSent(s *netlib.Session, packetid int, logicNo uint32, data []byte) bool {
	return true
}

func (ctf *ConnectionThrottleFilter) isConnectionOk(s *netlib.Session) bool {
	host, _, err := net.SplitHostPort(s.RemoteAddr())
	if err != nil {
		return false
	}

	tNow := time.Now()
	value := ctf.clients.Get(host)
	if value != nil {
		tLast := value.(time.Time)
		if tNow.Sub(tLast) < time.Duration(ctf.AllowedInterval)*time.Millisecond {
			ctf.clients.Set(host, tNow)
			return false
		}
	}

	ctf.clients.Set(host, tNow)
	return true
}

func init() {
	netlib.RegisteSessionFilterCreator(ConnectionThrottleFilterName, func() netlib.SessionFilter {
		return &ConnectionThrottleFilter{clients: container.NewSynchronizedMap(), AllowedInterval: DefaultAllowedInterval}
	})
}
