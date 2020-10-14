package basic

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

func TestSendCommand(t *testing.T) {
	n := 5
	opt := Options{
		Interval: time.Second,
		MaxDone:  n,
	}
	c := make(chan int)
	o := NewObject(1, "test1", opt, nil)
	o.Active()
	for i := 0; i < n*2; i++ {
		go func(tag int) {
			o.SendCommand(CommandWrapper(func(*Object) error {
				c <- tag
				return nil
			}), true)
		}(i)
	}

	go func() {
		i := 0
		for {
			i++
			if i%1000 == 0 {
				runtime.Gosched()
			}
		}
	}()

	slice := make([]int, 0, n*2)
	for i := 0; i < n*2; i++ {
		tag := <-c
		slice = append(slice, tag)
	}
	if len(slice) != n*2 {
		t.Fatal("Command be droped")
	}
	fmt.Println("TestSendCommand", slice)
}

func TestSendCommandLoop(t *testing.T) {
	n := 5
	m := n * 2
	opt := Options{
		Interval: time.Second,
		MaxDone:  n,
	}
	c := make(chan int)
	o := NewObject(1, "test1", opt, nil)
	o.Active()
	for i := 0; i < n; i++ {
		go func(tag int) {
			o.SendCommand(CommandWrapper(func(oo *Object) error {
				for j := 0; j < m; j++ {
					func(tag2 int) {
						oo.SendCommand(CommandWrapper(func(*Object) error {
							c <- tag*1000 + tag2
							return nil
						}), true)
					}(j)
				}
				return nil
			}), true)
		}(i)
	}
	go func() {
		i := 0
		for {
			i++
			if i%1000 == 0 {
				runtime.Gosched()
			}
		}
	}()
	slice := make([]int, 0, n*m)
	for i := 0; i < n*m; i++ {
		tag := <-c
		slice = append(slice, tag)
	}
	if len(slice) != n*m {
		t.Fatal("Command be droped")
	}
	fmt.Println("TestSendCommandLoop", slice, len(slice))
}
