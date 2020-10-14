package admin

import (
	"github.com/idealeak/goserver/core"
)

var Config = Configuration{}

type Configuration struct {
	SupportAdmin  bool
	AdminHttpAddr string
	AdminHttpPort int
	WhiteHttpAddr []string
}

func (c *Configuration) Name() string {
	return "admin"
}

func (c *Configuration) Init() error {
	if c.SupportAdmin {
		MyAdminApp.Start(c.AdminHttpAddr, c.AdminHttpPort)
	}
	return nil
}

func (c *Configuration) Close() error {
	return nil
}

func init() {
	core.RegistePackage(&Config)
}
