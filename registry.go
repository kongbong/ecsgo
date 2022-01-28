package ecsgo

import (
	"sync"
	"time"
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
	defferredMtx sync.Mutex
	defferredSys []sysInfo
	defferredCmp map[Entity][]*componentInfo
	deltaSeconds float64
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
	r.pipelines[PreTick].addSystem(makeSystem(r, processDeferredProcess))
	return r
}

func (r *Registry) Free() {
	r.storage.free()
}

// Create entity
func (r *Registry) Create() Entity {
	if len(r.freelist) > 0 {
		ent := r.freelist[len(r.freelist)-1]
		r.freelist = r.freelist[:len(r.freelist)-1]
		ent.setVersion(ent.version + 1)
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
		old.setVersion(old.version + 1)
		r.freelist = append(r.freelist, *old)
	}
	r.storage.eraseEntity(e)
}

func (r *Registry) DeltaSeconds() float64 {
	return r.deltaSeconds
}

// Run run systems
func (r *Registry) tick(deltaSeconds float64) {
	r.deltaSeconds = deltaSeconds
	var wg sync.WaitGroup
	for i := 0; i < int(ticktimeMax); i++ {
		wg.Add(1)
		r.pipelines[i].run(&wg)
		wg.Wait()
	}
}

func (r *Registry) Run(opts ...option) {
	var options options
	for _, o := range opts {
		o(&options)
	}

	var tick <-chan time.Time
	if options.fps > 0 {
		tick = time.Tick(time.Millisecond / time.Duration(options.fps))
	}
	lastTick := time.Now()
	for {
		if tick != nil {
			<-tick
		}

		now := time.Now()
		interval := now.Sub(lastTick).Seconds()
		if options.fixedTick {
			interval = 1 / float64(options.fps)
		}
		r.tick(interval)
		lastTick = now
	}
}

type sysInfo struct {
	time   Ticktime
	system isystem
}

// addsystem add system in pipeline
func (r *Registry) defferredAddsystem(time Ticktime, system isystem) {
	r.defferredMtx.Lock()
	defer r.defferredMtx.Unlock()
	r.defferredSys = append(r.defferredSys, sysInfo{time, system})
}

// addComponent is defferred until next pre tick
func (r *Registry) defferredAddComponent(ent Entity, cmpInfo *componentInfo) {
	r.defferredMtx.Lock()
	defer r.defferredMtx.Unlock()
	r.defferredCmp[ent] = append(r.defferredCmp[ent], cmpInfo)
}

func processDeferredProcess(r *Registry) {
	r.defferredMtx.Lock()
	defer r.defferredMtx.Unlock()

	for ent, cmps := range r.defferredCmp {
		r.storage.addComponents(ent, cmps)
	}
	r.defferredCmp = make(map[Entity][]*componentInfo)

	for _, sysInfo := range r.defferredSys {
		r.pipelines[sysInfo.time].addSystem(sysInfo.system)
	}
	r.defferredSys = nil
}
