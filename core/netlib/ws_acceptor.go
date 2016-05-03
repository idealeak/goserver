package netlib

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/utils"
)

const (
	// Time allowed to write a message to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next message from the client.
	readWait = 60 * time.Second

	// Send pings to client with this period. Must be less than readWait.
	pingPeriod = (readWait * 9) / 10
)

type WsAcceptor struct {
	e           *NetEngine
	sc          *SessionConfig
	idGen       utils.IdGen
	mapSessions map[int]*WsSession
	reaper      chan ISession
	acptChan    chan *WsSession
	waitor      *utils.Waitor
	upgrader    websocket.Upgrader
	quit        bool
	reaped      bool
	maxActive   int
	maxDone     int
}

func newWsAcceptor(e *NetEngine, sc *SessionConfig) *WsAcceptor {
	a := &WsAcceptor{
		e:           e,
		sc:          sc,
		quit:        false,
		mapSessions: make(map[int]*WsSession),
		waitor:      utils.NewWaitor(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  sc.RcvBuff,
			WriteBufferSize: sc.SndBuff,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
	}

	a.init()

	return a
}

func (a *WsAcceptor) init() {

	temp := int(a.sc.MaxConn + a.sc.ExtraConn)
	a.reaper = make(chan ISession, temp)
	a.acptChan = make(chan *WsSession, temp)
}

func (a *WsAcceptor) start() (err error) {
	http.Handle(a.sc.Path, http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		defer utils.DumpStackIfPanic("ws.HandlerFunc")
		if req.Method != "GET" {
			http.Error(res, "method not allowed", 405)
			return
		}
		ws, err := a.upgrader.Upgrade(res, req, nil)
		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(res, "Not a websocket handshake", 400)
			return
		} else if err != nil {
			http.Error(res, fmt.Sprintf("%v", err), 500)
			logger.Error(err)
			return
		}
		ws.SetPongHandler(func(string) error {
			ws.SetReadDeadline(time.Now().Add(readWait))
			return nil
		})
		s := newWsSession(a.idGen.NextId(), ws, a.sc, a)
		a.acptChan <- s
	}))
	go func() {
		service := a.sc.Ip + ":" + strconv.Itoa(int(a.sc.Port))
		err := http.ListenAndServe(service, nil)
		if err != nil {
			logger.Error(err)
		}
	}()

	return nil
}

func (a *WsAcceptor) update() {
	a.procActive()
	a.procChanEvent()
}

func (a *WsAcceptor) shutdown() {

	if a.quit {
		return
	}

	a.quit = true

	if len(a.mapSessions) == 0 {
		go a.reapRoutine()
	} else {
		for _, v := range a.mapSessions {
			v.Close()
		}
	}
}

func (a *WsAcceptor) onClose(s ISession) {
	a.reaper <- s
}

func (a *WsAcceptor) procReap(s *Session) {
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

func (a *WsAcceptor) reapRoutine() {
	if a.reaped {
		return
	}
	a.reaped = true
	a.waitor.Wait()

	a.e.childAck <- a.sc.Id
}

func (a *WsAcceptor) procAccepted(s *WsSession) {
	a.mapSessions[s.Id] = s
	s.FireConnectEvent()
	s.start()
}

func (a *WsAcceptor) procActive() {
	var i int
	//var nowork bool
	var doneCnt int
	for _, v := range a.mapSessions {
		//nowork = true
		for i = 0; i < a.sc.MaxDone; i++ {
			if v.IsConned() && len(v.recvBuffer) > 0 {
				data, ok := <-v.recvBuffer
				if !ok {
					break
				}
				data.do()
				//nowork = false
				doneCnt++
			} else {
				break
			}
		}
		/*
			if nowork && v.IsConned() && v.IsIdle() {
				v.FireSessionIdle()
			}
		*/
	}

	if doneCnt > a.maxDone {
		a.maxDone = doneCnt
	}
	if len(a.mapSessions) > a.maxActive {
		a.maxActive = len(a.mapSessions)
	}
}

func (a *WsAcceptor) dump() {
	logger.Info("=========wsaccept dump maxSessions=", a.maxActive, " maxDone=", a.maxDone)
	for sid, s := range a.mapSessions {
		logger.Info("=========wssession:", sid, " recvBuffer size=", len(s.recvBuffer), " sendBuffer size=", len(s.sendBuffer))
	}
}

func (a *WsAcceptor) procChanEvent() {
	for i := 0; i < a.sc.MaxDone; i++ {
		select {
		case s := <-a.acptChan:
			a.procAccepted(s)
		case s := <-a.reaper:
			if ss, ok := s.(*Session); ok {
				a.procReap(ss)
			}

		default:
			return
		}
	}
}

func (a *WsAcceptor) GetSessionConfig() *SessionConfig {
	return a.sc
}

type WsAddr struct {
	acceptor *WsAcceptor
}

// name of the network
func (a *WsAddr) Network() string {
	return "WS"
}

// string form of address
func (a *WsAddr) String() string {
	return fmt.Sprintf("%v:%v", a.acceptor.sc.Ip, a.acceptor.sc.Port)
}

func (a *WsAcceptor) Addr() net.Addr {
	return &WsAddr{acceptor: a}
}
