package ecsgo

import (
	"reflect"
	"slices"

	"github.com/pkg/errors"
)

type ArcheType struct {
	enitityIds []EntityId
	components map[reflect.Type]cmpInterface

	entityIdxMap map[EntityId]int

	debugComponentStr []string
}

type cmpInterface interface {
	onAddEntity(idx int)
	onRemoveEntity(idx, lastIdx int)
	copyDataToOtherArcheType(acc *ArcheTypeAccessor, other *ArcheType) error
}

func newArcheType(types ...reflect.Type) *ArcheType {
	a := &ArcheType{
		components:   make(map[reflect.Type]cmpInterface),
		entityIdxMap: make(map[EntityId]int),
	}
	for _, t := range types {
		a.components[t] = nil
		a.debugComponentStr = append(a.debugComponentStr, t.String())
	}
	return a
}

func (a *ArcheType) getEntityCount() int {
	return len(a.enitityIds)
}

func (a *ArcheType) hasComponent(t reflect.Type) bool {
	_, found := a.components[t]
	return found
}

func (a *ArcheType) hasEntity(entityId EntityId) bool {
	_, found := a.entityIdxMap[entityId]
	return found
}

func HasArcheTypeComponent[T any](a *ArcheType) bool {
	var t T
	return a.hasComponent(reflect.TypeOf(t))
}

func (a *ArcheType) getComponentTypeList() []reflect.Type {
	var typeList = make([]reflect.Type, len(a.components))
	var i int
	for t := range a.components {
		typeList[i] = t
	}
	return typeList
}

func (a *ArcheType) equalComponents(types []reflect.Type) bool {
	if len(types) != len(a.components) {
		return false
	}
	for _, t := range types {
		_, found := a.components[t]
		if !found {
			return false
		}
	}
	return true
}

func (a *ArcheType) addEntity(entityId EntityId) {
	_, added := a.entityIdxMap[entityId]
	if added {
		// already added
		return
	}
	idx := len(a.enitityIds)
	a.enitityIds = append(a.enitityIds, entityId)
	a.entityIdxMap[entityId] = idx

	for _, v := range a.components {
		if v != nil {
			v.onAddEntity(idx)
		}
	}
}

func (a *ArcheType) removeEntity(entityId EntityId) {
	idx, found := a.entityIdxMap[entityId]
	if !found {
		// not added
		return
	}
	lastIdx := len(a.enitityIds) - 1
	lastEntityId := a.enitityIds[lastIdx]

	delete(a.entityIdxMap, entityId)

	if idx == lastIdx {
		// remove last
		a.enitityIds = a.enitityIds[:lastIdx]
	} else {
		// swap
		a.enitityIds[idx] = lastEntityId
		a.entityIdxMap[lastEntityId] = idx
		slices.Delete(a.enitityIds, lastIdx, lastIdx+1)
	}
	for _, v := range a.components {
		if v != nil {
			v.onRemoveEntity(idx, lastIdx)
		}
	}
}

func (a *ArcheType) copyDataToOtherArcheType(entityId EntityId, ty reflect.Type, other *ArcheType) error {
	cmp := a.components[ty]
	if cmp == nil {
		// no data
		return nil
	}
	return cmp.copyDataToOtherArcheType(a.GetAccessor(entityId), other)
}

type compData[T any] struct {
	arr []T
}

func newCompData[T any](size int) *compData[T] {
	return &compData[T]{
		arr: make([]T, size),
	}
}

func (c *compData[T]) onAddEntity(idx int) {
	if idx != len(c.arr) {
		panic("added index should be same with length of array")
	}
	var t T
	c.arr = append(c.arr, t)
}

func (c *compData[T]) onRemoveEntity(idx, lastIdx int) {
	if lastIdx != len(c.arr)-1 {
		panic("lastIdx should be same with last index of array")
	}
	c.arr[idx] = c.arr[lastIdx]
	slices.Delete(c.arr, lastIdx, lastIdx+1)
}

func (c *compData[T]) copyDataToOtherArcheType(acc *ArcheTypeAccessor, other *ArcheType) error {
	otherAcc := other.GetAccessor(acc.entityId)
	if otherAcc == nil {
		return errors.Errorf("other Archetype doesn't have entityId %v", acc.entityId)
	}
	success := setArcheTypeComponentByIdx[T](other, otherAcc.idx, c.arr[acc.idx])
	if !success {
		return errors.Errorf("failed to set archetype data")
	}
	return nil
}

func getArcheTypeComponent[T any](a *ArcheType, entityId EntityId) *T {
	idx, found := a.entityIdxMap[entityId]
	if !found {
		return nil
	}

	return getArcheTypeComponentByIdx[T](a, idx)
}

func getArcheTypeComponentByIdx[T any](a *ArcheType, idx int) *T {
	if idx < 0 || idx >= len(a.enitityIds) {
		return nil
	}

	var t T
	var cmpData *compData[T]
	v, found := a.components[reflect.TypeOf(t)]
	if !found {
		return nil
	}
	if v == nil {
		cmpData = newCompData[T](len(a.enitityIds))
		a.components[reflect.TypeOf(t)] = cmpData
	} else {
		cmpData = v.(*compData[T])
	}
	return &cmpData.arr[idx]
}

func setArcheTypeComponent[T any](a *ArcheType, entityId EntityId, value T) bool {
	idx, found := a.entityIdxMap[entityId]
	if !found {
		// entity is not added
		return false
	}
	return setArcheTypeComponentByIdx[T](a, idx, value)
}

func setArcheTypeComponentByIdx[T any](a *ArcheType, idx int, value T) bool {
	var cmpData *compData[T]
	v, found := a.components[reflect.TypeOf(value)]
	if !found {
		return false
	}
	if v == nil {
		cmpData = newCompData[T](len(a.enitityIds))
		a.components[reflect.TypeOf(value)] = cmpData
	} else {
		cmpData = v.(*compData[T])
	}
	cmpData.arr[idx] = value
	return true
}

type ArcheTypeAccessor struct {
	idx       int
	entityId  EntityId
	archeType *ArcheType
}

func (a *ArcheType) GetAccessor(entityId EntityId) *ArcheTypeAccessor {
	idx, found := a.entityIdxMap[entityId]
	if !found {
		return nil
	}
	return a.getAceessorByIdx(idx)
}

func (a *ArcheType) getAceessorByIdx(idx int) *ArcheTypeAccessor {
	if idx < 0 || idx >= len(a.enitityIds) {
		return nil
	}
	return &ArcheTypeAccessor{
		idx:       idx,
		entityId:  a.enitityIds[idx],
		archeType: a,
	}
}

func (a *ArcheType) Foreach(fn func(accessor *ArcheTypeAccessor) error) error {
	for i := 0; i < len(a.enitityIds); i++ {
		err := fn(a.getAceessorByIdx(i))
		if err != nil {
			return err
		}
	}
	return nil
}

func (acc *ArcheTypeAccessor) GetArcheType() *ArcheType {
	return acc.archeType
}

func (acc *ArcheTypeAccessor) GetEntityId() EntityId {
	return acc.entityId
}

func GetComponentByAccessor[T any](acc *ArcheTypeAccessor) *T {
	return getArcheTypeComponentByIdx[T](acc.archeType, acc.idx)
}

func SetComponentData[T any](acc *ArcheTypeAccessor, value T) bool {
	return setArcheTypeComponentByIdx[T](acc.archeType, acc.idx, value)
}
