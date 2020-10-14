// recycler_test
package recycler

import (
	"runtime"
	"testing"
)

func makeBuffer() interface{} {
	buf := make([]byte, 0, 1024)
	return buf
}

var MyRecycler = NewRecycler(RecyclerBacklogDefault, makeBuffer, "test")

func TestGet(t *testing.T) {
	MyRecycler.Get()
}

func TestGive(t *testing.T) {
	MyRecycler.Give(nil)
}

func BenchmarkGet(b *testing.B) {
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		MyRecycler.Get()
	}
	b.StopTimer()
}
