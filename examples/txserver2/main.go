// main
package main

import (
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/module"
)

func main() {
	defer core.ClosePackages()
	core.LoadPackages("config.json")

	waiter := module.Start()
	waiter.Wait("main")
}
