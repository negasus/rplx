package rplx

import (
	"context"
	"go.uber.org/zap"
	"sync/atomic"
	"time"
)

const (
	syncCodeSuccess = int64(0)
)

// sync sends replication data from buffer to remote node
func (n *node) sync() {
	if !atomic.CompareAndSwapInt32(&n.syncing, 0, 1) {
		n.logger.Debug("call node.sync fail, active sync", zap.String("remote node id", n.remoteNodeID))
		if n.maxDeferSync > 0 && int(atomic.LoadInt32(&n.deferSyncCounter)) < n.maxDeferSync {
			n.logger.Debug("place defer sync")
			time.AfterFunc(n.deferSyncTimeout, func() {
				atomic.AddInt32(&n.deferSyncCounter, 1)
				go n.sync()
			})
			return
		}
		n.logger.Debug("max defer sync reached, abort")
		return
	}
	defer atomic.StoreInt32(&n.syncing, 0)
	atomic.StoreInt32(&n.deferSyncCounter, 0)

	n.bufferMx.Lock()

	// if replication called by ticker, but buffer is empty - return
	if len(n.buffer) == 0 {
		n.bufferMx.Unlock()
		return
	}

	req := SyncRequest{
		NodeID:    n.localNodeID,
		Variables: make(map[string]*SyncVariable),
	}

	replicatedVersions := make(map[string]int64)

	n.replicatedVersionsMx.RLock()
	for name, v := range n.buffer {
		sv := &SyncVariable{
			TTL:         v.TTL(),
			TTLVersion:  v.TTLVersion(),
			NodesValues: make(map[string]*SyncNodeValue),
		}

		lastReplicatedVersion, ok := n.replicatedVersions[name+"@"+n.localNodeID]
		if !ok {
			lastReplicatedVersion = 0
		}

		if lastReplicatedVersion < v.self.version() {
			sv.NodesValues[n.localNodeID] = &SyncNodeValue{
				Value:   v.self.value(),
				Version: v.self.version(),
			}
			replicatedVersions[name+"@"+n.localNodeID] = v.self.version()
		}

		v.remoteItemsMx.RLock()
		for nodeID, item := range v.remoteItems {

			// Dont send to remote node its data
			if nodeID == n.remoteNodeID {
				continue
			}

			lastReplicatedVersion, ok := n.replicatedVersions[name+"@"+nodeID]
			if !ok {
				lastReplicatedVersion = -1
			}

			if lastReplicatedVersion < item.version() {
				sv.NodesValues[nodeID] = &SyncNodeValue{
					Value:   item.value(),
					Version: item.version(),
				}
				replicatedVersions[name+"@"+nodeID] = item.version()
			}
		}
		v.remoteItemsMx.RUnlock()

		if len(sv.NodesValues) > 0 {
			req.Variables[name] = sv
		}

		delete(n.buffer, name)
	}
	n.replicatedVersionsMx.RUnlock()
	n.bufferMx.Unlock()

	if len(req.Variables) == 0 {
		n.logger.Debug("call node.sync cancelled, empty variables", zap.String("remote node id", n.remoteNodeID))
		return
	}

	n.logger.Debug("send sync message", zap.String("remote node ID", n.remoteNodeID), zap.Int("variables", len(req.Variables)), zap.Any("vars map", req.Variables))

	r, err := n.replicatorClient.Sync(context.Background(), &req)

	if err != nil {
		n.logger.Error("error send sync message", zap.String("remote node ID", n.remoteNodeID), zap.Error(err))
		return
	}

	if r.Code != syncCodeSuccess {
		n.logger.Error("error sync response code", zap.String("remote node ID", n.remoteNodeID), zap.Int64("code", r.Code))
		return
	}

	// mark variables as replicated
	n.replicatedVersionsMx.Lock()
	for key, stamp := range replicatedVersions {
		n.replicatedVersions[key] = stamp
	}
	n.replicatedVersionsMx.Unlock()
}
