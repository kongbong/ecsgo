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

func TestECSGo(t *testing.T) {
	registry := New()
	defer registry.Free()

	var called2 bool
	var wg sync.WaitGroup
	wg.Add(3)
	sys := AddSystem1[Position](registry, OnTick, func (r *Registry, entity Entity, pos *Position) {
		log.Println("Should not call this")
		assert.True(t, false)
	})
	Exclude[Velocity](sys)
	Exclude[EnemyTag](sys)
	
	sys = AddSystem1[Velocity](registry, OnTick, func (r *Registry, entity Entity, vel *Velocity) {
		log.Println("Velocity system")
		assert.Equal(t, vel.X, float32(10))
		assert.Equal(t, vel.Y, float32(10))
		time.Sleep(time.Second)
		called2 = true
		wg.Done()
		log.Println("Velocity system Done")
	})
	Exclude[HP](sys)
	Exclude[EnemyTag](sys)

	sys = AddSystem2[Position, Velocity](registry, OnTick, func (r *Registry, entity Entity, pos *Position, vel *Velocity) {
		log.Println("Position, Velocity system")
		assert.True(t, called2)
		assert.Equal(t, pos.X, float32(10))
		assert.Equal(t, pos.Y, float32(10))
		assert.Equal(t, vel.X, float32(10))
		assert.Equal(t, vel.Y, float32(10))
		wg.Done()
	})
	Exclude[EnemyTag](sys)
	
	sys = AddSystem1[Position](registry, OnTick, func (r *Registry, entity Entity, pos *Position) {
		log.Println("Position system With Enemy Tag")
		assert.Equal(t, pos.X, float32(100))
		assert.Equal(t, pos.Y, float32(100))
		wg.Done()
	})
	Tag[EnemyTag](sys)

	entity := registry.Create()
	AddComponent[Position](registry, entity, &Position{10, 10})
	AddComponent[Velocity](registry, entity, &Velocity{10, 10})
	err := AddComponent[ReferenceFieldComponent](registry, entity, &ReferenceFieldComponent{})
	assert.NotNil(t, err)
	err = AddComponent[EmptySizeComponent](registry, entity, &EmptySizeComponent{})
	assert.NotNil(t, err)

	entity = registry.Create()
	AddComponent[Position](registry, entity, &Position{100, 100})
	AddTag[EnemyTag](registry, entity)

	registry.tick(0.01)
	wg.Wait()
}


func TestECSGoReadOnly(t *testing.T) {
	registry := New()
	defer registry.Free()

	var called1 bool
	var wg sync.WaitGroup
	wg.Add(3)
	sys := AddSystem1[Position](registry, OnTick, func (r *Registry, entity Entity, pos *Position) {
		log.Println("First Position system")
		time.Sleep(time.Second)
		called1 = true
		wg.Done()
	})
	Readonly[Position](sys)

	sys = AddSystem1[Position](registry, OnTick, func (r *Registry, entity Entity, pos *Position) {
		log.Println("Second Position system")
		assert.False(t, called1)
		wg.Done()
	})
	Readonly[Position](sys)

	AddSystem2[Position, Velocity](registry, OnTick, func (r *Registry, entity Entity, pos *Position, vel *Velocity) {
		log.Println("Position, Velocity system")
		assert.True(t, called1)
		wg.Done()
	})

	entity := registry.Create()
	AddComponent[Position](registry, entity, &Position{10, 10})
	AddComponent[Velocity](registry, entity, &Velocity{10, 10})

	registry.tick(0.01)
	wg.Wait()
}

func TestECSGoPostTask(t *testing.T) {
	registry := New()
	defer registry.Free()

	var wg sync.WaitGroup
	var called1 bool
	wg.Add(2)
	PostTask1[Position](registry, OnTick, func (r *Registry, entity Entity, pos *Position) {
		log.Println("PostTask1 Position - this is called only one time")
		assert.False(t, called1)
		called1 = true
		wg.Done()
	})

	AddSystem2[Position, Velocity](registry, OnTick, func (r *Registry, entity Entity, pos *Position, vel *Velocity) {
		log.Println("Position, Velocity system - This is called every tick")
		assert.True(t, called1)
		wg.Done()
	})

	entity := registry.Create()
	AddComponent[Position](registry, entity, &Position{10, 10})
	AddComponent[Velocity](registry, entity, &Velocity{10, 10})

	registry.tick(0.01)
	wg.Wait()
	wg.Add(1)
	registry.tick(0.01)
	wg.Wait()
}