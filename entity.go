package ecsgo

type Entity uint32

// 4bytes EntityId + 4bytes Version
type EntityVer uint64

const entityMask = 0xffffffff00000000
const versinMask = 0x00000000ffffffff
const entityShift = 32

func newEntity(id Entity, version uint32) EntityVer {
	return (EntityVer(id) << entityShift) | EntityVer(version)
}

func (e EntityVer) ToEntity() Entity {
	return Entity((e & entityMask) >> entityShift)
}

func (e EntityVer) ToVersion() uint32 {
	return uint32(e & versinMask)
}

func (e *EntityVer) SetVersion(version uint32) {
	*e = (*e & entityMask) | EntityVer(version)
}

func SetEntityComponent1[T any](r *Registry, entity EntityVer, v *T) {
	t := getOrAddTable1[T](r)
	t.insert(entity, v)
}

func SetEntityComponent2[T any, U any](r *Registry, entity EntityVer, v1 *T, v2 *U) {
	t := getOrAddTable2[T, U](r)
	t.insert(entity, v1, v2)
}