package netlib

import "sync"

var ap = NewActionPool(1024)

func AllocAction() *action {
	return ap.Get()
}

func FreeAction(a *action) {
	ap.Give(a)
}

type ActionPool struct {
	free      *action
	lock      *sync.Mutex
	num       int
	allocNum  int
	remainNum int
}

func NewActionPool(num int) *ActionPool {
	ap := &ActionPool{
		lock: new(sync.Mutex),
		num:  num,
	}
	return ap
}

func (ap *ActionPool) grow() {
	var (
		i  int
		a  *action
		as = make([]action, ap.num)
	)
	ap.free = &(as[0])
	a = ap.free
	for i = 1; i < ap.num; i++ {
		a.next = &(as[i])
		a = a.next
	}
	a.next = nil
	ap.allocNum += ap.num
	ap.remainNum += ap.num
	return
}

func (ap *ActionPool) Get() (a *action) {
	ap.lock.Lock()
	if a = ap.free; a == nil {
		ap.grow()
		a = ap.free
	}
	ap.free = a.next
	a.next = nil
	ap.remainNum--
	ap.lock.Unlock()
	return
}

func (ap *ActionPool) Give(a *action) {
	ap.lock.Lock()
	a.next = ap.free
	ap.free = a
	ap.remainNum++
	ap.lock.Unlock()
}
