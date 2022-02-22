package ecsgo

import (
	"math"
	"reflect"
	"sync"
)

// pipeline running pipeline where has systems in dependency tree
type pipeline struct {
	depenMap map[reflect.Type]*pipeNode
	sysw     []*sysWrapper
}

// newPipeline make new pipeline
func newPipeline() *pipeline {
	return &pipeline{
		depenMap: make(map[reflect.Type]*pipeNode),
	}
}

// pipeNode single dependency line
type pipeNode struct {
	prev     *pipeNode
	next     *pipeNode
	sysw     []*sysWrapper
	donech   chan bool
	readonly bool
	priority int
}

// sysWrapper system wrapper for waiting dependent systems done
type sysWrapper struct {
	sys     isystem
	waitCnt int
	waitch  chan bool
	nodes   []*pipeNode
}

// addSystem adding system in pipeline, analyzing dependency and making tree
func (p *pipeline) addSystem(sys isystem) {
	types := sys.getDependencyTypes()
	sysw := newWrapper(sys)
	p.sysw = append(p.sysw, sysw)

	for _, tp := range types {
		node, ok := p.depenMap[tp]
		if !ok {
			// add empty root node
			node = &pipeNode{}
			p.depenMap[tp] = node
		}

		inserted := false
		for node.next != nil {
			node = node.next
			if sys.getPriority() < node.priority {
				if node.readonly && sys.isReadonly(tp) {
					// read can overlapped
					node.addSysw(sysw)
				} else {
					insertNode(sysw, sys.isReadonly(tp), node.prev, node)
				}
				inserted = true
				break
			}
		}

		if !inserted {
			if node.readonly && sys.isReadonly(tp) {
				// read can overlapped
				node.addSysw(sysw)
			} else {
				insertNode(sysw, sys.isReadonly(tp), node, nil)
			}
		}
		sysw.waitCnt++
	}
}

func insertNode(sysw *sysWrapper, readonly bool, prev, next *pipeNode) {
	newNode := &pipeNode{
		sysw:     []*sysWrapper{sysw},
		donech:   make(chan bool),
		readonly: readonly,
		priority: sysw.sys.getPriority(),
	}
	prev.next = newNode
	newNode.prev = prev
	newNode.next = next
	sysw.nodes = append(sysw.nodes, newNode)
}

func (p *pipeline) removeSystem(sys isystem) {
	var sw *sysWrapper
	for i, s := range p.sysw {
		if s.sys == sys {
			sw = s
			p.sysw = append(p.sysw[:i], p.sysw[i+1:]...)
			break
		}
	}

	for _, n := range sw.nodes {
		for i, s := range n.sysw {
			if s == sw {
				n.sysw = append(n.sysw[:i], n.sysw[i+1:]...)
				break
			}
		}
		priority := math.MaxInt
		for _, s := range n.sysw {
			if s.sys.getPriority() < priority {
				priority = s.sys.getPriority()
			}
		}
		n.priority = priority
		if len(n.sysw) == 0 {
			if n.prev == nil {
				panic("n.prev should be not nil as root doesn't have any system")
			}
			n.prev.next = n.next
			n.next.prev = n.prev
			n.prev = nil
			n.next = nil
		}
	}
}

// run run pipeline and waiting until all systems are done
func (p *pipeline) run(done *sync.WaitGroup) {
	for _, n := range p.depenMap {
		go runNodeline(n)
	}

	var wg sync.WaitGroup
	wg.Add(len(p.sysw))
	for _, s := range p.sysw {
		go s.run(&wg)
	}
	wg.Wait()
	done.Done()
}

// runNodeline single dependency line
func runNodeline(n *pipeNode) {
	for n != nil {
		wc := len(n.sysw)
		for _, s := range n.sysw {
			s.waitch <- true
		}
		for wc > 0 {
			<-n.donech
			wc--
		}

		n = n.next
	}
}

func (node *pipeNode) addSysw(sysw *sysWrapper) {
	// read can overlapped
	node.sysw = append(node.sysw, sysw)
	sysw.nodes = append(sysw.nodes, node)
	if sysw.sys.getPriority() < node.priority {
		node.priority = sysw.sys.getPriority()
	}
}

// run run systemWrapper
func (s *sysWrapper) run(wg *sync.WaitGroup) {
	wc := s.waitCnt
	for wc > 0 {
		<-s.waitch
		wc--
	}
	s.sys.run()
	for _, n := range s.nodes {
		n.donech <- true
	}
	wg.Done()
}

// newWrapper make new system wrapper
func newWrapper(sys isystem) *sysWrapper {
	return &sysWrapper{
		sys:    sys,
		waitch: make(chan bool),
	}
}
