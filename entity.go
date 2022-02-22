package ecsgo

type Entity struct {
	id      uint32
	version uint32
}

var EntityNil = Entity{0, 0}

func newEntity(id uint32, version uint32) Entity {
	return Entity{id: id, version: version}
}

func (e *Entity) setVersion(version uint32) {
	e.version = version
}
