// queue_test
package queue

import (
	"testing"
	"time"
)

func TestSyncQueneEnqueue(t *testing.T) {
	const CNT int = 10
	q := NewQueueS()
	for i := 0; i < CNT; i++ {
		q.Enqueue(i, 0)
	}
	if q.Len() != CNT {
		t.Error("sync queue Enqueue error")
	}
}

func TestSyncQueneDequeue(t *testing.T) {
	const CNT int = 10
	q := NewQueueS()
	for i := 0; i < CNT; i++ {
		q.Enqueue(i, 0)
	}

	var (
		b   bool = true
		d   interface{}
		cnt int
	)

	for b {
		d, b = q.Dequeue(0)
		if b {
			cnt++
			t.Log("Dequeue data:", d)
		}
	}
	if cnt != CNT {
		t.Error("sync queue Dequeue error")
	}
}

func BenchmarkSyncQueneEnqueue(b *testing.B) {
	q := NewQueueS()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(i, 0)
	}
	b.StopTimer()
}

func BenchmarkSyncQueneDequeue(b *testing.B) {
	q := NewQueueS()
	for i := 0; i < b.N; i++ {
		q.Enqueue(i, 0)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		q.Dequeue(0)
	}
	b.StopTimer()
}

func BenchmarkChanQueneEnqueue(b *testing.B) {
	q := NewQueueC(b.N)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(i, time.Millisecond)
	}
	b.StopTimer()
}

func BenchmarkChanQueneDequeue(b *testing.B) {
	q := NewQueueC(b.N)
	for i := 0; i < b.N; i++ {
		q.Enqueue(i, 0)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		q.Dequeue(0)
	}
	b.StopTimer()
}
