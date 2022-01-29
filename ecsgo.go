package ecsgo

import (
	"fmt"
	"reflect"
	"unsafe"
)

// AddSystem add none component system
func AddSystem(r *Registry, time Ticktime, fn func (r *Registry)) isystem {
	sys := makeSystem(r, time, false, fn)
	r.defferredAddsystem(time, sys)
	return sys
}

// PostTask add one-time called system - fn is called only one time
func PostTask(r *Registry, time Ticktime, fn func (r *Registry)) isystem {
	sys := makeSystem(r, time, true, fn)
	r.defferredAddsystem(time, sys)
	return sys
}

// AddSystem1 add single value system
func AddSystem1[T any](r *Registry, time Ticktime, fn func (r *Registry, entity Entity, t *T)) isystem {
	sys := makeSystem1[T](r, time, false, fn)
	r.defferredAddsystem(time, sys)
	return sys
}

// PostTask1 add one-time called single component system - fn is called only one time
func PostTask1[T any](r *Registry, time Ticktime, fn func (r *Registry, entity Entity, t *T)) isystem {
	sys := makeSystem1[T](r, time, true, fn)
	r.defferredAddsystem(time, sys)
	return sys
}

// AddSystem2 add two values system
func AddSystem2[T any, U any](r *Registry, time Ticktime, fn func (r *Registry, entity Entity, t *T, u *U)) isystem {
	sys := makeSystem2[T, U](r, time, false, fn)
	r.defferredAddsystem(time, sys)
	return sys
}

// PostTask2 add one-time called two components system - fn is called only one time
func PostTask2[T any, U any](r *Registry, time Ticktime, fn func (r *Registry, entity Entity, t *T, u *U)) isystem {
	sys := makeSystem2[T, U](r, time, true, fn)
	r.defferredAddsystem(time, sys)
	return sys
}

func Exclude[T any](sys isystem) {
	var zeroT T
	sys.addExcludeType(reflect.TypeOf(zeroT))
}

func Tag[T any](sys isystem) {
	var zeroT T
	if err := checkTagType(reflect.TypeOf(zeroT)); err != nil {
		panic(err)
	}
	sys.addTagType(reflect.TypeOf(zeroT))
}

func Readonly[T any](sys isystem) {
	var zeroT T
	sys.makeReadonly(reflect.TypeOf(zeroT))
}

// AddDependency add a dependency of T, system needs this when the system needs set/get other entity's other component value
func AddDependency[T any](sys isystem, readonly bool) {
	var zeroT T
	sys.addDependencyType(reflect.TypeOf(zeroT))
	if readonly {
		Readonly[T](sys)
	}
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

// Set sets component value
// This can be used for setting other entity's value
// The system called this function needs to add write Dependaency of T
func Set[T any](r *Registry, entity Entity, v *T) error {
	if err := checkType(reflect.TypeOf(*v)); err != nil {
		return err
	}
	cmpInfo := &componentInfo{
		tp: reflect.TypeOf(*v),
		v: v,
		ptr: unsafe.Pointer(v),
	}
	r.storage.setValue(entity, cmpInfo)
	return nil
}

// Get Gets component value
// This can be used for getting other entity's value
// The system called this function needs to add write Dependaency of T
func Get[T any](r *Registry, entity Entity) *T {
	var zeroT T
	tp := reflect.TypeOf(zeroT)
	if err := checkType(tp); err != nil {
		return nil
	}
	ptr := r.storage.getValue(entity, tp)
	return (*T)(ptr)
}

// GetValue Gets component value (readonly)
// This can be used for getting other entity's value
// The system called this function needs to add read Dependaency of T
func GetValue[T any](r *Registry, entity Entity) (T, error) {
	var zeroT T
	tp := reflect.TypeOf(zeroT)
	if err := checkType(tp); err != nil {
		return zeroT, err
	}
	ptr := r.storage.getValue(entity, tp)
	if ptr != nil {
		zeroT = (*(*T)(ptr))
		return zeroT, nil
	}
	return zeroT, fmt.Errorf("entity data is removed")
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