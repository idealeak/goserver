// acceptor
package netlib

import (
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/utils"
	"net"
	"strconv"
)

const (
	DefaultMaxConnect = 1024
)

type Acceptor struct {
	e           *NetEngine
	sc          *SessionConfig
	listener    net.Listener
	idGen       utils.IdGen
	mapSessions map[int]*Session
	acptChan    chan *Session
	connChan    chan net.Conn
	reaper      chan *Session
	waitor      *utils.Waitor
	quit        bool
	reaped      bool
}

func newAcceptor(e *NetEngine, sc *SessionConfig) *Acceptor {
	a := &Acceptor{
		e:           e,
		sc:          sc,
		quit:        false,
		mapSessions: make(map[int]*Session),
		waitor:      utils.NewWaitor(),
	}

	a.init()

	return a
}

func (a *Acceptor) init() {

	temp := int(a.sc.MaxConn + a.sc.ExtraConn)
	a.connChan = make(chan net.Conn, temp)
	a.acptChan = make(chan *Session, temp)
	a.reaper = make(chan *Session, temp)
}

func (a *Acceptor) start() (err error) {
	service := a.sc.Ip + ":" + strconv.Itoa(int(a.sc.Port))
	a.listener, err = net.Listen("tcp", service)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Info(a.sc.Name, " listen at ", a.listener.Addr().String())

	go a.acceptRoutine()
	go a.sessionRoutine()

	return nil
}

func (a *Acceptor) update() {

	a.procActive()

	a.procChanEvent()
}

func (a *Acceptor) shutdown() {

	if a.quit {
		return
	}

	a.quit = true

	if a.listener != nil {
		a.listener.Close()
		a.listener = nil
	}

	if a.connChan != nil {
		close(a.connChan)
		a.connChan = nil
	}

	if len(a.mapSessions) == 0 {
		go a.reapRoutine()
	} else {
		for _, v := range a.mapSessions {
			v.Close()
		}
	}
}

func (a *Acceptor) acceptRoutine() {

	a.waitor.Add(1)
	defer a.waitor.Done()

	for !a.quit {
		conn, err := a.listener.Accept()
		if err != nil {
			logger.Warn(err)
			continue
		}
		a.connChan <- conn
	}
}

func (a *Acceptor) sessionRoutine() {

	a.waitor.Add(1)
	defer a.waitor.Done()

	for !a.quit {
		select {
		case conn, ok := <-a.connChan:
			if !ok { //quiting(chan had closed)
				return
			}

			if tcpconn, ok := conn.(*net.TCPConn); ok {
				tcpconn.SetLinger(a.sc.SoLinger)
				tcpconn.SetNoDelay(a.sc.NoDelay)
				tcpconn.SetKeepAlive(a.sc.KeepAlive)
				tcpconn.SetReadBuffer(a.sc.RcvBuff)
				tcpconn.SetWriteBuffer(a.sc.SndBuff)
			}
			s := newSession(a.idGen.NextId(), conn, a.sc, a)
			a.acptChan <- s
		}
	}
}

func (a *Acceptor) onClose(s *Session) {
	a.reaper <- s
}

func (a *Acceptor) procReap(s *Session) {
	if _, exist := a.mapSessions[s.Id]; exist {
		delete(a.mapSessions, s.Id)
		s.destroy()
	}

	if a.quit {
		if len(a.mapSessions) == 0 {
			go a.reapRoutine()
		}
	}
}

func (a *Acceptor) reapRoutine() {
	if a.reaped {
		return
	}
	a.reaped = true
	a.waitor.Wait()

	a.e.childAck <- a.sc.Id
}

func (a *Acceptor) procAccepted(s *Session) {
	a.mapSessions[s.Id] = s
	s.FireConnectEvent()
	s.start()
}

func (a *Acceptor) procActive() {
	var i int
	var nowork bool
	for _, v := range a.mapSessions {
		nowork = true
		for i = 0; i < a.sc.MaxDone; i++ {
			if v.IsConned && len(v.recvBuffer) > 0 {
				data, ok := <-v.recvBuffer
				if !ok {
					break
				}
				data.do()
				nowork = false
			} else {
				break
			}
		}
		if nowork && v.IsConned && v.IsIdle() {
			v.FireSessionIdle()
		}
	}
}

func (a *Acceptor) procChanEvent() {
	for i := 0; i < a.sc.MaxDone; i++ {
		select {
		case s := <-a.acptChan:
			a.procAccepted(s)
		case s := <-a.reaper:
			a.procReap(s)
		default:
			return
		}
	}
}

func (a *Acceptor) GetSessionConfig() *SessionConfig {
	return a.sc
}

func (a *Acceptor) Addr() net.Addr {
	if a.listener != nil {
		return a.listener.Addr()
	}
	return nil
}
