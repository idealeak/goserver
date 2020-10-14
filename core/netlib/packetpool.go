package netlib

import "sync"

var pp = NewPacketPool(10240)

func AllocPacket() *packet {
	return pp.Get()
}

func FreePacket(p *packet) {
	pp.Give(p)
}

type PacketPool struct {
	free      *packet
	lock      *sync.Mutex
	num       int
	allocNum  int
	remainNum int
}

func NewPacketPool(num int) *PacketPool {
	pp := &PacketPool{
		lock: new(sync.Mutex),
		num:  num,
	}
	return pp
}

func (pp *PacketPool) grow() {
	var (
		i  int
		p  *packet
		ps = make([]packet, pp.num)
	)
	pp.free = &(ps[0])
	p = pp.free
	for i = 1; i < pp.num; i++ {
		p.next = &(ps[i])
		p = p.next
	}
	p.next = nil
	pp.allocNum += pp.num
	pp.remainNum += pp.num
	return
}

func (pp *PacketPool) Get() (p *packet) {
	pp.lock.Lock()
	if p = pp.free; p == nil {
		pp.grow()
		p = pp.free
	}
	pp.free = p.next
	p.next = nil
	pp.remainNum--
	pp.lock.Unlock()
	return
}

func (pp *PacketPool) Give(p *packet) {
	if p.next != nil {
		return
	}
	pp.lock.Lock()
	p.next = pp.free
	pp.free = p
	pp.remainNum++
	pp.lock.Unlock()
}
