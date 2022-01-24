package ecsgo

import (
	"reflect"
	"unsafe"

	"github.com/kongbong/ecsgo/sparseMap"
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

func (s *storage) query(types ...reflect.Type) []*unsafeTable {
	var rst []*unsafeTable

	for _, t := range s.tables {
		found := true
		for _, tp := range types {
			if !t.hasType(tp) {
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
	sortTypes(types)
	var ent Entity
	dataSize := int(unsafe.Sizeof(ent))
	tb := &unsafeTable{}
	for _, t := range types {
		column := &columnDesc{
			dataType: t,
			offset:   dataSize,
		}
		dataSize += int(t.Size())
		tb.columns = append(tb.columns, column)
	}
	tb.spMap = sparseMap.NewUnsafeAutoIncresing[uint32](dataSize, 100)
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
	if tb, ok := s.entityMap[ent]; ok {
		getter := tb.find(ent)
		if getter == nil {
			// already removed, weird
			panic("entity data is removed")
		}
		for _, c := range getter.columns {
			valMap[c.dataType] = unsafe.Add(getter.ptr, c.offset)
			types = append(types, c.dataType)
		}
	}

	for _, c := range cmpInfos {
		valMap[c.dataType] = c.ptr
		types = append(types, c.dataType)
	}

	tb := s.getOrAddTable(types...)
	tb.insert(ent, valMap)

}
