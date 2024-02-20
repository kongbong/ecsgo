package ecsgo

import (
	"context"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

type Registry struct {
	eg              *executionGroup
	deferredActions *deferredActions

	archeTypeList      []*ArcheType
	systems            []*System
	observers          []*Observer
	entityArcheTypeMap map[EntityId]*ArcheType

	duringTick int32

	// for issue new id
	mx         sync.Mutex
	lastId     uint32
	tombstones []EntityId
}

func NewRegistry() *Registry {
	r := &Registry{
		eg:                 newExecutionGroup(),
		entityArcheTypeMap: make(map[EntityId]*ArcheType),
	}
	r.deferredActions = newDeferredActions(r)
	return r
}

func (r *Registry) CreateEntity() EntityId {
	r.mx.Lock()

	var entityId EntityId
	lastIdx := len(r.tombstones) - 1
	if lastIdx >= 0 {
		lastId := r.tombstones[lastIdx]
		r.tombstones[lastIdx] = EntityId{}
		r.tombstones = r.tombstones[:lastIdx]
		lastId.version++
		entityId = lastId
	} else {
		r.lastId++
		entityId = EntityId{
			id:      r.lastId,
			version: 1,
		}
	}
	// setting archetype, currently it is just for checking issued or not
	r.entityArcheTypeMap[entityId] = nil
	r.mx.Unlock()

	r.deferredActions.createEntity(entityId)
	return entityId
}

func (r *Registry) RemoveEntity(entityId EntityId) {
	r.deferredActions.removeEntity(entityId)
}

func AddComponent[T any](r *Registry, entityId EntityId, val T) {
	addComponentDeferredAction[T](r.deferredActions, entityId, val)
}

func RemoveComponent[T any](r *Registry, entityId EntityId) {
	removeComponentDeferredAction[T](r.deferredActions, entityId)
}

func (r *Registry) AddSystem(name string, priority int, fn SystemFn) *System {
	s := newSystem(r, name, priority, fn)
	r.deferredActions.addSystem(s)
	return s
}

func (r *Registry) AddObserver(name string, fn ObserverFunc) *Observer {
	o := newObserver(r, name, fn)
	r.deferredActions.addObserver(o)
	return o
}

func (r *Registry) IsActiveEntity(entityId EntityId) bool {
	_, found := r.entityArcheTypeMap[entityId]
	return found
}

func (r *Registry) Tick(deltaTime time.Duration, ctx context.Context) error {
	atomic.StoreInt32(&r.duringTick, 1)
	defer func() {
		atomic.StoreInt32(&r.duringTick, 0)
	}()

	err := r.processDeferredActions()
	if err != nil {
		return err
	}
	err = r.eg.execute(deltaTime, ctx)
	if err != nil {
		return err
	}
	// processDeferred again that process deferred actions while processing Systems
	return r.processDeferredActions()
}

func (r *Registry) processDeferredActions() error {
	return r.deferredActions.process()
}

func (r *Registry) addObserverSync(o *Observer) {
	r.observers = append(r.observers, o)
}

func (r *Registry) addSystemSync(s *System) {
	r.systems = append(r.systems, s)
	for _, a := range r.archeTypeList {
		s.addArcheTypeIfInterest(a)
	}
	r.eg.addSystem(s)
}

func (r *Registry) removeEntitySync(entityId EntityId) error {
	a := r.entityArcheTypeMap[entityId]
	if a != nil {
		a.removeEntity(entityId)
	}
	delete(r.entityArcheTypeMap, entityId)
	// it is threadsafe so don't need to lock because it is only called on deferredActions
	r.tombstones = append(r.tombstones, entityId)

	// call observers
	removed := a.getComponentTypeList()
	for _, o := range r.observers {
		err := o.executeIfInterest(entityId, a, nil, removed)
		if err != nil {
			return nil
		}
	}
	return nil
}

func (r *Registry) processEntityActionSync(entityId EntityId, actions []entityAction) error {
	if len(actions) == 0 {
		return nil
	}

	var types []reflect.Type
	var added []reflect.Type
	var removed []reflect.Type
	archeType := r.entityArcheTypeMap[entityId]
	if archeType != nil {
		types = archeType.getComponentTypeList()
	}

	for _, action := range actions {
		action.modifyTypes(&types, &added, &removed)
	}

	targetArcheType := r.getOrMakeArcheTypeSync(types)
	if targetArcheType != nil {
		targetArcheType.addEntity(entityId)

		if archeType != nil && archeType != targetArcheType {
			// move component data from old to new archeType
			origtypes := archeType.getComponentTypeList()
			for _, t := range origtypes {
				if targetArcheType.hasComponent(t) {
					err := archeType.copyDataToOtherArcheType(entityId, t, targetArcheType)
					if err != nil {
						return errors.Errorf("failed to move data from old to new archetype %v", err)
					}
				}
			}
		}

		for _, action := range actions {
			action.apply(entityId, targetArcheType)
		}
	}
	if archeType != nil {
		archeType.removeEntity(entityId)
	}
	r.entityArcheTypeMap[entityId] = targetArcheType

	// call observers
	for _, o := range r.observers {
		err := o.executeIfInterest(entityId, targetArcheType, added, removed)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Registry) getOrMakeArcheTypeSync(types []reflect.Type) *ArcheType {
	if len(types) == 0 {
		return nil
	}
	for _, a := range r.archeTypeList {
		if a.equalComponents(types) {
			return a
		}
	}
	newArcheType := newArcheType(types...)
	r.archeTypeList = append(r.archeTypeList, newArcheType)
	r.onAddArcheType(newArcheType)
	return newArcheType
}

func (r *Registry) onAddArcheType(archeType *ArcheType) {
	for _, s := range r.systems {
		s.addArcheTypeIfInterest(archeType)
	}
}
