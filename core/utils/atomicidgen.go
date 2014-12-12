// AtomicIdGen
package utils

import (
	"sync/atomic"
)

type AtomicIdGen struct {
	cur uint32
	beg uint32
}

func (this *AtomicIdGen) NextId() uint32 {
	return atomic.AddUint32(&this.cur, 1)
}

func (this *AtomicIdGen) Reset() {
	atomic.StoreUint32(&this.cur, this.beg)
}

func (this *AtomicIdGen) SetStartPoint(startPoint uint32) {
	this.beg = startPoint
	this.Reset()
}

func (this *AtomicIdGen) CurrId() uint32 {
	return this.cur
}
