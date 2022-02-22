package ecsgo

import (
	"math"
	"reflect"
)

// isystem system interface
type isystem interface {
	
	SetTickInterval(intervalSecond float64)
	SetPriority(priority int)

	run()
	getIncludeTypes() []reflect.Type
	getDependencyTypes() []reflect.Type
	getExcludeTypes() []reflect.Type
	addExcludeType(tp reflect.Type)
	addIncludeType(tp reflect.Type)
	addTagType(tp reflect.Type)
	addDependencyType(tp reflect.Type)
	makeReadonly(tp reflect.Type)
	isReadonly(tp reflect.Type) bool
	isTemporary() bool
	getPriority() int
}

type baseSystem struct {
	r               *Registry
	includeTypes    []reflect.Type
	dependencyTypes []reflect.Type
	excludeTypes    []reflect.Type
	readonlyMap     map[reflect.Type]bool
	isTemp          bool
	time            Ticktime
	intervalSeconds float64
	elapsedSeconds  float64
	priority        int
}

func newBaseSystem(r *Registry, time Ticktime, isTemporary bool) *baseSystem {
	return &baseSystem{
		r: r,
		readonlyMap: make(map[reflect.Type]bool),
		isTemp: isTemporary,
		time: time,
		priority: math.MaxInt,
	}
}

func (s *baseSystem) getIncludeTypes() []reflect.Type {
	return s.includeTypes
}

func (s *baseSystem) getDependencyTypes() []reflect.Type {
	return s.dependencyTypes
}

func (s *baseSystem) getExcludeTypes() []reflect.Type {
	return s.excludeTypes
}

func (s *baseSystem) addExcludeType(tp reflect.Type) {
	s.excludeTypes = append(s.excludeTypes, tp)
}

func (s *baseSystem) addIncludeType(tp reflect.Type) {
	s.includeTypes = append(s.includeTypes, tp)
	s.dependencyTypes = append(s.dependencyTypes, tp)
}

func (s *baseSystem) addTagType(tp reflect.Type) {
	s.includeTypes = append(s.includeTypes, tp)
}

