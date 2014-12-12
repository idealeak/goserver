// signal
package signal

import (
	"errors"
	"fmt"
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
		this.mh[s] = make(map[Handler]interface{})
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
		if _, has := v[h]; !has {
			delete(v, h)
		}
	}

	return nil
}

func (this *SignalHandler) ProcessSignal() {
	logger.Info("(this *SignalHandler) ProcessSignal()")
	for {
		select {
		case s := <-this.sc:
			logger.Info("-------->receive UnHandle Signal:", s)
			this.lock.RLock()
			defer this.lock.RUnlock()
			if v, ok := this.mh[s]; ok {
				for hk, hv := range v {
					hk.Process(s, hv)
				}
			} else {
				logger.Info("-------->receive UnHandle Signal:", s)
			}
		}
	}
}
