package main

import (
	"fmt"
	"time"

	"github.com/idealeak/goserver/core/timer"
)

func init() {
	time.AfterFunc(time.Second*5, func() {
		var i int
		h, b := timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
			i++
			fmt.Println(i, time.Now())
			return true
		}), nil, time.Second, 10)
		fmt.Println("timer lauch ", h, b)
	})
}
