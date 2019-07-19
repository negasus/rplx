package rplx

import (
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"sync"
	"time"
)

// replication channel size
var nodeChSize = 1024

var defaultMaxBufferSize = 1024

// node describe remote node
type node struct {

	// local node ID
	localNodeID string

	// remote node ID
	ID string

	conn             *grpc.ClientConn
	replicatorClient ReplicatorClient
	logger           *zap.Logger

	replicationChan chan *variable

	bufferMx      sync.RWMutex
	buffer        map[string]*variable
	MaxBufferSize int

	syncMx sync.Mutex

	syncInterval time.Duration
}

func newNode(localNodeID, remoteNodeID string, addr string, syncInterval time.Duration, logger *zap.Logger, opts grpc.DialOption) (*node, error) {
	var err error

	n := &node{
		localNodeID:     localNodeID,
		ID:              remoteNodeID,
		logger:          logger,
		replicationChan: make(chan *variable, nodeChSize),
		buffer:          make(map[string]*variable),
		syncInterval:    syncInterval,
		MaxBufferSize:   defaultMaxBufferSize,
	}

	n.conn, err = grpc.Dial(addr, opts)
	if err != nil {
		return nil, err
	}

	n.replicatorClient = NewReplicatorClient(n.conn)

	return n, nil
}

func (n *node) serve() {
	t := time.NewTicker(n.syncInterval)

	for {
		select {
		case <-t.C:
			n.logger.Debug("start sync for node by ticker", zap.String("nodeID", n.ID))

			if err := n.sync(); err != nil {
				n.logger.Warn("error sync", zap.Error(err))
			}
		case v := <-n.replicationChan:
			n.logger.Debug("add var to buffer for remote node", zap.String("nodeID", n.ID), zap.String("var name", v.name))

			n.bufferMx.Lock()
			n.buffer[v.name] = v
			l := len(n.buffer)
			n.bufferMx.Unlock()

			if l > n.MaxBufferSize {
				n.logger.Debug("start sync for node by buffer size", zap.String("nodeID", n.ID))

				if err := n.sync(); err != nil {
					n.logger.Error("error sync", zap.Error(err))
				}
			}
		}
	}
}
