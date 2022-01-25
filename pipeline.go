package ecsgo

import (
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

// addSystem adding system in pipeline, analyzing dependency and making tree
func (p *pipeline) addSystem(sys isystem) {
	types := sys.getIncludeTypes()
	sysw := newWrapper(sys, len(types))
	p.sysw = append(p.sysw, sysw)

	for _, tp := range types {
		newNode := &pipeNode{
			sysw:   sysw,
			donech: make(chan bool),
		}
		sysw.nodes = append(sysw.nodes, newNode)

		node, ok := p.depenMap[tp]
		if !ok {
			p.depenMap[tp] = newNode
		} else {
			for node.next != nil {
				node = node.next
			}
			node.next = newNode
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
		n.sysw.waitch <- true
		<-n.donech
		n = n.next
	}
}

// pipeNode single dependency line
type pipeNode struct {
	next   *pipeNode
	sysw   *sysWrapper
	donech chan bool
}

// sysWrapper system wrapper for waiting dependent systems done
type sysWrapper struct {
	sys     isystem
	waitCnt int
	waitch  chan bool
	nodes   []*pipeNode
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
func newWrapper(sys isystem, waitCnt int) *sysWrapper {
	return &sysWrapper{
		sys:     sys,
		waitCnt: waitCnt,
		waitch:  make(chan bool),
	}
}
