package rplx

import (
	"context"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
	"time"
)

// replication channel size
var nodeChSize = 1024

const (
	defaultSendHelloRequestInterval = time.Second * 5
)

// node describe remote node
type node struct {
	syncing int32

	localNodeID  string
	remoteNodeID string

	replicatorClient ReplicatorClient
	logger           *zap.Logger

	replicationChan chan *variable

	bufferMx      sync.RWMutex
	buffer        map[string]*variable
	maxBufferSize int

	// replicatedVersions contains data for replicated variableItems versions
	// map key format: <VARIABLE_NAME>@<REMOTE_NODE_ID>
	// map value: last replicated variableItem version
	replicatedVersionsMx sync.RWMutex
	replicatedVersions   map[string]int64

	syncInterval time.Duration

	stopConnecting chan struct{}
}

//func newNode(localNodeID string, addr string, opts grpc.DialOption, syncInterval time.Duration, maxBufferSize int, logger *zap.Logger) (*node, error) {
//
//	n := &node{
//		localNodeID:        localNodeID,
//		logger:             logger,
//		replicationChan:    make(chan *variable, nodeChSize),
//		buffer:             make(map[string]*variable),
//		syncInterval:       syncInterval,
//		maxBufferSize:      maxBufferSize,
//		replicatedVersions: make(map[string]int64),
//		stopConnecting:     make(chan struct{}),
//	}
//
//	n.replicatorClient = NewReplicatorClient(conn)
//
//	return n, nil
//}

func (n *node) Stop() {
	close(n.stopConnecting)
}

func (n *node) Connect(addr string, r *Rplx) {
	t := time.NewTicker(defaultSendHelloRequestInterval)
	var connected bool

	for {
		select {
		case <-t.C:
			hello, err := n.replicatorClient.Hello(context.Background(), &HelloRequest{})
			if err == nil {
				n.remoteNodeID = hello.ID
				connected = true
				break
			}
			n.logger.Warn("error send hello request", zap.Error(err))
		case <-n.stopConnecting:
			return
		}
		if connected {
			break
		}
	}

	n.logger.Debug("connect to remote node", zap.String("addr", addr), zap.String("remote node ID", n.remoteNodeID))

	r.nodesMx.Lock()
	defer r.nodesMx.Unlock()

	if _, ok := r.nodes[n.remoteNodeID]; ok {
		n.Stop()
		n.logger.Error("node already exists", zap.String("addr", addr))
		return
	}

	r.variablesMx.RLock()

	for name, v := range r.variables {
		n.buffer[name] = v
	}
	r.nodes[n.remoteNodeID] = n

	n.logger.Debug("add variables for sync", zap.Int("count", len(n.buffer)), zap.Any("vars", r.variables))

	r.variablesMx.RUnlock()

	go n.sync()

	go n.syncByTicker()
	go n.listenReplicationChannel()
}

func (n *node) syncByTicker() {
	t := time.NewTicker(n.syncInterval)

	for range t.C {
		//n.logger.Debug("start sync for node by ticker", zap.String("remote node ID", n.remoteNodeID))
		n.sync()
	}
}

func (n *node) listenReplicationChannel() {
	for v := range n.replicationChan {
		n.bufferMx.Lock()
		n.buffer[v.name] = v
		l := len(n.buffer)
		n.bufferMx.Unlock()

		if l > n.maxBufferSize {
			if atomic.LoadInt32(&n.syncing) == 0 {
				go n.sync()
			}
		}
	}
}
