# ECSGo
ECSGo is an Entity Component System(ECS) in Go.
This is made with Generic Go, so it needs Go 1.18 version

- Cache friendly data storage
- Run systems in concurrently with analyzing dependency tree.


## Example
```go
package main

import (
	"context"
	"log"
	"time"

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
	registry := ecsgo.NewRegistry()

	sys1 := registry.AddSystem("VelocitySystem", 0, func(ctx *ecsgo.ExecutionContext) error {
		qr := ctx.GetQueryResult(0)
		log.Println("This system should have not any archtype", qr.GetArcheTypeCount())
		return nil
	})
	q1 := sys1.NewQuery()
	ecsgo.AddReadWriteComponent[Velocity](q1)
	ecsgo.AddExcludeComponent[EnemyTag](q1)

	o := registry.AddObserver("AddVelocityObserver", func(ctx *ecsgo.ObserverContext) error {
		vel := ecsgo.GetComponentObserver[Velocity](ctx)
		log.Println("This is one time called system", ctx.GetEntityId(), vel)
		return nil
	})
	ecsgo.AddComponentToObserver[Velocity](o)

	sys2 := registry.AddSystem("VelocitySystem2", 0, func(ctx *ecsgo.ExecutionContext) error {
		qr := ctx.GetQueryResult(0)
		qr.ForeachEntities(func(accessor *ecsgo.ArcheTypeAccessor) error {
			vel := ecsgo.GetComponentByAccessor[Velocity](accessor)
			log.Println("VelocitySystem2", accessor.GetEntityId(), vel)
			return nil
		})
		return nil
	})
	q2 := sys2.NewQuery()
	ecsgo.AddExcludeComponent[HP](q2)
	ecsgo.AddReadonlyComponent[Velocity](q2)

	sys3 := registry.AddSystem("PositionAndVelocity", 0, func(ctx *ecsgo.ExecutionContext) error {
		qr := ctx.GetQueryResult(0)
		qr.ForeachEntities(func(accessor *ecsgo.ArcheTypeAccessor) error {
			pos := ecsgo.GetComponentByAccessor[Position](accessor)
			vel := ecsgo.GetComponentByAccessor[Velocity](accessor)
			log.Println("Position, Velocity system", accessor.GetEntityId(), pos, vel, ctx.GetDeltaTime())
			pos.X++
			pos.Y++
			vel.X++
			vel.Y++
			return nil
		})
		return nil
	})
	q3 := sys3.NewQuery()
	ecsgo.AddReadWriteComponent[Position](q3)
	ecsgo.AddReadWriteComponent[Velocity](q3)
	ecsgo.AddReadonlyComponent[EnemyTag](q3)

	entity := registry.CreateEntity()
	ecsgo.AddComponent(registry, entity, Position{10, 10})
	ecsgo.AddComponent(registry, entity, Velocity{20, 20})
	ecsgo.AddComponent(registry, entity, EnemyTag{})

	ctx := context.Background()
	for i := 0; i < 10; i++ {
		registry.Tick(time.Second, ctx)
		time.Sleep(time.Second)
	}
}
```