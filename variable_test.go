package rplx

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVariable_Get(t *testing.T) {
	v := newVariable("var1")

	v.selfItem = &variableItem{v: 100}
	v.remoteItems["node2"] = &variableItem{v: 20}
	v.remoteItems["node3"] = &variableItem{v: 3}

	res, err := v.get()
	assert.NoError(t, err)
	assert.Equal(t, int64(123), res)
}

func TestVariable_Update(t *testing.T) {
	v := newVariable("var1")

	v.update(100)

	assert.Equal(t, int64(100), v.selfItem.v)
}

func TestVariable_UpdateNodeValue(t *testing.T) {
	v := newVariable("var1")

	_, ok := v.remoteItems["node1"]
	assert.False(t, ok)

	v.updateRemoteNode("node1", 100, 100)
	assert.Equal(t, int64(100), v.remoteItems["node1"].v)

	v.updateRemoteNode("node1", 200, 200)
	assert.Equal(t, int64(200), v.remoteItems["node1"].v)

	// updateRemoteNode with stamp less than current, does not save new incoming value
	v.updateRemoteNode("node1", 300, 150)
	assert.Equal(t, int64(200), v.remoteItems["node1"].v)
}

func TestVariable_Items(t *testing.T) {
	v := newVariable("var1")

	item1 := &variableItem{v: 100}
	item2 := &variableItem{v: 20}
	item3 := &variableItem{v: 3}

	v.remoteItems["node1"] = item1
	v.remoteItems["node2"] = item2
	v.remoteItems["node3"] = item3

	items := v.items()

	item, ok := items["node1"]
	assert.True(t, ok)
	assert.Equal(t, item1.v, item.v)
	assert.Equal(t, item1.s, item.s)
	assert.NotEqual(t, fmt.Sprintf("%p", item), fmt.Sprintf("%p", item1))

	item, ok = items["node2"]
	assert.True(t, ok)
	assert.Equal(t, item2.v, item.v)
	assert.Equal(t, item2.s, item.s)
	assert.NotEqual(t, fmt.Sprintf("%p", item), fmt.Sprintf("%p", item2))

	item, ok = items["node3"]
	assert.True(t, ok)
	assert.Equal(t, item3.v, item.v)
	assert.Equal(t, item3.s, item.s)
	assert.NotEqual(t, fmt.Sprintf("%p", item), fmt.Sprintf("%p", item3))
}

func TestVariable_Replicated(t *testing.T) {
	v := newVariable("var1")

	v.remoteItems["node2"] = &variableItem{}

	assert.False(t, v.isReplicated("node1")) // not exists

	assert.False(t, v.isReplicated("node2"))

	v.replicatedOn("node2")
	assert.True(t, v.isReplicated("node2"))

	v.replicatedOff("node2")
	assert.False(t, v.isReplicated("node2"))
}
