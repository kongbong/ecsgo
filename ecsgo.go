package ecsgo

import (
	"unsafe"
	"reflect"
)

// AddSystem add none component system
func AddSystem(r *Registry, time Ticktime, fn func (r *Registry)) isystem {
	sys := makeSystem(r, fn)
	r.defferredAddsystem(time, sys)
	return sys
}

// AddSystem1 add single value system
func AddSystem1[T any](r *Registry, time Ticktime, fn func (r *Registry, entity Entity, t *T)) isystem {
	sys := makeSystem1[T](r, time, fn)
	r.defferredAddsystem(time, sys)
	return sys
}

// AddSystem2 add two values system
func AddSystem2[T any, U any](r *Registry, time Ticktime, fn func (r *Registry, entity Entity, t *T, u *U)) isystem {
	sys := makeSystem2[T, U](r, time, fn)
	r.defferredAddsystem(time, sys)
	return sys
}

func Exclude[T any](sys isystem) {
	var zeroT T
	sys.addExcludeTypes(reflect.TypeOf(zeroT))
}

func Tag[T any](sys isystem) {
	var zeroT T
	if err := checkTagType(reflect.TypeOf(zeroT)); err != nil {
		panic(err)
	}
	sys.addIncludeTypes(reflect.TypeOf(zeroT), true)
}

func Readonly[T any](sys isystem) {
	var zeroT T
	sys.makeReadonly(reflect.TypeOf(zeroT))
}

// AddComponent add component into entity
func AddComponent[T any](r *Registry, entity Entity, v *T) error {
	if err := checkType(reflect.TypeOf(*v)); err != nil {
		return err
	}
	cmpInfo := &componentInfo{
		tp: reflect.TypeOf(*v),
		v: v,
		ptr: unsafe.Pointer(v),
	}
	r.defferredAddComponent(entity, cmpInfo)
	return nil
}

// AddTag add a tag into entity, tag is same with empty size Component
func AddTag[T any](r *Registry, entity Entity) error {
	var zeroT T
	if err := checkTagType(reflect.TypeOf(zeroT)); err != nil {
		return err
	}
	tagInfo := &componentInfo{
		tp: reflect.TypeOf(zeroT),
	}
	r.defferredAddComponent(entity, tagInfo)
	return nil
}