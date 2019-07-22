package rplx

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"testing"
	"time"
)

type mockReplicationClient struct {
	mock.Mock
}

func (m *mockReplicationClient) Sync(ctx context.Context, in *SyncRequest, opts ...grpc.CallOption) (*SyncResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*SyncResponse), args.Error(1)
}

func TestNode_Sync(t *testing.T) {

	replicatorClient := &mockReplicationClient{}

	replicatorClient.On("Sync", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		var n *SyncNodeValue

		in := args.Get(1).(*SyncRequest)

		assert.Equal(t, "node1", in.NodeID)
		assert.Equal(t, 2, len(in.Variables))

		// var1
		v, ok := in.Variables["var1"]
		assert.True(t, ok)
		assert.Equal(t, int64(100), v.TTL)
		assert.True(t, v.TTLStamp > 0 && v.TTLStamp <= time.Now().UTC().UnixNano())
		assert.Equal(t, 3, len(v.NodesValues))

		n, ok = v.NodesValues["node1"]
		assert.True(t, ok)
		assert.Equal(t, int64(0), n.Stamp)
		assert.Equal(t, int64(0), n.Value)

		n, ok = v.NodesValues["node2"]
		assert.True(t, ok)
		assert.Equal(t, int64(100), n.Stamp)
		assert.Equal(t, int64(100), n.Value)

		n, ok = v.NodesValues["node3"]
		assert.True(t, ok)
		assert.Equal(t, int64(200), n.Stamp)
		assert.Equal(t, int64(200), n.Value)

		// var2
		v, ok = in.Variables["var2"]
		assert.True(t, ok)
		assert.Equal(t, int64(100), v.TTL)
		assert.Equal(t, 3, len(v.NodesValues))

		n, ok = v.NodesValues["node1"]
		assert.True(t, ok)
		assert.Equal(t, int64(505), n.Stamp)
		assert.Equal(t, int64(500), n.Value)

		n, ok = v.NodesValues["node2"]
		assert.True(t, ok)
		assert.Equal(t, int64(100), n.Stamp)
		assert.Equal(t, int64(100), n.Value)

		n, ok = v.NodesValues["node3"]
		assert.True(t, ok)
		assert.Equal(t, int64(200), n.Stamp)
		assert.Equal(t, int64(200), n.Value)
	}).Return(&SyncResponse{Code: 0}, nil)

	n := &node{
		localNodeID:      "node1",
		replicatorClient: replicatorClient,
		buffer:           make(map[string]*variable),
		logger:           zap.NewNop(),
	}

	n.buffer["var1"] = newVariable("var1")
	n.buffer["var1"].updateRemoteNode("node2", 100, 100)
	n.buffer["var1"].updateRemoteNode("node3", 200, 200)
	n.buffer["var1"].updateTTL(time.Unix(0, 100))

	n.buffer["var2"] = newVariable("var2")
	n.buffer["var2"].selfItem.setValue(500)
	n.buffer["var2"].selfItem.setStamp(505)
	n.buffer["var2"].updateRemoteNode("node2", 100, 100)
	n.buffer["var2"].updateRemoteNode("node3", 200, 200)
	n.buffer["var2"].updateTTL(time.Unix(0, 100))

	err := n.sync()

	assert.NoError(t, err)
	replicatorClient.AssertExpectations(t)
}
