package rplx

import (
	"sync"
	"sync/atomic"
	"time"
)

type variable struct {
	name string

	// variable value for current node
	self *variableItem

	ttl        int64
	ttlVersion int64

	// variable values for remote nodes
	// map key - is remove node remoteNodeID
	remoteItemsMx sync.RWMutex
	remoteItems   map[string]*variableItem
}

func newVariable(name string) *variable {
	v := &variable{
		name:        name,
		self:        newVariableItem(),
		remoteItems: make(map[string]*variableItem),
	}

	return v
}

// get returns variable value
func (v *variable) get() int64 {
	result := v.self.value()

	v.remoteItemsMx.RLock()
	for _, item := range v.remoteItems {
		result += item.value()
	}
	v.remoteItemsMx.RUnlock()

	return result
}

func (v *variable) TTL() int64 {
	return atomic.LoadInt64(&v.ttl)
}

func (v *variable) TTLVersion() int64 {
	return atomic.LoadInt64(&v.ttlVersion)
}

func (v *variable) update(delta int64) int64 {
	return v.self.update(delta)
}

func (v *variable) updateTTL(ttl int64) {
	atomic.StoreInt64(&v.ttl, ttl)
	atomic.StoreInt64(&v.ttlVersion, time.Now().UTC().UnixNano())
	v.self.update(0) // обновляем текущее значение на 0, чтобы обновилась версия переменной и она ушла на репликацию
}

// updateItem updates value for selected node and returns flag: updated or not
func (v *variable) updateItem(nodeID string, value, version int64) bool {
	v.remoteItemsMx.Lock()
	defer v.remoteItemsMx.Unlock()

	updated := false

	i, ok := v.remoteItems[nodeID]
	if !ok {
		i = newVariableItem()
		v.remoteItems[nodeID] = i
		updated = true
	}

	if i.version() < version {
		i.set(value, version)
		updated = true
	}

	return updated
}
