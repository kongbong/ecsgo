package ecsgo

import (
	"fmt"
	"reflect"
	"unsafe"
)

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
	tp  reflect.Type
	v   interface{}
	ptr unsafe.Pointer
}
