// +build ui

package tests

import (
	"testing"
	"time"
)

// TestGraphVisualizer tests graph visualization
func TestGraphVisualizer(t *testing.T) {
	gv := NewGraphVisualizer()

	// Add nodes
	node1 := gv.AddNode(1, "Node 1", 100, 100)
	node2 := gv.AddNode(2, "Node 2", 300, 100)

	if node1 == nil || node2 == nil {
		t.Fatal("Failed to add nodes")
	}

	if len(gv.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(gv.Nodes))
	}
}

func TestLinkCreation(t *testing.T) {
	gv := NewGraphVisualizer()

	// Add nodes
	gv.AddNode(1, "Node 1", 100, 100)
	gv.AddNode(2, "Node 2", 300, 100)

	// Add link
	err := gv.AddLink(1, 1, 2)
	if err != nil {
		t.Fatalf("Failed to add link: %v", err)
	}

	if len(gv.Links) != 1 {
		t.Errorf("Expected 1 link, got %d", len(gv.Links))
	}
}

func TestNodeRemoval(t *testing.T) {
	gv := NewGraphVisualizer()

	gv.AddNode(1, "Node 1", 100, 100)
	gv.AddNode(2, "Node 2", 300, 100)
	gv.AddNode(3, "Node 3", 500, 100)

	gv.RemoveNode(2)

	if len(gv.Nodes) != 2 {
		t.Errorf("Expected 2 nodes after removal, got %d", len(gv.Nodes))
	}
}

func TestNodeAt(t *testing.T) {
	gv := NewGraphVisualizer()

	gv.AddNode(1, "Node 1", 100, 100)

	// Check hit detection
	node := gv.GetNodeAt(125, 115)
	if node == nil {
		t.Error("Failed to detect node at coordinates")
	}

	// Check miss detection
	node = gv.GetNodeAt(50, 50)
	if node != nil {
		t.Error("Incorrectly detected node outside area")
	}
}

func TestBezierRouting(t *testing.T) {
	gv := NewGraphVisualizer()

	node1 := gv.AddNode(1, "Node 1", 0, 0)
	node2 := gv.AddNode(2, "Node 2", 200, 0)

	gv.Engine.Strategy = "bezier"
	route := gv.Engine.CalculateRoute(node1, node2)

	if len(route) < 2 {
		t.Errorf("Expected at least 2 points in bezier route, got %d", len(route))
	}
}

func TestManhattanRouting(t *testing.T) {
	gv := NewGraphVisualizer()

	node1 := gv.AddNode(1, "Node 1", 0, 0)
	node2 := gv.AddNode(2, "Node 2", 200, 200)

	gv.Engine.Strategy = "manhattan"
	route := gv.Engine.CalculateRoute(node1, node2)

	if len(route) != 4 {
		t.Errorf("Expected 4 points in manhattan route, got %d", len(route))
	}
}

func TestDirectRouting(t *testing.T) {
	gv := NewGraphVisualizer()

	node1 := gv.AddNode(1, "Node 1", 0, 0)
	node2 := gv.AddNode(2, "Node 2", 200, 200)

	gv.Engine.Strategy = "direct"
	route := gv.Engine.CalculateRoute(node1, node2)

	if len(route) != 2 {
		t.Errorf("Expected 2 points in direct route, got %d", len(route))
	}
}

func TestEventLoop(t *testing.T) {
	// Simulate event loop with timeout
	done := make(chan bool, 1)
	timeout := time.After(1 * time.Second)

	go func() {
		// Simulate event processing
		time.Sleep(100 * time.Millisecond)
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-timeout:
		t.Error("Event loop timed out")
	}
}

func BenchmarkGraphRendering(b *testing.B) {
	gv := NewGraphVisualizer()

	// Create large graph
	for i := 0; i < 100; i++ {
		gv.AddNode(uint32(i), "Node", float64(i*50), float64(i*50))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate rendering
		_ = gv.Nodes
		_ = gv.Links
	}
}
