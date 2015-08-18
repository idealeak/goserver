// acceptor
package netlib

import (
	"net"
	"strconv"

	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/utils"
)

const (
	DefaultMaxConnect = 1024
)

type TcpAcceptor struct {
	e           *NetEngine
	sc          *SessionConfig
	listener    net.Listener
	idGen       utils.IdGen
	mapSessions map[int]*TcpSession
	acptChan    chan *TcpSession
	connChan    chan net.Conn
	reaper      chan ISession
	waitor      *utils.Waitor
	quit        bool
	reaped      bool
	maxActive   int
	maxDone     int
}

func newTcpAcceptor(e *NetEngine, sc *SessionConfig) *TcpAcceptor {
	a := &TcpAcceptor{
		e:           e,
		sc:          sc,
		quit:        false,
		mapSessions: make(map[int]*TcpSession),
		waitor:      utils.NewWaitor(),
	}

	a.init()

	return a
}

func (a *TcpAcceptor) init() {

	temp := int(a.sc.MaxConn + a.sc.ExtraConn)
	a.connChan = make(chan net.Conn, temp)
	a.acptChan = make(chan *TcpSession, temp)
	a.reaper = make(chan ISession, temp)
}

func (a *TcpAcceptor) start() (err error) {
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

func (a *TcpAcceptor) update() {

	a.procActive()

	a.procChanEvent()
}

func (a *TcpAcceptor) shutdown() {

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

func (a *TcpAcceptor) acceptRoutine() {

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

func (a *TcpAcceptor) sessionRoutine() {

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
			s := newTcpSession(a.idGen.NextId(), conn, a.sc, a)
			a.acptChan <- s
		}
	}
}

func (a *TcpAcceptor) onClose(s ISession) {
	a.reaper <- s
}

func (a *TcpAcceptor) procReap(s *TcpSession) {
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

func (a *TcpAcceptor) reapRoutine() {
	if a.reaped {
		return
	}
	a.reaped = true
	a.waitor.Wait()

	a.e.childAck <- a.sc.Id
}

func (a *TcpAcceptor) procAccepted(s *TcpSession) {
	a.mapSessions[s.Id] = s
	s.FireConnectEvent()
	s.start()
}

func (a *TcpAcceptor) procActive() {
	var i int
	var nowork bool
	var doneCnt int
	for _, v := range a.mapSessions {
		nowork = true
		for i = 0; i < a.sc.MaxDone; i++ {
			if v.IsConned() && len(v.recvBuffer) > 0 {
				data, ok := <-v.recvBuffer
				if !ok {
					break
				}
				data.do()
				nowork = false
				doneCnt++
			} else {
				break
			}
		}
		if nowork && v.IsConned() && v.IsIdle() {
			v.FireSessionIdle()
		}
	}

	if doneCnt > a.maxDone {
		a.maxDone = doneCnt
	}
	if len(a.mapSessions) > a.maxActive {
		a.maxActive = len(a.mapSessions)
	}
}

func (a *TcpAcceptor) dump() {
	logger.Info("=========accept dump maxSessions=", a.maxActive, " maxDone=", a.maxDone)
	for sid, s := range a.mapSessions {
		logger.Info("=========session:", sid, " recvBuffer size=", len(s.recvBuffer), " sendBuffer size=", len(s.sendBuffer))
	}
}

func (a *TcpAcceptor) procChanEvent() {
	for i := 0; i < a.sc.MaxDone; i++ {
		select {
		case s := <-a.acptChan:
			a.procAccepted(s)
		case s := <-a.reaper:
			if tcps, ok := s.(*TcpSession); ok {
				a.procReap(tcps)
			}
		default:
			return
		}
	}
}

func (a *TcpAcceptor) GetSessionConfig() *SessionConfig {
	return a.sc
}

func (a *TcpAcceptor) Addr() net.Addr {
	if a.listener != nil {
		return a.listener.Addr()
	}
	return nil
}
