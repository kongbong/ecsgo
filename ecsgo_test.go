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

func TestECSGo(t *testing.T) {
	registry := New()
	defer registry.Free()

	var called1 bool
	var called2 bool
	var wg sync.WaitGroup
	wg.Add(3)
	AddSystem1[Position](registry, onTick, func (r *Registry, entity Entity, pos *Position) {
		log.Println("Position system")
		assert.Equal(t, pos.X, float32(10))
		assert.Equal(t, pos.Y, float32(10))
		time.Sleep(time.Second)
		called1 = true
		wg.Done()
		log.Println("Position system Done")
	})
	AddSystem1[Velocity](registry, onTick, func (r *Registry, entity Entity, vel *Velocity) {
		log.Println("Velocity system")
		assert.False(t, called1)
		assert.Equal(t, vel.X, float32(10))
		assert.Equal(t, vel.Y, float32(10))
		time.Sleep(time.Second)
		called2 = true
		wg.Done()
		log.Println("Velocity system Done")
	})
	AddSystem2[Position, Velocity](registry, onTick, func (r *Registry, entity Entity, pos *Position, vel *Velocity) {
		log.Println("Position, Velocity system")
		assert.True(t, called1)
		assert.True(t, called2)
		assert.Equal(t, pos.X, float32(10))
		assert.Equal(t, pos.Y, float32(10))
		assert.Equal(t, vel.X, float32(10))
		assert.Equal(t, vel.Y, float32(10))
		wg.Done()
	})

	entity := registry.Create()
	AddComponent[Position](registry, entity, &Position{10, 10})
	AddComponent[Velocity](registry, entity, &Velocity{10, 10})

	registry.Run()
	wg.Wait()
}
