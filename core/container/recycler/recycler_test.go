// recycler_test
package recycler

import (
	"testing"
	"time"
)

func makeBuffer() interface{} {
	buf := make([]byte, 0, 1024)
	return buf
}

var MyRecycler = NewRecycler(RecyclerBacklogDefault, makeBuffer)

func TestGet(t *testing.T) {
	if len(MyRecycler.get) != RecyclerBacklogDefault {
		t.Fatal("Recycler get size error")
	}
	if MyRecycler.que.Len() != 1 {
		t.Fatal("Recycler inner que error")
	}
	MyRecycler.Get()
}

func TestGive(t *testing.T) {
	MyRecycler.Give(nil)

	time.Sleep(time.Second)

	if MyRecycler.que.Len() != 2 {
		t.Fatal("Recycler inner que size error")
	}
}

func BenchmarkGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MyRecycler.Get()
	}
}
