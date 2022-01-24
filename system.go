package ecsgo

import "reflect"

// isystem system interface
type isystem interface {
	run()
	getCmpTypes() []reflect.Type
}

// AddSystem add none component system
func AddSystem(r *Registry, time Ticktime, fn func (r *Registry)) {
	r.addsystem(time, &system{
		r: r,
		fn: fn,
	})
}

// system non componenet system
type system struct {
	r *Registry
	fn func (r *Registry)
}


// run run system
func (s *system) run() {
	s.fn(s.r)
}

func (s *system) getCmpTypes() []reflect.Type {
	return nil
}

// AddSystem1 add single value system
func AddSystem1[T any](r *Registry, time Ticktime, fn func (r *Registry, entity Entity, t *T)) {
	r.addsystem(time, &system1[T]{
		r: r,
		fn: fn,
	})
}

// system1 single value system
type system1[T any] struct {
	r *Registry
	fn func (*Registry, Entity, *T)
}

// query1 find tables where has type T value
func query1[T any](r *Registry) []*unsafeTable {
	var zeroT T
	return r.storage.query(reflect.TypeOf(zeroT))
}

// run run system
func (s *system1[T]) run() {
	var zeroT T

	tables := query1[T](s.r)
	for _, t := range tables {
		for iter := t.iterator(); !iter.isNil(); iter.next() {
			if !s.r.IsAlive(iter.entity()) {
				continue
			}			
			ptrT := (*T)(iter.get(reflect.TypeOf(zeroT)))
			s.fn(s.r, iter.entity(), ptrT)
		}
	}
}

// getCmpTypes return relying component list
func (s *system1[T]) getCmpTypes() []reflect.Type {
	var zeroT T
	return []reflect.Type{reflect.TypeOf(zeroT)}
}

// AddSystem2 add two values system
func AddSystem2[T any, U any](r *Registry, time Ticktime, fn func (r *Registry, entity Entity, t *T, u *U)) {
	r.addsystem(time, &system2[T, U]{
		r: r,
		fn: fn,
	})
}

// system2 two values system
type system2[T any, U any] struct {
	r *Registry
	fn func (*Registry, Entity, *T, *U)
}

// query2 find tables where has type T and U values
func query2[T any, U any](r *Registry) []*unsafeTable {
	var zeroT T
	var zeroU U
	return r.storage.query(reflect.TypeOf(zeroT), reflect.TypeOf(zeroU))
}

// run run system
func (s *system2[T, U]) run() {
	var zeroT T
	var zeroU U

	tables := query2[T, U](s.r)
	for _, t := range tables {
		for iter := t.iterator(); !iter.isNil(); iter.next() {
			if !s.r.IsAlive(iter.entity()) {
				continue
			}	
			ptrT := (*T)(iter.get(reflect.TypeOf(zeroT)))
			ptrU := (*U)(iter.get(reflect.TypeOf(zeroU)))
			s.fn(s.r, iter.entity(), ptrT, ptrU)
		}
	}
}

// getCmpTypes return relying component list
func (s *system2[T, U]) getCmpTypes() []reflect.Type {
	var zeroT T
	var zeroU U
	return []reflect.Type{reflect.TypeOf(zeroT), reflect.TypeOf(zeroU)}
}
