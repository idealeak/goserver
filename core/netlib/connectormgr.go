package netlib

import (
	"fmt"
	"sync"
)

var (
	ConnectorMgr = &connectorMgr{
		pool: make(map[string]Connector),
	}
)

type connectorMgr struct {
	pool map[string]Connector
	lock sync.Mutex
}

func (cm *connectorMgr) IsConnecting(sc *SessionConfig) bool {
	strKey := fmt.Sprintf("%v:%v", sc.Ip, sc.Port)
	cm.lock.Lock()
	defer cm.lock.Unlock()
	if _, exist := cm.pool[strKey]; exist {
		return true
	}
	return false
}

func (cm *connectorMgr) registeConnector(c Connector) {
	sc := c.GetSessionConfig()
	strKey := fmt.Sprintf("%v:%v", sc.Ip, sc.Port)
	cm.lock.Lock()
	defer cm.lock.Unlock()
	cm.pool[strKey] = c
}

func (cm *connectorMgr) unregisteConnector(c Connector) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	for k, v := range cm.pool {
		if v == c {
			delete(cm.pool, k)
			return
		}
	}
}
