package rplx

import (
	"sync/atomic"
	"time"
)

// variableItem stores data about variable for single node
// variableItem is concurrent safe
type variableItem struct {
	v int64 // value
	s int64 // last update stamp in UnixNano (UTC)
}

func (item *variableItem) update(delta int64) int64 {
	atomic.StoreInt64(&item.s, time.Now().UTC().UnixNano())
	return atomic.AddInt64(&item.v, delta)
}

func (item *variableItem) setStamp(stamp int64) {
	atomic.StoreInt64(&item.s, stamp)
}

func (item *variableItem) setValue(value int64) {
	atomic.StoreInt64(&item.v, value)
}

func (item *variableItem) value() int64 {
	return atomic.LoadInt64(&item.v)
}

func (item *variableItem) stamp() int64 {
	return atomic.LoadInt64(&item.s)
}
