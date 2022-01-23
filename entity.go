package ecsgo

type Entity struct {
	id      uint32
	version uint32
}

func newEntity(id uint32, version uint32) Entity {
	return Entity{id: id, version: version}
}

func (e *Entity) SetVersion(version uint32) {
	e.version = version
}
