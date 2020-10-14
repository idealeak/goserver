package netlib

import "testing"

func TestAllocAction(t *testing.T) {
	AllocAction()
}

func TestFreeAction(t *testing.T) {
	a := AllocAction()
	FreeAction(a)
}

func BenchmarkAllocAction(b *testing.B) {
	tt := make([]*action, 0, b.N)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tt = append(tt, AllocAction())
	}
	b.StopTimer()
}
