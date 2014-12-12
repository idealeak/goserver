package basic

import (
	"time"
)

const (
	QueueType_List int = iota
	QueueType_Chan
)

type Options struct {
	//  HeartBeat interval
	Interval time.Duration
	//	The maximum number of processing each heartbeat
	MaxDone int
	//
	QueueBacklog int
}
