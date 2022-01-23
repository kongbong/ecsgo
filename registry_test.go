package ecsgo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEntities(t *testing.T) {
	registry := &Registry{}

	for i := 0; i < 10; i++ {
		e := registry.Create()
		assert.Equal(t, e.ToEntity(), Entity(i))
		assert.Equal(t, e.ToVersion(), uint32(1))
	}

	e5 := newEntity(4, 1)
	registry.Release(e5)
	assert.Equal(t, false, registry.IsAlive(e5))
	recycle := registry.Create()
	assert.Equal(t, recycle.ToEntity(), e5.ToEntity())
	assert.NotEqual(t, recycle.ToVersion(), e5.ToVersion())
	assert.Equal(t, true, registry.IsAlive(recycle))
}
