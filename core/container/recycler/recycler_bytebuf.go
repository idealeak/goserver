package recycler

import (
	"bytes"
)

const (
	BytebufRecyclerBacklog int = 128
)

var BytebufRecycler = NewRecycler(
	BytebufRecyclerBacklog,
	func() interface{} {
		return bytes.NewBuffer(nil)
	},
	"bytebuf_recycler",
)

func AllocBytebuf() *bytes.Buffer {
	b := BytebufRecycler.Get()
	buf := b.(*bytes.Buffer)
	buf.Reset()
	return buf
}

func FreeBytebuf(buf *bytes.Buffer) {
	BytebufRecycler.Give(buf)
}
