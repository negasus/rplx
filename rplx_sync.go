package rplx

import (
	"context"
	"go.uber.org/zap"
)

// Sync is GRPC function, fired on incoming sync message
func (rplx *Rplx) Sync(ctx context.Context, mes *SyncRequest) (*SyncResponse, error) {
	rplx.logger.Debug("get SyncRequest", zap.Any("request", mes))

	for name, v := range mes.Variables {
		rplx.variablesMx.Lock()
		localVar, ok := rplx.variables[name]
		if !ok {
			localVar = newVariable(name)
			rplx.variables[name] = localVar
		}
		rplx.variablesMx.Unlock()

		updated := false

		for nodeID, n := range v.NodesValues {
			if localVar.updateRemoteNode(nodeID, n.Value, n.Stamp) {
				updated = true
			}
		}

		if localVar.ttlStamp < v.TTL {
			localVar.ttl = v.TTL
			localVar.ttlStamp = v.TTLStamp
			updated = true
		}

		if updated {
			go rplx.sendToReplication(localVar)
		}

		localVar.updateReplicationStamp(mes.NodeID)
	}

	return &SyncResponse{Code: 0}, nil
}
