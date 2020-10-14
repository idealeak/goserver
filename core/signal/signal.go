// signal
package signal

import (
	"errors"
	"fmt"
	"github.com/idealeak/goserver/core/utils"
	"os"
	"os/signal"
	"sync"

	"github.com/idealeak/goserver/core/logger"
)

var SignalHandlerModule = NewSignalHandler()

type Handler interface {
	Process(s os.Signal, ud interface{}) error
}

type SignalHandler struct {
	lock sync.RWMutex
	sc   chan os.Signal
	mh   map[os.Signal]map[Handler]interface{}
}

func NewSignalHandler() *SignalHandler {
	sh := &SignalHandler{
		sc: make(chan os.Signal, 10),
		mh: make(map[os.Signal]map[Handler]interface{}),
	}

	signal.Notify(sh.sc)
	return sh
}

func (this *SignalHandler) RegisteHandler(s os.Signal, h Handler, ud interface{}) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	if v, ok := this.mh[s]; !ok {
		m := make(map[Handler]interface{})
		this.mh[s] = m
		m[h] = ud
	} else {
		if _, has := v[h]; !has {
			v[h] = ud
		} else {
			return errors.New(fmt.Sprintf("SignalHandler.RegisterHandler repeate registe handle %v %v", s, h))
		}
	}

	return nil
}

func (this *SignalHandler) UnregisteHandler(s os.Signal, h Handler) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	if v, ok := this.mh[s]; ok {
		if _, has := v[h]; has {
			delete(v, h)
		}
	}

	return nil
}

func (this *SignalHandler) ClearHandler(s os.Signal) int {
	this.lock.Lock()
	defer this.lock.Unlock()
	if v, ok := this.mh[s]; ok {
		cnt := len(v)
		delete(this.mh, s)
		return cnt
	}
	return 0
}

func (this *SignalHandler) ProcessSignal() {
	logger.Logger.Trace("(this *SignalHandler) ProcessSignal()")
	for {
		select {
		case s, ok := <-this.sc:
			if !ok {
				logger.Logger.Trace("(this *SignalHandler) ProcessSignal() quit!!!")
				return
			}
			//logger.Logger.Warn("-------->receive Signal:", s)
			handlers := map[Handler]interface{}{}
			this.lock.RLock()
			v, ok := this.mh[s]
			if ok && len(v) > 0 {
				for hk, hv := range v {
					handlers[hk] = hv
				}
			}
			this.lock.RUnlock()
			if ok && len(handlers) > 0 {
				for hk, hv := range handlers {
					utils.CatchPanic(func() { hk.Process(s, hv) })
				}
			//} else {
			//	logger.Logger.Warn("-------->UnHandle Signal:", s)
			}
		}
	}
}
