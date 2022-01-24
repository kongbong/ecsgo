package main

import (
	"log"

	"github.com/kongbong/ecsgo"
)

type Position struct {
	X float32
	Y float32
}

type Velocity struct {
	X float32
	Y float32
}

func main() {
	registry := ecsgo.New()
	defer registry.Free()

	ecsgo.AddSystem1(registry, ecsgo.OnTick, func(r *ecsgo.Registry, entity ecsgo.Entity, pos *Position) {
		log.Println("Position system", entity, pos)
	})
	ecsgo.AddSystem1(registry, ecsgo.OnTick, func(r *ecsgo.Registry, entity ecsgo.Entity, vel *Velocity) {
		log.Println("Velocity system", entity, vel)
	})
	ecsgo.AddSystem2(registry, ecsgo.OnTick, func(r *ecsgo.Registry, entity ecsgo.Entity, pos *Position, vel *Velocity) {
		log.Println("Position, Velocity system", entity, pos, vel)
	})

	entity := registry.Create()
	ecsgo.AddComponent(registry, entity, &Position{10, 10})
	ecsgo.AddComponent(registry, entity, &Velocity{20, 20})

	registry.Run()
}
