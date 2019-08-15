package rplx

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRplx_Upsert_TTL_Upsert(t *testing.T) {
	var err error
	var v int64

	r := New()

	r.Upsert("VAR-1", 100)

	v, err = r.Get("VAR-1")
	assert.NoError(t, err)
	assert.Equal(t, int64(100), v)

	// make variable expired
	err = r.UpdateTTL("VAR-1", time.Now().Add(-time.Second))
	assert.NoError(t, err)

	// Variable not garbage collected!
	r.Upsert("VAR-1", 200)

	v, err = r.Get("VAR-1")
	require.NoError(t, err)
	assert.Equal(t, int64(200), v)
}
