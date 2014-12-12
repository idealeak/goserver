package profile

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/idealeak/goserver/core/utils"
)

var TimeStatisticMgr = &timeStatisticMgr{
	elements: make(map[string]*timeElement),
}

type timeElement struct {
	name      string
	times     int64
	totalTick time.Duration
	maxTick   time.Duration
	minTick   time.Duration
	lastTick  time.Duration
}

type timeStatisticMgr struct {
	elements map[string]*timeElement
	l        sync.Mutex
}

func (this *timeStatisticMgr) WatchStart(name string) *TimeWatcher {
	tw := newTimeWatcher(name)
	return tw
}

func (this *timeStatisticMgr) addStatistic(name string, d time.Duration) {
	this.l.Lock()
	defer this.l.Unlock()

	if te, exist := this.elements[name]; exist {
		te.times++
		te.totalTick += d
		if d > te.maxTick {
			te.maxTick = d
		}
		if d < te.minTick {
			te.minTick = d
		}
		te.lastTick = d

	} else {
		this.elements[name] = &timeElement{
			name:      name,
			times:     1,
			totalTick: d,
			maxTick:   d,
			minTick:   d,
			lastTick:  d,
		}
	}
}

func (this *timeStatisticMgr) dump(w io.Writer) {
	this.l.Lock()
	defer this.l.Unlock()
	fmt.Fprintf(w, "| % -30s| % -10s | % -16s | % -16s | % -16s | % -16s |\n", "name", "times", "used", "max used", "min used", "avg used")
	for k, v := range this.elements {
		fmt.Fprintf(w, "| % -30s| % -10d | % -16s | % -16s | % -16s | % -16s |\n", strings.ToLower(k), v.times, utils.ToS(v.totalTick), utils.ToS(v.maxTick), utils.ToS(v.minTick), utils.ToS(time.Duration(int64(v.totalTick)/v.times)))
	}
}
