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
			// Если в пришедшем запросе мы получили наш nodeID, то пропускем. Предполагаем, что у нас всегда данные самые свежие
			// todo: может при отправке на конкретную ноду - не отправлять ее данные? Добавить эту проверку в node.sync()?
			if nodeID == rplx.nodeID {
				continue
			}
			localVar.updateRemoteNode(nodeID, n.Value, n.Stamp)
		}
	}

	return &SyncResponse{Code: 0}, nil
}
