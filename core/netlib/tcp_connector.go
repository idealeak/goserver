// connector
package netlib

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/utils"
	"sync/atomic"
)

type TcpConnector struct {
	sc         *SessionConfig
	e          *NetEngine
	s          *TcpSession
	idGen      utils.IdGen
	connChan   chan net.Conn
	reaper     chan ISession
	waitor     *utils.Waitor
	createTime time.Time
	quit       bool
	reaped     bool
	maxActive  int
	maxDone    int
}

func newTcpConnector(e *NetEngine, sc *SessionConfig) *TcpConnector {
	c := &TcpConnector{
		sc:         sc,
		e:          e,
		s:          nil,
		connChan:   make(chan net.Conn, 2),
		reaper:     make(chan ISession, 1),
		waitor:     utils.NewWaitor("netlib.TcpConnector"),
		createTime: time.Now(),
	}

	ConnectorMgr.registeConnector(c)
	return c
}

func (c *TcpConnector) connectRoutine() {
	name := fmt.Sprintf("TcpConnector.connectRoutine(%v_%v)", c.sc.Name, c.sc.Id)
	c.waitor.Add(name, 1)
	defer c.waitor.Done(name)

	service := c.sc.Ip + ":" + strconv.Itoa(int(c.sc.Port))
	conn, err := net.Dial("tcp", service)
	if err == nil {
		c.connChan <- conn
		return
	}
	for {
		select {
		case <-time.After(ReconnectInterval):
			if c.quit {
				return
			}
			conn, err := net.Dial("tcp", service)
			if err == nil {
				if c.quit {
					conn.Close()
					return
				}
				c.connChan <- conn
				return
			}
		}
	}
}

func (c *TcpConnector) start() error {

	go c.connectRoutine()
	return nil
}

func (c *TcpConnector) update() {
	c.procActive()
	c.procChanEvent()
}

func (c *TcpConnector) shutdown() {

	if c.quit {
		return
	}
	c.quit = true

	if c.s != nil {
		c.s.Close()
	} else {
		go c.reapRoutine()
	}
}

func (c *TcpConnector) procActive() {
	var i int
	var doneCnt int
	if c.s != nil && c.s.canShutdown() {
		return
	} else if c.s != nil && c.s.IsConned() {
		if len(c.s.recvBuffer) > 0 {
			for i = 0; i < c.sc.MaxDone; i++ {
				select {
				case data, ok := <-c.s.recvBuffer:
					if !ok {
						goto NEXT
					}
					data.do()
					doneCnt++
				default:
					goto NEXT
				}
			}
		}
	}
NEXT:
	if doneCnt > c.maxDone {
		c.maxDone = doneCnt
	}
}

func (c *TcpConnector) dump() {
	logger.Logger.Info("=========connector dump maxDone=", c.maxDone)
	logger.Logger.Info("=========session recvBuffer size=", len(c.s.recvBuffer), " sendBuffer size=", len(c.s.sendBuffer))
}

func (c *TcpConnector) procChanEvent() {
	for {
		select {
		case conn := <-c.connChan:
			c.procConnected(conn)
		case s := <-c.reaper:
			if tcps, ok := s.(*Session); ok {
				c.procReap(tcps)
			}

		default:
			return
		}
	}
}

func (c *TcpConnector) onClose(s ISession) {
	c.reaper <- s
}

func (c *TcpConnector) procConnected(conn net.Conn) {
	if tcpconn, ok := conn.(*net.TCPConn); ok {
		tcpconn.SetLinger(c.sc.SoLinger)
		tcpconn.SetNoDelay(c.sc.NoDelay)
		tcpconn.SetReadBuffer(c.sc.RcvBuff)
		tcpconn.SetWriteBuffer(c.sc.SndBuff)
		tcpconn.SetKeepAlive(c.sc.KeepAlive)
		if c.sc.KeepAlive {
			tcpconn.SetKeepAlivePeriod(c.sc.KeepAlivePeriod)
			//			err := tcpkeepalive.SetKeepAlive(conn, c.sc.KeepAliveIdle, c.sc.KeepAliveCount, c.sc.KeepAlivePeriod)
			//			if err != nil {
			//				logger.Logger.Warnf("(a *TcpConnector) procConnected SetKeepAlive err:%v", err)
			//			}
		}
	}

	c.s = newTcpSession(c.idGen.NextId(), conn, c.sc, c)
	c.s.FireConnectEvent()
	c.s.start()
}

func (c *TcpConnector) procReap(s *Session) {
	for len(s.recvBuffer) > 0 {
		data, ok := <-s.recvBuffer
		if !ok {
			break
		}
		data.do()
	}

	s.destroy()

	if (c.sc.IsAutoReconn == false && c.s.Id == s.Id) || c.quit {
		c.s = nil
		go c.reapRoutine()
	} else if c.sc.IsAutoReconn && c.s.Id == s.Id {
		c.s = nil
		if !c.quit {
			go c.connectRoutine()
		}
	}
}

func (c *TcpConnector) reapRoutine() {
	if c.reaped {
		return
	}

	c.reaped = true

	c.waitor.Wait(fmt.Sprintf("TcpConnector.reapRoutine_%v", c.sc.Id))
	select {
	case conn := <-c.connChan:
		conn.Close()
	default:
	}
	c.e.childAck <- c.sc.Id
	ConnectorMgr.unregisteConnector(c)
}

func (c *TcpConnector) GetSessionConfig() *SessionConfig {
	return c.sc
}

func (c *TcpConnector) stats() ServiceStats {
	tNow := time.Now()
	stats := ServiceStats{
		Id:          c.sc.Id,
		Type:        c.sc.Type,
		Name:        c.sc.Name,
		MaxActive:   1,
		MaxDone:     c.maxDone,
		RunningTime: int64(tNow.Sub(c.createTime) / time.Second),
	}

	if c.s != nil {
		stats.Addr = c.s.LocalAddr()
		stats.SessionStats = []SessionStats{
			{
				Id:           c.s.Id,
				GroupId:      c.s.GroupId,
				SendedBytes:  atomic.LoadInt64(&c.s.sendedBytes),
				RecvedBytes:  atomic.LoadInt64(&c.s.recvedBytes),
				SendedPack:   atomic.LoadInt64(&c.s.sendedPack),
				RecvedPack:   atomic.LoadInt64(&c.s.recvedPack),
				PendSendPack: len(c.s.sendBuffer),
				PendRecvPack: len(c.s.recvBuffer),
				RemoteAddr:   c.s.RemoteAddr(),
				RunningTime:  int64(tNow.Sub(c.s.createTime) / time.Second),
			},
		}

	}
	return stats
}
