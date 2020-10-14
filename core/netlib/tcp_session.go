// session
package netlib

import (
	"fmt"
	"net"
	"runtime"
	"strconv"
	"time"

	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/utils"
)

var (
	SendRoutinePoison *packet = nil
)

type TcpSession struct {
	Session
	conn net.Conn
}

func newTcpSession(id int, conn net.Conn, sc *SessionConfig, scl SessionCloseListener) *TcpSession {
	s := &TcpSession{
		conn: conn,
	}
	s.Session.Id = id
	s.Session.sc = sc
	s.Session.scl = scl
	s.Session.createTime = time.Now()
	s.Session.waitor = utils.NewWaitor("netlib.TcpSession")
	s.Session.impl = s
	s.init()

	return s
}

func (s *TcpSession) init() {
	s.Session.init()
}

func (s *TcpSession) LocalAddr() string {
	return s.conn.LocalAddr().String()
}

func (s *TcpSession) RemoteAddr() string {
	return s.conn.RemoteAddr().String()
}

func (s *TcpSession) start() {
	s.lastRcvTime = time.Now()
	go s.recvRoutine()
	go s.sendRoutine()
}

func (s *TcpSession) sendRoutine() {
	name := fmt.Sprintf("TcpSession.sendRoutine(%v_%v)", s.sc.Name, s.Id)
	s.waitor.Add(name, 1)
	defer func() {
		if err := recover(); err != nil {
			if !s.sc.IsClient && s.sc.IsInnerLink {
				logger.Logger.Warn(s.Id, " ->close: TcpSession.sendRoutine err: ", err)
			} else {
				logger.Logger.Trace(s.Id, " ->close: TcpSession.sendRoutine err: ", err)
			}
		}
		s.sc.encoder.FinishEncode(&s.Session)
		s.shutWrite()
		s.shutRead()
		s.Close()
		s.waitor.Done(name)
	}()

	var (
		err  error
		data []byte
	)

	for !s.quit || len(s.sendBuffer) != 0 {
		if s.PendingSnd {
			runtime.Gosched()
			continue
		}
		select {
		case packet, ok := <-s.sendBuffer:
			if !ok {
				panic("[comm expt]sendBuffer chan closed")
			}

			if packet == nil {
				panic("[comm expt]normal close send")
			}

			if s.sc.IsInnerLink {
				var timeZero time.Time
				s.conn.SetWriteDeadline(timeZero)
			} else {
				if s.sc.WriteTimeout != 0 {
					s.conn.SetWriteDeadline(time.Now().Add(s.sc.WriteTimeout))
				}
			}

			data, err = s.sc.encoder.Encode(&s.Session, packet.packetid, packet.logicno, packet.data, s.conn)
			if err != nil {
				logger.Logger.Trace("s.sc.encoder.Encode err", err)
				if s.sc.IsInnerLink == false {
					FreePacket(packet)
					panic(err)
				}
			}
			FreePacket(packet)
			s.FirePacketSent(packet.packetid, packet.logicno, data)
			s.lastSndTime = time.Now()
		}
	}
}

func (s *TcpSession) recvRoutine() {
	name := fmt.Sprintf("TcpSession.recvRoutine(%v_%v)", s.sc.Name, s.Id)
	s.waitor.Add(name, 1)
	defer func() {
		if err := recover(); err != nil {
			if !s.sc.IsClient && s.sc.IsInnerLink {
				logger.Logger.Warn(s.Id, " ->close: TcpSession.recvRoutine err: ", err)
			} else {
				logger.Logger.Trace(s.Id, " ->close: TcpSession.recvRoutine err: ", err)
			}
		}
		s.sc.decoder.FinishDecode(&s.Session)
		s.shutRead()
		s.Close()
		s.waitor.Done(name)
	}()

	var (
		err      error
		pck      interface{}
		packetid int
		logicNo  uint32
		raw      []byte
	)

	for {
		if s.PendingRcv {
			runtime.Gosched()
			continue
		}
		if s.sc.IsInnerLink {
			var timeZero time.Time
			s.conn.SetReadDeadline(timeZero)
		} else {
			if s.sc.ReadTimeout != 0 {
				s.conn.SetReadDeadline(time.Now().Add(s.sc.ReadTimeout))
			}
		}

		packetid, logicNo, pck, err, raw = s.sc.decoder.Decode(&s.Session, s.conn)
		if err != nil {
			bUnproc := true
			bPackErr := false
			if _, ok := err.(*UnparsePacketTypeErr); ok {
				bPackErr = true
				if s.sc.eph != nil && s.sc.eph.OnErrorPacket(&s.Session, packetid, logicNo, raw) {
					bUnproc = false
				}
			}
			if bUnproc {
				logger.Logger.Tracef("s.sc.decoder.Decode(packetid:%v) err:%v ", packetid, err)
				if s.sc.IsInnerLink == false {
					panic(err)
				} else if !bPackErr {
					panic(err)
				}
			}
		}
		if pck != nil {
			if s.FirePacketReceived(packetid, logicNo, pck) {
				act := AllocAction()
				act.s = &s.Session
				act.p = pck
				act.packid = packetid
				act.logicNo = logicNo
				act.n = "packet:" + strconv.Itoa(packetid)
				s.recvBuffer <- act
			}
		}
		s.lastRcvTime = time.Now()
	}
}

func (s *TcpSession) shutRead() {
	if s.shutRecv {
		return
	}
	logger.Logger.Trace(s.Id, " shutRead")
	s.shutRecv = true
	if tcpconn, ok := s.conn.(*net.TCPConn); ok {
		tcpconn.CloseRead()
	}
}

func (s *TcpSession) shutWrite() {
	if s.shutSend {
		return
	}
	logger.Logger.Trace(s.Id, " shutWrite")
	rest := len(s.sendBuffer)
	for rest > 0 {
		packet := <-s.sendBuffer
		if packet != nil {
			FreePacket(packet)
		}
		rest--
	}

	s.shutSend = true
	if tcpconn, ok := s.conn.(*net.TCPConn); ok {
		tcpconn.CloseWrite()
	}
}

func (s *TcpSession) canShutdown() bool {
	return s.shutRecv && s.shutSend
}
