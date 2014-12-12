package profile

import (
	"github.com/idealeak/goserver/core"
)

var Config = Configuration{}

type Configuration struct {
	SlowMS int
}

func (c *Configuration) Name() string {
	return "profile"
}

func (c *Configuration) Init() error {
	if c.SlowMS <= 0 {
		c.SlowMS = 1000
	}
	return nil
}

func (c *Configuration) Close() error {
	return nil
}

func init() {
	core.RegistePackage(&Config)
}
