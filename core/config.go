package core

import (
	"runtime"
)

var Config = Configuration{}

type Configuration struct {
	MaxProcs int
}

func (c *Configuration) Name() string {
	return "core"
}

func (c *Configuration) Init() error {
	if c.MaxProcs <= 0 {
		c.MaxProcs = 1
	}
	runtime.GOMAXPROCS(c.MaxProcs)
	AppCtx.Object.Active()
	return nil
}

func (c *Configuration) Close() error {
	return nil
}

func init() {
	RegistePackage(&Config)
}
