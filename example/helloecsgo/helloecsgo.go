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

type HP struct {
	Hp    float32
	MaxHp float32
}

type EnemyTag struct{}

func main() {
	registry := ecsgo.New()
	defer registry.Free() // need to call before remove registry to free C malloc memory

	sys := ecsgo.AddSystem1(registry, ecsgo.OnTick, func(r *ecsgo.Registry, entity ecsgo.Entity, pos *Position) {
		log.Println("This should not called as Entity has Velocity component")
	})
	ecsgo.Exclude[EnemyTag](sys)

	sys = ecsgo.AddSystem1(registry, ecsgo.OnTick, func(r *ecsgo.Registry, entity ecsgo.Entity, vel *Velocity) {
		log.Println("Velocity system", entity, vel)
		vel.X++
		vel.Y++
	})
	ecsgo.Exclude[HP](sys)
	ecsgo.Readonly[Velocity](sys)

	sys = ecsgo.AddSystem2(registry, ecsgo.OnTick, func(r *ecsgo.Registry, entity ecsgo.Entity, pos *Position, vel *Velocity) {
		log.Println("Position, Velocity system", entity, pos, vel)
		pos.X++
		pos.Y++
		vel.X++
		vel.Y++
	})
	ecsgo.Tag[EnemyTag](sys)
	ecsgo.Readonly[Position](sys)

	entity := registry.Create()
	ecsgo.AddComponent(registry, entity, &Position{10, 10})
	ecsgo.AddComponent(registry, entity, &Velocity{20, 20})
	ecsgo.AddTag[EnemyTag](registry, entity)

	registry.Run()
}
