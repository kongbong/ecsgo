package ecsgo

import (
	"reflect"
	"ecsgo/sparseSet"
)

// table has rows and columns
type elem1[T any] struct {
	entity EntityVer
	v1 T
}

// table has rows and columns
type elem2[T any, U any] struct {
	entity EntityVer
	v1 T
	v2 U
}

type itable interface {
	hasType(t reflect.Type) bool
	iterator() iiterrow
}

type table1[T any] struct {
	set *sparseSet.Set[elem1[T]]
}

func getOrAddTable1[T any](r *Registry) *table1[T] {
	for _, v := range r.tables {
		if t, ok := v.(*table1[T]); ok {
			// already made
			return t
		}
	}
	t := &table1[T]{
		set: sparseSet.NewAutoIncresing[elem1[T]](100),
	}
	r.tables = append(r.tables, t)
	return t
} 

func (t *table1[T]) insert(entity EntityVer, v *T) {
	t.set.Insert(uint32(entity.ToEntity()), &elem1[T]{entity, *v})
}

func (*table1[T]) hasType(t reflect.Type) bool {
	var zeroT T
	return reflect.TypeOf(zeroT) == t
}

func (t *table1[T]) iterator() iiterrow {
	return &iter1[T]{
		elems: t.set.Iterate(),
		idx: -1,
	}
}

type table2[T any, U any] struct {
	set *sparseSet.Set[elem2[T, U]]
}

func (t *table2[T, U]) insert(entity EntityVer, v1 *T, v2 *U) {
	t.set.Insert(uint32(entity.ToEntity()), &elem2[T, U]{entity, *v1, *v2})
}

func (*table2[T, U]) hasType(t reflect.Type) bool {
	var zeroT T
	var zeroU U
	return reflect.TypeOf(zeroT) == t || reflect.TypeOf(zeroU) == t
}

func (t *table2[T, U]) iterator() iiterrow {
	return &iter2[T, U]{
		elems: t.set.Iterate(),
		idx: -1,
	}
}

func getOrAddTable2[T any, U any](r *Registry) *table2[T, U] {
	for _, v := range r.tables {
		if t, ok := v.(*table2[T, U]); ok {
			// already made
			return t
		}
	}
	t := &table2[T, U]{
		set: sparseSet.NewAutoIncresing[elem2[T, U]](100),
	}
	r.tables = append(r.tables, t)
	return t
} 

type iiterrow interface {
	next() bool
	get(t reflect.Type) interface{}
	entity() EntityVer
}

type iter1[T any] struct {
	idx int
	elems []elem1[T]
}

func (i *iter1[T]) next() bool {
	i.idx++
	return i.idx < len(i.elems)
}

func (i *iter1[T]) get(t reflect.Type) interface{} {
	var zero T
	if t != reflect.TypeOf(zero) {
		panic("t should be same with T type")
	}
	return &i.elems[i.idx].v1
}

func (i *iter1[T]) entity() EntityVer {
	return i.elems[i.idx].entity
}

type iter2[T any, U any] struct {
	idx int
	elems []elem2[T, U]
}

func (i *iter2[T, U]) next() bool {
	i.idx++
	return i.idx < len(i.elems)
}

func (i *iter2[T, U]) get(t reflect.Type) interface{} {
	var zeroT T
	var zeroU U
	if t == reflect.TypeOf(zeroT) {
		return &i.elems[i.idx].v1
	}
	if t == reflect.TypeOf(zeroU) {
		return &i.elems[i.idx].v2
	}
	panic("t should be same with T or U type")
}

func (i *iter2[T, U]) entity() EntityVer {
	return i.elems[i.idx].entity
}
