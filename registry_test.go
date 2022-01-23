package ecsgo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEntities(t *testing.T) {
	registry := &Registry{}

	for i := 0; i < 10; i++ {
		e := registry.Create()
		assert.Equal(t, e.id, uint32(i))
		assert.Equal(t, e.version, uint32(1))
	}

	e5 := newEntity(4, 1)
	registry.Release(e5)
	assert.Equal(t, false, registry.IsAlive(e5))
	recycle := registry.Create()
	assert.Equal(t, recycle.id, e5.id)
	assert.NotEqual(t, recycle.version, e5.version)
	assert.Equal(t, true, registry.IsAlive(recycle))
}
