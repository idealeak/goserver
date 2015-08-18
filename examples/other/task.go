package main

import (
	"fmt"
	"time"

	"github.com/idealeak/goserver/core/task"
)

type TaskExample struct {
}

//in task.Worker goroutine
func (this *TaskExample) Call() interface{} {
	fmt.Println("TaskExample start execute")
	return nil
}

// in laucher goroutine
func (this *TaskExample) Done(i interface{}) {
	fmt.Println("TaskExample execute over")
}

func init() {
	time.AfterFunc(time.Second*5, func() {
		th := &TaskExample{}
		t := task.New(nil, th, th)
		if b := t.StartByExecutor("test"); !b {
			fmt.Println("task lauch failed")
		} else {
			fmt.Println("task lauch success")
		}
	})
}
