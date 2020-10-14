// acceptor
package netlib

import (
	"fmt"
	"net"
	"strconv"

	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/utils"
	"sync/atomic"
	"time"
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
	createTime  time.Time
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
		waitor:      utils.NewWaitor("netlib.TcpAcceptor"),
		createTime:  time.Now(),
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
		logger.Logger.Error(err)
		return err
	}
	logger.Logger.Info(a.sc.Name, " listen at ", a.listener.Addr().String())

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
	name := fmt.Sprintf("TcpAcceptor.acceptRoutine(%v_%v)", a.sc.Name, a.sc.Id)
	a.waitor.Add(name, 1)
	defer a.waitor.Done(name)

	for !a.quit {
		conn, err := a.listener.Accept()
		if err != nil {
			logger.Logger.Warn(err)
			continue
		}
		a.connChan <- conn
	}
}

func (a *TcpAcceptor) sessionRoutine() {
	name := fmt.Sprintf("TcpAcceptor.sessionRoutine(%v_%v)", a.sc.Name, a.sc.Id)
	a.waitor.Add(name, 1)
	defer a.waitor.Done(name)

	for !a.quit {
		select {
		case conn, ok := <-a.connChan:
			if !ok { //quiting(chan had closed)
				return
			}
			if tcpconn, ok := conn.(*net.TCPConn); ok {
				tcpconn.SetLinger(a.sc.SoLinger)
				tcpconn.SetNoDelay(a.sc.NoDelay)
				tcpconn.SetReadBuffer(a.sc.RcvBuff)
				tcpconn.SetWriteBuffer(a.sc.SndBuff)
				tcpconn.SetKeepAlive(a.sc.KeepAlive)
				if a.sc.KeepAlive {
					tcpconn.SetKeepAlivePeriod(a.sc.KeepAlivePeriod)
				}
			}
			s := newTcpSession(a.idGen.NextId(), conn, a.sc, a)
			a.acptChan <- s
		}
	}
}

func (a *TcpAcceptor) onClose(s ISession) {
	a.reaper <- s
}

func (a *TcpAcceptor) procReap(s *Session) {
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
	a.waitor.Wait(fmt.Sprintf("TcpAcceptor.reapRoutine_%v", a.sc.Id))

	a.e.childAck <- a.sc.Id
}

func (a *TcpAcceptor) procAccepted(s *TcpSession) {
	a.mapSessions[s.Id] = s
	s.FireConnectEvent()
	s.start()
}

func (a *TcpAcceptor) procActive() {
	var i int
	//var nowork bool
	var doneCnt int
	for _, v := range a.mapSessions {
		//nowork = true
		if v.IsConned() && len(v.recvBuffer) > 0 {
			for i = 0; i < a.sc.MaxDone; i++ {
				select {
				case data, ok := <-v.recvBuffer:
					if !ok {
						goto NEXT
					}
					data.do()
					//nowork = false
					doneCnt++
				default:
					goto NEXT
				}
			}
		}
	NEXT:
		//关闭idle
		//		if nowork && v.IsConned() && v.IsIdle() {
		//			v.FireSessionIdle()
		//		}
	}

	if doneCnt > a.maxDone {
		a.maxDone = doneCnt
	}
	if len(a.mapSessions) > a.maxActive {
		a.maxActive = len(a.mapSessions)
	}
}

func (a *TcpAcceptor) dump() {
	logger.Logger.Info("=========accept dump maxSessions=", a.maxActive, " maxDone=", a.maxDone)
	for sid, s := range a.mapSessions {
		logger.Logger.Info("=========session:", sid, " recvBuffer size=", len(s.recvBuffer), " sendBuffer size=", len(s.sendBuffer))
	}
}

func (a *TcpAcceptor) stats() ServiceStats {
	tNow := time.Now()
	stats := ServiceStats{
		Id:          a.sc.Id,
		Type:        a.sc.Type,
		Name:        a.sc.Name,
		Addr:        a.listener.Addr().String(),
		MaxActive:   a.maxActive,
		MaxDone:     a.maxDone,
		RunningTime: int64(tNow.Sub(a.createTime) / time.Second),
	}

	stats.SessionStats = make([]SessionStats, 0, len(a.mapSessions))
	for _, s := range a.mapSessions {
		ss := SessionStats{
			Id:           s.Id,
			GroupId:      s.GroupId,
			SendedBytes:  atomic.LoadInt64(&s.sendedBytes),
			RecvedBytes:  atomic.LoadInt64(&s.recvedBytes),
			SendedPack:   atomic.LoadInt64(&s.sendedPack),
			RecvedPack:   atomic.LoadInt64(&s.recvedPack),
			PendSendPack: len(s.sendBuffer),
			PendRecvPack: len(s.recvBuffer),
			RemoteAddr:   s.RemoteAddr(),
			RunningTime:  int64(tNow.Sub(s.createTime) / time.Second),
		}
		stats.SessionStats = append(stats.SessionStats, ss)
	}
	return stats
}

func (a *TcpAcceptor) procChanEvent() {
	for i := 0; i < a.sc.MaxDone; i++ {
		select {
		case s := <-a.acptChan:
			a.procAccepted(s)
		case s := <-a.reaper:
			if tcps, ok := s.(*Session); ok {
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
