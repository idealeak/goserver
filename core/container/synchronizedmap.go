package container

import (
	"sync"
)

type SynchronizedMap struct {
	lock *sync.RWMutex
	m    map[interface{}]interface{}
}

// NewSynchronizedMap return new SynchronizedMap
func NewSynchronizedMap() *SynchronizedMap {
	return &SynchronizedMap{
		lock: new(sync.RWMutex),
		m:    make(map[interface{}]interface{}),
	}
}

// Get from maps return the k's value
func (m *SynchronizedMap) Get(k interface{}) interface{} {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if val, ok := m.m[k]; ok {
		return val
	}
	return nil
}

// Maps the given key and value. Returns false
// if the key is already in the map and changes nothing.
func (m *SynchronizedMap) Set(k interface{}, v interface{}) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	if val, ok := m.m[k]; !ok {
		m.m[k] = v
	} else if val != v {
		m.m[k] = v
	} else {
		return false
	}
	return true
}

// Returns true if k is exist in the map.
func (m *SynchronizedMap) IsExist(k interface{}) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if _, ok := m.m[k]; !ok {
		return false
	}
	return true
}

// Delete the given key and value.
func (m *SynchronizedMap) Delete(k interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.m, k)
}

// Items returns all items in SynchronizedMap.
func (m *SynchronizedMap) Items() map[interface{}]interface{} {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.m
}
