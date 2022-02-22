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

type flag uint8
const (
	Alive flag = 0
	Deleted flag = 1
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
		dataSize += int(t.Size()+1) // for flag
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

func (t *unsafeTable) find(ent Entity) *typeGetAndSet {
	ptr, _ := t.spMap.Find(ent.id)
	if *((*Entity)(ptr)) != ent {
		return nil
	}
	return &typeGetAndSet{
		ptr: ptr,
		typeOffMap: t.typeOffMap,
	}
}

func memCopy(dst, src unsafe.Pointer, size uintptr) {
	if uintptr(src) == 0 {
		C.memset(dst, C.int(Deleted), C.size_t(1))
	} else {
		C.memset(dst, C.int(Alive), C.size_t(1))
		C.memcpy(unsafe.Add(dst, 1), src, C.size_t(size))
	}
}

func (t *unsafeTable) insert(ent Entity, valMap map[reflect.Type]unsafe.Pointer) {
	sz := t.spMap.ValueSize()
	ptr := unsafe.Pointer(C.malloc(C.size_t(sz)))
	C.memcpy(ptr, unsafe.Pointer(&ent), C.size_t(unsafe.Sizeof(ent)))

	for tp, off := range t.typeOffMap {
		if tp.Size() > 0 {
			memCopy(unsafe.Add(ptr, off), valMap[tp], tp.Size())
		}
	}

	success := t.spMap.Insert(ent.id, ptr)
	if !success {
		panic("failed to insert")
	}
	C.free(ptr)
}

type typeGetAndSet struct {
	ptr unsafe.Pointer
	typeOffMap map[reflect.Type]int
}

func (g *typeGetAndSet) getPtr(t reflect.Type) unsafe.Pointer {
	if uintptr(g.ptr) == 0 {
		panic("ptr is nil")
	}
	off, ok := g.typeOffMap[t]
	if !ok {
		panic("table doesn't have the type")
	}
	return unsafe.Add(g.ptr, off)
}

func (g *typeGetAndSet) get(t reflect.Type) unsafe.Pointer {
	ptr := g.getPtr(t)
	flag := flag(*(*uint8)(ptr))
	if flag == Deleted {
		return nil
	}
	return unsafe.Add(ptr, 1)
}

func (g *typeGetAndSet) set(t reflect.Type, val unsafe.Pointer) {
	ptr := g.getPtr(t)
	memCopy(ptr, val, t.Size())
}

// iterator get iterator
func (t *unsafeTable) iterator() *tableIter {
	return newTableIter(t.spMap.Iterate(), t.typeOffMap)
}

// iiter iterator interface
type tableIter struct {
	iter *sparseMap.UnsafeIter
	getter typeGetAndSet
}

func newTableIter(iter *sparseMap.UnsafeIter, typeOffMap map[reflect.Type]int) *tableIter {
	return &tableIter{
		iter: iter,
		getter: typeGetAndSet{
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
