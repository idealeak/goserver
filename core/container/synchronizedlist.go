package container

import (
	"container/list"
	"sync"
)

type SynchronizedList struct {
	list *list.List
	lock *sync.Mutex
}

func NewSynchronizedList() *SynchronizedList {
	sl := &SynchronizedList{
		list: list.New(),
		lock: new(sync.Mutex),
	}
	return sl
}

func (sl *SynchronizedList) PushFront(v interface{}) {
	sl.lock.Lock()
	sl.list.PushFront(v)
	sl.lock.Unlock()
}

func (sl *SynchronizedList) PopFront() (v interface{}) {
	sl.lock.Lock()
	e := sl.list.Front()
	if e != nil {
		v = e.Value
		sl.list.Remove(e)
	}
	sl.lock.Unlock()
	return v
}

func (sl *SynchronizedList) PushBack(v interface{}) {
	sl.lock.Lock()
	sl.list.PushBack(v)
	sl.lock.Unlock()
}

func (sl *SynchronizedList) PopBack() (v interface{}) {
	sl.lock.Lock()
	e := sl.list.Back()
	if e != nil {
		v = e.Value
		sl.list.Remove(e)
	}
	sl.lock.Unlock()
	return v
}

func (sl *SynchronizedList) Len() (n int) {
	sl.lock.Lock()
	n = sl.list.Len()
	sl.lock.Unlock()
	return
}
