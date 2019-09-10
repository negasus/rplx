package rplx

import (
	"go.uber.org/zap"
	"time"
)

// Option describe Rplx option
type Option func(*Rplx)

// WithNodeID option for specify rplx node name
func WithNodeID(nodeID string) Option {
	return func(rplx *Rplx) {
		rplx.nodeID = nodeID
	}
}

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

// WithReadOnly option sets read only mode
func WithReadOnly() Option {
	return func(rplx *Rplx) {
		rplx.readOnly = 1
	}
}

// WithRemoteNodesProvider option
func WithRemoteNodesProvider(provider RemoteNodesProvider) Option {
	return func(rplx *Rplx) {
		rplx.remoteNodesProvider = provider
	}
}

// WithRemoteNodesCheckInterval option
func WithRemoteNodesCheckInterval(interval time.Duration) Option {
	return func(rplx *Rplx) {
		rplx.remoteNodesCheckInterval = interval
	}
}
