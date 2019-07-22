package rplx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReplicated(t *testing.T) {
	v := &variableItem{}

	assert.False(t, v.isReplicated())

	v.replicatedOn()
	assert.True(t, v.isReplicated())

	v.replicatedOff()
	assert.False(t, v.isReplicated())
}
