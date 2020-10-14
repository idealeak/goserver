package mongo

import (
	"fmt"
	"github.com/idealeak/goserver/core/logger"
	"sync"
	"sync/atomic"
	"time"

	"github.com/globalsign/mgo"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/container"
)

var Config = Configuration{
	Dbs: make(map[string]DbConfig),
}

var autoPingInterval time.Duration = 30 * time.Second
var mgoSessions = container.NewSynchronizedMap()
var databases = container.NewSynchronizedMap()
var cLock sync.RWMutex
var collections = make(map[string]*Collection)
var resetingSess int32

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
	//auto ping, ensure net is connected
	go func() {
		for {
			select {
			case <-time.After(autoPingInterval):
				Ping()
			}
		}
	}()
	return nil
}

func (c *Configuration) Close() error {
	sessions := mgoSessions.Items()
	for k, s := range sessions {
		if session, ok := s.(*mgo.Session); ok && session != nil {
			logger.Logger.Warnf("mongo.Close!!! (%v)", k)
			session.Close()
		}
	}
	return nil
}

func init() {
	core.RegistePackage(&Config)
}

type Collection struct {
	*mgo.Collection
	Ref   int32
	valid bool
}

func (c *Collection) Hold() {
	if atomic.AddInt32(&c.Ref, 1) == 1 {
		key := c.Database.Name + c.FullName
		cLock.Lock()
		if old, exist := collections[key]; exist {
			old.valid = false
		}
		collections[key] = c
		cLock.Unlock()
	}
}

func (c *Collection) Unhold() {
	if atomic.AddInt32(&c.Ref, -1) == 0 {
		key := c.Database.Name + c.FullName
		cLock.Lock()
		delete(collections, key)
		cLock.Unlock()
	}
}

func (c *Collection) IsValid() bool {
	return c.valid
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
	url := fmt.Sprintf("mongodb://%s%s/admin", login, host)
	//fmt.Println(url)
	session, err := mgo.Dial(url)
	if err != nil {
		return
	}
	session.SetSafe(&dbc.Safe)
	s = session
	return
}

func Ping() {
	if atomic.LoadInt32(&resetingSess) == 1 {
		return
	}
	var err error
	sessions := mgoSessions.Items()
	for k, s := range sessions {
		if session, ok := s.(*mgo.Session); ok && session != nil {
			if atomic.LoadInt32(&resetingSess) == 1 {
				return
			}
			err = session.Ping()
			if err != nil {
				logger.Logger.Errorf("mongo.Ping (%v) err:%v", k, err)
				if atomic.LoadInt32(&resetingSess) == 1 {
					return
				}
				session.Refresh()
			} else {
				logger.Logger.Tracef("mongo.Ping (%v) suc", k)
			}
		}
	}
}

func SetAutoPing(interv time.Duration) {
	autoPingInterval = interv
}

func Database(dbName string) *mgo.Database {
	if atomic.LoadInt32(&resetingSess) == 1 {
		return nil
	}
	var dbc DbConfig
	var exist bool
	if dbc, exist = Config.Dbs[dbName]; !exist {
		return nil
	}
	d := databases.Get(dbName)
	if d == nil {
		s, err := newDBSession(&dbc)
		if err != nil {
			fmt.Println("Database:", dbName, " error:", err)
			return nil
		}
		mgoSessions.Set(dbName, s)
		db := s.DB(dbc.Database)
		if db == nil {
			return nil
		}
		databases.Set(dbName, db)
		return db
	} else {
		if db, ok := d.(*mgo.Database); ok {
			return db
		}
	}
	return nil
}

func DatabaseC(dbName, collectionName string) *Collection {
	if atomic.LoadInt32(&resetingSess) == 1 {
		return nil
	}
	//一个库共享一个连接池
	db := Database(dbName)
	if db != nil {
		c := db.C(collectionName)
		if c != nil {
			return &Collection{Collection: c, valid: true}
		}
	}
	return nil
}

//不严格的多线程保护
func ResetAllSession() {
	atomic.StoreInt32(&resetingSess, 1)
	defer atomic.StoreInt32(&resetingSess, 0)
	tstart := time.Now()
	logger.Logger.Warnf("ResetAllSession!!! start.")
	sessions := mgoSessions.Items()
	mgoSessions = container.NewSynchronizedMap()
	databases = container.NewSynchronizedMap()

	//使缓存无效
	cLock.Lock()
	for k, c := range collections {
		c.valid = false
		logger.Logger.Warnf("%s collections reset.", k)
	}
	collections = make(map[string]*Collection)
	cLock.Unlock()

	//关闭旧的session
	for k, s := range sessions {
		if session, ok := s.(*mgo.Session); ok && session != nil {
			logger.Logger.Warnf("mongo.Close!!! (%v)", k)
			session.Close()
		}
	}
	logger.Logger.Warnf("ResetAllSession!!! end. take:%v", time.Now().Sub(tstart))
}
