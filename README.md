# ECSGo
ECSGo is an Entity Component System(ECS) in Go.
This is made with Generic Go, so it needs Go 1.18 version

- Cache friendly data storage
- Run systems in concurrently with analyzing dependency tree.


## Example
```go
package main

import (
    "github.com/kongbong/ecsgo"
    "log"
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

    ecsgo.AddSystem1(registry, func (entity ecsgo.Entity, pos *Position) {		
        log.Println("Position system Done")
    })

    ecsgo.AddSystem1(registry, func (entity ecsgo.Entity, vel *Velocity) {
        log.Println("Velocity system Done")
    })

    ecsgo.AddSystem2(registry, func (entity ecsgo.Entity, pos *Position, vel *Velocity) {
        log.Println("Position, Velocity system")		
    })

    entity := registry.Create()
    ecsgo.SetEntityComponent2(registry, entity, &Position{10, 10}, &Velocity{10, 10})

    registry.Run()
}
```