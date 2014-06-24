// connector
package netlib

import (
	"net"
	"strconv"
	"time"

	"github.com/idealeak/goserver/core/utils"
)

const (
	ReconnectInterval time.Duration = 5 * time.Second
)

type Connector struct {
	sc       *SessionConfig
	e        *NetEngine
	s        *Session
	idGen    utils.IdGen
	connChan chan net.Conn
	reaper   chan *Session
	waitor   *utils.Waitor
	quit     bool
	reaped   bool
}

func newConnector(e *NetEngine, sc *SessionConfig) *Connector {
	c := &Connector{
		sc:       sc,
		e:        e,
		s:        nil,
		connChan: make(chan net.Conn, 2),
		reaper:   make(chan *Session, 1),
		waitor:   utils.NewWaitor(),
	}

	ConnectorMgr.registeConnector(c)
	return c
}

func (c *Connector) connectRoutine() {

	c.waitor.Add(1)
	defer c.waitor.Done()

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

func (c *Connector) start() error {

	go c.connectRoutine()
	return nil
}

func (c *Connector) update() {
	c.procActive()
	c.procChanEvent()
}

func (c *Connector) shutdown() {

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

func (c *Connector) procActive() {
	var i int
	if c.s != nil && c.s.canShutdown() {
		return
	} else if c.s != nil && c.s.IsConned {
		for i = 0; i < c.sc.MaxDone; i++ {
			if len(c.s.recvBuffer) > 0 {
				data, ok := <-c.s.recvBuffer
				if !ok {
					break
				}
				data.do()
			}
		}
	}
}

func (c *Connector) procChanEvent() {
	for {
		select {
		case conn := <-c.connChan:
			c.procConnected(conn)
		case s := <-c.reaper:
			c.procReap(s)
		default:
			return
		}
	}
}

func (c *Connector) onClose(s *Session) {
	c.reaper <- s
}

func (c *Connector) procConnected(conn net.Conn) {

	if tcpconn, ok := conn.(*net.TCPConn); ok {
		tcpconn.SetLinger(c.sc.SoLinger)
		tcpconn.SetNoDelay(c.sc.NoDelay)
		tcpconn.SetKeepAlive(c.sc.KeepAlive)
		tcpconn.SetReadBuffer(c.sc.RcvBuff)
		tcpconn.SetWriteBuffer(c.sc.SndBuff)
	}

	c.s = newSession(c.idGen.NextId(), conn, c.sc, c)
	c.s.FireConnectEvent(false)
	c.s.start()
}

func (c *Connector) procReap(s *Session) {
	for len(s.recvBuffer) > 0 {
		data, ok := <-s.recvBuffer
		if !ok {
			break
		}
		data.do()
	}

	s.destroy()

	if (c.sc.IsAutoReconn == false && c.s == s) || c.quit {
		c.s = nil
		go c.reapRoutine()
	} else if c.sc.IsAutoReconn && c.s == s {
		c.s = nil
		go c.connectRoutine()
	}
}

func (c *Connector) reapRoutine() {
	if c.reaped {
		return
	}

	c.reaped = true

	c.waitor.Wait()
	select {
	case conn := <-c.connChan:
		conn.Close()
	default:
	}
	c.e.childAck <- c.sc.Id
	ConnectorMgr.unregisteConnector(c)
}
