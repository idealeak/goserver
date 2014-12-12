package netlib

import (
	"errors"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/profile"
	"github.com/idealeak/goserver/core/utils"
)

const (
	IoServiceMaxCount int = 10
)

var (
	NetModule = newNetEngine()
)

type NetEngine struct {
	pool     map[int]ioService
	childAck chan int
	quit     bool
	reaped   bool
}

func newNetEngine() *NetEngine {
	e := &NetEngine{
		pool:     make(map[int]ioService),
		childAck: make(chan int, IoServiceMaxCount),
	}

	return e
}

func (e *NetEngine) newIoService(sc *SessionConfig) ioService {
	var s ioService
	if sc.IsClient {
		if !sc.AllowMultiConn && ConnectorMgr.isConnecting(sc) {
			return nil
		}
		s = newConnector(e, sc)
	} else {
		s = newAcceptor(e, sc)
	}
	return s
}

func (e *NetEngine) GetAcceptors() []*Acceptor {
	acceptors := make([]*Acceptor, 0, len(e.pool))
	for _, v := range e.pool {
		if a, is := v.(*Acceptor); is {
			acceptors = append(acceptors, a)
		}
	}

	return acceptors
}

func (e *NetEngine) Connect(s *basic.Object, sc *SessionConfig) error {
	if e.quit {
		return errors.New("NetEngine already quiting")
	}
	SendStartNetIoService(s, sc)
	return nil
}

func (e *NetEngine) Listen(s *basic.Object, sc *SessionConfig) error {
	if e.quit {
		return errors.New("NetEngine already quiting")
	}
	SendStartNetIoService(s, sc)
	return nil
}

func (e *NetEngine) ShutConnector(ip string, port int) {
	for _, v := range e.pool {
		if c, is := v.(*Connector); is {
			if c.sc.Ip == ip && c.sc.Port == port {
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
				logger.Error(err)
			}
		}
	}
}

func (e *NetEngine) Update() {
	defer utils.DumpStackIfPanic("NetEngine.Update")

	watch := profile.TimeStatisticMgr.WatchStart(e.ModuleName())
	defer watch.Stop()

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

func init() {
	module.RegisteModule(NetModule, 0, 0)
}

func Connect(o *basic.Object, sc *SessionConfig) error {
	return NetModule.Connect(o, sc)
}

func Listen(o *basic.Object, sc *SessionConfig) error {
	return NetModule.Listen(o, sc)
}

func GetAcceptors() []*Acceptor {
	return NetModule.GetAcceptors()
}

func ShutConnector(ip string, port int) {
	NetModule.ShutConnector(ip, port)
}
