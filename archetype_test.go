package ecsgo

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type PositionComponent struct {
	X int
	Y int
}

type VelocityComponent struct {
	Vel int
}

func TestArcheType(t *testing.T) {
	a := newArcheType(reflect.TypeOf(PositionComponent{}), reflect.TypeOf(VelocityComponent{}))
	assert.NotNil(t, a)

	entityId := EntityId{
		id:      1,
		version: 1,
	}
	a.addEntity(entityId)
	posComp := getArcheTypeComponent[PositionComponent](a, entityId)
	assert.NotNil(t, posComp)

	posComp.X = 100
	posComp.Y = 100

	entityId2 := EntityId{
		id:      2,
		version: 2,
	}
	a.addEntity(entityId2)
	posComp2 := getArcheTypeComponent[PositionComponent](a, entityId2)
	assert.NotNil(t, posComp2)

	posComp2.X = 200
	posComp2.Y = 200

	posComp = getArcheTypeComponent[PositionComponent](a, entityId)
	assert.NotNil(t, posComp)

	assert.Equal(t, 100, posComp.X)
	assert.Equal(t, 100, posComp.Y)

	posComp2 = getArcheTypeComponent[PositionComponent](a, entityId2)
	assert.NotNil(t, posComp)

	assert.Equal(t, 200, posComp2.X)
	assert.Equal(t, 200, posComp2.Y)

	setArcheTypeComponent[VelocityComponent](a, entityId, VelocityComponent{
		Vel: 100,
	})

	velComp := getArcheTypeComponent[VelocityComponent](a, entityId)
	assert.NotNil(t, velComp)
	assert.Equal(t, 100, velComp.Vel)
}
