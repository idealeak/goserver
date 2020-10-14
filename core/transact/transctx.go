// transctx
package transact

import (
	"sync"
)

type TransCtx struct {
	fields map[interface{}]interface{}
	lock   *sync.RWMutex
}

func NewTransCtx() *TransCtx {
	tc := &TransCtx{
		lock: new(sync.RWMutex),
	}
	return tc
}

func (this *TransCtx) SetField(k, v interface{}) {
	this.lock.Lock()
	if this.fields == nil {
		this.fields = make(map[interface{}]interface{})
	}
	this.fields[k] = v
	this.lock.Unlock()
}

func (this *TransCtx) GetField(k interface{}) interface{} {
	this.lock.RLock()
	if this.fields != nil {
		if v, exist := this.fields[k]; exist {
			this.lock.RUnlock()
			return v
		}
	}
	this.lock.RUnlock()
	return nil
}
