package ecsgo

import (
	"sync"
	"testing"
	"time"
	"log"

	"github.com/stretchr/testify/assert"
)

type Position struct {
	X float32
	Y float32
}

type Velocity struct {
	X float32
	Y float32
}

type HP struct {
	Hp float32
	MaxHp float32
}

type ReferenceFieldComponent struct {
	sliceIsNotAllowed []int
}

type EmptySizeComponent struct {}

type EnemyTag struct {}

func TestBasic(t *testing.T) {
	registry := New()
	defer registry.Free()

	var called2 bool
	var wg sync.WaitGroup
	wg.Add(3)
	sys := AddSystem1[Position](registry, OnTick, func (r *Registry, iter *Iterator) {
		log.Println("Should not call this")
		assert.True(t, false)
	})
	Exclude[Velocity](sys)
	Exclude[EnemyTag](sys)
	
	sys = AddSystem1[Velocity](registry, OnTick, func (r *Registry, iter *Iterator) {
		log.Println("Velocity system")
		assert.False(t, iter.IsNil())
		for ; !iter.IsNil(); iter.Next() {
			vel := Get[Velocity](iter)
			assert.Equal(t, vel.X, float32(10))
			assert.Equal(t, vel.Y, float32(10))
		}
		
		time.Sleep(time.Second)
		called2 = true
		wg.Done()
		log.Println("Velocity system Done")
	})
	Exclude[HP](sys)
	Exclude[EnemyTag](sys)

	sys = AddSystem2[Position, Velocity](registry, OnTick, func (r *Registry, iter *Iterator) {
		log.Println("Position, Velocity system")
		assert.True(t, called2)
		assert.False(t, iter.IsNil())
		for ; !iter.IsNil(); iter.Next() {
			pos := Get[Position](iter)
			vel := Get[Velocity](iter)

			assert.Equal(t, pos.X, float32(10))
			assert.Equal(t, pos.Y, float32(10))
			assert.Equal(t, vel.X, float32(10))
			assert.Equal(t, vel.Y, float32(10))
		}
		wg.Done()
	})
	Exclude[EnemyTag](sys)
	
	sys = AddSystem1[Position](registry, OnTick, func (r *Registry, iter *Iterator) {
		log.Println("Position system With Enemy Tag")
		assert.False(t, iter.IsNil())
		for ; !iter.IsNil(); iter.Next() {
			pos := Get[Position](iter)
			assert.Equal(t, pos.X, float32(100))
			assert.Equal(t, pos.Y, float32(100))
		}
		wg.Done()
	})
	Tag[EnemyTag](sys)

	entity := registry.Create()
	AddComponent[Position](registry, entity, &Position{10, 10})
	AddComponent[Velocity](registry, entity, &Velocity{10, 10})
	assert.Panics(t, func() {
		AddComponent[ReferenceFieldComponent](registry, entity, &ReferenceFieldComponent{})
	})
	assert.Panics(t, func() {
		AddComponent[EmptySizeComponent](registry, entity, &EmptySizeComponent{})
	})
	

	entity = registry.Create()
	AddComponent[Position](registry, entity, &Position{100, 100})
	AddTag[EnemyTag](registry, entity)

	registry.Tick(0.01)
	wg.Wait()
}

func TestReadOnly(t *testing.T) {
	registry := New()
	defer registry.Free()

	var called1 bool
	var wg sync.WaitGroup
	wg.Add(3)
	sys := AddSystem1[Position](registry, OnTick, func (r *Registry, iter *Iterator) {
		log.Println("First Position system")
		time.Sleep(time.Second)
		called1 = true
		wg.Done()
	})
	Readonly[Position](sys)

	sys = AddSystem1[Position](registry, OnTick, func (r *Registry, iter *Iterator) {
		log.Println("Second Position system")
		assert.False(t, called1)
		wg.Done()
	})
	Readonly[Position](sys)

	AddSystem2[Position, Velocity](registry, OnTick, func (r *Registry, iter *Iterator) {
		log.Println("Position, Velocity system")
		assert.True(t, called1)
		assert.Equal(t, 0.01, r.DeltaSeconds())
		wg.Done()
	})

	entity := registry.Create()
	AddComponent[Position](registry, entity, &Position{10, 10})
	AddComponent[Velocity](registry, entity, &Velocity{10, 10})

	registry.Tick(0.01)
	wg.Wait()
}

func TestPostTask(t *testing.T) {
	registry := New()
	defer registry.Free()

	var wg sync.WaitGroup
	var called1 bool
	wg.Add(2)
	PostTask1[Position](registry, OnTick, func (r *Registry, iter *Iterator) {
		log.Println("PostTask1 Position - this is called only one time")
		assert.False(t, called1)
		called1 = true
		wg.Done()
	})

	AddSystem2[Position, Velocity](registry, OnTick, func (r *Registry, iter *Iterator) {
		log.Println("Position, Velocity system - This is called every tick")
		assert.True(t, called1)
		wg.Done()
	})

	entity := registry.Create()
	AddComponent[Position](registry, entity, &Position{10, 10})
	AddComponent[Velocity](registry, entity, &Velocity{10, 10})

	registry.Tick(0.01)
	wg.Wait()
	wg.Add(1)
	registry.Tick(0.01)
	wg.Wait()
}

type ArrayComponent struct {
	intArr Array[int]
	nameStr Name
	dynamicStr String
}

