package rplx

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"sync"
	"time"
)

var (
	// ErrNodeAlreadyExists describe error for add already exists node
	ErrNodeAlreadyExists = errors.New("node already exists")

	defaultSendVariableToNodeTimeout   = time.Millisecond * 500
	defaultSendToReplicationTimeout    = time.Millisecond * 500
	defaultGCInterval                  = time.Second * 60
	defaultLogger                      = zap.NewNop()
	defaultReplicationChanCap          = 10240
	defaultNodeMaxBufferSize           = 1024
	defaultByTickerReplicationInterval = time.Minute
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

	runByTickerReplication      bool
	byTickerReplicationInterval time.Duration
}

// New creates new Rplx
func New(nodeID string, opts ...Option) *Rplx {
	r := &Rplx{
		nodeID:                      nodeID,
		logger:                      defaultLogger,
		variables:                   make(map[string]*variable),
		replicationChan:             make(chan *variable, defaultReplicationChanCap),
		nodes:                       make(map[string]*node),
		gcInterval:                  defaultGCInterval,
		nodeMaxBufferSize:           defaultNodeMaxBufferSize,
		runByTickerReplication:      false,
		byTickerReplicationInterval: defaultByTickerReplicationInterval,
	}

	for _, o := range opts {
		o(r)
	}

	if r.runByTickerReplication {
		go r.byTickerReplication()
	}

	go r.listenReplicationChannel()
	go r.startGC()

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

// AddNode adds and serve new remote node
func (rplx *Rplx) AddNode(nodeID string, addr string, syncInterval time.Duration, opts grpc.DialOption) error {
	rplx.nodesMx.Lock()
	defer rplx.nodesMx.Unlock()

	_, ok := rplx.nodes[nodeID]
	if ok {
		return ErrNodeAlreadyExists
	}

	n, err := newNode(rplx.nodeID, nodeID, addr, syncInterval, rplx.logger, rplx.nodeMaxBufferSize, opts)
	if err != nil {
		return errors.Wrap(err, "error create node")
	}

	go n.serve()

	rplx.nodes[nodeID] = n

	return nil
}

// sendToReplication send variable to replication channel
// if reached timeout defaultSendToReplicationTimeout while send, log error
func (rplx *Rplx) sendToReplication(v *variable) {
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
			if !v.isReplicated(node.ID) {
				go rplx.sendVariableToNode(v, nodeID, node)
			}
		}
		rplx.nodesMx.RUnlock()
	}
}

func (rplx *Rplx) byTickerReplication() {
	t := time.NewTicker(rplx.byTickerReplicationInterval)

	for range t.C {
		rplx.variablesMx.RLock()
		for _, v := range rplx.variables {
			if v.isReplicatedAll() {
				continue
			}
			go rplx.sendToReplication(v)
		}

		rplx.variablesMx.RUnlock()
	}
}

func (rplx *Rplx) sendVariableToNode(v *variable, nodeID string, n *node) {
	select {
	case n.replicationChan <- v:
	case <-time.After(defaultSendVariableToNodeTimeout):
		rplx.logger.Error("error send variable to node replication channel", zap.String("remote node ID", nodeID))
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
	vars := make([]string, 0)

	now := time.Now().UTC().UnixNano()

	rplx.variablesMx.RLock()

	currentLen := len(rplx.variables)

	for name, v := range rplx.variables {
		if v.ttl > 0 && v.ttl < now {
			vars = append(vars, name)
		}
	}
	rplx.variablesMx.RUnlock()

	rplx.variablesMx.Lock()
	for _, name := range vars {
		delete(rplx.variables, name)
	}
	rplx.variablesMx.Unlock()

	rplx.logger.Debug("GC done", zap.Int("deleted", len(vars)), zap.Int("init count", currentLen))
}
