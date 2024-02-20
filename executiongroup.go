package ecsgo

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type executionGroup struct {
	executeList []*System

	depRootNode *depTreeNode
	dirty       bool
}

func newExecutionGroup() *executionGroup {
	return &executionGroup{}
}

func (e *executionGroup) addSystem(sys *System) *executionGroup {
	e.dirty = true
	e.executeList = append(e.executeList, sys)
	return e
}

// dependent node
type depNode struct {
	sys   *System
	edges []*depNode
}

func (e *executionGroup) build() error {
	if !e.dirty {
		return nil
	}
	e.dirty = false

	slices.SortFunc(e.executeList, func(a, b *System) int {
		if a.GetPriority() == b.GetPriority() {
			return b.getInterestComponentCount() - a.getInterestComponentCount()
		}
		return a.GetPriority() - b.GetPriority()
	})

	// make dependency graph
	nodes := make([]*depNode, len(e.executeList))
	for i, sys := range e.executeList {
		nodes[i] = &depNode{sys: sys}
	}

	// double loop to make dependency graph
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if nodes[i].sys.dependent(nodes[j].sys) {
				nodes[i].edges = append(nodes[i].edges, nodes[j])
			}
		}
	}

	// change dependency graph to dependency tree
	var err error
	e.depRootNode, err = changeToDependencyTree(nodes)
	return errors.Errorf("failed to change to dependency tree: %v", err)
}

// dependency tree node
type depTreeNode struct {
	sys *System

	waitCount int
	wg        sync.WaitGroup
	edges     []*depTreeNode
}

func (n *depTreeNode) addEdge(edge *depTreeNode) {
	edge.wg.Add(1)
	edge.waitCount++
	n.edges = append(n.edges, edge)
}

func changeToDependencyTree(nodes []*depNode) (*depTreeNode, error) {
	root := &depTreeNode{}

	resolved := make(map[*depNode]bool)
	depNodeMap := make(map[*depNode]*depTreeNode)

	for {
		if len(nodes) == 0 {
			return root, nil
		}

		unresolved := make(map[*depNode]bool)
		leastDepNode, err := depResolve(nodes[0], resolved, unresolved)
		if err != nil {
			return nil, errors.Errorf("failed to resolve dependency: %v", err)
		}

		treeNode := &depTreeNode{
			sys: leastDepNode.sys,
		}
		depNodeMap[leastDepNode] = treeNode

		// inverse node
		if len(leastDepNode.edges) == 0 {
			root.addEdge(treeNode)
		} else {
			for _, edge := range leastDepNode.edges {
				edgeTreeNode := depNodeMap[edge]
				edgeTreeNode.addEdge(treeNode)
			}
		}

		// remove leastDepNode from node list
		nodes = slices.DeleteFunc(nodes, func(n *depNode) bool {
			return n == leastDepNode
		})
	}
}

func depResolve(node *depNode, resolved, unresolved map[*depNode]bool) (*depNode, error) {
	unresolved[node] = true
	for _, edge := range node.edges {
		if !resolved[edge] {
			if unresolved[edge] {
				return nil, errors.Errorf("circular dependency detected: %v -> %v", node.sys, edge.sys)
			}
			return depResolve(edge, resolved, unresolved)
		}
	}
	resolved[node] = true
	delete(unresolved, node)
	return node, nil
}

func (e *executionGroup) execute(deltaTime time.Duration, ctx context.Context) error {
	err := e.build()
	if err != nil {
		return errors.Errorf("failed to build dependency tree %v", err)
	}
	errs, ctx := errgroup.WithContext(ctx)

	err = runTree(e.depRootNode, deltaTime, &sync.Map{}, errs, ctx)
	if err != nil {
		return errors.Errorf("failed to run dependency tree %v", err)
	}
	return errs.Wait()
}

// runTree - run dependency tree parallely
func runTree(node *depTreeNode, deltaTime time.Duration, visited *sync.Map, errs *errgroup.Group, ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	node.wg.Wait()
	// reset wait group count for next round
	node.wg.Add(node.waitCount)

	var err error
	if node.sys != nil {
		err = node.sys.execute(deltaTime)
	}

	for _, edge := range node.edges {
		edge.wg.Done()
	}

	// shold return error after edge node waitGroup done for not starving edge
	if err != nil {
		return err
	}

	for _, edge := range node.edges {
		_, loaded := visited.LoadOrStore(edge, true)
		if !loaded {
			e := edge
			errs.Go(func() error {
				return runTree(e, deltaTime, visited, errs, ctx)
			})
		}
	}
	return nil
}
