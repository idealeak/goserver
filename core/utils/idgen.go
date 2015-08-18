// idgen
package utils

import (
	"sync/atomic"
)

type IdGen struct {
	beg int32
	seq int32
}

func (this *IdGen) NextId() int {
	seq := atomic.AddInt32(&this.seq, 1)
	return int(seq)
}

func (this *IdGen) Reset() {
	atomic.StoreInt32(&this.seq, this.beg)
}

func (this *IdGen) SetSeq(seq int) {
	atomic.StoreInt32(&this.seq, int32(seq))
}

func (this *IdGen) SetStartPoint(startPoint int) {
	atomic.StoreInt32(&this.beg, int32(startPoint))
}

func (this *IdGen) CurrId() int {
	return int(atomic.LoadInt32(&this.seq))
}
