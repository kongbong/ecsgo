package ecsgo

import (
	"reflect"
	"sync"
)

type pipeline struct {
	depenMap map[reflect.Type]*pipeNode
	sysw     []*sysWrapper
}

func newPipeline() *pipeline {
	return &pipeline{
		depenMap: make(map[reflect.Type]*pipeNode),
	}
}

func (p *pipeline) addSystem(sys isystem) {
	types := sys.getCmpTypes()
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

func (p *pipeline) run() {

	for _, n := range p.depenMap {
		go runNodeline(n)
	}

	var wg sync.WaitGroup
	wg.Add(len(p.sysw))
	for _, s := range p.sysw {
		go s.run(&wg)
	}
	wg.Wait()
}

func runNodeline(n *pipeNode) {
	for n != nil {
		n.sysw.waitch <- true
		<-n.donech
		n = n.next
	}
}

type pipeNode struct {
	next   *pipeNode
	sysw   *sysWrapper
	donech chan bool
}

type sysWrapper struct {
	sys     isystem
	waitCnt int
	waitch  chan bool
	nodes   []*pipeNode
}

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

func newWrapper(sys isystem, waitCnt int) *sysWrapper {
	return &sysWrapper{
		sys:     sys,
		waitCnt: waitCnt,
		waitch:  make(chan bool),
	}
}
