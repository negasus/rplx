package rplx

import (
	"context"
	"go.uber.org/zap"
)

// Sync is GRPC function, fired on incoming sync message
func (rplx *Rplx) Sync(ctx context.Context, mes *SyncRequest) (*SyncResponse, error) {
	rplx.logger.Debug("get SyncRequest", zap.Any("request", mes))

	rplx.variablesMx.Lock()
	defer rplx.variablesMx.Unlock()

	for name, v := range mes.Variables {
		localVar, ok := rplx.variables[name]
		if !ok {
			localVar = newVariable(name)
			rplx.variables[name] = localVar
		}

		for nodeID, n := range v.NodesValues {
			// if got data for current node, skip it
			if nodeID == rplx.nodeID {
				continue
			}
			localVar.updateRemoteNode(nodeID, n.Value, n.Stamp)
		}

		if localVar.ttlStamp < v.TTL {
			localVar.ttl = v.TTL
			localVar.ttlStamp = v.TTLStamp
		}
	}

	return &SyncResponse{Code: 0}, nil
}
