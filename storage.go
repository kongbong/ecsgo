package ecsgo

import (
	"reflect"
	"unsafe"
)

// storage has tables
type storage struct {
	tables    []*unsafeTable
	entityMap map[Entity]*unsafeTable
}

func newStorage() *storage {
	return &storage{
		entityMap: make(map[Entity]*unsafeTable),
	}
}

func (s *storage) free() {
	for _, t := range s.tables {
		t.free()
	}
}

func (s *storage) query(includeTypes []reflect.Type, excludeTypes []reflect.Type) []*unsafeTable {
	var rst []*unsafeTable

	for _, t := range s.tables {
		found := true
		for _, tp := range includeTypes {
			if !t.hasType(tp) {
				found = false
				break
			}
		}
		if !found {
			continue
		}
		for _, tp := range excludeTypes {
			if t.hasType(tp) {
				found = false
				break
			}
		}
		if found {
			rst = append(rst, t)
		}
	}
	return rst
}

// getOrAddTable making single value table
func (s *storage) getOrAddTable(types ...reflect.Type) *unsafeTable {
	for _, t := range s.tables {
		if t.Same(types...) {
			return t
		}
	}
	tb := newTable(types...)
	s.tables = append(s.tables, tb)
	return tb
}

func (s *storage) eraseEntity(ent Entity) {
	if tb, ok := s.entityMap[ent]; ok {
		tb.erase(ent)
	}
	delete(s.entityMap, ent)
}

func (s *storage) addComponents(ent Entity, cmpInfos []*componentInfo) {
	valMap := make(map[reflect.Type]unsafe.Pointer)
	types := []reflect.Type{}
	prevTb, ok := s.entityMap[ent]
	if ok {
		getter := prevTb.find(ent)
		if getter == nil {
			// already removed, weird
			panic("entity data is removed")
		}
		for tp, off := range getter.typeOffMap {
			valMap[tp] = unsafe.Add(getter.ptr, off+1) // Add flag offset
			types = append(types, tp)
		}
	}

	for _, c := range cmpInfos {
		valMap[c.tp] = c.ptr
		types = append(types, c.tp)
	}

	tb := s.getOrAddTable(types...)
	tb.insert(ent, valMap)
	if prevTb != nil {
		prevTb.erase(ent)
	}
	s.entityMap[ent] = tb
}

func (s *storage) setValue(ent Entity, c *componentInfo) {
	prevTb, ok := s.entityMap[ent]
	if !ok {
		panic("Entity doesn't have any components")
	}
	getter := prevTb.find(ent)
	if getter == nil {
		// already removed, weird
		panic("entity data is removed")
	}
	if c.tp.Size() > 0 {
		getter.set(c.tp, c.ptr)
	}
}

func (s *storage) getValue(ent Entity, tp reflect.Type) unsafe.Pointer {
	prevTb, ok := s.entityMap[ent]
	if !ok {
		panic("Entity doesn't have any components")
	}
	getter := prevTb.find(ent)
	if getter == nil {
		// already removed, weird
		panic("entity data is removed")
	}
	if tp.Size() > 0 {
		return getter.get(tp)
	}
	return nil
}

func (s *storage) hasType(ent Entity, tp reflect.Type) bool {
	tb, ok := s.entityMap[ent]
	if !ok {
		panic("Entity doesn't have any components")
	}
	return tb.hasType(tp)
}
