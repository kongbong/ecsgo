package ecsgo

import (
	"reflect"
	"github.com/kongbong/ecsgo/sparseSet"
)

// elem1 table row element for single value
type elem1[T any] struct {
	entity Entity
	v1 T
}

// elem2 table row element for two values
type elem2[T any, U any] struct {
	entity Entity
	v1 T
	v2 U
}

// itable table interface
type itable interface {
	hasType(t reflect.Type) bool
	iterator() iiter
}

// table1 single value table
type table1[T any] struct {
	set *sparseSet.Set[elem1[T]]
}

// getOrAddTable1 making single value table
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

// insert entity value into table
func (t *table1[T]) insert(entity Entity, v *T) {
	t.set.Insert(uint32(entity.id), &elem1[T]{entity, *v})
}

// hasType type check if it has specific type value
func (*table1[T]) hasType(t reflect.Type) bool {
	var zeroT T
	return reflect.TypeOf(zeroT) == t
}

// iterator get iterator
func (t *table1[T]) iterator() iiter {
	return &iter1[T]{
		elems: t.set.Iterate(),
		idx: -1,
	}
}

// table2 two values table
type table2[T any, U any] struct {
	set *sparseSet.Set[elem2[T, U]]
}

// insert entity value into table
func (t *table2[T, U]) insert(entity Entity, v1 *T, v2 *U) {
	t.set.Insert(uint32(entity.id), &elem2[T, U]{entity, *v1, *v2})
}

// hasType type check if it has specific type value
func (*table2[T, U]) hasType(t reflect.Type) bool {
	var zeroT T
	var zeroU U
	return reflect.TypeOf(zeroT) == t || reflect.TypeOf(zeroU) == t
}

// iterator get iterator
func (t *table2[T, U]) iterator() iiter {
	return &iter2[T, U]{
		elems: t.set.Iterate(),
		idx: -1,
	}
}

// getOrAddTable1 making two values table
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

// iiter iterator interface
type iiter interface {
	next() bool
	get(t reflect.Type) interface{}
	entity() Entity
}

// iter1 single value iterator
type iter1[T any] struct {
	idx int
	elems []elem1[T]
}

// next advance to next element
func (i *iter1[T]) next() bool {
	i.idx++
	return i.idx < len(i.elems)
}

// get current element specific type value
func (i *iter1[T]) get(t reflect.Type) interface{} {
	var zero T
	if t != reflect.TypeOf(zero) {
		panic("t should be same with T type")
	}
	return &i.elems[i.idx].v1
}

// entity current element entity
func (i *iter1[T]) entity() Entity {
	return i.elems[i.idx].entity
}

// iter1 two values iterator
type iter2[T any, U any] struct {
	idx int
	elems []elem2[T, U]
}

// next advance to next element
func (i *iter2[T, U]) next() bool {
	i.idx++
	return i.idx < len(i.elems)
}

// get current element specific type value
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

// entity current element entity
func (i *iter2[T, U]) entity() Entity {
	return i.elems[i.idx].entity
}
