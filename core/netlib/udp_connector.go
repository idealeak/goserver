// connector
package netlib

import (
	"fmt"
	"github.com/xtaci/kcp-go"
	"strconv"
	"time"

	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/utils"
	"sync/atomic"
)

type UdpConnector struct {
	sc         *SessionConfig
	e          *NetEngine
	s          *UdpSession
	idGen      utils.IdGen
	connChan   chan *kcp.UDPSession
	reaper     chan ISession
	waitor     *utils.Waitor
	createTime time.Time
	quit       bool
	reaped     bool
	maxActive  int
	maxDone    int
}

func newUdpConnector(e *NetEngine, sc *SessionConfig) *UdpConnector {
	c := &UdpConnector{
		sc:         sc,
		e:          e,
		s:          nil,
		connChan:   make(chan *kcp.UDPSession, 2),
		reaper:     make(chan ISession, 1),
		waitor:     utils.NewWaitor("netlib.UdpConnector"),
		createTime: time.Now(),
	}

	ConnectorMgr.registeConnector(c)
	return c
}

func (c *UdpConnector) connectRoutine() {
	name := fmt.Sprintf("UdpConnector.connectRoutine(%v_%v)", c.sc.Name, c.sc.Id)
	c.waitor.Add(name, 1)
	defer c.waitor.Done(name)

	service := c.sc.Ip + ":" + strconv.Itoa(int(c.sc.Port))
	conn, err := kcp.DialWithOptions(service, nil, 0, 0)
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
			conn, err := kcp.DialWithOptions(service, nil, 0, 0)
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

func (c *UdpConnector) start() error {

	go c.connectRoutine()
	return nil
}

func (c *UdpConnector) update() {
	c.procActive()
	c.procChanEvent()
}

func (c *UdpConnector) shutdown() {

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

func (c *UdpConnector) procActive() {
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

func (c *UdpConnector) dump() {
	logger.Logger.Info("=========connector dump maxDone=", c.maxDone)
	logger.Logger.Info("=========session recvBuffer size=", len(c.s.recvBuffer), " sendBuffer size=", len(c.s.sendBuffer))
}

func (c *UdpConnector) procChanEvent() {
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

func (c *UdpConnector) onClose(s ISession) {
	c.reaper <- s
}

func (c *UdpConnector) procConnected(conn *kcp.UDPSession) {
	c.s = newUdpSession(c.idGen.NextId(), conn, c.sc, c)
	if c.s != nil {
		c.s.conn.SetWindowSize(c.sc.MaxPend, c.sc.MaxPend)
		if c.sc.NoDelay {
			c.s.conn.SetNoDelay(1, 10, 2, 1)
		} else {
			c.s.conn.SetNoDelay(0, 40, 0, 0)
		}
		if c.sc.MTU > 128 && c.sc.MTU <= 1500 { //粗略的估算ip(最长60)+udp(8)+kcp(24)+proto(12)
			c.s.conn.SetMtu(c.sc.MTU)
		}
		c.s.FireConnectEvent()
		c.s.start()
	}
}

func (c *UdpConnector) procReap(s *Session) {
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

func (c *UdpConnector) reapRoutine() {
	if c.reaped {
		return
	}

	c.reaped = true

	c.waitor.Wait(fmt.Sprintf("UdpConnector.reapRoutine_%v", c.sc.Id))
	select {
	case conn := <-c.connChan:
		conn.Close()
	default:
	}
	c.e.childAck <- c.sc.Id
	ConnectorMgr.unregisteConnector(c)
}

func (c *UdpConnector) GetSessionConfig() *SessionConfig {
	return c.sc
}

func (c *UdpConnector) stats() ServiceStats {
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
