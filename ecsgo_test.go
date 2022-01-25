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

func TestECSGo(t *testing.T) {
	registry := New()
	defer registry.Free()

	var called2 bool
	var wg sync.WaitGroup
	wg.Add(2)
	Exclude1[Velocity](
		AddSystem1[Position](registry, OnTick, func (r *Registry, entity Entity, pos *Position) {
			log.Println("Should not call this")
			assert.True(t, false)
		}),
	)
	Exclude1[HP](
		AddSystem1[Velocity](registry, OnTick, func (r *Registry, entity Entity, vel *Velocity) {
			log.Println("Velocity system")
			assert.Equal(t, vel.X, float32(10))
			assert.Equal(t, vel.Y, float32(10))
			time.Sleep(time.Second)
			called2 = true
			wg.Done()
			log.Println("Velocity system Done")
		}),
	)
	AddSystem2[Position, Velocity](registry, OnTick, func (r *Registry, entity Entity, pos *Position, vel *Velocity) {
		log.Println("Position, Velocity system")
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
