package rplx

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestAPI_Get_with_cache(t *testing.T) {
	v := newVariable("A")
	v.CacheDuration = 1
	v.self.val = 100

	val := v.get()
	assert.Equal(t, int64(100), val)

	v.self.val = 200

	val = v.get()
	assert.Equal(t, int64(100), val)

	time.Sleep(time.Second)

	val = v.get()
	assert.Equal(t, int64(200), val)
}

func TestAPI_Upsert_NewVariable(t *testing.T) {
	r := New()

	tm := time.Now().UTC()

	newValue := r.Upsert("VAR-1", 100)
	assert.Equal(t, int64(100), newValue)

	v, ok := r.variables["VAR-1"]
	require.True(t, ok)

	assert.Equal(t, int64(100), v.self.val)
	assert.True(t, v.self.ver >= tm.UnixNano())
	assert.Len(t, v.remoteItems, 0)
}

func TestAPI_Upsert_ExistsVariable(t *testing.T) {
	r := New()
	v := newVariable("VAR-1")

	vt := time.Now().UTC().Add(-time.Second).UnixNano()

	v.self.val = 150
	v.self.ver = vt
	r.variables["VAR-1"] = v

	newValue := r.Upsert("VAR-1", 100)
	assert.Equal(t, int64(250), newValue)

	v, ok := r.variables["VAR-1"]
	require.True(t, ok)

	assert.Equal(t, int64(250), v.self.val)
	assert.True(t, v.self.ver > vt)
	assert.Len(t, v.remoteItems, 0)
}

func TestAPI_Upsert_ExpiredVariable(t *testing.T) {
	r := New()

	v := newVariable("VAR-1")
	v.CacheDuration = 0
	v.ttl = time.Now().UTC().Add(-time.Second).UnixNano()

	v.self.val = 150
	r.variables["VAR-1"] = v

	newValue := r.Upsert("VAR-1", 100)
	assert.Equal(t, int64(100), newValue)

	v, ok := r.variables["VAR-1"]
	require.True(t, ok)

	assert.Equal(t, int64(100), v.self.val)
	assert.Equal(t, int64(0), v.ttl)
	assert.Len(t, v.remoteItems, 0)
}

func TestAPI_UpdateTTL_NotExistsVariable(t *testing.T) {
	r := New()

	err := r.UpdateTTL("VAR-1", time.Now())

	require.Error(t, err)
	assert.Equal(t, "variable not exists", err.Error())
}

func TestAPI_UpdateTTL_ExistsVariable(t *testing.T) {
	r := New()

	tt := time.Now().UTC().Add(time.Second)

	v := newVariable("VAR-1")
	v.ttl = time.Now().UTC().Add(-time.Second).UnixNano()
	r.variables["VAR-1"] = v

	err := r.UpdateTTL("VAR-1", tt)
	require.NoError(t, err)

	assert.Equal(t, tt.UnixNano(), r.variables["VAR-1"].ttl)
}

func TestAPI_All(t *testing.T) {
	r := New()
	r.Upsert("VAR1", 100)
	err := r.UpdateTTL("VAR1", time.Now().UTC().Add(-time.Second))
	require.NoError(t, err)
	r.Upsert("VAR2", 200)

	varsNotExpired, varsExpired := r.All()
	assert.Equal(t, 1, len(varsNotExpired))
	assert.Equal(t, 1, len(varsExpired))

	var ok bool
	var v int64

	v, ok = varsNotExpired["VAR2"]
	assert.True(t, ok)
	assert.Equal(t, int64(200), v)

	v, ok = varsExpired["VAR1"]
	assert.True(t, ok)
	assert.Equal(t, int64(100), v)
}
