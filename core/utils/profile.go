package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strconv"
	"time"
)

var startTime = time.Now()
var pid int

func init() {
	pid = os.Getpid()
}

type RuntimeStats struct {
	CountGoroutine int
	CountHeap      int
	CountThread    int
	CountBlock     int
}

func ProcessInput(input string, w io.Writer) {
	switch input {
	case "lookup goroutine":
		p := pprof.Lookup("goroutine")
		p.WriteTo(w, 2)
	case "lookup heap":
		p := pprof.Lookup("heap")
		p.WriteTo(w, 2)
	case "lookup threadcreate":
		p := pprof.Lookup("threadcreate")
		p.WriteTo(w, 2)
	case "lookup block":
		p := pprof.Lookup("block")
		p.WriteTo(w, 2)
	case "start cpuprof":
		StartCPUProfile()
	case "stop cpuprof":
		StopCPUProfile()
	case "get memprof":
		MemProf()
	case "gc summary":
		PrintGCSummary(w)
	}
}

func MemProf() {
	if f, err := os.Create("mem-" + strconv.Itoa(pid) + ".memprof"); err != nil {
		log.Fatal("record memory profile failed: %v", err)
	} else {
		runtime.GC()
		defer f.Close()
		pprof.WriteHeapProfile(f)
	}
}

func StartCPUProfile() {
	f, err := os.Create("cpu-" + strconv.Itoa(pid) + ".pprof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
}

func StopCPUProfile() {
	pprof.StopCPUProfile()
}

func PrintGCSummary(w io.Writer) {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	gcstats := &debug.GCStats{PauseQuantiles: make([]time.Duration, 100)}
	debug.ReadGCStats(gcstats)

	printGC(memStats, gcstats, w)
}

func printGC(memStats *runtime.MemStats, gcstats *debug.GCStats, w io.Writer) {

	if gcstats.NumGC > 0 {
		lastPause := gcstats.Pause[0]
		elapsed := time.Now().Sub(startTime)
		overhead := float64(gcstats.PauseTotal) / float64(elapsed) * 100
		allocatedRate := float64(memStats.TotalAlloc) / elapsed.Seconds()

		fmt.Fprintf(w, "NumGC:%d Pause:%s Pause(Avg):%s Overhead:%3.2f%% Alloc:%s Sys:%s Alloc(Rate):%s/s Histogram:%s %s %s \n",
			gcstats.NumGC,
			ToS(lastPause),
			ToS(Avg(gcstats.Pause)),
			overhead,
			ToH(memStats.Alloc),
			ToH(memStats.Sys),
			ToH(uint64(allocatedRate)),
			ToS(gcstats.PauseQuantiles[94]),
			ToS(gcstats.PauseQuantiles[98]),
			ToS(gcstats.PauseQuantiles[99]))
	} else {
		// while GC has disabled
		elapsed := time.Now().Sub(startTime)
		allocatedRate := float64(memStats.TotalAlloc) / elapsed.Seconds()

		fmt.Fprintf(w, "Alloc:%s Sys:%s Alloc(Rate):%s/s\n",
			ToH(memStats.Alloc),
			ToH(memStats.Sys),
			ToH(uint64(allocatedRate)))
	}
}

func StatsRuntime() RuntimeStats {
	stats := RuntimeStats{}
	stats.CountGoroutine = runtime.NumGoroutine()
	stats.CountThread, _ = runtime.ThreadCreateProfile(nil)
	stats.CountHeap, _ = runtime.MemProfile(nil, true)
	stats.CountBlock, _ = runtime.BlockProfile(nil)
	return stats
}
