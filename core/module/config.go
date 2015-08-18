package module

import (
	"time"

	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
)

var Config = Configuration{}

type Configuration struct {
	Options basic.Options
}

func (c *Configuration) Name() string {
	return "module"
}

func (c *Configuration) Init() error {
	if c.Options.QueueBacklog <= 0 {
		c.Options.QueueBacklog = 1024
	}
	if c.Options.MaxDone <= 0 {
		c.Options.MaxDone = 1024
	}
	if c.Options.Interval <= 0 {
		c.Options.Interval = time.Millisecond * 10
	} else {
		c.Options.Interval = time.Millisecond * c.Options.Interval
	}

	return nil
}

func (c *Configuration) Close() error {
	return nil
}

func init() {
	core.RegistePackage(&Config)
}
