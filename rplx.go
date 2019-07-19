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

const replicationChannelCap = 1024

var ErrNodeAlreadyExists = errors.New("node already exists")

var defaultSendVariableToNodeTimeout = time.Millisecond * 500

var defaultSendToReplicationTimeout = time.Millisecond * 500

// Rplx describe main Rplx object
type Rplx struct {
	nodeID string
	logger *zap.Logger

	replicationChan chan *variable

	nodesMx sync.RWMutex
	nodes   map[string]*node

	variablesMx sync.RWMutex
	variables   map[string]*variable
}

// New creates new Rplx
func New(nodeID string, logger *zap.Logger) *Rplx {
	r := &Rplx{
		nodeID:          nodeID,
		logger:          logger,
		variables:       make(map[string]*variable),
		replicationChan: make(chan *variable, replicationChannelCap),
		nodes:           make(map[string]*node),
	}

	go r.listenReplicationChannel()

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

	n, err := newNode(rplx.nodeID, nodeID, addr, syncInterval, rplx.logger, opts)
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
			go rplx.sendVariableToNode(v, nodeID, node)
		}
		rplx.nodesMx.RUnlock()
	}
}

func (rplx *Rplx) sendVariableToNode(v *variable, nodeID string, n *node) {
	select {
	case n.replicationChan <- v:
	case <-time.After(defaultSendVariableToNodeTimeout):
		rplx.logger.Error("error send variable to node replication channel", zap.String("remote node ID", nodeID))
	}
}
