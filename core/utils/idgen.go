// idgen
package utils

type IdGen struct {
	beg int
	seq int
}

func (this *IdGen) NextId() int {
	this.seq++
	return this.seq
}

func (this *IdGen) Reset() {
	this.seq = this.beg
}

func (this *IdGen) SetSeq(seq int) {
	this.seq = seq
}

func (this *IdGen) SetStartPoint(startPoint int) {
	this.beg = startPoint
	if this.seq < this.beg {
		this.seq = this.beg
	}
}

func (this *IdGen) CurrId() int {
	return this.seq
}
