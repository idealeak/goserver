// recycler
package recycler

import (
	"container/list"
	"time"
)

var RecyclerBacklogDefault int = 5

type element struct {
	when time.Time
	data interface{}
}

type Recycler struct {
	get     chan interface{}
	give    chan interface{}
	ocf     func() interface{}
	que     *list.List
	timeout *time.Timer
	makecnt int
	name    string
	running bool
}

func NewRecycler(backlog int, ocf func() interface{}, name string) *Recycler {
	r := &Recycler{
		get:     make(chan interface{}, backlog),
		give:    make(chan interface{}, backlog),
		ocf:     ocf,
		que:     list.New(),
		timeout: time.NewTimer(time.Minute),
		name:    name,
		running: true,
	}

	go r.run()
	return r
}

func (this *Recycler) run() {
	RecyclerMgr.registe(this)
	defer RecyclerMgr.unregiste(this)

	for this.running {
		if this.que.Len() == 0 {
			this.que.PushFront(element{when: time.Now(), data: this.ocf()})
			this.makecnt++
		}

		this.timeout.Reset(time.Minute)
		e := this.que.Front()
		select {
		case d := <-this.give:
			this.timeout.Stop()
			this.que.PushFront(element{when: time.Now(), data: d})
		case this.get <- e.Value.(element).data:
			this.timeout.Stop()
			this.que.Remove(e)
		case <-this.timeout.C:
			e := this.que.Front()
			for e != nil {
				n := e.Next()
				if time.Since(e.Value.(element).when) > time.Minute {
					this.que.Remove(e)
					e.Value = nil
					this.makecnt--
				}
				e = n
			}
		}
	}
}

func (this *Recycler) Get() interface{} {
	i := <-this.get
	return i
}

func (this *Recycler) Give(i interface{}) {
	this.give <- i
}

func (this *Recycler) Close() {
	this.running = false
}
