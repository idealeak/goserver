package netlib

import (
	"github.com/idealeak/goserver/core/container/recycler"
)

const (
	RWBufRecyclerBacklog int = 128
)

var RWRecycler = recycler.NewRecycler(
	RWBufRecyclerBacklog,
	func() interface{} {
		rb := &RWBuffer{
			buf: make([]byte, 0, MaxPacketSize),
		}

		return rb
	},
	"rwbuf_recycler",
)

func AllocRWBuf() *RWBuffer {
	b := RWRecycler.Get()
	rb := b.(*RWBuffer)
	rb.Init()
	return rb
}

func FreeRWBuf(buf *RWBuffer) {
	RWRecycler.Give(buf)
}
