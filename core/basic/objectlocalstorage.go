package basic

import (
	"math"
	"sync"

	"github.com/idealeak/goserver/core/container"
)

//  Be similar to (Windows, Thread Local Storage)

const OLS_MAX_SLOT uint = 64
const OLS_INVALID_SLOT = math.MaxUint32

type OlsSlotCleanHandler func(interface{})

var objSlotFlag uint64
var objSlotLock sync.Mutex
var objSlotCleanHandler [OLS_MAX_SLOT]OlsSlotCleanHandler
var objSlotHolder = container.NewSynchronizedMap()

func OlsAlloc() uint {
	objSlotLock.Lock()
	for i := uint(0); i < 64; i++ {
		if ((1 << i) & objSlotFlag) == 0 {
			objSlotFlag |= (1 << i)
			objSlotLock.Unlock()
			return i
		}
	}
	objSlotLock.Unlock()
	return OLS_INVALID_SLOT
}

func OlsFree(slot uint) {
	objSlotLock.Lock()
	defer objSlotLock.Unlock()
	if slot < OLS_MAX_SLOT {
		handler := objSlotCleanHandler[slot]
		flag := objSlotFlag & (1 << slot)
		if handler != nil && flag != 0 {
			objSlotFlag ^= (1 << slot)
			objSlotHolder.Foreach(func(k, v interface{}) {
				if o, ok := k.(*Object); ok && o != nil {
					v := o.ols[slot]
					if v != nil {
						o.ols[slot] = nil
						handler(v)
					}
				}
			})
		}
	}
}

func OlsInstallSlotCleanHandler(slot uint, handler OlsSlotCleanHandler) {
	if slot < OLS_MAX_SLOT {
		objSlotCleanHandler[slot] = handler
	}
}

func (o *Object) OlsGetValue(slot uint) interface{} {
	if slot < OLS_MAX_SLOT {
		return o.ols[slot]
	}
	return nil
}

func (o *Object) OlsSetValue(slot uint, val interface{}) {
	if slot < OLS_MAX_SLOT {
		old := o.ols[slot]
		o.ols[slot] = val
		if old != nil {
			handler := objSlotCleanHandler[slot]
			if handler != nil {
				handler(old)
			}
		}
		objSlotHolder.Set(o, struct{}{})
	}
}

func (o *Object) OlsClrValue() {
	for i := uint(0); i < OLS_MAX_SLOT; i++ {
		v := o.ols[i]
		if v != nil {
			handler := objSlotCleanHandler[i]
			if handler != nil {
				handler(v)
			}
		}
	}
}
