package utils

import (
	"runtime"

	"encoding/json"
	"fmt"
	"github.com/idealeak/goserver/core/logger"
	"sync"
	"sync/atomic"
	"time"
)

var _panicStackMgr = &PanicStackMgr{
	items: make(map[string]*PanicStackInfo),
}

type PanicStackMgr struct {
	sync.RWMutex
	items map[string]*PanicStackInfo
}

type PanicStackInfo struct {
	FirstTime time.Time
	LastTime  time.Time
	Times     int64
	ErrorMsg  string
	StackBuf  string
}

func DumpStackIfPanic(f string) {
	if err := recover(); err != nil {
		defer func() { //防止二次panic
			if err := recover(); err != nil {
				logger.Logger.Error(f, " panic.panic,error=", err)
			}
		}()
		logger.Logger.Error(f, " panic,error=", err)
		errMsg := fmt.Sprintf("%v", err)
		var buf [4096]byte
		n := runtime.Stack(buf[:], false)
		logger.Logger.Error("stack--->", string(buf[:n]))
		stk := make([]uintptr, 32)
		m := runtime.Callers(0, stk[:])
		stk = stk[:m]
		if len(stk) > 0 {
			d, err := json.Marshal(stk)
			if err == nil && len(d) > 0 {
				key := string(d)
				_panicStackMgr.Lock()
				defer _panicStackMgr.Unlock()
				tNow := time.Now()
				if ps, exist := _panicStackMgr.items[key]; exist {
					atomic.AddInt64(&ps.Times, 1)
					ps.LastTime = tNow
				} else {
					ps = &PanicStackInfo{
						ErrorMsg:  errMsg,
						Times:     1,
						StackBuf:  string(buf[:n]),
						FirstTime: tNow,
						LastTime:  tNow,
					}
					_panicStackMgr.items[key] = ps
				}
			}
		}
	}
}

func DumpStack(f string) {
	logger.Logger.Error(f)
	var buf [4096]byte
	len := runtime.Stack(buf[:], false)
	logger.Logger.Error("stack--->", string(buf[:len]))
}

func GetPanicStats() map[string]PanicStackInfo {
	stats := make(map[string]PanicStackInfo)
	_panicStackMgr.RLock()
	defer _panicStackMgr.RUnlock()
	for k, v := range _panicStackMgr.items {
		stats[k] = *v
	}
	return stats
}
