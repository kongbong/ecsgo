package ecsgo

import (
	"unsafe"
	"reflect"
)

// SetEntityComponent1 store entity data into storage
func AddComponent[T any](r *Registry, entity Entity, v *T) {
	cmpInfo := &componentInfo{
		dataType: reflect.TypeOf(*v),
		v: v,
		ptr: unsafe.Pointer(v),
	}
	r.defferredAddComponent(entity, cmpInfo)
	//t := getOrAddTable(r, reflect.TypeOf(*v))
	//t.insert(entity, v)
}

type componentInfo struct {
	dataType reflect.Type
	v interface{}
	ptr unsafe.Pointer
}

func processDeferredComponent(r *Registry) {
	for ent, cmps := range r.defferredCmp {
		r.storage.addComponents(ent, cmps)
	}
	r.defferredCmp = make(map[Entity][]*componentInfo)
}