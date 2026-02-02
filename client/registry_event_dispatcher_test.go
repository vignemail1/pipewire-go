package client

import (
	"testing"
	"sync"
)

func TestRegistryThreadSafety(t *testing.T) {
	reg := NewRegistry()

	// Create test nodes
	nodes := make([]*Node, 100)
	for i := 0; i < 100; i++ {
		nodes[i] = &Node{
			ID:   uint32(i),
			Name: "test-node",
		}
	}

	// Concurrent writes
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			reg.AddNode(nodes[idx])
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			reg.GetNode(uint32(idx))
		}(i)
	}

	wg.Wait()

	// Verify all nodes added
	allNodes := reg.GetNodes()
	if len(allNodes) != 100 {
		t.Errorf("Expected 100 nodes, got %d", len(allNodes))
	}
}

func TestRegistryAddGetRemove(t *testing.T) {
	reg := NewRegistry()

	node := &Node{ID: 1, Name: "test"}
	reg.AddNode(node)

	if got := reg.GetNode(1); got == nil || got.ID != 1 {
		t.Errorf("Failed to retrieve node")
	}

	reg.RemoveNode(1)
	if got := reg.GetNode(1); got != nil {
		t.Errorf("Node should be removed but still exists")
	}
}

func TestEventDispatcherThreadSafety(t *testing.T) {
	ed := NewEventDispatcher()

	// Subscribe from multiple goroutines
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ed.Subscribe("test-event", func(e Event) {})
		}()
	}

	// Dispatch from multiple goroutines
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ed.Dispatch(Event{Type: "test-event"})
		}()
	}

	wg.Wait()

	// Verify subscribers added
	if count := ed.GetSubscriberCount("test-event"); count != 50 {
		t.Errorf("Expected 50 subscribers, got %d", count)
	}
}

func TestEventDispatcherSubscribeUnsubscribe(t *testing.T) {
	ed := NewEventDispatcher()

	counter := 0
	handler := func(e Event) {
		counter++
	}

	ed.Subscribe("test", handler)
	ed.Dispatch(Event{Type: "test"})

	if counter != 1 {
		t.Errorf("Handler not called, counter = %d", counter)
	}

	ed.Unsubscribe("test")
	ed.Dispatch(Event{Type: "test"})

	if counter != 1 {
		t.Errorf("Handler called after unsubscribe, counter = %d", counter)
	}
}
