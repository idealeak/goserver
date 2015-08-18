// session
package netlib

import (
	"net"
	"strconv"
	"time"

	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/utils"
)

var (
	SendRoutinePoison interface{} = nil
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
	s.Session.waitor = utils.NewWaitor()
	s.init()

	return s
}

func (s *TcpSession) init() {
	s.Session.init()
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

	defer func() {
		if err := recover(); err != nil {
			logger.Trace(s.Id, " ->close: Session.procSend err: ", err)
		}
		s.sc.encoder.FinishEncode(&s.Session)
		s.shutWrite()
		s.shutRead()
		s.Close()
	}()

	s.waitor.Add(1)
	defer s.waitor.Done()

	var (
		err  error
		data []byte
	)

	for !s.quit || len(s.sendBuffer) != 0 {
		select {
		case msg, ok := <-s.sendBuffer:
			if !ok {
				panic("[comm expt]sendBuffer chan closed")
			}

			if msg == nil {
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

			data, err = s.sc.encoder.Encode(&s.Session, msg, s.conn)
			if err != nil {
				logger.Trace("s.sc.encoder.Encode err", err)
				if s.sc.IsInnerLink == false {
					panic(err)
				}
			}
			s.FirePacketSent(data)
			s.lastSndTime = time.Now()
		}
	}
}

func (s *TcpSession) recvRoutine() {

	defer func() {
		if err := recover(); err != nil {
			logger.Trace(s.Id, " ->close: Session.procRecv err: ", err)
		}
		s.sc.decoder.FinishDecode(&s.Session)
		s.shutRead()
		s.Close()
	}()

	s.waitor.Add(1)
	defer s.waitor.Done()

	var (
		err      error
		pck      interface{}
		packetid int
		raw      []byte
	)

	for /*!s.quit */ true {
		if s.sc.IsInnerLink {
			var timeZero time.Time
			s.conn.SetReadDeadline(timeZero)
		} else {
			if s.sc.ReadTimeout != 0 {
				s.conn.SetReadDeadline(time.Now().Add(s.sc.ReadTimeout))
			}
		}

		packetid, pck, err, raw = s.sc.decoder.Decode(&s.Session, s.conn)
		if err != nil {
			bUnproc := true
			bPackErr := false
			if _, ok := err.(*UnparsePacketTypeErr); ok {
				bPackErr = true
				if s.sc.eph != nil && s.sc.eph.OnErrorPacket(&s.Session, packetid, raw) {
					bUnproc = false
				}
			}
			if bUnproc {
				logger.Trace("s.sc.decoder.Decode err ", err)
				if s.sc.IsInnerLink == false {
					panic(err)
				} else if !bPackErr {
					panic(err)
				}
			}
		}
		if pck != nil {
			if s.FirePacketReceived(packetid, pck) {
				act := AllocAction()
				act.s = &s.Session
				act.p = pck
				act.packid = packetid
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
	logger.Trace(s.Id, " shutRead")
	s.shutRecv = true
	if tcpconn, ok := s.conn.(*net.TCPConn); ok {
		tcpconn.CloseRead()
	}
}

func (s *TcpSession) shutWrite() {
	if s.shutSend {
		return
	}
	logger.Trace(s.Id, " shutWrite")
	rest := len(s.sendBuffer)
	for rest > 0 {
		<-s.sendBuffer
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
