package rplx

import (
	"context"
	"go.uber.org/zap"
)

// Sync is GRPC function, fired on incoming sync message
func (rplx *Rplx) Sync(ctx context.Context, req *SyncRequest) (*SyncResponse, error) {
	rplx.logger.Debug("get SyncRequest", zap.Int("variables", len(req.Variables)), zap.String("from node", req.NodeID))

	go rplx.sync(req)

	return &SyncResponse{Code: 0}, nil
}

func (rplx *Rplx) sync(req *SyncRequest) {
	rplx.variablesMx.RLock()
	rplx.logger.Debug("rplx.variables before sync", zap.Int("count", len(rplx.variables)))
	rplx.variablesMx.RUnlock()

	for name, v := range req.Variables {
		rplx.variablesMx.Lock()
		localVar, ok := rplx.variables[name]
		if !ok {
			localVar = newVariable(name)
			rplx.variables[name] = localVar
		}
		rplx.variablesMx.Unlock()

		varWasUpdated := false

		rplx.nodesMx.RLock()
		remoteNodeInstance := rplx.nodes[req.NodeID]
		rplx.nodesMx.RUnlock()

		for nodeID, n := range v.NodesValues {
			// Если мы получили данные с нашим remoteNodeID, пропускаем
			if nodeID == rplx.nodeID {
				continue
			}

			if localVar.updateItem(nodeID, n.Value, n.Version) {
				varWasUpdated = true

				// Если нода, от которой пришли данные, есть у нас в списке - куда мы шлем обновления,
				// то для данной переменной для данной ноды запишем, что у нас самая свежая версия
				if remoteNodeInstance != nil {
					remoteNodeInstance.replicatedVersionsMx.Lock()
					remoteNodeInstance.replicatedVersions[name+"@"+nodeID] = n.Version
					remoteNodeInstance.replicatedVersionsMx.Unlock()
				}
			}

		}

		if localVar.ttlVersion < v.TTLVersion {
			localVar.ttl = v.TTL
			localVar.ttlVersion = v.TTLVersion
			varWasUpdated = true
		}

		if varWasUpdated {
			go rplx.sendToReplication(localVar)
		}
	}

	rplx.variablesMx.RLock()
	rplx.logger.Debug("rplx.variables after sync", zap.Int("count", len(rplx.variables)))
	rplx.variablesMx.RUnlock()
}