func TestArrayName(t *testing.T) {
	registry := New()
	defer registry.Free()

	var wg sync.WaitGroup
	called := false
	wg.Add(1)
	AddSystem1[ArrayComponent](registry, OnTick, func (r *Registry, iter *Iterator) {
		arr := Get[ArrayComponent](iter)
		if !called {
			called = true
			assert.Equal(t, 1, arr.intArr.Len())
			assert.Equal(t, 10, arr.intArr.Get(0))
			assert.Equal(t, "Hello", arr.nameStr.String())
			assert.Equal(t, "Hello", arr.dynamicStr.String())
			arr.intArr.Add(100)
			arr.nameStr = NewName("World")
			arr.dynamicStr.Set("World")
		} else {
			assert.Equal(t, 2, arr.intArr.Len())
			assert.Equal(t, 10, arr.intArr.Get(0))
			assert.Equal(t, 100, arr.intArr.Get(1))
			assert.Equal(t, "World", arr.nameStr.String())
			assert.Equal(t, "World", arr.dynamicStr.String())
		}	
		iter.Next()
		assert.True(t, iter.IsNil())
		wg.Done()
	})

	entity := registry.Create()
	AddComponent[ArrayComponent](registry, entity, &ArrayComponent{
		intArr: NewArrayFromSlice([]int{10}),
		nameStr: NewName("Hello"),
		dynamicStr: NewString("Hello"),
	})

	registry.Tick(0.01)
	wg.Wait()
	wg.Add(1)
	registry.Tick(0.01)
	wg.Wait()
}

type OtherEntity struct {
	other Entity
}

func TestDependency(t *testing.T) {
	registry := New()
	defer registry.Free()

	var wg sync.WaitGroup
	called := false
	wg.Add(2)
	sys := AddSystem1[OtherEntity](registry, OnTick, func (r *Registry, iter *Iterator) {
		assert.False(t, called)
		other := Get[OtherEntity](iter)
		pos := Get[Position](r, other.other)
		assert.NotNil(t, pos)
		pos.X += 100
		pos.Y += 100
		iter.Next()
		assert.True(t, iter.IsNil())
		time.Sleep(time.Second)
		called = true
		wg.Done()
	})
	AddDependency[Position](sys)

	AddSystem1[Position](registry, OnTick, func (r *Registry, iter *Iterator) {
		assert.True(t, called)
		pos := Get[Position](iter)
		assert.Equal(t, float32(200), pos.X)
		assert.Equal(t, float32(200), pos.Y)
		iter.Next()
		assert.True(t, iter.IsNil())
		wg.Done()
	})

	entity := registry.Create()
	AddComponent[Position](registry, entity, &Position{
		X: 100,
		Y: 100,
	})

	entity2 := registry.Create()
	AddComponent[OtherEntity](registry, entity2, &OtherEntity{
		other: entity,
	})

	registry.Tick(0.01)
	wg.Wait()
}

func TestPriority(t *testing.T) {
	registry := New()
	defer registry.Free()

	var wg sync.WaitGroup
	called1 := false
	called2 := false
	wg.Add(3)
	AddSystem1[Position](registry, OnTick, func (r *Registry, iter *Iterator) {
		assert.True(t, called2)
		pos := Get[Position](iter)
		assert.Equal(t, float32(300), pos.X)
		assert.Equal(t, float32(300), pos.Y)
		iter.Next()
		assert.True(t, iter.IsNil())
		wg.Done()
	})

	sys := PostTask1[Position](registry, OnTick, func (r *Registry, iter *Iterator) {
		assert.True(t, called1)
		assert.False(t, called2)
		pos := Get[Position](iter)
		assert.Equal(t, float32(200), pos.X)
		assert.Equal(t, float32(200), pos.Y)
		pos.X += 100
		pos.Y += 100
		time.Sleep(time.Second)
		iter.Next()
		assert.True(t, iter.IsNil())
		called2 = true
		wg.Done()
	})
	sys.SetPriority(999)
	
	sys = PostTask1[Position](registry, OnTick, func (r *Registry, iter *Iterator) {
		// should first call
		assert.False(t, called1)
		pos := Get[Position](iter)
		pos.X += 100
		pos.Y += 100
		time.Sleep(time.Second)
		iter.Next()
		assert.True(t, iter.IsNil())
		called1 = true
		wg.Done()
	})
	sys.SetPriority(1)

	entity := registry.Create()
	AddComponent[Position](registry, entity, &Position{
		X: 100,
		Y: 100,
	})

	registry.Tick(0.01)
	wg.Wait()
	wg.Add(1)
	registry.Tick(0.01)
	wg.Wait()
}

func TestZeroSet(t *testing.T) {
	registry := New()
	defer registry.Free()

	var wg sync.WaitGroup
	wg.Add(1)
	called := false
	AddSystem1[Position](registry, OnTick, func (r *Registry, iter *Iterator) {
		var i int
		for ; !iter.IsNil(); iter.Next() {
			pos := Get[Position](iter)
			if i == 0 {
				assert.Nil(t, pos)
			} else {
				if !called {
					assert.NotNil(t, pos)
					Set[Position](r, iter.Entity(), nil)
				} else {
					assert.Nil(t, pos)
				}
			}
			i++
		}
		called = true
		wg.Done()
	})

	entity := registry.Create()
	AddComponent[Position](registry, entity, nil)

	entity = registry.Create()
	AddComponent[Position](registry, entity, &Position{
		X: 100,
		Y: 100,
	})
	registry.Tick(0.01)
	wg.Wait()
	wg.Add(1)
	registry.Tick(0.01)
	wg.Wait()
}

