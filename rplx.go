package rplx

import (
	"context"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	defaultGCInterval               = time.Second * 60
	defaultLogger                   = zap.NewNop()
	defaultReplicationChanCap       = 1024 * 100
	defaultRemoteNodesCheckInterval = time.Minute
)

// RemoteNodesProvider is type for function, called automatically and returns info about remote nodes
type RemoteNodesProvider func() []*RemoteNodeOption

// Rplx describe main Rplx object
type Rplx struct {
	nodeID string
	logger *zap.Logger

	replicationChan chan *variable

	nodesMx       sync.RWMutex
	nodes         map[string]*node
	nodesIDToAddr map[string]string

	variablesMx sync.RWMutex
	variables   map[string]*variable

	gcInterval time.Duration

	remoteNodesTicker        *time.Ticker
	remoteNodesProvider      RemoteNodesProvider
	remoteNodesCheckInterval time.Duration
	grpcServer               *grpc.Server

	gcTicker *time.Ticker

	readOnly int32

	withMetrics bool
	metrics     *metrics
}

// New creates new Rplx
func New(opts ...Option) *Rplx {
	r := &Rplx{
		logger:                   defaultLogger,
		variables:                make(map[string]*variable),
		replicationChan:          make(chan *variable, defaultReplicationChanCap),
		nodes:                    make(map[string]*node),
		nodesIDToAddr:            make(map[string]string),
		gcInterval:               defaultGCInterval,
		remoteNodesCheckInterval: defaultRemoteNodesCheckInterval,
	}

	// apply options
	for _, o := range opts {
		o(r)
	}

	r.metrics = newMetrics()
	if r.withMetrics {
		r.metrics.register()
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

// Stop Rplx
func (rplx *Rplx) Stop() {
	atomic.StoreInt32(&rplx.readOnly, 1)

	close(rplx.replicationChan)

	if rplx.grpcServer != nil {
		rplx.grpcServer.GracefulStop()
	}

	if rplx.remoteNodesTicker != nil {
		rplx.remoteNodesTicker.Stop()
	}

	if rplx.gcTicker != nil {
		rplx.gcTicker.Stop()
	}

	for _, n := range rplx.nodes {
		n.Stop()
	}
}

// StartReplicationServer starts grpc server for receive sync messages from remote nodes
func (rplx *Rplx) StartReplicationServer(ln net.Listener, grpcOptions ...grpc.ServerOption) error {
	if rplx.withMetrics {
		grpcOptions = append(grpcOptions, grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor))
		grpcOptions = append(grpcOptions, grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor))
	}

	rplx.grpcServer = grpc.NewServer(grpcOptions...)

	RegisterReplicatorServer(rplx.grpcServer, rplx)
	reflection.Register(rplx.grpcServer)

	if rplx.withMetrics {
		grpc_prometheus.Register(rplx.grpcServer)
	}

	rplx.logger.Debug("start grpc server", zap.String("address", ln.Addr().String()))

	return rplx.grpcServer.Serve(ln)
}

// Hello is implementation grpc method for get Hello request
func (rplx *Rplx) Hello(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
	return &HelloResponse{ID: rplx.nodeID}, nil
}

func (rplx *Rplx) startRemoteNodesListener() {
	rplx.remoteNodesTicker = time.NewTicker(rplx.remoteNodesCheckInterval)

	for range rplx.remoteNodesTicker.C {
		nodesOptions := rplx.remoteNodesProvider()

		newNodesAddresses := make(map[string]struct{})

		for _, nodeOption := range nodesOptions {
			newNodesAddresses[nodeOption.Addr] = struct{}{}

			rplx.nodesMx.RLock()
			_, ok := rplx.nodes[nodeOption.Addr]
			rplx.nodesMx.RUnlock()

			if ok {
				continue
			}

			rplx.logger.Info("add remote node to rplx", zap.String("addr", nodeOption.Addr))

			rplx.nodesMx.Lock()
			rplx.nodes[nodeOption.Addr] = newNode(nodeOption, rplx.nodeID, rplx.logger, rplx.metrics)
			go rplx.nodes[nodeOption.Addr].connect(nodeOption.DialOpts, rplx)
			rplx.nodesMx.Unlock()
		}

		// if exists nodes not contains in new list, stop and remove it
		rplx.nodesMx.Lock()
		for addr, node := range rplx.nodes {
			if _, ok := newNodesAddresses[addr]; !ok {
				rplx.logger.Info("stop and remove remote node", zap.String("id", node.remoteNodeID), zap.String("addr", addr))
				node.Stop()
				delete(rplx.nodesIDToAddr, node.remoteNodeID)
				delete(rplx.nodes, addr)
			}
		}
		rplx.nodesMx.Unlock()
	}
}

// sendToReplication send variable to replication channel
// if reached timeout defaultSendToReplicationTimeout while send, log error
func (rplx *Rplx) sendToReplication(v *variable) {
	if atomic.LoadInt32(&rplx.readOnly) == 1 {
		return
	}

	select {
	case rplx.replicationChan <- v:
	default:
		rplx.logger.Error("error send variable to replication channel")
	}
}

// listenReplicationChannel starts loop for listen replication channel and send to each node
func (rplx *Rplx) listenReplicationChannel() {
	for v := range rplx.replicationChan {
		rplx.nodesMx.RLock()
		for _, node := range rplx.nodes {
			select {
			case node.replicationChan <- v:
			default:
				rplx.logger.Error("error send variable to node replication channel", zap.String("remote node id", node.remoteNodeID))
			}
		}
		rplx.nodesMx.RUnlock()
	}
}

// startGC start GC loop
func (rplx *Rplx) startGC() {
	rplx.logger.Debug("start GC loop", zap.Duration("interval", rplx.gcInterval))

	rplx.gcTicker = time.NewTicker(rplx.gcInterval)

	for range rplx.gcTicker.C {
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
