package ecsgo

import "reflect"

// isystem system interface
type isystem interface {
	run()
	getIncludeTypes() []reflect.Type
	getExcludeTypes() []reflect.Type
	setExcludeTypes(types ...reflect.Type)
}

type baseSystem struct {
	r *Registry
	includeTypes []reflect.Type
	excludeTypes []reflect.Type
}

func (s *baseSystem) getIncludeTypes() []reflect.Type {
	return s.includeTypes
}

func (s *baseSystem) getExcludeTypes() []reflect.Type {
	return s.excludeTypes
}

func (s *baseSystem) setExcludeTypes(types ...reflect.Type) {
	s.excludeTypes = types
}

func (s *baseSystem) setIncludeTypes(types ...reflect.Type) {
	s.includeTypes = types
}

func (s *baseSystem) query() []*unsafeTable {
	return s.r.storage.query(s.includeTypes, s.excludeTypes)
}

// AddSystem add none component system
func AddSystem(r *Registry, time Ticktime, fn func (r *Registry)) isystem {
	sys := &system{
		fn: fn,
	}
	sys.r = r
	r.addsystem(time, sys)
	return sys
}

// system non componenet system
type system struct {
	baseSystem
	fn func (r *Registry)
}

// run run system
func (s *system) run() {
	s.fn(s.r)
}

// system1 single value system
type system1[T any] struct {
	baseSystem
	fn func (*Registry, Entity, *T)
}

// AddSystem1 add single value system
func AddSystem1[T any](r *Registry, time Ticktime, fn func (r *Registry, entity Entity, t *T)) isystem {
	sys := &system1[T]{
		fn: fn,
	}	
	sys.r = r
	var zeroT T
	sys.setIncludeTypes(reflect.TypeOf(zeroT))
	r.addsystem(time, sys)
	return sys
}

// run run system
func (s *system1[T]) run() {
	var zeroT T
	tables := s.query()
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

// system2 two values system
type system2[T any, U any] struct {
	baseSystem
	fn func (*Registry, Entity, *T, *U)
}

// AddSystem2 add two values system
func AddSystem2[T any, U any](r *Registry, time Ticktime, fn func (r *Registry, entity Entity, t *T, u *U)) isystem {
	sys := &system2[T, U]{
		fn: fn,
	}	
	sys.r = r
	var zeroT T
	var zeroU U
	sys.setIncludeTypes(reflect.TypeOf(zeroT), reflect.TypeOf(zeroU))
	r.addsystem(time, sys)
	return sys
}

// run run system
func (s *system2[T, U]) run() {
	var zeroT T
	var zeroU U
	tables := s.query()
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

func Exclude1[T any](sys isystem) {
	var zeroT T
	sys.setExcludeTypes(reflect.TypeOf(zeroT))
}

func Exclude2[T any, U any](sys isystem) {
	var zeroT T
	var zeroU U
	sys.setExcludeTypes(reflect.TypeOf(zeroT), reflect.TypeOf(zeroU))
}