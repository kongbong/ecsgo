package ecsgo

import (
	"unsafe"
	"reflect"
	"fmt"
)

// SetEntityComponent1 store entity data into storage
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

func checkType(tp reflect.Type) error {
	if tp.Size() == 0 {
		// empty size is not allowed
		return fmt.Errorf("Empty size component is not allowed. You can use Tag instead Component")
	}
	switch tp.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.String, reflect.UnsafePointer:
		return fmt.Errorf("Component type should be not reference type %v", tp)
	case reflect.Struct:
		// need to check every fields
		for i := 0; i < tp.NumField(); i++ {
			err := checkType(tp.Field(i).Type)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func checkTagType(tp reflect.Type) error {
	if tp.Size() != 0 {
		// empty size is not allowed
		return fmt.Errorf("Tag type should be empty type")
	}
	return nil
}

type componentInfo struct {
	tp reflect.Type
	v interface{}
	ptr unsafe.Pointer
}
