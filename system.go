package ecsgo

import "reflect"

// isystem system interface
type isystem interface {
	run()
	getIncludeTypes() []includeTypeInfo
	getExcludeTypes() []reflect.Type
	addExcludeTypes(tp reflect.Type)
	addIncludeTypes(tp reflect.Type, tag bool)
	makeReadonly(tp reflect.Type)
	isReadonly(tp reflect.Type) bool
}

type includeTypeInfo struct {
	tp reflect.Type
	tag bool
}

type baseSystem struct {
	r *Registry
	includeTypes []includeTypeInfo
	excludeTypes []reflect.Type
	readonlyMap map[reflect.Type]bool
}

func newBaseSystem(r *Registry) *baseSystem {
	return &baseSystem{
		r: r,
		readonlyMap: make(map[reflect.Type]bool),
	}
}

func (s *baseSystem) getIncludeTypes() []includeTypeInfo {
	return s.includeTypes
}

func (s *baseSystem) getExcludeTypes() []reflect.Type {
	return s.excludeTypes
}

func (s *baseSystem) addExcludeTypes(tp reflect.Type) {
	s.excludeTypes = append(s.excludeTypes, tp)
}

func (s *baseSystem) addIncludeTypes(tp reflect.Type, tag bool) {
	s.includeTypes = append(s.includeTypes, includeTypeInfo{tp, tag})
}


func (s *baseSystem) makeReadonly(tp reflect.Type) {
	s.readonlyMap[tp] = true
}

func (s *baseSystem) isReadonly(tp reflect.Type) bool {
	return s.readonlyMap[tp]
}

func (s *baseSystem) query() []*unsafeTable {
	return s.r.storage.query(s.includeTypes, s.excludeTypes)
}

// system non componenet system
type system struct {
	baseSystem
	fn func (r *Registry)
}

func MakeSystem(r *Registry, fn func (r *Registry)) isystem {
	sys := &system{
		baseSystem: *newBaseSystem(r),
		fn: fn,
	}
	return sys
}

// AddSystem add none component system
func AddSystem(r *Registry, time Ticktime, fn func (r *Registry)) isystem {
	sys := MakeSystem(r, fn)
	r.defferredAddsystem(time, sys)
	return sys
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
	var zeroT T
	if err := checkType(reflect.TypeOf(zeroT)); err != nil {
		panic(err)
	}
	sys := &system1[T]{
		baseSystem: *newBaseSystem(r),
		fn: fn,
	}
	sys.addIncludeTypes(reflect.TypeOf(zeroT), false)
	r.defferredAddsystem(time, sys)
	return sys
}

// run run system
func (s *system1[T]) run() {
	var zeroT T
	typeT := reflect.TypeOf(zeroT)
	tables := s.query()
	for _, t := range tables {
		for iter := t.iterator(); !iter.isNil(); iter.next() {
			if !s.r.IsAlive(iter.entity()) {
				continue
			}			
			ptrT := (*T)(iter.get(typeT))
			if s.isReadonly(typeT) {
				// copy to temp data
				zeroT = *ptrT
				ptrT = &zeroT
			}
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
	var zeroT T
	var zeroU U
	if err := checkType(reflect.TypeOf(zeroT)); err != nil {
		panic(err)
	}
	if err := checkType(reflect.TypeOf(zeroU)); err != nil {
		panic(err)
	}

	sys := &system2[T, U]{
		baseSystem: *newBaseSystem(r),
		fn: fn,
	}
	sys.addIncludeTypes(reflect.TypeOf(zeroT), false)
	sys.addIncludeTypes(reflect.TypeOf(zeroU), false)
	r.defferredAddsystem(time, sys)
	return sys
}

// run run system
func (s *system2[T, U]) run() {
	var zeroT T
	var zeroU U
	typeT := reflect.TypeOf(zeroT)
	typeU := reflect.TypeOf(zeroU)
	tables := s.query()
	for _, t := range tables {
		for iter := t.iterator(); !iter.isNil(); iter.next() {
			if !s.r.IsAlive(iter.entity()) {
				continue
			}	
			ptrT := (*T)(iter.get(typeT))
			ptrU := (*U)(iter.get(typeU))
			if s.isReadonly(typeT) {
				// copy to temp data
				zeroT = *ptrT
				ptrT = &zeroT
			}
			if s.isReadonly(typeU) {
				// copy to temp data
				zeroU = *ptrU
				ptrU = &zeroU
			}
			s.fn(s.r, iter.entity(), ptrT, ptrU)
		}
	}
}

func Exclude[T any](sys isystem) {
	var zeroT T
	sys.addExcludeTypes(reflect.TypeOf(zeroT))
}

func ExcludeTag[T any](sys isystem) {
	var zeroT T
	if err := checkTagType(reflect.TypeOf(zeroT)); err != nil {
		panic(err)
	}
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
