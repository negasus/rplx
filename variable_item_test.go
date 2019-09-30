package rplx

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestVariableItemUpdate(t *testing.T) {
	item := newVariableItem()

	tm := time.Now().UTC().UnixNano()

	item.update(100)

	assert.True(t, item.ver >= tm)
	assert.Equal(t, int64(100), item.val)
}

func TestVariableItemSet(t *testing.T) {
	item := newVariableItem()

	item.set(100, 200)

	assert.Equal(t, int64(100), item.val)
	assert.Equal(t, int64(200), item.ver)
}

func TestVariableItemValue(t *testing.T) {
	item := newVariableItem()

	item.val = 100

	assert.Equal(t, int64(100), item.value())
}

func TestVariableItemVersion(t *testing.T) {
	item := newVariableItem()

	item.ver = 100

	assert.Equal(t, int64(100), item.version())
}