func (s *baseSystem) addDependencyType(tp reflect.Type) {
	s.dependencyTypes = append(s.dependencyTypes, tp)
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

func (s *baseSystem) isTemporary() bool {
	return s.isTemp
}

func (s *baseSystem) SetTickInterval(intervalSeconds float64) {
	s.intervalSeconds = intervalSeconds
}

func (s *baseSystem) exceedTickInterval() bool {
	s.elapsedSeconds += s.r.deltaSeconds
	if s.elapsedSeconds >= s.intervalSeconds {
		s.r.setSystemDeltaSeconds(s.elapsedSeconds)
		s.elapsedSeconds = 0
		return true
	}
	return false
}

func (s *baseSystem) SetPriority(priority int) {
	s.priority = priority
}

func (s *baseSystem) getPriority() int {
	return s.priority
}

// system non componenet system
type nonComponentSystem struct {
	baseSystem
	fn          func (r *Registry)
}

func makeNonComponentSystem(r *Registry, time Ticktime, isTemporary bool, fn func (r *Registry)) isystem {
	sys := &nonComponentSystem{
		baseSystem: *newBaseSystem(r, time, isTemporary),
		fn: fn,
	}
	return sys
}

// run run system
func (s *nonComponentSystem) run() {
	if !s.exceedTickInterval() {
		return
	}

	s.fn(s.r)
	if s.isTemp {
		s.r.defferredRemovesystem(s.time, s)
	}
}

// system non componenet system
type system struct {
	baseSystem
	fn          func (r *Registry, iter *Iterator)
}

func makeSystem(r *Registry, time Ticktime, isTemporary bool, fn func (r *Registry, iter *Iterator)) isystem {
	sys := &system{
		baseSystem: *newBaseSystem(r, time, isTemporary),
		fn: fn,
	}
	return sys
}

func makeSystem1[T any](r *Registry, time Ticktime, isTemporary bool, fn func (r *Registry, iter *Iterator)) isystem {
	sys := makeSystem(r, time, isTemporary, fn)
	var zeroT T
	sys.addIncludeType(reflect.TypeOf(zeroT))
	return sys
}

func makeSystem2[T1, T2 any](r *Registry, time Ticktime, isTemporary bool, fn func (r *Registry, iter *Iterator)) isystem {
	sys := makeSystem(r, time, isTemporary, fn)
	var zeroT1 T1
	var zeroT2 T2
	sys.addIncludeType(reflect.TypeOf(zeroT1))
	sys.addIncludeType(reflect.TypeOf(zeroT2))
	return sys
}

func makeSystem3[T1, T2, T3 any](r *Registry, time Ticktime, isTemporary bool, fn func (r *Registry, iter *Iterator)) isystem {
	sys := makeSystem(r, time, isTemporary, fn)
	var zeroT1 T1
	var zeroT2 T2
	var zeroT3 T3
	sys.addIncludeType(reflect.TypeOf(zeroT1))
	sys.addIncludeType(reflect.TypeOf(zeroT2))
	sys.addIncludeType(reflect.TypeOf(zeroT3))
	return sys
}

func makeSystem4[T1, T2, T3, T4 any](r *Registry, time Ticktime, isTemporary bool, fn func (r *Registry, iter *Iterator)) isystem {
	sys := makeSystem(r, time, isTemporary, fn)
	var zeroT1 T1
	var zeroT2 T2
	var zeroT3 T3
	var zeroT4 T4
	sys.addIncludeType(reflect.TypeOf(zeroT1))
	sys.addIncludeType(reflect.TypeOf(zeroT2))
	sys.addIncludeType(reflect.TypeOf(zeroT3))
	sys.addIncludeType(reflect.TypeOf(zeroT4))
	return sys
}

func makeSystem5[T1, T2, T3, T4, T5 any](r *Registry, time Ticktime, isTemporary bool, fn func (r *Registry, iter *Iterator)) isystem {
	sys := makeSystem(r, time, isTemporary, fn)
	var zeroT1 T1
	var zeroT2 T2
	var zeroT3 T3
	var zeroT4 T4
	var zeroT5 T5
	sys.addIncludeType(reflect.TypeOf(zeroT1))
	sys.addIncludeType(reflect.TypeOf(zeroT2))
	sys.addIncludeType(reflect.TypeOf(zeroT3))
	sys.addIncludeType(reflect.TypeOf(zeroT4))
	sys.addIncludeType(reflect.TypeOf(zeroT5))
	return sys
}

// run run system
func (s *system) run() {
	if !s.exceedTickInterval() {
		return
	}
	
	tables := s.query()
	iter := makeIterator(s, s.r, tables)
	if !iter.IsNil() {
		s.fn(s.r, iter)
	}
	if s.isTemp {
		s.r.defferredRemovesystem(s.time, s)
	}
}

type Iterator struct {
	s isystem
	r *Registry
	tables []*unsafeTable
	tabIdx int
	tabIter *tableIter
}

func makeIterator(s isystem, r *Registry, tables []*unsafeTable) *Iterator {
	itr := &Iterator{
		s: s,
		r: r,
		tables: tables,
	}
	itr.Next()
	return itr
}

func (i *Iterator) Next() {
	for {
		if i.tabIter == nil || i.tabIter.isNil() {
			if i.tabIdx == len(i.tables) {
				return
			}
			i.tabIter = i.tables[i.tabIdx].iterator()
			i.tabIdx++
		} else {
			i.tabIter.next()
		}
		if i.tabIter.isNil() {
			continue
		}
		if i.r.IsAlive(i.tabIter.entity()) {
			return
		}
	}
}

func (i *Iterator) IsNil() bool {
	return i.tabIdx == len(i.tables) && (i.tabIter == nil || i.tabIter.isNil())
}

func (i *Iterator) Entity() Entity {
	return i.tabIter.entity()
}
