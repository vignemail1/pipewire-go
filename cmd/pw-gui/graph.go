package main

import (
	"fmt"
	"math"
)

// Point represents a 2D coordinate
type Point struct {
	X, Y float64
}

// NodeVisual represents visual properties of a node
type NodeVisual struct {
	X, Y, Width, Height float64
	Label string
	ID uint32
}

// LinkVisual represents visual properties of a link
type LinkVisual struct {
	Source, Target uint32
	Route []Point
}

// GraphVisualizer handles graph visualization
type GraphVisualizer struct {
	Nodes map[uint32]*NodeVisual
	Links map[uint32]*LinkVisual
	Engine *RoutingEngine
}

// NewGraphVisualizer creates a new graph visualizer
func NewGraphVisualizer() *GraphVisualizer {
	return &GraphVisualizer{
		Nodes: make(map[uint32]*NodeVisual),
		Links: make(map[uint32]*LinkVisual),
		Engine: NewRoutingEngine(),
	}
}

// AddNode adds a node to the graph
func (gv *GraphVisualizer) AddNode(id uint32, label string, x, y float64) *NodeVisual {
	node := &NodeVisual{
		ID: id,
		Label: label,
		X: x,
		Y: y,
		Width: 100,
		Height: 40,
	}
	gv.Nodes[id] = node
	return node
}

// RemoveNode removes a node from the graph
func (gv *GraphVisualizer) RemoveNode(id uint32) {
	delete(gv.Nodes, id)

	// Remove connected links
	var linksToRemove []uint32
	for linkID, link := range gv.Links {
		if link.Source == id || link.Target == id {
			linksToRemove = append(linksToRemove, linkID)
		}
	}
	for _, linkID := range linksToRemove {
		delete(gv.Links, linkID)
	}
}

// AddLink adds a link between nodes
func (gv *GraphVisualizer) AddLink(id uint32, source, target uint32) error {
	if _, ok := gv.Nodes[source]; !ok {
		return fmt.Errorf("source node %d not found", source)
	}
	if _, ok := gv.Nodes[target]; !ok {
		return fmt.Errorf("target node %d not found", target)
	}

	link := &LinkVisual{
		Source: source,
		Target: target,
	}

	// Calculate route
	link.Route = gv.Engine.CalculateRoute(gv.Nodes[source], gv.Nodes[target])

	gv.Links[id] = link
	return nil
}

// RemoveLink removes a link
func (gv *GraphVisualizer) RemoveLink(id uint32) {
	delete(gv.Links, id)
}

// GetNodeAt returns the node at the given coordinates
func (gv *GraphVisualizer) GetNodeAt(x, y float64) *NodeVisual {
	for _, node := range gv.Nodes {
		if x >= node.X && x <= node.X+node.Width &&
		   y >= node.Y && y <= node.Y+node.Height {
			return node
		}
	}
	return nil
}

// RoutingEngine handles link routing calculations
type RoutingEngine struct {
	Strategy string // "direct", "manhattan", "bezier"
}

// NewRoutingEngine creates a new routing engine
func NewRoutingEngine() *RoutingEngine {
	return &RoutingEngine{
		Strategy: "bezier",
	}
}

// CalculateRoute calculates the route between two nodes
func (re *RoutingEngine) CalculateRoute(from, to *NodeVisual) []Point {
	switch re.Strategy {
	case "direct":
		return re.directRoute(from, to)
	case "manhattan":
		return re.manhattanRoute(from, to)
	default:
		return re.bezierRoute(from, to)
	}
}

// directRoute returns a direct line
func (re *RoutingEngine) directRoute(from, to *NodeVisual) []Point {
	fromX := from.X + from.Width/2
	fromY := from.Y + from.Height/2
	toX := to.X + to.Width/2
	toY := to.Y + to.Height/2

	return []Point{
		{fromX, fromY},
		{toX, toY},
	}
}

// manhattanRoute returns a manhattan-style path
func (re *RoutingEngine) manhattanRoute(from, to *NodeVisual) []Point {
	fromX := from.X + from.Width/2
	fromY := from.Y + from.Height/2
	toX := to.X + to.Width/2
	toY := to.Y + to.Height/2

	midX := (fromX + toX) / 2

	return []Point{
		{fromX, fromY},
		{midX, fromY},
		{midX, toY},
		{toX, toY},
	}
}

// bezierRoute returns a bezier curve
func (re *RoutingEngine) bezierRoute(from, to *NodeVisual) []Point {
	fromX := from.X + from.Width/2
	fromY := from.Y + from.Height/2
	toX := to.X + to.Width/2
	toY := to.Y + to.Height/2

	var points []Point
	steps := 20

	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)

		// Quadratic Bezier with control point
		controlX := (fromX + toX) / 2
		controlY := (fromY + toY) / 2 - 50

		x := math.Pow(1-t, 2)*fromX +
			2*(1-t)*t*controlX +
			t*t*toX

		y := math.Pow(1-t, 2)*fromY +
			2*(1-t)*t*controlY +
			t*t*toY

		points = append(points, Point{x, y})
	}

	return points
}
