package rplx

import (
	"context"
	"go.uber.org/zap"
)

const (
	syncCodeSuccess = int64(0)
)

// sync sends replication data from buffer to remote node
func (n *node) sync() error {
	n.syncMx.Lock()
	defer n.syncMx.Unlock()

	req := SyncRequest{
		NodeID:    n.localNodeID,
		Variables: make(map[string]*SyncVariable),
	}

	n.bufferMx.Lock()
	for name, v := range n.buffer {

		req.Variables[name] = &SyncVariable{
			TTL:         v.getTTL(),
			TTLStamp:    v.getTTLStamp(),
			NodesValues: make(map[string]*SyncNodeValue),
		}

		req.Variables[name].NodesValues[n.localNodeID] = &SyncNodeValue{
			ID:    n.localNodeID,
			Value: v.selfItem.value(),
			Stamp: v.selfItem.stamp(),
		}

		for nodeID, item := range v.items() {
			req.Variables[name].NodesValues[nodeID] = &SyncNodeValue{
				ID:    nodeID,
				Value: item.value(),
				Stamp: item.stamp(),
			}
		}

		delete(n.buffer, name)
	}
	n.bufferMx.Unlock()

	n.logger.Debug("send sync message", zap.String("nodeID", n.ID))
	r, err := n.replicatorClient.Sync(context.Background(), &req)
	n.logger.Debug("complete sync message", zap.String("nodeID", n.ID))
	if err != nil {
		return err
	}

	n.logger.Debug("message sent", zap.Int64("sync response code", r.Code), zap.String("nodeID", n.ID))

	if r.Code != syncCodeSuccess {
		n.logger.Warn("error sync response code", zap.Int64("code", r.Code))
	}

	return nil
}
