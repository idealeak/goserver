package mongo

import (
	"fmt"
	"time"

	"github.com/idealeak/goserver/core"
	"labix.org/v2/mgo"
)

var Config = Configuration{
	Dbs: make(map[string]DbConfig),
}

var autoPingInterval time.Duration = 30
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
		session, err := newDBSession(&cf)
		if err == nil {
			Database[name] = session.DB(cf.Database)
		} else {
			fmt.Println(err)
		}
	}

	//auto ping, ensure net is connected
	go func() {
		for {
			select {
			case <-time.After(time.Second * autoPingInterval):
				Ping()
			}
		}
	}()
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

func newDBSession(dbc *DbConfig) (s *mgo.Session, err error) {
	login := ""
	if dbc.User != "" {
		login = dbc.User + ":" + dbc.Password + "@"
	}
	host := "localhost"
	if dbc.Host != "" {
		host = dbc.Host
	}

	// http://goneat.org/pkg/labix.org/v2/mgo/#Session.Mongo
	// [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
	url := fmt.Sprintf("mongodb://%s%s/%s", login, host, dbc.Database)
	fmt.Println(url)
	session, err := mgo.Dial(url)
	if err != nil {
		return
	}
	session.SetSafe(&dbc.Safe)
	s = session
	return
}

func Ping() {
	var err error
	for _, db := range Database {
		err = db.Session.Ping()
		if err != nil {
			fmt.Println("mongo.Ping error:", err)
			db.Session.Refresh()
		}
	}
}

func SetAutoPing(interv time.Duration) {
	autoPingInterval = interv
}
