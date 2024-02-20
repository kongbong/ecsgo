package ecsgo

import "reflect"

type ObserverContext struct {
	registry          *Registry
	entityId          EntityId
	archeType         *ArcheType
	addedComponents   []reflect.Type
	removedComponents []reflect.Type
}

type ObserverFunc func(ctx *ObserverContext) error

type Observer struct {
	registry *Registry
	name     string
	fn       ObserverFunc

	addComponents    map[reflect.Type]bool
	removeComponents map[reflect.Type]bool
}

func newObserver(registry *Registry, name string, fn ObserverFunc) *Observer {
	return &Observer{
		registry: registry,
		name:     name,
		fn:       fn,
	}
}

func (o *Observer) GetName() string {
	return o.name
}

func AddComponentToObserver[T any](o *Observer) {
	var t T
	if o.addComponents == nil {
		o.addComponents = make(map[reflect.Type]bool)
	}
	o.addComponents[reflect.TypeOf(t)] = true
}

func RemoveComponentFromObserver[T any](o *Observer) {
	var t T
	if o.removeComponents == nil {
		o.removeComponents = make(map[reflect.Type]bool)
	}
	o.removeComponents[reflect.TypeOf(t)] = true
}

func (o *Observer) executeIfInterest(entityId EntityId, archeType *ArcheType, addedComponents, removedComponents []reflect.Type) error {
	interested, interestedAdd, interestedRemove := o.interestedIn(addedComponents, removedComponents)
	if interested {
		return o.execute(entityId, archeType, interestedAdd, interestedRemove)
	}
	return nil
}

func (o *Observer) interestedIn(addedComponents, removedComponents []reflect.Type) (interested bool, added []reflect.Type, removed []reflect.Type) {
	for _, t := range addedComponents {
		_, found := o.addComponents[t]
		if found {
			added = append(added, t)
			interested = true
		}
	}
	for _, t := range removedComponents {
		_, found := o.removeComponents[t]
		if found {
			removed = append(removed, t)
			interested = true
		}
	}
	return
}

func (o *Observer) execute(entityId EntityId, archeType *ArcheType, addedComponents, removedComponents []reflect.Type) error {
	return o.fn(&ObserverContext{
		registry:          o.registry,
		entityId:          entityId,
		archeType:         archeType,
		addedComponents:   addedComponents,
		removedComponents: removedComponents,
	})
}

func (ctx *ObserverContext) GetEntityId() EntityId {
	return ctx.entityId
}

func (ctx *ObserverContext) GetArcheType() *ArcheType {
	return ctx.archeType
}

func (ctx *ObserverContext) CreateEntity() EntityId {
	return ctx.registry.CreateEntity()
}

func GetComponentObserver[T any](ctx *ObserverContext) *T {
	return getArcheTypeComponent[T](ctx.archeType, ctx.entityId)
}
