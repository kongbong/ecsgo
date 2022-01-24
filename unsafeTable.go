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
	columns []*columnDesc
}

type columnDesc struct {
	dataType reflect.Type
	offset int
}

func (t *unsafeTable) free() {
	t.spMap.Free()
}

func (t *unsafeTable) Same(types ...reflect.Type) bool {
	if len(types) != len(t.columns) {
		return false
	}
	sortTypes(types)
	for i, c := range t.columns {
		if types[i] != c.dataType {
			return false
		}
	}
	return true
}

// hasType type check if it has specific type value
func (t *unsafeTable) hasType(tp reflect.Type) bool {
	for _, c := range t.columns {
		if tp == c.dataType {
			return true
		}
	}
	return false
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
		columns: t.columns,
	}
}

func (t *unsafeTable) insert(ent Entity, valMap map[reflect.Type]unsafe.Pointer) {
	sz := t.spMap.ValueSize()
	ptr := unsafe.Pointer(C.malloc(C.size_t(sz)))
	C.memcpy(ptr, unsafe.Pointer(&ent), C.size_t(unsafe.Sizeof(ent)))

	for _, c := range t.columns {
		C.memcpy(unsafe.Add(ptr, c.offset), valMap[c.dataType], C.size_t(c.dataType.Size()))
	}

	success := t.spMap.Insert(ent.id, ptr)
	if !success {
		panic("failed to insert")
	}
	C.free(ptr)
}

type typeGetter struct {
	ptr unsafe.Pointer
	columns []*columnDesc
}

func (g *typeGetter) get(t reflect.Type) unsafe.Pointer {
	if uintptr(g.ptr) == 0 {
		panic("ptr is nil")
	}
	for _, c := range g.columns {
		if c.dataType == t {
			return unsafe.Add(g.ptr, c.offset)
		}
	}
	return nil
}

// iterator get iterator
func (t *unsafeTable) iterator() *tableIter {
	return newTableIter(t.spMap.Iterate(), t.columns)
}

// iiter iterator interface
type tableIter struct {
	iter *sparseMap.UnsafeIter
	getter typeGetter
}

func newTableIter(iter *sparseMap.UnsafeIter, columns []*columnDesc) *tableIter {
	return &tableIter{
		iter: iter,
		getter: typeGetter{
			ptr: iter.Get(),
			columns: columns,
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
