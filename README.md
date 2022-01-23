# ECSGo
ECSGo is an Entity Component System(ECS) in Go.
This is made with Generic Go, so it needs Go 1.18 version

- Cache friendly data storage
- Run systems in concurrently with analyzing dependency tree.


## Example
```go
package main

import (
    "kongbong/ecsgo"
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
    registry := New()

    AddSystem1[Position](registry, func (entity EntityVer, pos *Position) {		
		log.Println("Position system Done")
	})

    AddSystem1[Velocity](registry, func (entity EntityVer, vel *Velocity) {
        log.Println("Velocity system Done")
	})
	AddSystem2[Position, Velocity](registry, func (entity EntityVer, pos *Position, vel *Velocity) {
        log.Println("Position, Velocity system")		
	})

	entity := registry.Create()
	SetEntityComponent2[Position, Velocity](registry, entity, &Position{10, 10}, &Velocity{10, 10})

	registry.Run()
}
```