package profile

import "sync"

var wp = NewWatcherPool(1024)

func AllocWatcher() *TimeWatcher {
	return wp.Get()
}

func FreeWatcher(t *TimeWatcher) {
	wp.Give(t)
}

type WatcherPool struct {
	free      *TimeWatcher
	lock      *sync.Mutex
	num       int
	allocNum  int
	remainNum int
}

func NewWatcherPool(num int) *WatcherPool {
	wp := &WatcherPool{
		lock: new(sync.Mutex),
		num:  num,
	}
	return wp
}

func (wp *WatcherPool) grow() {
	var (
		i  int
		t  *TimeWatcher
		ts = make([]TimeWatcher, wp.num)
	)
	wp.free = &(ts[0])
	t = wp.free
	for i = 1; i < wp.num; i++ {
		t.next = &(ts[i])
		t = t.next
	}
	t.next = nil
	wp.allocNum += wp.num
	wp.remainNum += wp.num
	return
}

func (wp *WatcherPool) Get() (t *TimeWatcher) {
	wp.lock.Lock()
	if t = wp.free; t == nil {
		wp.grow()
		t = wp.free
	}
	wp.free = t.next
	t.next = nil
	wp.remainNum--
	wp.lock.Unlock()
	return
}

func (wp *WatcherPool) Give(t *TimeWatcher) {
	wp.lock.Lock()
	t.next = wp.free
	wp.free = t
	wp.remainNum++
	wp.lock.Unlock()
}
