package ecsgo

import "reflect"

type isystem interface {
	run()
	getCmpTypes() []reflect.Type
}

func AddSystem1[T any](r *Registry, fn func (entity EntityVer, t *T)) {
	r.addsystem(&system1[T]{
		r: r,
		fn: fn,
	})
}

type system1[T any] struct {
	r *Registry
	fn func (EntityVer, *T)
}

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

func (s *system1[T]) run() {
	var zeroT T

	tables := query1[T](s.r)
	for _, t := range tables {
		for iter := t.iterator(); iter.next(); {
			ptrT := iter.get(reflect.TypeOf(zeroT)).(*T)
			s.fn(iter.entity(), ptrT)
		}
	}
}

func (s *system1[T]) getCmpTypes() []reflect.Type {
	var zeroT T
	return []reflect.Type{reflect.TypeOf(zeroT)}
}

func AddSystem2[T any, U any](r *Registry, fn func (entity EntityVer, t *T, u *U)) {
	r.addsystem(&system2[T, U]{
		r: r,
		fn: fn,
	})
}

type system2[T any, U any] struct {
	r *Registry
	fn func (EntityVer, *T, *U)
}

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

func (s *system2[T, U]) run() {
	var zeroT T
	var zeroU U

	tables := query2[T, U](s.r)
	for _, t := range tables {
		for iter := t.iterator(); iter.next(); {
			ptrT := iter.get(reflect.TypeOf(zeroT)).(*T)
			ptrU := iter.get(reflect.TypeOf(zeroU)).(*U)
			s.fn(iter.entity(), ptrT, ptrU)
		}
	}
}

func (s *system2[T, U]) getCmpTypes() []reflect.Type {
	var zeroT T
	var zeroU U
	return []reflect.Type{reflect.TypeOf(zeroT), reflect.TypeOf(zeroU)}
}
