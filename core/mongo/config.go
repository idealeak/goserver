package mongo

import (
	"fmt"

	"github.com/idealeak/goserver/core"
	"labix.org/v2/mgo"
)

var Config = Configuration{
	Dbs: make(map[string]DbConfig),
}

var Database = make(map[string]*mgo.Database)

type Configuration struct {
	Dbs map[string]DbConfig
}
type DbConfig struct {
	Host     string
	Database string
	User     string
	Password string
	Safe     mgo.Safe
}

func (c *Configuration) Name() string {
	return "mongo"
}

func (c *Configuration) Init() error {
	for name, cf := range Config.Dbs {
		login := ""
		if cf.User != "" {
			login = cf.User + ":" + cf.Password + "@"
		}
		host := "localhost"
		if cf.Host != "" {
			host = cf.Host
		}

		// http://goneat.org/pkg/labix.org/v2/mgo/#Session.Mongo
		// [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
		url := fmt.Sprintf("mongodb://%s%s/%s", login, host, cf.Database)
		fmt.Println(url)
		session, err := mgo.Dial(url)
		if err != nil {
			return err
		}
		session.SetSafe(&cf.Safe)
		Database[name] = session.DB(cf.Database)
	}
	return nil
}

func (c *Configuration) Close() error {
	for _, db := range Database {
		if db != nil && db.Session != nil {
			db.Session.Close()
			db.Session = nil
			db = nil
		}
	}
	return nil
}
func init() {
	core.RegistePackage(&Config)
}
