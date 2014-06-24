// transctx
package transact

import (
	"sync"
)

type TransCtx struct {
	fields map[interface{}]interface{}
	mutex  *sync.Mutex
}

func NewTransCtx() *TransCtx {
	tc := &TransCtx{
		mutex: new(sync.Mutex),
	}
	return tc
}

func (this *TransCtx) SetField(k, v interface{}) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if this.fields == nil {
		this.fields = make(map[interface{}]interface{})
	}
	this.fields[k] = v
}

func (this *TransCtx) GetField(k interface{}) interface{} {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if this.fields != nil {
		if v, exist := this.fields[k]; exist {
			return v
		}
	}

	return nil
}
