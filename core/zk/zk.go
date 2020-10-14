package zk

import (
	"errors"
	"path"
	"strings"
	"time"

	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/samuel/go-zookeeper/zk"
)

var (
	// error
	ErrNoChild      = errors.New("zk: children is nil")
	ErrNodeNotExist = errors.New("zk: node not exist")
)

// Connect connect to zookeeper, and start a goroutine log the event.
func Connect(addr []string, timeout time.Duration) (*zk.Conn, error) {
	conn, session, err := zk.Connect(addr, timeout)
	if err != nil {
		logger.Logger.Errorf("zk.Connect(\"%v\", %d) error(%v)", addr, timeout, err)
		return nil, err
	}
	go func() {
		for {
			event := <-session
			logger.Logger.Tracef("zookeeper get a event: %s", event.State.String())
		}
	}()
	return conn, nil
}

// Create create zookeeper path, if path exists ignore error
func Create(conn *zk.Conn, fpath string) error {
	// create zk root path
	tpath := ""
	for _, str := range strings.Split(fpath, "/")[1:] {
		tpath = path.Join(tpath, "/", str)
		logger.Logger.Tracef("create zookeeper path: \"%s\"", tpath)
		_, err := conn.Create(tpath, []byte(""), 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			if err == zk.ErrNodeExists {
				logger.Logger.Warnf("zk.create(\"%s\") exists", tpath)
			} else {
				logger.Logger.Errorf("zk.create(\"%s\") error(%v)", tpath, err)
				return err
			}
		}
	}

	return nil
}

// RegisterTmp create a ephemeral node, and watch it, if node droped then shutdown.
func RegisterTemp(conn *zk.Conn, fpath string, data []byte) error {
	tpath, err := conn.Create(path.Join(fpath)+"/", data, zk.FlagEphemeral|zk.FlagSequence, zk.WorldACL(zk.PermAll))
	if err != nil {
		logger.Logger.Errorf("conn.Create(\"%s\", \"%s\", zk.FlagEphemeral|zk.FlagSequence) error(%v)", fpath, string(data), err)
		return err
	}
	logger.Logger.Tracef("create a zookeeper node:%s", tpath)
	// watch self
	go func() {
		for {
			logger.Logger.Infof("zk path: \"%s\" set a watch", tpath)
			exist, _, watch, err := conn.ExistsW(tpath)
			if err != nil {
				logger.Logger.Errorf("zk.ExistsW(\"%s\") error(%v)", tpath, err)
				logger.Logger.Warnf("zk path: \"%s\" set watch failed, shutdown", tpath)
				module.Stop()
				return
			}
			if !exist {
				logger.Logger.Warnf("zk path: \"%s\" not exist, shutdown", tpath)
				module.Stop()
				return
			}
			event := <-watch
			logger.Logger.Infof("zk path: \"%s\" receive a event %v", tpath, event)
		}
	}()
	return nil
}

// GetNodesW get all child from zk path with a watch.
func GetNodesW(conn *zk.Conn, path string) ([]string, <-chan zk.Event, error) {
	nodes, stat, watch, err := conn.ChildrenW(path)
	if err != nil {
		if err == zk.ErrNoNode {
			return nil, nil, ErrNodeNotExist
		}
		logger.Logger.Errorf("zk.ChildrenW(\"%s\") error(%v)", path, err)
		return nil, nil, err
	}
	if stat == nil {
		return nil, nil, ErrNodeNotExist
	}
	if len(nodes) == 0 {
		return nil, nil, ErrNoChild
	}
	return nodes, watch, nil
}

// GetNodes get all child from zk path.
func GetNodes(conn *zk.Conn, path string) ([]string, error) {
	nodes, stat, err := conn.Children(path)
	if err != nil {
		if err == zk.ErrNoNode {
			return nil, ErrNodeNotExist
		}
		logger.Logger.Errorf("zk.Children(\"%s\") error(%v)", path, err)
		return nil, err
	}
	if stat == nil {
		return nil, ErrNodeNotExist
	}
	if len(nodes) == 0 {
		return nil, ErrNoChild
	}
	return nodes, nil
}
