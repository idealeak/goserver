// session
package netlib

import (
	"bytes"
	"fmt"
	"github.com/xtaci/kcp-go"
	"runtime"
	"strconv"
	"time"

	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/utils"
)

type UdpSession struct {
	Session
	conn *kcp.UDPSession
}

func newUdpSession(id int, conn *kcp.UDPSession, sc *SessionConfig, scl SessionCloseListener) *UdpSession {
	s := &UdpSession{
		conn: conn,
	}
	s.Session.Id = id
	s.Session.sc = sc
	s.Session.scl = scl
	s.Session.createTime = time.Now()
	s.Session.waitor = utils.NewWaitor("netlib.UdpSession")
	s.Session.impl = s
	s.init()

	return s
}

func (s *UdpSession) init() {
	s.Session.init()
}

func (s *UdpSession) LocalAddr() string {
	return s.conn.LocalAddr().String()
}

func (s *UdpSession) RemoteAddr() string {
	return s.conn.RemoteAddr().String()
}

func (s *UdpSession) start() {
	s.lastRcvTime = time.Now()
	go s.recvRoutine()
	go s.sendRoutine()
}

func (s *UdpSession) sendRoutine() {
	name := fmt.Sprintf("UdpSession.sendRoutine(%v_%v)", s.sc.Name, s.Id)
	s.waitor.Add(name, 1)
	defer func() {
		if err := recover(); err != nil {
			if !s.sc.IsClient && s.sc.IsInnerLink {
				logger.Logger.Warn(s.Id, " ->close: UdpSession.sendRoutine err: ", err)
			} else {
				logger.Logger.Trace(s.Id, " ->close: UdpSession.sendRoutine err: ", err)
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

func (s *UdpSession) recvRoutine() {
	name := fmt.Sprintf("UdpSession.recvRoutine(%v_%v)", s.sc.Name, s.Id)
	s.waitor.Add(name, 1)
	defer func() {
		if err := recover(); err != nil {
			if !s.sc.IsClient && s.sc.IsInnerLink {
				logger.Logger.Warn(s.Id, " ->close: UdpSession.recvRoutine err: ", err)
			} else {
				logger.Logger.Trace(s.Id, " ->close: UdpSession.recvRoutine err: ", err)
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
		n        int
	)

	buf := make([]byte, s.sc.MaxPacket)
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

		n, err = s.conn.Read(buf)
		if err != nil {
			panic(err)
		}

		packetid, logicNo, pck, err, raw = s.sc.decoder.Decode(&s.Session, bytes.NewBuffer(buf[:n]))
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

func (s *UdpSession) shutRead() {
	if s.shutRecv {
		return
	}
	logger.Logger.Trace(s.Id, " shutRead")
	s.shutRecv = true
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
}

func (s *UdpSession) shutWrite() {
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
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
}

func (s *UdpSession) canShutdown() bool {
	return s.shutRecv && s.shutSend
}
