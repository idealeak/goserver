package timer

import (
	"container/heap"
	"testing"
	"time"
)

func TestTimerQueuePush(t *testing.T) {
	tq := NewTimerQueue()
	tNow := time.Now()
	te1 := &TimerEntity{
		ud:   int(2),
		next: tNow.Add(time.Minute),
	}
	te2 := &TimerEntity{
		ud:   int(1),
		next: tNow.Add(time.Second),
	}
	te3 := &TimerEntity{
		ud:   int(3),
		next: tNow.Add(time.Hour),
	}
	heap.Push(tq, te2)
	heap.Push(tq, te1)
	heap.Push(tq, te3)

	if tq.Len() != 3 {
		t.Fatal("Timer Queue Size error")
	}
	var (
		tee interface{}
		te  *TimerEntity
		ok  bool
	)
	tee = heap.Pop(tq)
	if te, ok = tee.(*TimerEntity); ok {
		if te.ud.(int) != 1 {
			t.Fatal("First Must 1.")
		}
	}

	tee = heap.Pop(tq)
	if te, ok = tee.(*TimerEntity); ok {
		if te.ud.(int) != 2 {
			t.Fatal("Second Must 2.")
		}
	}

	tee = heap.Pop(tq)
	if te, ok = tee.(*TimerEntity); ok {
		if te.ud.(int) != 3 {
			t.Fatal("Third Must 3.")
		}
	}
}

func BenchmarkTimerQueuePush(b *testing.B) {
	tq := NewTimerQueue()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		h := generateTimerHandle()
		te := &TimerEntity{
			h: h,
		}
		tq.Push(te)
	}
	b.StopTimer()
}

func BenchmarkTimerQueuePop(b *testing.B) {
	tq := NewTimerQueue()

	for i := 0; i < b.N; i++ {
		h := generateTimerHandle()
		te := &TimerEntity{
			h: h,
		}
		tq.Push(te)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		te := tq.Pop()
		tq.Push(te)
	}
	b.StopTimer()
}
