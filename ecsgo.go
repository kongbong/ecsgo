package ecsgo

import (
	"fmt"
	"reflect"
	"unsafe"
)

// AddSystem add none component system
func AddSystem(r *Registry, time Ticktime, fn func (r *Registry)) isystem {
	sys := makeNonComponentSystem(r, time, false, fn)
	r.defferredAddsystem(time, sys)
	return sys
}

// PostTask add one-time called system - fn is called only one time
func PostTask(r *Registry, time Ticktime, fn func (r *Registry)) isystem {
	sys := makeNonComponentSystem(r, time, true, fn)
	r.defferredAddsystem(time, sys)
	return sys
}

// AddSystem1 add single value system
func AddSystem1[T any](r *Registry, time Ticktime, fn func (r *Registry, iter *Iterator)) isystem {
	sys := makeSystem1[T](r, time, false, fn)
	r.defferredAddsystem(time, sys)
	return sys
}

// PostTask1 add one-time called single component system - fn is called only one time
func PostTask1[T any](r *Registry, time Ticktime, fn func (r *Registry, iter *Iterator)) isystem {
	sys := makeSystem1[T](r, time, true, fn)
	r.defferredAddsystem(time, sys)
	return sys
}

// AddSystem2 add two values system
func AddSystem2[T1, T2 any](r *Registry, time Ticktime, fn func (r *Registry, iter *Iterator)) isystem {
	sys := makeSystem2[T1, T2](r, time, false, fn)
	r.defferredAddsystem(time, sys)
	return sys
}

// PostTask2 add one-time called two components system - fn is called only one time
func PostTask2[T1, T2 any](r *Registry, time Ticktime, fn func (r *Registry, iter *Iterator)) isystem {
	sys := makeSystem2[T1, T2](r, time, true, fn)
	r.defferredAddsystem(time, sys)
	return sys
}

// AddSystem3 add three values system
func AddSystem3[T1, T2, T3 any](r *Registry, time Ticktime, fn func (r *Registry, iter *Iterator)) isystem {
	sys := makeSystem3[T1, T2, T3](r, time, false, fn)
	r.defferredAddsystem(time, sys)
	return sys
}

// PostTask3 add one-time called three components system - fn is called only one time
func PostTask3[T1, T2, T3 any](r *Registry, time Ticktime, fn func (r *Registry, iter *Iterator)) isystem {
	sys := makeSystem3[T1, T2, T3](r, time, true, fn)
	r.defferredAddsystem(time, sys)
	return sys
}

// AddSystem4 add four values system
func AddSystem4[T1, T2, T3, T4 any](r *Registry, time Ticktime, fn func (r *Registry, iter *Iterator)) isystem {
	sys := makeSystem4[T1, T2, T3, T4](r, time, false, fn)
	r.defferredAddsystem(time, sys)
	return sys
}

// PostTask4 add one-time called three components system - fn is called only one time
func PostTask4[T1, T2, T3, T4 any](r *Registry, time Ticktime, fn func (r *Registry, iter *Iterator)) isystem {
	sys := makeSystem4[T1, T2, T3, T4](r, time, true, fn)
	r.defferredAddsystem(time, sys)
	return sys
}

// AddSystem5 add five values system
func AddSystem5[T1, T2, T3, T4, T5 any](r *Registry, time Ticktime, fn func (r *Registry, iter *Iterator)) isystem {
	sys := makeSystem5[T1, T2, T3, T4, T5](r, time, false, fn)
	r.defferredAddsystem(time, sys)
	return sys
}

// PostTask5 add one-time called three components system - fn is called only one time
func PostTask5[T1, T2, T3, T4, T5 any](r *Registry, time Ticktime, fn func (r *Registry, iter *Iterator)) isystem {
	sys := makeSystem5[T1, T2, T3, T4, T5](r, time, true, fn)
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
func AddDependency[T any](sys isystem) {
	var zeroT T
	sys.addDependencyType(reflect.TypeOf(zeroT))
}

// AddReadDependency add a dependency of T, system needs this when the system needs set/get other entity's other component value
func AddReadDependency[T any](sys isystem) {
	var zeroT T
	sys.addDependencyType(reflect.TypeOf(zeroT))
	Readonly[T](sys)
}

func Get[T any](args ...interface{}) *T {
	if len(args) == 1 {
		if iter, ok := args[0].(*Iterator); ok {
			return getIterValue[T](iter)
		} else {
			panic("Get() first argument should be iterator when one argument")
		}
	} else if len(args) == 2 {
		r, ok := args[0].(*Registry)
		if !ok {
			panic("Get() first argument should be *Registry when two arguments")
		}
		entity, ok := args[1].(Entity)
		if !ok {
			panic("Get() second argument should be Entity when two arguments")
		}
		return getEntityPtrValue[T](r, entity)
	}
	panic("Get() can be called with single argument or two arguments")
	return nil
}

func getIterValue[T any](i *Iterator) *T {
	var zeroT T
	typeT := reflect.TypeOf(zeroT)
	ptrT := (*T)(i.tabIter.get(typeT))
	if i.s.isReadonly(typeT) {
		// copy to temp data
		zeroT = *ptrT
		ptrT = &zeroT
	}
	return ptrT
}

// Get Gets component value
// This can be used for getting other entity's value
// The system called this function needs to add write Dependaency of T
func getEntityPtrValue[T any](r *Registry, entity Entity) *T {
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
func GetReadonly[T any](r *Registry, entity Entity) (T, error) {
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

// AddComponent add component into entity
func AddComponent[T any](r *Registry, entity Entity, v *T) {
	var zeroT T
	if err := checkType(reflect.TypeOf(zeroT)); err != nil {
		panic(err)
	}
	cmpInfo := &componentInfo{
		tp: reflect.TypeOf(zeroT),
		v: v,
		ptr: unsafe.Pointer(v),
	}
	r.defferredAddComponent(entity, cmpInfo)
}

// Set sets component value
// This can be used for setting other entity's value
// The system called this function needs to add write Dependaency of T
func Set[T any](r *Registry, entity Entity, v *T) {
	var zeroT T
	if err := checkType(reflect.TypeOf(zeroT)); err != nil {
		panic(err)
	}
	cmpInfo := &componentInfo{
		tp: reflect.TypeOf(zeroT),
		v: v,
		ptr: unsafe.Pointer(v),
	}
	r.storage.setValue(entity, cmpInfo)
}

// AddTag add a tag into entity, tag is same with empty size Component
func AddTag[T any](r *Registry, entity Entity) {
	var zeroT T
	if err := checkTagType(reflect.TypeOf(zeroT)); err != nil {
		panic(err)
	}
	tagInfo := &componentInfo{
		tp: reflect.TypeOf(zeroT),
	}
	r.defferredAddComponent(entity, tagInfo)
}

// HasTag whether entity has Tag or not
func HasTag[T any](r *Registry, entity Entity) bool {
	var zeroT T
	return r.storage.hasType(entity, reflect.TypeOf(zeroT))
}