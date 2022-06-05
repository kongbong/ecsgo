package ecsgo

import (
	"math"
	"sync"
	"sync/atomic"
	"time"
)

type Ticktime int

const (
	prepreTick Ticktime = iota
	PreTick
	OnTick
	PostTick
	postpostTick
	ticktimeMax
)

// Registry main struct where has entities, systems and storages
type Registry struct {
	entities           []Entity
	freelist           []Entity
	storage            *storage
	pipelines          [ticktimeMax]*pipeline
	defferredMtx       sync.Mutex
	defferredAddSys    []sysInfo
	defferredAddCmp    map[Entity][]*componentInfo
	defferredRemoveSys []sysInfo
	deltaSeconds       float64
	sysDeltaSeconds    uint64
}

// New make new Registry
func New() *Registry {
	r := &Registry{
		storage:         newStorage(),
		defferredAddCmp: make(map[Entity][]*componentInfo),
	}
	for i := 0; i < int(ticktimeMax); i++ {
		r.pipelines[i] = newPipeline()
	}
	r.pipelines[prepreTick].addSystem(makeNonComponentSystem(r, prepreTick, false, processPreProcess))
	r.pipelines[postpostTick].addSystem(makeNonComponentSystem(r, postpostTick, false, processPostProcess))
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
	return math.Float64frombits(atomic.LoadUint64(&r.sysDeltaSeconds))
}

func (r *Registry) setSystemDeltaSeconds(val float64) {
	atomic.StoreUint64(&r.sysDeltaSeconds, math.Float64bits(val))
}

// Run run systems
func (r *Registry) Tick(deltaSeconds float64) {
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
		tick = time.Tick(time.Second / time.Duration(options.fps))
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
		r.Tick(interval)
		lastTick = now
	}
}

type sysInfo struct {
	time   Ticktime
	system isystem
}

// defferredAddsystem add system in pipeline
func (r *Registry) defferredAddsystem(time Ticktime, system isystem) {
	r.defferredMtx.Lock()
	defer r.defferredMtx.Unlock()
	r.defferredAddSys = append(r.defferredAddSys, sysInfo{time, system})
}

// defferredRemovesystem remove system in pipeline
func (r *Registry) defferredRemovesystem(time Ticktime, system isystem) {
	r.defferredMtx.Lock()
	defer r.defferredMtx.Unlock()
	r.defferredRemoveSys = append(r.defferredRemoveSys, sysInfo{time, system})
}

// defferredAddComponent is adding new component into entity, it is defferred until next pre tick
func (r *Registry) defferredAddComponent(ent Entity, cmpInfo *componentInfo) {
	r.defferredMtx.Lock()
	defer r.defferredMtx.Unlock()
	r.defferredAddCmp[ent] = append(r.defferredAddCmp[ent], cmpInfo)
}

func processPreProcess(r *Registry) {
	r.defferredMtx.Lock()
	defer r.defferredMtx.Unlock()

	for ent, cmps := range r.defferredAddCmp {
		r.storage.addComponents(ent, cmps)
	}
	r.defferredAddCmp = make(map[Entity][]*componentInfo)

	for _, sysInfo := range r.defferredAddSys {
		r.pipelines[sysInfo.time].addSystem(sysInfo.system)
	}
	r.defferredAddSys = nil
}

func processPostProcess(r *Registry) {
	r.defferredMtx.Lock()
	defer r.defferredMtx.Unlock()

	for _, sysInfo := range r.defferredRemoveSys {
		r.pipelines[sysInfo.time].removeSystem(sysInfo.system)
	}
	r.defferredRemoveSys = nil
}
