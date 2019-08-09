package rplx

import (
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
					Value:   300,
					Version: 300,
				},
			}},
		},
	}

	rplx.sync(&req)

	assert.Equal(t, 1, len(rplx.variables))

	v, ok := rplx.variables["var1"]
	require.True(t, ok)

	value := v.get()
	assert.Equal(t, int64(300), value)
}
