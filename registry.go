package ecsgo

import (
	"sync"
)

type Ticktime int

const (
	PreTick Ticktime = iota
	OnTick
	PostTick
	ticktimeMax
)

// Registry main struct where has entities, systems and storages
type Registry struct {
	entities     []Entity
	freelist     []Entity
	storage      *storage
	pipelines    [ticktimeMax]*pipeline
	defferredCmp map[Entity][]*componentInfo
}

// New make new Registry
func New() *Registry {
	r := &Registry{
		storage:      newStorage(),
		defferredCmp: make(map[Entity][]*componentInfo),
	}
	for i := 0; i < int(ticktimeMax); i++ {
		r.pipelines[i] = newPipeline()
	}
	AddSystem(r, PreTick, processDeferredComponent)
	return r
}

func (r *Registry) Free() {

}

// Create entity
func (r *Registry) Create() Entity {
	if len(r.freelist) > 0 {
		ent := r.freelist[len(r.freelist)-1]
		r.freelist = r.freelist[:len(r.freelist)-1]
		ent.SetVersion(ent.version + 1)
		r.entities[ent.id] = ent
		return ent
	}
	ent := newEntity(uint32(len(r.entities)), 1)
	r.entities = append(r.entities, ent)
	return ent
}

// IsAlive checks entity is in Registry
func (r *Registry) IsAlive(e Entity) bool {
	return r.entities[uint32(e.id)] == e
}

// Release remove entity from Registry
func (r *Registry) Release(e Entity) {
	old := &r.entities[uint32(e.id)]
	if *old == e {
		old.SetVersion(old.version + 1)
		r.freelist = append(r.freelist, *old)
	}
	r.storage.eraseEntity(e)
}

// Run run systems
func (r *Registry) Run() {
	var wg sync.WaitGroup
	for i := 0; i < int(ticktimeMax); i++ {
		wg.Add(1)
		r.pipelines[i].run(&wg)
		wg.Wait()
	}
}

// addsystem add system in pipeline
func (r *Registry) addsystem(time Ticktime, system isystem) {
	r.pipelines[time].addSystem(system)
}

// addComponent is defferred until next pre tick
func (r *Registry) defferredAddComponent(ent Entity, cmpInfo *componentInfo) {
	r.defferredCmp[ent] = append(r.defferredCmp[ent], cmpInfo)
}
