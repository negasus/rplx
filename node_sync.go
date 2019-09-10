package rplx

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"sync/atomic"
)

const (
	syncCodeSuccess = int64(0)
)

func (n *node) sync() {
	select {
	case n.syncQueue <- struct{}{}:
	default:
	}
}

func (n *node) listenSyncQueue() {
	for range n.syncQueue {
		if err := n.sendSyncRequest(); err != nil {
			n.logger.Error("error send sync request", zap.Error(err), zap.String("remote node ID", n.remoteNodeID))
		}
	}
}

// sync sends replication data from buffer to remote node
func (n *node) sendSyncRequest() error {
	if atomic.LoadInt32(&n.connected) == 0 {
		return nil
	}

	if !atomic.CompareAndSwapInt32(&n.syncing, 0, 1) {
		return nil
	}
	defer atomic.StoreInt32(&n.syncing, 0)

	n.bufferMx.Lock()

	// if replication called by ticker, but buffer is empty - return
	if len(n.buffer) == 0 {
		n.bufferMx.Unlock()
		return nil
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
			lastReplicatedVersion = -1
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
		return nil
	}

	n.logger.Debug("send sync message", zap.String("remote node ID", n.remoteNodeID), zap.Int("variables", len(req.Variables)), zap.Any("vars map", req.Variables))

	r, err := n.replicatorClient.Sync(context.Background(), &req)

	if err != nil {
		// todo: check error, check off 'n.connected' flag and send node to reconnect?
		return fmt.Errorf("error call sync method, %v", err)
	}

	if r.Code != syncCodeSuccess {
		return fmt.Errorf("error sync response code %d", r.Code)
	}

	// mark variables as replicated
	n.replicatedVersionsMx.Lock()
	for key, stamp := range replicatedVersions {
		n.replicatedVersions[key] = stamp
	}
	n.replicatedVersionsMx.Unlock()

	return nil
}
