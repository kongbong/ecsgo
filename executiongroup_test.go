package ecsgo

import (
	"context"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestComponent1 struct {
	X int
	Y int
}

type TestComponent2 struct {
	V float64
}

type TestComponent3 struct {
	str string
	vec []int
}

type TestComponent4 struct {
	f1 float64
	f2 float64
}

func TestExecutionGroup(t *testing.T) {
	eg := newExecutionGroup()

	var executed [5]bool
	sys1 := newSystem(nil, "sys1", 0, func(ctx *ExecutionContext) error {
		executed[0] = true
		assert.True(t, executed[1])
		assert.True(t, executed[2])
		assert.Equal(t, 1, ctx.GetQueryResultCount())
		qr := ctx.GetQueryResult(0)
		qr.ForeachEntities(func(accessor *ArcheTypeAccessor) error {
			id := accessor.GetEntityId()
			assert.Equal(t, uint32(1), id.id)
			assert.Equal(t, uint32(1), id.version)
			t1 := GetComponentByAccessor[TestComponent1](accessor)
			assert.NotNil(t, t1)
			t1.X = 200
			t1.Y = 200
			t2 := GetComponentByAccessor[TestComponent2](accessor)
			assert.NotNil(t, t2)
			assert.Equal(t, 3.1415, t2.V)
			t3 := GetComponentByAccessor[TestComponent3](accessor)
			assert.Nil(t, t3)
			var testT3 TestComponent3
			success := SetComponentData(accessor, testT3)
			assert.False(t, success)
			return nil
		})
		t.Log("sys1 executed")
		time.Sleep(time.Duration(rand.Int31n(100)) * time.Millisecond)
		return nil
	})
	q := sys1.NewQuery()
	AddReadWriteComponent[TestComponent1](q)
	AddOptionalReadWriteComponent[TestComponent2](q)

	sys2 := newSystem(nil, "sys2", 0, func(ctx *ExecutionContext) error {
		executed[1] = true
		assert.False(t, executed[0])
		assert.Equal(t, 1, ctx.GetQueryResultCount())
		qr := ctx.GetQueryResult(0)
		qr.ForeachEntities(func(accessor *ArcheTypeAccessor) error {
			id := accessor.GetEntityId()
			assert.Equal(t, uint32(1), id.id)
			assert.Equal(t, uint32(1), id.version)

			t2 := GetComponentByAccessor[TestComponent2](accessor)
			assert.NotNil(t, t2)
			t2.V = 3.1415
			return nil
		})
		t.Log("sys2 executed")
		time.Sleep(time.Duration(rand.Int31n(100)) * time.Millisecond)
		return nil
	})
	q2 := sys2.NewQuery()
	AddReadWriteComponent[TestComponent2](q2)

	sys3 := newSystem(nil, "sys3", 0, func(ctx *ExecutionContext) error {
		executed[2] = true
		assert.False(t, executed[4])
		assert.Equal(t, 1, ctx.GetQueryResultCount())
		qr := ctx.GetQueryResult(0)
		qr.ForeachEntities(func(accessor *ArcheTypeAccessor) error {
			id := accessor.GetEntityId()
			assert.Equal(t, uint32(2), id.id)
			assert.Equal(t, uint32(1), id.version)
			t3 := GetComponentByAccessor[TestComponent3](accessor)
			assert.NotNil(t, t3)
			t3.str = "TestTest"
			t3.vec = append(t3.vec, 1, 2, 3, 4, 5)
			return nil
		})
		t.Log("sys3 executed")
		time.Sleep(time.Duration(rand.Int31n(100)) * time.Millisecond)
		return nil
	})
	q3 := sys3.NewQuery()
	AddReadWriteComponent[TestComponent3](q3)
	AddExcludeComponent[TestComponent1](q3)

	sys4 := newSystem(nil, "sys4", 0, func(ctx *ExecutionContext) error {
		executed[3] = true
		assert.False(t, executed[4])
		assert.Equal(t, 1, ctx.GetQueryResultCount())
		qr := ctx.GetQueryResult(0)
		qr.ForeachEntities(func(accessor *ArcheTypeAccessor) error {
			id := accessor.GetEntityId()
			assert.Equal(t, uint32(2), id.id)
			assert.Equal(t, uint32(1), id.version)
			t4 := GetComponentByAccessor[TestComponent4](accessor)
			assert.NotNil(t, t4)
			t4.f1 = 1234.1234
			t4.f2 = 3.141516
			return nil
		})
		t.Log("sys4 executed")
		time.Sleep(time.Duration(rand.Int31n(100)) * time.Millisecond)
		return nil
	})
	q4 := sys4.NewQuery()
	AddReadWriteComponent[TestComponent4](q4)
	AddExcludeComponent[TestComponent1](q4)

	sys5 := newSystem(nil, "sys5", 0, func(ctx *ExecutionContext) error {
		executed[4] = true
		assert.True(t, executed[3])
		assert.Equal(t, 1, ctx.GetQueryResultCount())
		qr := ctx.GetQueryResult(0)
		qr.ForeachEntities(func(accessor *ArcheTypeAccessor) error {
			id := accessor.GetEntityId()
			assert.Equal(t, uint32(2), id.id)
			assert.Equal(t, uint32(1), id.version)
			t3 := GetComponentByAccessor[TestComponent3](accessor)
			assert.NotNil(t, t3)
			assert.Equal(t, "TestTest", t3.str)
			assert.Equal(t, []int{1, 2, 3, 4, 5}, t3.vec)
			t4 := GetComponentByAccessor[TestComponent4](accessor)
			assert.NotNil(t, t4)
			assert.Equal(t, 1234.1234, t4.f1)
			assert.Equal(t, 3.141516, t4.f2)
			return nil
		})
		t.Log("sys5 executed")
		time.Sleep(time.Duration(rand.Int31n(100)) * time.Millisecond)
		return nil
	})
	q5 := sys5.NewQuery()
	AddReadWriteComponent[TestComponent3](q5)
	AddReadWriteComponent[TestComponent4](q5)

	eg.addSystem(sys1)
	eg.addSystem(sys2)
	eg.addSystem(sys3)
	eg.addSystem(sys4)
	eg.addSystem(sys5)

	eg.build()

	var t1 TestComponent1
	var t2 TestComponent2
	var t3 TestComponent3
	var t4 TestComponent4
	a1 := newArcheType(reflect.TypeOf(t1), reflect.TypeOf(t2))
	a2 := newArcheType(reflect.TypeOf(t3), reflect.TypeOf(t4))

	a1.addEntity(EntityId{
		id:      1,
		version: 1,
	})
	a2.addEntity(EntityId{
		id:      2,
		version: 1,
	})

	added := sys1.addArcheTypeIfInterest(a1)
	assert.True(t, added)
	added = sys2.addArcheTypeIfInterest(a1)
	assert.True(t, added)
	added = sys3.addArcheTypeIfInterest(a1)
	assert.False(t, added)
	added = sys4.addArcheTypeIfInterest(a1)
	assert.False(t, added)
	added = sys5.addArcheTypeIfInterest(a1)
	assert.False(t, added)

	added = sys1.addArcheTypeIfInterest(a2)
	assert.False(t, added)
	added = sys2.addArcheTypeIfInterest(a2)
	assert.False(t, added)
	added = sys3.addArcheTypeIfInterest(a2)
	assert.True(t, added)
	added = sys4.addArcheTypeIfInterest(a2)
	assert.True(t, added)
	added = sys5.addArcheTypeIfInterest(a2)
	assert.True(t, added)

	// dependency tree should be
	//        root
	//      /   \   \
	//     2    3    4
	//     |     \  /
	//     1      5
	eg.execute(time.Second, context.Background())
	for i := 0; i < 5; i++ {
		assert.True(t, executed[i])
		executed[i] = false
	}
}
