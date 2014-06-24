// main
package main

import (
	"runtime"

	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/module"
)

func main() {
	defer core.ClosePackages()
	core.LoadPackages("config.json")

	runtime.GOMAXPROCS(core.Config.MaxProcs)

	waiter := module.Start()
	waiter.Wait()
}
