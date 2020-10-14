// acceptor
package netlib

import (
	"fmt"
	"github.com/xtaci/kcp-go"
	"net"
	"strconv"

	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/utils"
	"sync/atomic"
	"time"
)

type UdpAcceptor struct {
	e           *NetEngine
	sc          *SessionConfig
	listener    *kcp.Listener
	idGen       utils.IdGen
	mapSessions map[int]*UdpSession
	acptChan    chan *UdpSession
	connChan    chan *kcp.UDPSession
	reaper      chan ISession
	waitor      *utils.Waitor
	createTime  time.Time
	quit        bool
	reaped      bool
	restart     bool
	maxActive   int
	maxDone     int
}

func newUdpAcceptor(e *NetEngine, sc *SessionConfig) *UdpAcceptor {
	a := &UdpAcceptor{
		e:  e,
		sc: sc,
	}

	a.init()

	return a
}

func (a *UdpAcceptor) init() {
	backlog := int(a.sc.MaxConn + a.sc.ExtraConn)
	a.connChan = make(chan *kcp.UDPSession, backlog)
	a.acptChan = make(chan *UdpSession, backlog)
	a.reaper = make(chan ISession, backlog)
	a.mapSessions = make(map[int]*UdpSession)
	a.waitor = utils.NewWaitor("netlib.UdpAcceptor")
	a.createTime = time.Now()
	a.quit = false
}

func (a *UdpAcceptor) start() (err error) {
	service := a.sc.Ip + ":" + strconv.Itoa(int(a.sc.Port))
	a.listener, err = kcp.ListenWithOptions(service, nil, 0, 0)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	logger.Logger.Info(a.sc.Name, " listen at ", a.listener.Addr().String())

	go a.acceptRoutine()
	go a.sessionRoutine()

	return nil
}

func (a *UdpAcceptor) update() {

	a.procActive()

	a.procChanEvent()
}

func (a *UdpAcceptor) shutdown() {

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

func (a *UdpAcceptor) acceptRoutine() {
	name := fmt.Sprintf("UdpAcceptor.acceptRoutine(%v_%v)", a.sc.Name, a.sc.Id)
	a.waitor.Add(name, 1)
	defer a.waitor.Done(name)

	for !a.quit {
		conn, err := a.listener.AcceptKCP()
		if err != nil {
			logger.Logger.Warn(err)
			if err.Error() == "timeout" {
				continue
			}
			break
		}
		a.connChan <- conn
	}

	//异常退出，需要重新启动
	if !a.quit {
		a.shutdown()
		a.restart = true
	}
}

func (a *UdpAcceptor) sessionRoutine() {
	name := fmt.Sprintf("UdpAcceptor.sessionRoutine(%v_%v)", a.sc.Name, a.sc.Id)
	a.waitor.Add(name, 1)
	defer a.waitor.Done(name)

	for !a.quit {
		select {
		case conn, ok := <-a.connChan:
			if !ok { //quiting(chan had closed)
				return
			}
			s := newUdpSession(a.idGen.NextId(), conn, a.sc, a)
			if s != nil {
				s.conn.SetWindowSize(a.sc.MaxPend, a.sc.MaxPend)
				if a.sc.NoDelay {
					s.conn.SetNoDelay(1, 10, 2, 1)
				} else {
					s.conn.SetNoDelay(0, 40, 0, 0)
				}
				if a.sc.MTU > 128 && a.sc.MTU <= 1500 { //粗略的估算ip(最长60)+udp(8)+kcp(24)+proto(12)
					s.conn.SetMtu(a.sc.MTU)
				}
				a.acptChan <- s
			}
		}
	}
}

func (a *UdpAcceptor) onClose(s ISession) {
	a.reaper <- s
}

func (a *UdpAcceptor) procReap(s *Session) {
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

func (a *UdpAcceptor) reapRoutine() {
	if a.reaped {
		return
	}
	a.reaped = true
	a.waitor.Wait(fmt.Sprintf("UdpAcceptor.reapRoutine_%v", a.sc.Id))

	a.e.childAck <- a.sc.Id
	if a.restart { //延迟1s后,重新启动
		time.Sleep(time.Second)
		a.e.backlogSc <- a.sc
	}
}

func (a *UdpAcceptor) procAccepted(s *UdpSession) {
	a.mapSessions[s.Id] = s
	s.FireConnectEvent()
	s.start()
}

func (a *UdpAcceptor) procActive() {
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

func (a *UdpAcceptor) dump() {
	logger.Logger.Info("=========accept dump maxSessions=", a.maxActive, " maxDone=", a.maxDone)
	for sid, s := range a.mapSessions {
		logger.Logger.Info("=========session:", sid, " recvBuffer size=", len(s.recvBuffer), " sendBuffer size=", len(s.sendBuffer))
	}
}

func (a *UdpAcceptor) stats() ServiceStats {
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

func (a *UdpAcceptor) procChanEvent() {
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

func (a *UdpAcceptor) GetSessionConfig() *SessionConfig {
	return a.sc
}

func (a *UdpAcceptor) Addr() net.Addr {
	if a.listener != nil {
		return a.listener.Addr()
	}
	return nil
}
