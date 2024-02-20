package ecsgo

import (
	"reflect"
	"slices"
	"sync"
)

type entityAction interface {
	modifyTypes(types, added, removed *[]reflect.Type)
	apply(entityId EntityId, a *ArcheType)
}

type deferredActions struct {
	r *Registry

	entityActions      map[EntityId][]entityAction
	addSystemActions   []*System
	addObserverActions []*Observer

	mx sync.Mutex
}

func newDeferredActions(r *Registry) *deferredActions {
	return &deferredActions{
		r:             r,
		entityActions: make(map[EntityId][]entityAction),
	}
}

type createEntityAction struct{}

func (a *createEntityAction) modifyTypes(types, added, removed *[]reflect.Type) {}
func (a *createEntityAction) apply(entityId EntityId, archeType *ArcheType)     {}

type removeEntityAction struct{}

func (a *removeEntityAction) modifyTypes(types, added, removed *[]reflect.Type) {}
func (a *removeEntityAction) apply(entityId EntityId, archeType *ArcheType)     {}

type addComponentAction[T any] struct {
	val T
}

func (a *addComponentAction[T]) modifyTypes(types, added, removed *[]reflect.Type) {
	var ty reflect.Type = reflect.TypeOf(a.val)
	*types = append(*types, ty)
	*added = append(*added, ty)
}

func (a *addComponentAction[T]) apply(entityId EntityId, archeType *ArcheType) {
	setArcheTypeComponent[T](archeType, entityId, a.val)
}

type removeComponentAction[T any] struct{}

func (a *removeComponentAction[T]) modifyTypes(types, added, removed *[]reflect.Type) {
	var t T
	var ty reflect.Type = reflect.TypeOf(t)
	*types = slices.DeleteFunc(*types, func(t reflect.Type) bool {
		return ty == t
	})
	*removed = append(*removed, ty)
}

func (a *removeComponentAction[T]) apply(entityId EntityId, archeType *ArcheType) {
	// nothing to change
}

func (d *deferredActions) createEntity(entityId EntityId) {
	d.mx.Lock()
	defer d.mx.Unlock()

	_, found := d.entityActions[entityId]
	if found {
		// create entity should be called at first
		panic("createEntity should be called at first")
	}
	d.entityActions[entityId] = []entityAction{&createEntityAction{}}
}

func (d *deferredActions) removeEntity(entityId EntityId) {
	d.mx.Lock()
	defer d.mx.Unlock()

	d.entityActions[entityId] = []entityAction{&removeEntityAction{}}
}

func addComponentDeferredAction[T any](d *deferredActions, entityId EntityId, val T) {
	d.mx.Lock()
	defer d.mx.Unlock()

	actions, found := d.entityActions[entityId]
	if !found {
		d.entityActions[entityId] = []entityAction{&addComponentAction[T]{val: val}}
	}
	// check if it is already removed
	if len(actions) == 1 {
		if _, ok := (actions)[0].(*removeEntityAction); ok {
			// it is already removed
			return
		}
	}

	actions = append(actions, &addComponentAction[T]{val: val})
	d.entityActions[entityId] = actions
}

func removeComponentDeferredAction[T any](d *deferredActions, entityId EntityId) {
	d.mx.Lock()
	defer d.mx.Unlock()

	actions, found := d.entityActions[entityId]
	if !found {
		d.entityActions[entityId] = []entityAction{&removeComponentAction[T]{}}
	}
	// check if it is already removed
	if len(actions) == 1 {
		if _, ok := (actions)[0].(*removeEntityAction); ok {
			// it is already removed
			return
		}
	}

	actions = append(actions, &removeComponentAction[T]{})
	d.entityActions[entityId] = actions
}

func (d *deferredActions) addSystem(sys *System) {
	d.mx.Lock()
	defer d.mx.Unlock()

	d.addSystemActions = append(d.addSystemActions, sys)
}

func (d *deferredActions) addObserver(o *Observer) {
	d.mx.Lock()
	defer d.mx.Unlock()

	d.addObserverActions = append(d.addObserverActions, o)
}

func (d *deferredActions) process() error {
	for _, o := range d.addObserverActions {
		d.r.addObserverSync(o)
	}
	d.addObserverActions = d.addObserverActions[:0]

	var err error
	for entityId, actions := range d.entityActions {
		if len(actions) == 0 {
			continue
		}
		if len(actions) == 1 {
			if _, ok := actions[0].(*removeEntityAction); ok {
				err = d.r.removeEntitySync(entityId)
				if err != nil {
					return err
				}
			}
		}

		err = d.r.processEntityActionSync(entityId, actions)
		if err != nil {
			return err
		}
		delete(d.entityActions, entityId)
	}

	for _, sys := range d.addSystemActions {
		d.r.addSystemSync(sys)
	}
	d.addSystemActions = d.addSystemActions[:0]
	return nil
}
