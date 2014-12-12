// queue_test
package queue

import (
	"testing"
)

func TestEnqueue(t *testing.T) {
	const CNT int = 10
	q := NewSyncQueue()
	for i := 0; i < CNT; i++ {
		q.Enqueue(i)
	}
	if q.Len() != CNT {
		t.Error("queue Enqueue error")
	}
}

func TestDequeue(t *testing.T) {
	const CNT int = 10
	q := NewSyncQueue()
	for i := 0; i < CNT; i++ {
		q.Enqueue(i)
	}

	var (
		b   bool = true
		d   interface{}
		cnt int
	)

	for b {
		d, b = q.Dequeue()
		if b {
			cnt++
			t.Log("Dequeue data:", d)
		}
	}
	if cnt != CNT {
		t.Error("queue Dequeue error")
	}
}
