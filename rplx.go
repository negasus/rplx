package rplx

import (
	"context"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"sync"
	"time"
)

var (
	// ErrNodeAlreadyExists returns from AddRemoteNode if node already exists
	ErrNodeAlreadyExists = errors.New("node already exists")

	defaultSendVariableToNodeTimeout = time.Millisecond * 500
	defaultSendToReplicationTimeout  = time.Millisecond * 500
	defaultGCInterval                = time.Second * 60
	defaultLogger                    = zap.NewNop()
	defaultReplicationChanCap        = 10240
	defaultNodeMaxBufferSize         = 1024
)

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

	readOnly bool
}

// New creates new Rplx
func New(opts ...Option) *Rplx {
	r := &Rplx{
		logger:            defaultLogger,
		variables:         make(map[string]*variable),
		replicationChan:   make(chan *variable, defaultReplicationChanCap),
		nodes:             make(map[string]*node),
		gcInterval:        defaultGCInterval,
		nodeMaxBufferSize: defaultNodeMaxBufferSize,
	}

	r.nodeID = uuid.New().String()

	// apply options
	for _, o := range opts {
		o(r)
	}

	go r.listenReplicationChannel()
	go r.startGC()

	r.logger.Debug("start rplx", zap.String("remoteNodeID", r.nodeID))

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

// AddRemoteNode adds and listenReplicationChannel new remote node
func (rplx *Rplx) AddRemoteNode(addr string, syncInterval time.Duration, opts grpc.DialOption) error {
	rplx.nodesMx.Lock()
	defer rplx.nodesMx.Unlock()

	if _, ok := rplx.nodes[addr]; ok {
		return ErrNodeAlreadyExists
	}

	n, err := newNode(rplx.nodeID, addr, opts, syncInterval, rplx.nodeMaxBufferSize, rplx.logger)
	if err != nil {
		return err
	}
	rplx.nodes[n.remoteNodeID] = n

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
		for nodeID, node := range rplx.nodes {
			go rplx.sendVariableToNode(v, nodeID, node)
		}
		rplx.nodesMx.RUnlock()
	}
}

func (rplx *Rplx) sendVariableToNode(v *variable, nodeID string, n *node) {
	select {
	case n.replicationChan <- v:
	case <-time.After(defaultSendVariableToNodeTimeout):
		rplx.logger.Error("error send variable to node replication channel", zap.String("remote node remoteNodeID", nodeID))
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
		delete(rplx.variables, name)
	}
	rplx.variablesMx.Unlock()

	if len(namesToDelete) > 0 {
		rplx.logger.Debug("gc collect variables", zap.Int("count", len(namesToDelete)), zap.Strings("names", namesToDelete))
	}
}
