package rplx

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultRemoteNodeConnectionInterval = time.Second * 5 // interval for connect to remote node
	defaultRemoteNodeSyncInterval       = time.Second * 2 // interval for sync with remote node
	defaultRemoteNodeMaxBufferSize      = 1024            // by reach this limit, inits sync process
	defaultRemoteNodeWaitSyncCount      = 5               // count sync tasks in queue, while current sync in progress
)

var (
	nodeChSize = 1024 * 1024
)

// node describe remote node
type node struct {
	connected int32
	syncing   int32

	addr         string
	localNodeID  string
	remoteNodeID string

	replicatorClient ReplicatorClient

	replicationChan chan *variable

	bufferMx      sync.RWMutex
	buffer        map[string]*variable
	maxBufferSize int

	// replicatedVersions contains data for replicated variableItems versions
	// map key format: <VARIABLE_NAME>@<REMOTE_NODE_ID>
	// map value: last replicated variableItem version
	replicatedVersionsMx sync.RWMutex
	replicatedVersions   map[string]int64

	syncInterval       time.Duration
	connectionInterval time.Duration

	conn *grpc.ClientConn

	syncQueue chan struct{}

	stopChan chan struct{}

	logger *zap.Logger
}

// RemoteNodeOption describe options for RemoteNode, returns from RemoteNodeProvider
type RemoteNodeOption struct {
	Addr               string
	DialOpts           grpc.DialOption
	SyncInterval       time.Duration
	MaxBufferSize      int
	ConnectionInterval time.Duration
	WaitSyncCount      int
}

// DefaultRemoteNodeOption returns default remoteNodeOption with provided address
func DefaultRemoteNodeOption(addr string) *RemoteNodeOption {
	option := &RemoteNodeOption{
		Addr:               addr,
		DialOpts:           grpc.WithInsecure(),
		SyncInterval:       defaultRemoteNodeSyncInterval,
		MaxBufferSize:      defaultRemoteNodeMaxBufferSize,
		ConnectionInterval: defaultRemoteNodeConnectionInterval,
		WaitSyncCount:      defaultRemoteNodeWaitSyncCount,
	}

	return option
}

func newNode(options *RemoteNodeOption, localNodeID string, logger *zap.Logger) *node {
	n := &node{
		addr:               options.Addr,
		localNodeID:        localNodeID,
		replicationChan:    make(chan *variable, nodeChSize),
		buffer:             make(map[string]*variable),
		maxBufferSize:      options.MaxBufferSize,
		replicatedVersions: make(map[string]int64),
		syncInterval:       options.SyncInterval,
		connectionInterval: options.ConnectionInterval,
		syncQueue:          make(chan struct{}, options.WaitSyncCount),
		stopChan:           make(chan struct{}),
		logger:             logger,
	}

	go n.listenSyncQueue()
	go n.listenReplicationChannel()
	go n.syncByTicker()

	return n
}

func (n *node) Stop() {
	close(n.syncQueue)
	close(n.replicationChan)
	close(n.stopChan)
	if err := n.conn.Close(); err != nil {
		n.logger.Warn("error close grpc connection", zap.Error(err), zap.String("remote node addr", n.addr))
	}
}

func (n *node) dial(dialOpts grpc.DialOption) error {
	var err error

	if n.replicatorClient != nil {
		return nil
	}

	n.conn, err = grpc.Dial(n.addr, dialOpts)
	if err != nil {
		return err
	}

	n.replicatorClient = NewReplicatorClient(n.conn)
	return nil
}

func (n *node) connect(dialOpts grpc.DialOption, rplx *Rplx) {
	t := time.NewTicker(n.connectionInterval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if err := n.dial(dialOpts); err != nil {
				n.logger.Error("error dial to remote node", zap.String("addr", n.addr), zap.Error(err))
				continue
			}

			hello, err := n.replicatorClient.Hello(context.Background(), &HelloRequest{})
			if err != nil {
				n.logger.Warn("error send hello request to remote node", zap.String("addr", n.addr), zap.Error(err))
				continue
			}

			n.remoteNodeID = hello.ID

			n.logger.Debug("connected to remote node", zap.String("addr", n.addr), zap.String("remote node ID", n.remoteNodeID))

			// send all current variables to replication for new connected node
			rplx.variablesMx.RLock()
			for _, v := range rplx.variables {
				n.replicationChan <- v
			}
			rplx.variablesMx.RUnlock()

			rplx.nodesMx.Lock()
			rplx.nodesIDToAddr[n.remoteNodeID] = n.addr
			rplx.nodesMx.Unlock()

			atomic.StoreInt32(&n.connected, 1)

			go n.sync()

			return
		case <-n.stopChan:
			return
		}
	}
}

func (n *node) syncByTicker() {
	t := time.NewTicker(n.syncInterval)

	for {
		select {
		case <-t.C:
			n.sync()
		case <-n.stopChan:
			t.Stop()
			return
		}
	}
}

func (n *node) listenReplicationChannel() {
	for v := range n.replicationChan {
		n.bufferMx.Lock()
		n.buffer[v.name] = v
		l := len(n.buffer)
		n.bufferMx.Unlock()

		if l > n.maxBufferSize {
			go n.sync()
		}
	}
}
