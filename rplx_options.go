package rplx

import (
	"go.uber.org/zap"
	"time"
)

type Option func(*Rplx)

func WithLogger(logger *zap.Logger) Option {
	return func(rplx *Rplx) {
		rplx.logger = logger
	}
}

func WithGCInterval(interval time.Duration) Option {
	return func(rplx *Rplx) {
		rplx.gcInterval = interval
	}
}

func WithReplicationChanCap(c int) Option {
	return func(rplx *Rplx) {
		rplx.replicationChan = make(chan *variable, c)
	}
}

func WithNodeMaxBufferSize(s int) Option {
	return func(rplx *Rplx) {
		rplx.nodeMaxBufferSize = s
	}
}
