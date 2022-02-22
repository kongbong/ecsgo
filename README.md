# ECSGo
ECSGo is an Entity Component System(ECS) in Go.
This is made with Generic Go, so it needs Go 1.18 version

- Cache friendly data storage
- Run systems in concurrently with analyzing dependency tree.


## Example
```go
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

    sys := ecsgo.AddSystem1[Velocity](registry, ecsgo.OnTick, func(r *ecsgo.Registry, iter *ecsgo.Iterator) {
        log.Println("This should not called as Entity has Velocity component")
    })
    ecsgo.Exclude[EnemyTag](sys)

    ecsgo.PostTask1[Velocity](registry, ecsgo.OnTick, func(r *ecsgo.Registry, iter *ecsgo.Iterator) {
        for ; !iter.IsNil(); iter.Next() {
            vel := ecsgo.Get[Velocity](iter)
            log.Println("This is one time called system", iter.Entity(), vel)
        }
    })

    sys = ecsgo.AddSystem1[Velocity](registry, ecsgo.OnTick, func(r *ecsgo.Registry, iter *ecsgo.Iterator) {
        for ; !iter.IsNil(); iter.Next() {
            vel := ecsgo.Get[Velocity](iter)
            log.Println("Velocity system", iter.Entity(), vel)
            // Velocity value of Entity is not changed as it is Readonly
            vel.X++
            vel.Y++
        }
    })
    ecsgo.Exclude[HP](sys)
    ecsgo.Readonly[Velocity](sys)

    sys = ecsgo.AddSystem2[Position, Velocity](registry, ecsgo.OnTick, func(r *ecsgo.Registry, iter *ecsgo.Iterator) {
        for ; !iter.IsNil(); iter.Next() {
            pos := ecsgo.Get[Position](iter)
            vel := ecsgo.Get[Velocity](iter)
            log.Println("Position, Velocity system", iter.Entity(), pos, vel, r.DeltaSeconds())
            // Position value is not changed only Velocity value is changed
            pos.X++
            pos.Y++
            vel.X++
            vel.Y++
        }
    })
    sys.SetTickInterval(1)
    ecsgo.Tag[EnemyTag](sys)
    ecsgo.Readonly[Position](sys)

    entity := registry.Create()
    ecsgo.AddComponent(registry, entity, &Position{10, 10})
    ecsgo.AddComponent(registry, entity, &Velocity{20, 20})
    ecsgo.AddTag[EnemyTag](registry, entity)

    registry.Run(ecsgo.FPS(10), ecsgo.FixedTick(true))
}
```