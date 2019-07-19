package rplx

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrVariableExpired = errors.New("variable expired")
)

type variable struct {
	name string

	// variable value for current node
	selfItem *variableItem

	ttl int64

	// variable values for remote nodes
	// map key - is remove node ID
	remoteItemsMx sync.RWMutex
	remoteItems   map[string]*variableItem
}

func newVariable(name string) *variable {
	v := &variable{
		name:        name,
		selfItem:    &variableItem{},
		remoteItems: make(map[string]*variableItem),
	}

	return v
}

// get returns variable value or error, if variable is expired
func (v *variable) get() (int64, error) {
	ttl := v.getTTL()
	if ttl > 0 && ttl < time.Now().UTC().UnixNano() {
		return 0, ErrVariableExpired
	}

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

func (v *variable) update(delta int64) int64 {
	return v.selfItem.update(delta)
}

func (v *variable) updateTTL(ttl time.Time) {
	atomic.StoreInt64(&v.ttl, ttl.UnixNano())
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
