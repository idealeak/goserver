package main

import (
	"fmt"
	"time"

	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/timer"
)

var TimerExampleSington = &TimerExample{}

type TimerExample struct {
}

////////////////////////////////////////////////////////////////////
/// Module Implement [beg]
////////////////////////////////////////////////////////////////////
func (this *TimerExample) ModuleName() string {
	return "timerexample"
}

func (this *TimerExample) Init() {
	var i int
	h, b := timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		i++
		fmt.Println(i, time.Now())

		if i > 5 {
			return false
		}

		return true
	}), nil, time.Second, 10)
	fmt.Println("timer lauch ", h, b)
}

func (this *TimerExample) Update() {
	fmt.Println("timer queue len=", timer.TimerModule.TimerCount())
}

func (this *TimerExample) Shutdown() {
	module.UnregisteModule(this)
}

////////////////////////////////////////////////////////////////////
/// Module Implement [end]
////////////////////////////////////////////////////////////////////

func init() {
	module.RegisteModule(TimerExampleSington, time.Second, 0)
}
