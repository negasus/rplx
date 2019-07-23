package rplx

import (
	"sync"
	"sync/atomic"
	"time"
)

type variable struct {
	name string

	// variable value for current node
	selfItem *variableItem

	ttl      int64
	ttlStamp int64

	// variable values for remote nodes
	// map key - is remove node ID
	remoteItemsMx sync.RWMutex
	remoteItems   map[string]*variableItem
}

func newVariable(name string) *variable {
	v := &variable{
		name:        name,
		selfItem:    newVariableItem(),
		remoteItems: make(map[string]*variableItem),
	}

	return v
}

// isReplicatedAll returns true, if all nodes are replicated
func (v *variable) isReplicatedAll() bool {
	if !v.selfItem.isReplicated() {
		return false
	}

	v.remoteItemsMx.RLock()
	defer v.remoteItemsMx.RUnlock()
	for _, i := range v.remoteItems {
		if !i.isReplicated() {
			return false
		}
	}

	return true
}

func (v *variable) isReplicated(nodeID string) bool {
	if nodeID == "" {
		return v.selfItem.isReplicated()
	}

	v.remoteItemsMx.RLock()
	defer v.remoteItemsMx.RUnlock()

	i, ok := v.remoteItems[nodeID]
	if !ok {
		return false
	}

	return i.isReplicated()
}

func (v *variable) updateReplicationStamp(nodeID string) {
	if nodeID == "" {
		v.selfItem.updateReplicationStamp()
		return
	}

	v.remoteItemsMx.RLock()
	defer v.remoteItemsMx.RUnlock()

	i, ok := v.remoteItems[nodeID]
	if !ok {
		return
	}

	i.updateReplicationStamp()
}

// get returns variable value
func (v *variable) get() (int64, error) {
	result := v.selfItem.value()

	v.remoteItemsMx.RLock()
	for _, item := range v.remoteItems {
		result += item.value()
	}
	v.remoteItemsMx.RUnlock()

	return result, nil
}

// items returns copy of v.remoteItems
func (v *variable) items() map[string]*variableItem {
	v.remoteItemsMx.Lock()
	defer v.remoteItemsMx.Unlock()

	r := make(map[string]*variableItem, len(v.remoteItems))

	for nodeID, i := range v.remoteItems {
		r[nodeID] = &variableItem{v: i.v, s: i.s}
	}

	return r
}

func (v *variable) getTTL() int64 {
	return atomic.LoadInt64(&v.ttl)
}

func (v *variable) getTTLStamp() int64 {
	return atomic.LoadInt64(&v.ttlStamp)
}

func (v *variable) update(delta int64) int64 {
	return v.selfItem.update(delta)
}

func (v *variable) updateTTL(ttl time.Time) {
	atomic.StoreInt64(&v.ttl, ttl.UnixNano())
	atomic.StoreInt64(&v.ttlStamp, time.Now().UTC().UnixNano())
}

// updateRemoteNode updates value for selected node and returns flag: updated or not
func (v *variable) updateRemoteNode(nodeID string, value, stamp int64) bool {
	v.remoteItemsMx.Lock()
	defer v.remoteItemsMx.Unlock()

	updated := false

	i, ok := v.remoteItems[nodeID]
	if !ok {
		i = newVariableItem()
		v.remoteItems[nodeID] = i
		updated = true
	}

	if i.stamp() < stamp {
		i.setValue(value)
		i.setStamp(stamp)
		updated = true
	}

	return updated
}
