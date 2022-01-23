package ecsgo

// Registry main struct where has entities, systems and storages
type Registry struct {
	entities []EntityVer
	freelist []EntityVer
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
func (r *Registry) Create() EntityVer {
	if len(r.freelist) > 0 {
		ent := r.freelist[len(r.freelist)-1]
		r.freelist = r.freelist[:len(r.freelist)-1]
		ent.SetVersion(ent.ToVersion() + 1)
		r.entities[uint32(ent.ToEntity())] = ent
		return ent
	}
	ent := newEntity(Entity(len(r.entities)), 1)
	r.entities = append(r.entities, ent)
	return ent
}

// IsAlive checks entity is in Registry
func (r *Registry) IsAlive(e EntityVer) bool {
	return r.entities[uint32(e.ToEntity())] == e
}

// Release remove entity from Registry
func (r *Registry) Release(e EntityVer) {
	old := &r.entities[uint32(e.ToEntity())]
	if *old == e {
		old.SetVersion(old.ToVersion() + 1)
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
