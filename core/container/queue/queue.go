// queue
package queue

import (
	"time"
)

type Queue interface {
	Len() int
	Enqueue(interface{}, time.Duration) bool
	Dequeue(time.Duration) (interface{}, bool)
}
