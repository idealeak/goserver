package netlib

import (
	"errors"
	"math"
	"time"

	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/utils"
)

const (
	IoServiceMaxCount int = 10
)

var (
	NetModule = newNetEngine()
)

type NetEngine struct {
	pool      map[int]ioService
	childAck  chan int
	backlogSc chan *SessionConfig
	quit      bool
	reaped    bool
}

func newNetEngine() *NetEngine {
	e := &NetEngine{
		pool:      make(map[int]ioService),
		childAck:  make(chan int, IoServiceMaxCount),
		backlogSc: make(chan *SessionConfig, IoServiceMaxCount),
	}

	return e
}

func (e *NetEngine) newIoService(sc *SessionConfig) ioService {
	var s ioService
	if sc.IsClient {
		if !sc.AllowMultiConn && ConnectorMgr.IsConnecting(sc) {
			return nil
		}
		switch sc.Protocol {
		case "ws", "wss":
			s = newWsConnector(e, sc)
		case "udp":
			s = newUdpConnector(e, sc)
		default:
			s = newTcpConnector(e, sc)
		}
	} else {
		switch sc.Protocol {
		case "ws", "wss":
			s = newWsAcceptor(e, sc)
		case "udp":
			s = newUdpAcceptor(e, sc)
		default:
			s = newTcpAcceptor(e, sc)
		}
	}
	return s
}

func (e *NetEngine) GetAcceptors() []Acceptor {
	acceptors := make([]Acceptor, 0, len(e.pool))
	for _, v := range e.pool {
		if a, is := v.(Acceptor); is {
			acceptors = append(acceptors, a)
		}
	}

	return acceptors
}

func (e *NetEngine) Connect(sc *SessionConfig) error {
	if e.quit {
		return errors.New("NetEngine already quiting")
	}
	SendStartNetIoService(sc)
	return nil
}

func (e *NetEngine) Listen(sc *SessionConfig) error {
	if e.quit {
		return errors.New("NetEngine already quiting")
	}
	SendStartNetIoService(sc)
	return nil
}

func (e *NetEngine) ShutConnector(ip string, port int) {
	for _, v := range e.pool {
		if c, is := v.(Connector); is {
			sc := c.GetSessionConfig()
			if sc.Ip == ip && sc.Port == port {
				c.shutdown()
				return
			}
		}
	}
}

////////////////////////////////////////////////////////////////////
/// Module Implement [beg]
////////////////////////////////////////////////////////////////////
func (e *NetEngine) ModuleName() string {
	return module.ModuleName_Net
}

func (e *NetEngine) Init() {
	var err error
	for i := 0; i < len(Config.IoServices); i++ {
		s := e.newIoService(&Config.IoServices[i])
		if s != nil {
			e.pool[Config.IoServices[i].Id] = s
			err = s.start()
			if err != nil {
				logger.Logger.Error(err)
			}
		}
	}

	//time.AfterFunc(time.Minute*5, func() { e.dump() })
}

func (e *NetEngine) Update() {
	defer utils.DumpStackIfPanic("NetEngine.Update")

	e.clearClosedIo()

	for _, v := range e.pool {
		v.update()
	}
}

func (e *NetEngine) Shutdown() {
	if e.quit {
		return
	}

	e.quit = true

	if len(e.pool) > 0 {
		for _, v := range e.pool {
			v.shutdown()
		}
		go e.reapRoutine()
	} else {
		e.destroy()
	}
}

////////////////////////////////////////////////////////////////////
/// Module Implement [end]
////////////////////////////////////////////////////////////////////

func (e *NetEngine) clearClosedIo() {
	for {
		select {
		case k := <-e.childAck:
			delete(e.pool, k)
		case sc := <-e.backlogSc:
			s := e.newIoService(sc)
			if s != nil {
				e.pool[sc.Id] = s
				err := s.start()
				if err != nil {
					logger.Logger.Error(err)
				}
			}
		default:
			return
		}
	}
}

func (e *NetEngine) reapRoutine() {
	if e.reaped {
		return
	}

	e.reaped = true

	for {
		select {
		case k := <-e.childAck:
			delete(e.pool, k)
			if len(e.pool) == 0 {
				e.destroy()
				return
			}
		}
	}
}

func (e *NetEngine) destroy() {
	module.UnregisteModule(e)
}

func (e *NetEngine) dump() {
	for _, v := range e.pool {
		v.dump()
	}
	time.AfterFunc(time.Minute*5, func() { e.dump() })
}

func (e *NetEngine) stats() map[int]ServiceStats {
	stats := make(map[int]ServiceStats)
	for k, v := range e.pool {
		s := v.stats()
		stats[k] = s
	}
	return stats
}

func init() {
	module.RegisteModule(NetModule, 0, math.MaxInt32)
}

func Connect(sc *SessionConfig) error {
	return NetModule.Connect(sc)
}

func Listen(sc *SessionConfig) error {
	return NetModule.Listen(sc)
}

func GetAcceptors() []Acceptor {
	return NetModule.GetAcceptors()
}

func ShutConnector(ip string, port int) {
	NetModule.ShutConnector(ip, port)
}

func Stats() map[int]ServiceStats {
	return NetModule.stats()
}
