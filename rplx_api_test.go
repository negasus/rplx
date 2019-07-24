package rplx

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRplx_Get_NotExists(t *testing.T) {
	rplx := New("node1")

	_, err := rplx.Get("var1")

	assert.Error(t, err)
	assert.Equal(t, ErrVariableNotExists, err)
}

func TestRplx_Get_Expired(t *testing.T) {
	rplx := New("node1")

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
	rplx := New("node1")

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
	rplx := New("node1")

	err := rplx.Delete("var1")

	assert.Error(t, err)
	assert.Equal(t, ErrVariableNotExists, err)
}

func TestRplx_Delete(t *testing.T) {
	rplx := New("node1")

	v := newVariable("var1")

	rplx.variables["var1"] = v

	assert.True(t, v.ttl == 0)

	err := rplx.Delete("var1")

	assert.NoError(t, err)
	assert.True(t, v.ttl > 0 && v.ttl < time.Now().UTC().UnixNano())
}

func TestRplx_Upsert_NotExists(t *testing.T) {
	rplx := New("node1")

	rplx.Upsert("var1", 100)

	v, ok := rplx.variables["var1"]

	assert.True(t, ok)
	assert.Equal(t, int64(100), v.selfItem.v)
}

func TestRplx_Upsert(t *testing.T) {
	rplx := New("node1")

	v := newVariable("var1")
	v.selfItem.v = 100
	rplx.variables["var1"] = v

	rplx.Upsert("var1", 200)

	v, ok := rplx.variables["var1"]

	assert.True(t, ok)
	assert.Equal(t, int64(300), v.selfItem.v)
}

func TestRplx_UpdateTTL_NotExists(t *testing.T) {
	rplx := New("node1")

	err := rplx.UpdateTTL("var1", time.Now())

	assert.Error(t, err)
	assert.Equal(t, ErrVariableNotExists, err)
}

func TestRplx_UpdateTTL(t *testing.T) {
	rplx := New("node1")

	t1 := time.Now().UTC().UnixNano()

	v := newVariable("var1")
	rplx.variables["var1"] = v

	newTTL := time.Now().Add(time.Minute)
	err := rplx.UpdateTTL("var1", newTTL)

	t2 := time.Now().UTC().UnixNano()

	assert.NoError(t, err)

	assert.True(t, rplx.variables["var1"].ttlStamp >= t1 && rplx.variables["var1"].ttlStamp <= t2)
	assert.Equal(t, newTTL.UnixNano(), rplx.variables["var1"].ttl)
}

func TestRplx_UpdateTTL_BadTLL(t *testing.T) {
	rplx := New("node1")

	err := rplx.UpdateTTL("var1", time.Now().UTC().Add(-time.Second))

	assert.Error(t, err)
	assert.Equal(t, ErrTTLLessThanNow, err)
}
