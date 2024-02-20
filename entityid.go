package ecsgo

type EntityId struct {
	id      uint32
	version uint32
}

func (e EntityId) NotNil() bool {
	return e.id != 0 && e.version != 0
}
