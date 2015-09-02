// connector
package netlib

import (
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/utils"
)

type WsConnector struct {
	sc        *SessionConfig
	e         *NetEngine
	s         *WsSession
	idGen     utils.IdGen
	connChan  chan *websocket.Conn
	reaper    chan ISession
	waitor    *utils.Waitor
	dialer    websocket.Dialer
	quit      bool
	reaped    bool
	maxActive int
	maxDone   int
}

func newWsConnector(e *NetEngine, sc *SessionConfig) *WsConnector {
	c := &WsConnector{
		sc:       sc,
		e:        e,
		s:        nil,
		connChan: make(chan *websocket.Conn, 2),
		reaper:   make(chan ISession, 1),
		waitor:   utils.NewWaitor(),
		dialer: websocket.Dialer{
			ReadBufferSize:  sc.RcvBuff,
			WriteBufferSize: sc.SndBuff,
		},
	}

	ConnectorMgr.registeConnector(c)
	return c
}

func (c *WsConnector) connectRoutine() {

	c.waitor.Add(1)
	defer c.waitor.Done()

	service := "ws://" + c.sc.Ip + ":" + strconv.Itoa(int(c.sc.Port)) + c.sc.Path
	conn, _, err := c.dialer.Dial(service, nil)
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
			conn, _, err := c.dialer.Dial(service, nil)
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

func (c *WsConnector) start() error {
	go c.connectRoutine()
	return nil
}

func (c *WsConnector) update() {
	c.procActive()
	c.procChanEvent()
}

func (c *WsConnector) shutdown() {
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

func (c *WsConnector) procActive() {
	var i int
	var doneCnt int
	if c.s != nil && c.s.IsConned() {
		for i = 0; i < c.sc.MaxDone; i++ {
			if len(c.s.recvBuffer) > 0 {
				data, ok := <-c.s.recvBuffer
				if !ok {
					break
				}
				data.do()
				doneCnt++
			}
		}
	}

	if doneCnt > c.maxDone {
		c.maxDone = doneCnt
	}
}

func (c *WsConnector) dump() {
	logger.Info("=========wsconnector dump maxDone=", c.maxDone)
	logger.Info("=========wssession recvBuffer size=", len(c.s.recvBuffer), " sendBuffer size=", len(c.s.sendBuffer))
}

func (c *WsConnector) procChanEvent() {
	for {
		select {
		case conn := <-c.connChan:
			c.procConnected(conn)
		case s := <-c.reaper:
			if wss, ok := s.(*Session); ok {
				c.procReap(wss)
			}

		default:
			return
		}
	}
}

func (c *WsConnector) onClose(s ISession) {
	c.reaper <- s
}

func (c *WsConnector) procConnected(conn *websocket.Conn) {
	c.s = newWsSession(c.idGen.NextId(), conn, c.sc, c)
	c.s.FireConnectEvent()
	c.s.start()
}

func (c *WsConnector) procReap(s *Session) {
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
		go c.connectRoutine()
	}
}

func (c *WsConnector) reapRoutine() {
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

func (c *WsConnector) GetSessionConfig() *SessionConfig {
	return c.sc
}
