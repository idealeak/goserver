package recycler

import (
	"fmt"
	"io"
	"sync"
)

var RecyclerMgr = &recyclerMgr{
	recyclers: make(map[interface{}]*Recycler),
	lock:      new(sync.Mutex),
}

type recyclerMgr struct {
	recyclers map[interface{}]*Recycler
	lock      *sync.Mutex
}

func (this *recyclerMgr) registe(r *Recycler) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.recyclers[r] = r
}

func (this *recyclerMgr) unregiste(r *Recycler) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.recyclers, r)
}

func (this *recyclerMgr) CloseAll() {
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, r := range this.recyclers {
		r.Close()
	}
}

func (this *recyclerMgr) Dump(w io.Writer) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for _, r := range this.recyclers {
		w.Write([]byte(fmt.Sprintf("(%s) alloc object (%d)", r.name, r.makecnt)))
	}
}
