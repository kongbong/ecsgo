package ecsgo

import "reflect"

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
type system struct {
	baseSystem
	fn          func (r *Registry)
}

func makeSystem(r *Registry, time Ticktime, isTemporary bool, fn func (r *Registry)) isystem {
	sys := &system{
		baseSystem: *newBaseSystem(r, time, isTemporary),
		fn: fn,
	}
	return sys
}

// run run system
func (s *system) run() {
	if !s.exceedTickInterval() {
		return
	}
	s.fn(s.r)
	if s.isTemp {
		s.r.defferredRemovesystem(s.time, s)
	}
}

// system1 single value system
type system1[T any] struct {
	baseSystem
	fn func (*Registry, Entity, *T)
}
// makeSystem1 add single value system
func makeSystem1[T any](r *Registry, time Ticktime, isTemporary bool, fn func (r *Registry, entity Entity, t *T)) isystem {
	var zeroT T
	if err := checkType(reflect.TypeOf(zeroT)); err != nil {
		panic(err)
	}
	sys := &system1[T]{
		baseSystem: *newBaseSystem(r, time, isTemporary),
		fn: fn,
	}
	sys.addIncludeType(reflect.TypeOf(zeroT))
	return sys
}

// run run system
func (s *system1[T]) run() {
	if !s.exceedTickInterval() {
		return
	}
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
	if s.isTemp {
		s.r.defferredRemovesystem(s.time, s)
	}
}

// system2 two values system
type system2[T any, U any] struct {
	baseSystem
	fn func (*Registry, Entity, *T, *U)
}

// AddSystem2 add two values system
func makeSystem2[T any, U any](r *Registry, time Ticktime, isTemporary bool, fn func (r *Registry, entity Entity, t *T, u *U)) isystem {
	var zeroT T
	var zeroU U
	if err := checkType(reflect.TypeOf(zeroT)); err != nil {
		panic(err)
	}
	if err := checkType(reflect.TypeOf(zeroU)); err != nil {
		panic(err)
	}

	sys := &system2[T, U]{
		baseSystem: *newBaseSystem(r, time, isTemporary),
		fn: fn,
	}
	sys.addIncludeType(reflect.TypeOf(zeroT))
	sys.addIncludeType(reflect.TypeOf(zeroU))
	return sys
}

// run run system
func (s *system2[T, U]) run() {
	if !s.exceedTickInterval() {
		return
	}
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
	if s.isTemp {
		s.r.defferredRemovesystem(s.time, s)
	}
}
