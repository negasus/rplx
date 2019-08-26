package rplx

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"sync"
	"time"
)

var (
	defaultSendVariableToNodeTimeout = time.Millisecond * 500
	defaultSendToReplicationTimeout  = time.Millisecond * 500
	defaultGCInterval                = time.Second * 60
	defaultLogger                    = zap.NewNop()
	defaultReplicationChanCap        = 10240
	defaultNodeMaxBufferSize         = 1024
	defaultNodeMaxDeferSync          = 5
	defaultNodeDeferSyncTimeout      = time.Second
	defaultRemoteNodesCheckInterval  = time.Minute
)

type RemoteNodeOption struct {
	Addr         string
	SyncInterval time.Duration
	DialOpts     grpc.DialOption
}

type RemoteNodesProvider func() []*RemoteNodeOption

// Rplx describe main Rplx object
type Rplx struct {
	nodeID string
	logger *zap.Logger

	replicationChan chan *variable

	nodesMx sync.RWMutex
	nodes   map[string]*node

	variablesMx sync.RWMutex
	variables   map[string]*variable

	gcInterval        time.Duration
	nodeMaxBufferSize int

	remoteNodesCurrentList   map[string]struct{}
	remoteNodesProvider      RemoteNodesProvider
	remoteNodesCheckInterval time.Duration

	readOnly bool
}

// New creates new Rplx
func New(opts ...Option) *Rplx {
	r := &Rplx{
		logger:                   defaultLogger,
		variables:                make(map[string]*variable),
		replicationChan:          make(chan *variable, defaultReplicationChanCap),
		nodes:                    make(map[string]*node),
		gcInterval:               defaultGCInterval,
		nodeMaxBufferSize:        defaultNodeMaxBufferSize,
		remoteNodesCheckInterval: defaultRemoteNodesCheckInterval,
		remoteNodesCurrentList:   make(map[string]struct{}),
	}

	// apply options
	for _, o := range opts {
		o(r)
	}

	if r.nodeID == "" {
		r.nodeID = uuid.New().String()
	}

	go r.listenReplicationChannel()
	go r.startGC()

	if r.remoteNodesProvider != nil {
		go r.startRemoteNodesListener()
	}

	r.logger.Debug("rplx start", zap.String("local node id", r.nodeID))

	return r
}

// StartReplicationServer starts grpc server for receive sync messages from remote nodes
func (rplx *Rplx) StartReplicationServer(ln net.Listener) error {

	server := grpc.NewServer()

	RegisterReplicatorServer(server, rplx)
	reflection.Register(server)

	rplx.logger.Debug("start grpc server", zap.String("address", ln.Addr().String()))

	return server.Serve(ln)
}

// Hello is implementation grpc method for get Hello request
func (rplx *Rplx) Hello(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
	return &HelloResponse{ID: rplx.nodeID}, nil
}

func (rplx *Rplx) startRemoteNodesListener() {
	t := time.NewTicker(rplx.remoteNodesCheckInterval)

	for range t.C {
		nodes := rplx.remoteNodesProvider()

		for _, node := range nodes {
			if _, ok := rplx.remoteNodesCurrentList[node.Addr]; ok {
				continue
			}

			rplx.logger.Debug("add remote node", zap.String("addr", node.Addr))

			if err := rplx.addRemoteNode(node.Addr, node.SyncInterval, node.DialOpts); err != nil {
				rplx.logger.Error("error add remote node", zap.String("addr", node.Addr))
				continue
			}

			rplx.remoteNodesCurrentList[node.Addr] = struct{}{}
		}

		// todo: stop and remove nodes, not exists in new list 'nodes'
	}
}

// addRemoteNode adds and listenReplicationChannel new remote node
func (rplx *Rplx) addRemoteNode(addr string, syncInterval time.Duration, opts grpc.DialOption) error {

	conn, err := grpc.Dial(addr, opts)
	if err != nil {
		return err
	}

	n := &node{
		localNodeID:        rplx.nodeID,
		logger:             rplx.logger,
		replicationChan:    make(chan *variable, nodeChSize),
		buffer:             make(map[string]*variable),
		syncInterval:       syncInterval,
		maxBufferSize:      rplx.nodeMaxBufferSize,
		replicatedVersions: make(map[string]int64),
		stopChan:           make(chan struct{}),
		maxDeferSync:       defaultNodeMaxDeferSync,
		deferSyncTimeout:   defaultNodeDeferSyncTimeout,
	}

	n.replicatorClient = NewReplicatorClient(conn)

	go n.Connect(addr, rplx)

	return nil
}

// sendToReplication send variable to replication channel
// if reached timeout defaultSendToReplicationTimeout while send, log error
func (rplx *Rplx) sendToReplication(v *variable) {
	if rplx.readOnly {
		return
	}

	select {
	case rplx.replicationChan <- v:
	case <-time.After(defaultSendToReplicationTimeout):
		rplx.logger.Error("error send variable to replication channel", zap.String("variable name", v.name), zap.Float64("timeout, sec", defaultSendToReplicationTimeout.Seconds()))
	}
}

// listenReplicationChannel starts loop for listen replication channel and send to each node
func (rplx *Rplx) listenReplicationChannel() {
	for v := range rplx.replicationChan {
		rplx.nodesMx.RLock()
		for _, node := range rplx.nodes {
			go rplx.sendVariableToNode(v, node)
		}
		rplx.nodesMx.RUnlock()
	}
}

func (rplx *Rplx) sendVariableToNode(v *variable, n *node) {
	select {
	case n.replicationChan <- v:
	case <-time.After(defaultSendVariableToNodeTimeout):
		rplx.logger.Error("error send variable to node replication channel", zap.String("remote node remoteNodeID", n.remoteNodeID))
	}
}

// startGC start GC loop
func (rplx *Rplx) startGC() {
	rplx.logger.Debug("start GC loop")

	t := time.NewTicker(rplx.gcInterval)

	for range t.C {
		rplx.gc()
	}
}

// gc collects expired variables and remove it from rplx.variable map
func (rplx *Rplx) gc() {
	namesToDelete := make([]string, 0)

	now := time.Now().UTC().UnixNano()

	rplx.variablesMx.RLock()

	for name, v := range rplx.variables {
		if v.ttl > 0 && v.ttl < now {
			namesToDelete = append(namesToDelete, name)
		}
	}
	rplx.variablesMx.RUnlock()

	rplx.variablesMx.Lock()
	for _, name := range namesToDelete {
		v, ok := rplx.variables[name]
		if ok && v.ttl < time.Now().UTC().UnixNano() {
			delete(rplx.variables, name)
		}
	}
	rplx.variablesMx.Unlock()

	if len(namesToDelete) > 0 {
		rplx.logger.Debug("gc collect variables", zap.Int("count", len(namesToDelete)), zap.Strings("names", namesToDelete))
	}
}
