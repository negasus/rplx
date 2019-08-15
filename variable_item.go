package rplx

import (
	"sync/atomic"
	"time"
)

type variableItem struct {
	val int64
	ver int64
}

func newVariableItem() *variableItem {
	return &variableItem{}
}

func (item *variableItem) update(delta int64) int64 {
	atomic.StoreInt64(&item.ver, time.Now().UTC().UnixNano())
	return atomic.AddInt64(&item.val, delta)
}

func (item *variableItem) set(value, version int64) {
	atomic.StoreInt64(&item.val, value)
	atomic.StoreInt64(&item.ver, version)
}

func (item *variableItem) value() int64 {
	return atomic.LoadInt64(&item.val)
}

func (item *variableItem) version() int64 {
	return atomic.LoadInt64(&item.ver)
}
