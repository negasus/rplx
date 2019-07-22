package rplx

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestRplx_Get_NotExists(t *testing.T) {
	rplx := New("node1", zap.NewNop())

	_, err := rplx.Get("var1")

	assert.Error(t, err)
	assert.Equal(t, ErrVariableNotExists, err)
}

func TestRplx_Get_Expired(t *testing.T) {
	rplx := New("node1", zap.NewNop())

	v := newVariable("var1")
	v.ttl = time.Now().UTC().Add(-time.Second).UnixNano()

	rplx.variables["var1"] = v

	_, err := rplx.Get("var1")

	assert.Error(t, err)
	assert.Equal(t, ErrVariableNotExists, err)

	_, ok := rplx.variables["var1"]
	assert.False(t, ok)
}

func TestRplx_Get(t *testing.T) {
	rplx := New("node1", zap.NewNop())

	v := newVariable("var1")
	v.selfItem.v = 42

	rplx.variables["var1"] = v

	value, err := rplx.Get("var1")

	assert.NoError(t, err)
	assert.Equal(t, int64(42), value)

	vv, ok := rplx.variables["var1"]
	assert.True(t, ok)
	assert.Equal(t, vv, v)
}

func TestRplx_Delete_NotExists(t *testing.T) {
	rplx := New("node1", zap.NewNop())

	err := rplx.Delete("var1")

	assert.Error(t, err)
	assert.Equal(t, ErrVariableNotExists, err)
}

func TestRplx_Delete(t *testing.T) {
	rplx := New("node1", zap.NewNop())

	v := newVariable("var1")

	rplx.variables["var1"] = v

	assert.True(t, v.ttl == 0)

	err := rplx.Delete("var1")

	assert.NoError(t, err)
	assert.True(t, v.ttl > 0 && v.ttl < time.Now().UTC().UnixNano())
}
