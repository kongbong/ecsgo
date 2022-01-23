package ecsgo

// Registry main struct where has entities, systems and storages
type Registry struct {
	entities []Entity
	freelist []Entity
	tables   []itable
	pipeline *pipeline
}

// New make new Registry
func New() *Registry {
	return &Registry{
		pipeline: newPipeline(),
	}
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
}

// Run run systems
func (r *Registry) Run() {
	r.pipeline.run()
}

// addsystem add system in pipeline
func (r *Registry) addsystem(system isystem) {
	r.pipeline.addSystem(system)
}
