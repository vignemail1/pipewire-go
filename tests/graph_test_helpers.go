package tests

import (
	"fmt"
)

// GraphVisualizer test helper
type GraphVisualizer struct {
	Nodes map[uint32]*NodeVisual
	Links map[uint32]*LinkVisual
	Engine *RoutingEngine
}

type NodeVisual struct {
	X, Y, Width, Height float64
	Label string
	ID uint32
}

type LinkVisual struct {
	Source, Target uint32
	Route []Point
}

type Point struct {
	X, Y float64
}

type RoutingEngine struct {
	Strategy string
}

func NewGraphVisualizer() *GraphVisualizer {
	return &GraphVisualizer{
		Nodes: make(map[uint32]*NodeVisual),
		Links: make(map[uint32]*LinkVisual),
		Engine: &RoutingEngine{Strategy: "bezier"},
	}
}

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

func (gv *GraphVisualizer) RemoveNode(id uint32) {
	delete(gv.Nodes, id)
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
		Route: []Point{{gv.Nodes[source].X, gv.Nodes[source].Y}},
	}
	gv.Links[id] = link
	return nil
}

func (gv *GraphVisualizer) GetNodeAt(x, y float64) *NodeVisual {
	for _, node := range gv.Nodes {
		if x >= node.X && x <= node.X+node.Width &&
		   y >= node.Y && y <= node.Y+node.Height {
			return node
		}
	}
	return nil
}

func (re *RoutingEngine) CalculateRoute(from, to *NodeVisual) []Point {
	switch re.Strategy {
	case "direct":
		return []Point{
			{from.X + from.Width/2, from.Y + from.Height/2},
			{to.X + to.Width/2, to.Y + to.Height/2},
		}
	case "manhattan":
		midX := (from.X + to.X) / 2
		return []Point{
			{from.X + from.Width/2, from.Y + from.Height/2},
			{midX, from.Y + from.Height/2},
			{midX, to.Y + to.Height/2},
			{to.X + to.Width/2, to.Y + to.Height/2},
		}
	default:
		return []Point{
			{from.X + from.Width/2, from.Y + from.Height/2},
			{to.X + to.Width/2, to.Y + to.Height/2},
		}
	}
}
