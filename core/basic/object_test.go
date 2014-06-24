package basic

import (
	"testing"
	"time"
)

func TestObject(t *testing.T) {
	opt := Options{
		Timeout: time.Second,
		MaxDone: 10,
	}
	r := NewObject(opt)
	c := NewObject(opt)
	r.LaunchChild(c)

	if c.owner != r {
		t.Fatal("Object LaunchChild failed. Child Object's owner not correct")
	}

	r.ProcessCommand()
	if len(r.childs) == 0 {
		t.Fatal("Object LaunchChild failed. Parent Object's child not correct")
	}

	c.Terminate()
	r.ProcessCommand()
	if len(r.childs) != 0 {
		t.Fatal("Object Terminate Failed")
	}

	c.ProcessCommand()
	if !c.terminated {
		t.Fatal("Object Terminate Failed. Child Object statue error")
	}

	r.ProcessCommand()
	if r.termAcks != 0 {
		t.Fatal("Object Terminate Failed. Parent Object statue error")
	}
}
