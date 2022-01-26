package ecsgo

/*
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"reflect"
	"unsafe"
	
	"github.com/kongbong/ecsgo/sparseMap"
)

type unsafeTable struct {
	spMap *sparseMap.UnsafeMap[uint32]	
	typeOffMap map[reflect.Type]int
}

func newTable(types ...reflect.Type) *unsafeTable {
	tb := &unsafeTable{
		typeOffMap: make(map[reflect.Type]int),
	}

	var ent Entity
	dataSize := int(unsafe.Sizeof(ent))
	for _, t := range types {			
		tb.typeOffMap[t] = dataSize	
		dataSize += int(t.Size())
	}
	tb.spMap = sparseMap.NewUnsafeAutoIncresing[uint32](dataSize, 100)
	return tb
}

func (t *unsafeTable) free() {
	t.spMap.Free()
}

func (t *unsafeTable) Same(types ...reflect.Type) bool {
	if len(types) != len(t.typeOffMap) {
		return false
	}
	for _, tp := range types {
		if _, ok := t.typeOffMap[tp]; !ok {
			return false
		}
	}
	return true
}

// hasType type check if it has specific type value
func (t *unsafeTable) hasType(tp reflect.Type) bool {	
	_, ok := t.typeOffMap[tp]
	return ok
}

func (t *unsafeTable) erase(ent Entity) {
	t.spMap.Erase(ent.id)
}

func (t *unsafeTable) find(ent Entity) *typeGetter {
	ptr, _ := t.spMap.Find(ent.id)
	if *((*Entity)(ptr)) != ent {
		return nil
	}
	return &typeGetter{
		ptr: ptr,
		typeOffMap: t.typeOffMap,
	}
}

func (t *unsafeTable) insert(ent Entity, valMap map[reflect.Type]unsafe.Pointer) {
	sz := t.spMap.ValueSize()
	ptr := unsafe.Pointer(C.malloc(C.size_t(sz)))
	C.memcpy(ptr, unsafe.Pointer(&ent), C.size_t(unsafe.Sizeof(ent)))

	for tp, off := range t.typeOffMap {
		if tp.Size() > 0 {
			C.memcpy(unsafe.Add(ptr, off), valMap[tp], C.size_t(tp.Size()))
		}
	}

	success := t.spMap.Insert(ent.id, ptr)
	if !success {
		panic("failed to insert")
	}
	C.free(ptr)
}

type typeGetter struct {
	ptr unsafe.Pointer
	typeOffMap map[reflect.Type]int
}

func (g *typeGetter) get(t reflect.Type) unsafe.Pointer {
	if uintptr(g.ptr) == 0 {
		panic("ptr is nil")
	}
	off, ok := g.typeOffMap[t]
	if !ok {
		panic("table doesn't have the type")
	}
	return unsafe.Add(g.ptr, off)
}

// iterator get iterator
func (t *unsafeTable) iterator() *tableIter {
	return newTableIter(t.spMap.Iterate(), t.typeOffMap)
}

// iiter iterator interface
type tableIter struct {
	iter *sparseMap.UnsafeIter
	getter typeGetter
}

func newTableIter(iter *sparseMap.UnsafeIter, typeOffMap map[reflect.Type]int) *tableIter {
	return &tableIter{
		iter: iter,
		getter: typeGetter{
			ptr: iter.Get(),
			typeOffMap: typeOffMap,
		},
	}
}

func (i *tableIter) isNil() bool {
	return i.iter.IsNil()
}

func (i *tableIter) next() {
	if i.isNil() {
		panic("current is nil")
	}
	i.iter.Next()
	i.getter.ptr = i.iter.Get()
}

func (i *tableIter) get(t reflect.Type) unsafe.Pointer {
	if i.isNil() {
		panic("current is nil")
	}
	return i.getter.get(t)
}

func (i *tableIter) entity() Entity {
	if i.isNil() {
		panic("current is nil")
	}
	return *((*Entity)(i.getter.ptr))
}
