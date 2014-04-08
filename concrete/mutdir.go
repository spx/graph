// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package concrete

import (
	"math"
	"sort"

	"github.com/gonum/graph"
)

// A GonumGraph is a very generalized graph that can handle an arbitrary number of vertices and
// edges -- as well as act as either directed or undirected.
//
// Internally, it uses a map of successors AND predecessors, to speed up some operations (such as
// getting all successors/predecessors). It also speeds up thing like adding edges (assuming both
// edges exist).
//
// However, its generality is also its weakness (and partially a flaw in needing to satisfy
// MutableGraph). For most purposes, creating your own graph is probably better. For instance,
// see TileGraph for an example of an immutable 2D grid of tiles that also implements the Graph
// interface, but would be more suitable if all you needed was a simple undirected 2D grid.
type MutableDirectedGraph struct {
	successors   map[int]map[int]WeightedEdge
	predecessors map[int]map[int]WeightedEdge
	nodeMap      map[int]graph.Node
}

func NewMutableDirectedGraph() *MutableDirectedGraph {
	return &MutableDirectedGraph{
		successors:   make(map[int]map[int]WeightedEdge),
		predecessors: make(map[int]map[int]WeightedEdge),
		nodeMap:      make(map[int]graph.Node),
	}
}

/* Mutable Graph implementation */

func (g *MutableDirectedGraph) NewNode() graph.Node {
	nodeList := g.NodeList()
	ids := make([]int, len(nodeList))
	for i, n := range nodeList {
		ids[i] = n.ID()
	}

	nodes := sort.IntSlice(ids)
	sort.Sort(&nodes)
	for i, n := range nodes {
		if i != n {
			g.AddNode(Node(i))
			return Node(i)
		}
	}

	newID := len(nodes)
	g.AddNode(Node(newID))
	return Node(newID)
}

func (g *MutableDirectedGraph) AddNode(n graph.Node) {
	if _, ok := g.nodeMap[n.ID()]; ok {
		return
	}

	g.nodeMap[n.ID()] = n
	g.successors[n.ID()] = make(map[int]WeightedEdge)
	g.predecessors[n.ID()] = make(map[int]WeightedEdge)
}

func (g *MutableDirectedGraph) AddEdgeTo(e graph.Edge, cost float64) {
	head, tail := e.Head(), e.Tail()
	g.AddNode(head)
	g.AddNode(tail)

	g.successors[head.ID()][tail.ID()] = WeightedEdge{Edge: e, Cost: cost}
	g.predecessors[tail.ID()][head.ID()] = WeightedEdge{Edge: e, Cost: cost}
}

func (g *MutableDirectedGraph) RemoveNode(n graph.Node) {
	if _, ok := g.nodeMap[n.ID()]; !ok {
		return
	}
	delete(g.nodeMap, n.ID())

	for succ, _ := range g.successors[n.ID()] {
		delete(g.predecessors[succ], n.ID())
	}
	delete(g.successors, n.ID())

	for pred, _ := range g.predecessors[n.ID()] {
		delete(g.successors[pred], n.ID())
	}
	delete(g.predecessors, n.ID())

}

func (g *MutableDirectedGraph) RemoveEdgeTo(e graph.Edge) {
	head, tail := e.Head(), e.Tail()
	if _, ok := g.nodeMap[head.ID()]; !ok {
		return
	} else if _, ok := g.nodeMap[tail.ID()]; !ok {
		return
	}

	delete(g.successors[head.ID()], tail.ID())
	delete(g.predecessors[tail.ID()], head.ID())
}

func (g *MutableDirectedGraph) EmptyGraph() {
	g.successors = make(map[int]map[int]WeightedEdge)
	g.predecessors = make(map[int]map[int]WeightedEdge)
	g.nodeMap = make(map[int]graph.Node)
}

/* Graph implementation */

func (g *MutableDirectedGraph) Successors(n graph.Node) []graph.Node {
	if _, ok := g.successors[n.ID()]; !ok {
		return nil
	}

	successors := make([]graph.Node, len(g.successors[n.ID()]))
	i := 0
	for succ, _ := range g.successors[n.ID()] {
		successors[i] = g.nodeMap[succ]
		i++
	}

	return successors
}

func (g *MutableDirectedGraph) EdgeTo(n, succ graph.Node) graph.Edge {
	if _, ok := g.nodeMap[n.ID()]; !ok {
		return nil
	} else if _, ok := g.nodeMap[succ.ID()]; !ok {
		return nil
	}

	edge, ok := g.successors[n.ID()][succ.ID()]
	if !ok {
		return nil
	}
	return edge
}

func (g *MutableDirectedGraph) Predecessors(n graph.Node) []graph.Node {
	if _, ok := g.successors[n.ID()]; !ok {
		return nil
	}

	predecessors := make([]graph.Node, len(g.predecessors[n.ID()]))
	i := 0
	for succ, _ := range g.predecessors[n.ID()] {
		predecessors[i] = g.nodeMap[succ]
		i++
	}

	return predecessors
}

func (g *MutableDirectedGraph) Neighbors(n graph.Node) []graph.Node {
	if _, ok := g.successors[n.ID()]; !ok {
		return nil
	}

	neighbors := make([]graph.Node, len(g.predecessors[n.ID()])+len(g.successors[n.ID()]))
	i := 0
	for succ, _ := range g.successors[n.ID()] {
		neighbors[i] = g.nodeMap[succ]
		i++
	}

	for pred, _ := range g.predecessors[n.ID()] {
		// We should only add the predecessor if it wasn't already added from successors
		if _, ok := g.successors[n.ID()][pred]; !ok {
			neighbors[i] = g.nodeMap[pred]
			i++
		}
	}

	return neighbors
}

func (g *MutableDirectedGraph) EdgeBetween(n, neigh graph.Node) graph.Edge {
	e := g.EdgeTo(n, neigh)
	if e != nil {
		return e
	}

	e = g.EdgeTo(neigh, n)
	if e != nil {
		return e
	}

	return nil
}

func (g *MutableDirectedGraph) NodeExists(n graph.Node) bool {
	_, ok := g.nodeMap[n.ID()]

	return ok
}

func (g *MutableDirectedGraph) Degree(n graph.Node) int {
	if _, ok := g.nodeMap[n.ID()]; !ok {
		return 0
	}

	return len(g.successors[n.ID()]) + len(g.predecessors[n.ID()])
}

func (g *MutableDirectedGraph) NodeList() []graph.Node {
	nodes := make([]graph.Node, len(g.successors))
	i := 0
	for _, n := range g.nodeMap {
		nodes[i] = n
		i++
	}

	return nodes
}

func (g *MutableDirectedGraph) Cost(e graph.Edge) float64 {
	if s, ok := g.successors[e.Head().ID()]; ok {
		if we, ok := s[e.Tail().ID()]; ok {
			return we.Cost
		}
	}
	return math.Inf(1)
}

func (g *MutableDirectedGraph) EdgeList() []graph.Edge {
	edgeList := make([]graph.Edge, 0, len(g.successors))
	edgeMap := make(map[int]map[int]struct{}, len(g.successors))
	for n, succMap := range g.successors {
		edgeMap[n] = make(map[int]struct{}, len(succMap))
		for succ, edge := range succMap {
			if doneMap, ok := edgeMap[succ]; ok {
				if _, ok := doneMap[n]; ok {
					continue
				}
			}
			edgeList = append(edgeList, edge)
			edgeMap[n][succ] = struct{}{}
		}
	}

	return edgeList
}