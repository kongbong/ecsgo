package ecsgo

import "reflect"

// isystem system interface
type isystem interface {
	run()
	getCmpTypes() []reflect.Type
}

// AddSystem1 add single value system
func AddSystem1[T any](r *Registry, fn func (entity Entity, t *T)) {
	r.addsystem(&system1[T]{
		r: r,
		fn: fn,
	})
}

// system1 single value system
type system1[T any] struct {
	r *Registry
	fn func (Entity, *T)
}

// query1 find tables where has type T value
func query1[T any](r *Registry) []itable {
	var rst []itable
	var zeroT T
	for _, t := range r.tables {
		if !t.hasType(reflect.TypeOf(zeroT)) {
			continue
		}
		rst = append(rst, t)
	}
	return rst
}

// run run system
func (s *system1[T]) run() {
	var zeroT T

	tables := query1[T](s.r)
	for _, t := range tables {
		for iter := t.iterator(); iter.next(); {
			if !s.r.IsAlive(iter.entity()) {
				continue
			}			
			ptrT := iter.get(reflect.TypeOf(zeroT)).(*T)
			s.fn(iter.entity(), ptrT)
		}
	}
}

// getCmpTypes return relying component list
func (s *system1[T]) getCmpTypes() []reflect.Type {
	var zeroT T
	return []reflect.Type{reflect.TypeOf(zeroT)}
}

// AddSystem2 add two values system
func AddSystem2[T any, U any](r *Registry, fn func (entity Entity, t *T, u *U)) {
	r.addsystem(&system2[T, U]{
		r: r,
		fn: fn,
	})
}

// system2 two values system
type system2[T any, U any] struct {
	r *Registry
	fn func (Entity, *T, *U)
}

// query2 find tables where has type T and U values
func query2[T any, U any](r *Registry) []itable {
	var rst []itable
	var zeroT T
	var zeroU U
	for _, t := range r.tables {
		if !t.hasType(reflect.TypeOf(zeroT)) {
			continue
		}
		if !t.hasType(reflect.TypeOf(zeroU)) {
			continue
		}
		rst = append(rst, t)
	}
	return rst
}

// run run system
func (s *system2[T, U]) run() {
	var zeroT T
	var zeroU U

	tables := query2[T, U](s.r)
	for _, t := range tables {
		for iter := t.iterator(); iter.next(); {
			if !s.r.IsAlive(iter.entity()) {
				continue
			}	
			ptrT := iter.get(reflect.TypeOf(zeroT)).(*T)
			ptrU := iter.get(reflect.TypeOf(zeroU)).(*U)
			s.fn(iter.entity(), ptrT, ptrU)
		}
	}
}

// getCmpTypes return relying component list
func (s *system2[T, U]) getCmpTypes() []reflect.Type {
	var zeroT T
	var zeroU U
	return []reflect.Type{reflect.TypeOf(zeroT), reflect.TypeOf(zeroU)}
}
