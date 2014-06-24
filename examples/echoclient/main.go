// main
package main

import (
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/module"
)

func main() {
	defer core.ClosePackages()
	core.LoadPackages("config.json")

	runtime.GOMAXPROCS(core.Config.MaxProcs)

	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	waiter := module.Start()
	waiter.Wait()
}
