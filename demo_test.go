package rplx

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"testing"
)

type replicatorClientMock struct {
	mock.Mock
}

func (m *replicatorClientMock) Hello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*HelloResponse), args.Error(1)
}

func (m *replicatorClientMock) Sync(ctx context.Context, in *SyncRequest, opts ...grpc.CallOption) (*SyncResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*SyncResponse), args.Error(1)
}

func TestEmptySyncRequestIfEmptyVariables(t *testing.T) {

	mockClient := &replicatorClientMock{}

	v := newVariable("VAR-1")

	node1 := &node{
		logger:           zap.NewNop(),
		connected:        1,
		localNodeID:      "localNodeID",
		replicatorClient: mockClient,
		buffer: map[string]*variable{
			"VAR-1": v,
		},
		replicatedVersions: map[string]int64{},
	}

	err := node1.sendSyncRequest()
	require.NoError(t, err)

	mockClient.AssertNotCalled(t, "Sync", mock.Anything, mock.Anything, mock.Anything)

	mockClient.AssertExpectations(t)
}

func TestErrorNodeSyncWithBadSyncResponseCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockReplicatorClient(ctrl)

	mockClient.EXPECT().Sync(
		gomock.Any(),
		gomock.Any(),
	).Return(&SyncResponse{
		Code: 1,
	}, nil)

	v := newVariable("VAR-1")
	v.update(100)

	node1 := &node{
		logger:           zap.NewNop(),
		connected:        1,
		localNodeID:      "localNodeID",
		replicatorClient: mockClient,
		buffer: map[string]*variable{
			"VAR-1": v,
		},
		metrics: newMetrics(),
	}

	err := node1.sendSyncRequest()
	require.Error(t, err)
	assert.Equal(t, "error sync response code 1", err.Error())
}

func TestFillReplicatedVersionsAfterNodeSync(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockReplicatorClient(ctrl)

	mockClient.EXPECT().Sync(
		gomock.Any(),
		&SyncRequest{
			NodeID: "localNodeID",
			Variables: map[string]*SyncVariable{
				"VAR-1": &SyncVariable{
					NodesValues: map[string]*SyncNodeValue{
						"localNodeID": &SyncNodeValue{
							Value:   100,
							Version: 1,
						},
						"remoteNode1": &SyncNodeValue{
							Value:   200,
							Version: 2,
						},
					},
					TTL:        500,
					TTLVersion: 5,
				},
			},
		},
	).Return(&SyncResponse{
		Code: 0,
	}, nil)

	var1 := newVariable("VAR-1")
	var1.self.val = 100
	var1.self.ver = 1
	var1.ttl = 500
	var1.ttlVersion = 5
	var1.remoteItems = map[string]*variableItem{
		"remoteNode1": &variableItem{
			val: 200,
			ver: 2,
		},
	}

	node1 := &node{
		logger:           zap.NewNop(),
		connected:        1,
		localNodeID:      "localNodeID",
		replicatorClient: mockClient,
		buffer: map[string]*variable{
			"VAR-1": var1,
		},
		replicatedVersions: map[string]int64{},
		metrics:            newMetrics(),
	}

	err := node1.sendSyncRequest()

	assert.NoError(t, err)

	replicatedVersion, ok := node1.replicatedVersions["VAR-1@localNodeID"]
	require.True(t, ok)
	assert.Equal(t, int64(1), replicatedVersion)

	replicatedVersion, ok = node1.replicatedVersions["VAR-1@remoteNode1"]
	require.True(t, ok)
	assert.Equal(t, int64(2), replicatedVersion)
}
