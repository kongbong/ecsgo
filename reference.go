package ecsgo

import (
	"sync/atomic"
	"sync"
)

// reference reference type for ecsgo component array variable
// reference is stored in sync.Map for preventing GC
// Component cannot have reference variable directly because it is not fixed size and no guarantee not GC 
type reference[T any] struct {
	idx int64
}

var refMap sync.Map
var lastRefIdx int64

func newRef[T any](v *T) reference[T] {
	idx := atomic.AddInt64(&lastRefIdx, 1)
	refMap.Store(idx, v)
	return reference[T]{
		idx: idx,
	}
}

func (r reference[T]) get() *T {
	if val, ok := refMap.Load(r.idx); ok {
		return val.(*T)
	} else {
		panic("no reference in the map")
	}
	return nil
}

func (r *reference[T]) free() {
	refMap.Delete(r.idx)
	r.idx = 0
}

type Array[T any] reference[[]T]

func NewArray[T any](ln int) Array[T] {
	arr := make([]T, ln)
	return Array[T](newRef[[]T](&arr))
}

func NewArrayWithCap[T any](ln int, cap int) Array[T] {
	arr := make([]T, ln, cap)
	return Array[T](newRef[[]T](&arr))
}

func NewArrayFromSlice[T any](slice []T) Array[T] {
	return Array[T](newRef[[]T](&slice))
}

func (a Array[T]) Len() int {
	arr := (reference[[]T](a)).get()
	return len(*arr)
}

func (a Array[T]) Cap() int {
	arr := (reference[[]T](a)).get()
	return cap(*arr)
}

func (a Array[T]) Get(idx int) T {
	arr := (reference[[]T](a)).get()
	if idx >= len(*arr) {
		panic("out of index")
	}
	return (*arr)[idx]
}

func (a Array[T]) Set(idx int, val T) {
	arr := (reference[[]T](a)).get()
	if idx >= len(*arr) {
		panic("out of index")
	}
	(*arr)[idx] = val
}

func (a Array[T]) Add(val T) {
	arr := (reference[[]T](a)).get()
	*arr = append(*arr, val)
}

func (a Array[T]) Nil() bool {
	return a.idx == 0
}

func (a *Array[T]) Free() {
	((*reference[[]T])(a)).free()
}


type String reference[string]

func NewString(str string) String {
	return String(newRef[string](&str))
}

func (s String) String() string {
	return *((reference[string](s)).get())
}

func (s String) Set(str string) {
	*((reference[string](s)).get()) = str
}

func (s *String) Free() {
	((*reference[string])(s)).free()
}