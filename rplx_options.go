package rplx

import (
	"go.uber.org/zap"
	"time"
)

// Option describe Rplx option
type Option func(*Rplx)

// WithLogger option for specify logger
func WithLogger(logger *zap.Logger) Option {
	return func(rplx *Rplx) {
		rplx.logger = logger
	}
}

// WithGCInterval option for set garbage collect interval
func WithGCInterval(interval time.Duration) Option {
	return func(rplx *Rplx) {
		rplx.gcInterval = interval
	}
}

// WithReplicationChanCap option for set replication channel capacity
func WithReplicationChanCap(c int) Option {
	return func(rplx *Rplx) {
		rplx.replicationChan = make(chan *variable, c)
	}
}

// WithNodeMaxBufferSize option for set node max buffer size
func WithNodeMaxBufferSize(s int) Option {
	return func(rplx *Rplx) {
		rplx.nodeMaxBufferSize = s
	}
}

// WithByTickerReplication option for run background by ticker replication
func WithByTickerReplication() Option {
	return func(rplx *Rplx) {
		rplx.runByTickerReplication = true
	}
}

// WithByTickerReplicationInterval option for set by ticker replication interval
func WithByTickerReplicationInterval(i time.Duration) Option {
	return func(rplx *Rplx) {
		rplx.byTickerReplicationInterval = i
	}
}

// WithReadOnly option sets read only mode
func WithReadOnly() Option {
	return func(rplx *Rplx) {
		rplx.readOnly = true
	}
}
