package rplx

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
)

func TestRplx_Sync(t *testing.T) {
	rplx := &Rplx{
		nodeID:    "node1",
		logger:    zap.NewNop(),
		variables: map[string]*variable{},
	}

	req := SyncRequest{
		NodeID: "node2",
		Variables: map[string]*SyncVariable{
			"var1": {NodesValues: map[string]*SyncNodeValue{
				"node3": {
					ID:    "node3",
					Value: 300,
					Stamp: 300,
				},
			}},
		},
	}

	response, err := rplx.Sync(context.Background(), &req)

	assert.NoError(t, err)
	assert.Equal(t, int64(0), response.Code)
	assert.Equal(t, 1, len(rplx.variables))

	v, ok := rplx.variables["var1"]
	require.True(t, ok)

	value, err := v.get()
	assert.NoError(t, err)
	assert.Equal(t, int64(300), value)

}
