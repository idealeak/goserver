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
	if val, ok := m.m[k]; ok {
		m.lock.RUnlock()
		return val
	}
	m.lock.RUnlock()
	return nil
}

// Maps the given key and value. Returns false
// if the key is already in the map and changes nothing.
func (m *SynchronizedMap) Set(k interface{}, v interface{}) bool {
	m.lock.Lock()
	if val, ok := m.m[k]; !ok {
		m.m[k] = v
	} else if val != v {
		m.m[k] = v
	} else {
		m.lock.Unlock()
		return false
	}
	m.lock.Unlock()
	return true
}

// Returns true if k is exist in the map.
func (m *SynchronizedMap) IsExist(k interface{}) bool {
	m.lock.RLock()
	if _, ok := m.m[k]; !ok {
		m.lock.RUnlock()
		return false
	}
	m.lock.RUnlock()
	return true
}

// Delete the given key and value.
func (m *SynchronizedMap) Delete(k interface{}) {
	m.lock.Lock()
	delete(m.m, k)
	m.lock.Unlock()
}

// Items returns all items in SynchronizedMap.
func (m *SynchronizedMap) Items() map[interface{}]interface{} {
	mm := make(map[interface{}]interface{})
	m.lock.RLock()
	for k, v := range m.m {
		mm[k] = v
	}
	m.lock.RUnlock()
	return mm
}

func (m *SynchronizedMap) Foreach(cb func(k, v interface{})) {
	m.lock.Lock()
	defer m.lock.Unlock()
	for k, v := range m.m {
		cb(k, v)
	}
}
