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

	replicated int32
}

func newVariable(name string) *variable {
	v := &variable{
		name:        name,
		selfItem:    &variableItem{},
		remoteItems: make(map[string]*variableItem),
	}

	return v
}

func (v *variable) isReplicated() bool {
	return atomic.LoadInt32(&v.replicated) == 1
}

func (v *variable) replicatedOn() {
	atomic.StoreInt32(&v.replicated, 1)
}

func (v *variable) replicatedOff() {
	atomic.StoreInt32(&v.replicated, 0)
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

func (v *variable) updateRemoteNode(nodeID string, value, stamp int64) {
	v.remoteItemsMx.Lock()
	defer v.remoteItemsMx.Unlock()

	i, ok := v.remoteItems[nodeID]
	if !ok {
		i = &variableItem{}
		v.remoteItems[nodeID] = i
	}

	if i.stamp() < stamp {
		i.setValue(value)
		i.setStamp(stamp)
	}
}
