// +build integration

package tests

import (
	"testing"
	"time"

	"github.com/vignemail1/pipewire-go/client"
	"github.com/vignemail1/pipewire-go/core"
)

func TestConnectionWithRealPipeWire(t *testing.T) {
	// Connect to PipeWire daemon
	conn, err := core.NewConnection("pipewire-0")
	if err != nil {
		t.Fatalf("Failed to connect to PipeWire: %v", err)
	}
	defer conn.Close()

	// Verify connection is alive
	if conn == nil {
		t.Fatal("Connection is nil")
	}

	t.Logf("Successfully connected to PipeWire")
}

func TestClientCreation(t *testing.T) {
	c, err := client.NewClient("test-client")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Disconnect()

	if c == nil {
		t.Fatal("Client is nil")
	}

	t.Logf("Client created successfully")
}

func TestGraphEnumeration(t *testing.T) {
	c, err := client.NewClient("test-enum")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Disconnect()

	// Wait for registry sync
	time.Sleep(100 * time.Millisecond)

	// Get registry
	registry := c.GetRegistry()
	if registry == nil {
		t.Fatal("Registry is nil")
	}

	// List all objects
	objects := registry.ListAll()
	t.Logf("Found %d objects in registry", len(objects))

	// We should have at least Core object
	if len(objects) == 0 {
		t.Error("Expected at least one object in registry")
	}
}

func TestNodeEnumeration(t *testing.T) {
	c, err := client.NewClient("test-nodes")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Disconnect()

	time.Sleep(100 * time.Millisecond)

	registry := c.GetRegistry()
	nodes := registry.ListNodes()

	t.Logf("Found %d nodes", len(nodes))

	for _, node := range nodes {
		t.Logf("Node: ID=%d Type=%s", node.ID, node.Type)
	}
}

func TestPortEnumeration(t *testing.T) {
	c, err := client.NewClient("test-ports")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Disconnect()

	time.Sleep(100 * time.Millisecond)

	registry := c.GetRegistry()
	ports := registry.ListPorts()

	t.Logf("Found %d ports", len(ports))

	for _, port := range ports {
		t.Logf("Port: ID=%d Type=%s", port.ID, port.Type)
	}
}

func TestLinkEnumeration(t *testing.T) {
	c, err := client.NewClient("test-links")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Disconnect()

	time.Sleep(100 * time.Millisecond)

	registry := c.GetRegistry()
	links := registry.ListLinks()

	t.Logf("Found %d links", len(links))

	for _, link := range links {
		t.Logf("Link: ID=%d Type=%s", link.ID, link.Type)
	}
}

func TestRegistryEventSubscription(t *testing.T) {
	c, err := client.NewClient("test-events")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Disconnect()

	// Subscribe to registry events
	eventReceived := false
	registry := c.GetRegistry()
	registry.OnGlobalAdded(func(obj *client.GlobalObject) {
		t.Logf("Event: Object added - ID=%d Type=%s", obj.ID, obj.Type)
		eventReceived = true
	})

	// Wait for events
	time.Sleep(500 * time.Millisecond)

	if !eventReceived {
		t.Log("No events received (this may be expected if no changes occur)")
	}
}

func TestConcurrentAccess(t *testing.T) {
	c, err := client.NewClient("test-concurrent")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Disconnect()

	time.Sleep(100 * time.Millisecond)

	// Spawn multiple goroutines accessing registry
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			registry := c.GetRegistry()
			for j := 0; j < 100; j++ {
				_ = registry.ListAll()
				_ = registry.CountObjects()
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	t.Log("Concurrent access test passed")
}

func BenchmarkRegistryListAll(b *testing.B) {
	c, err := client.NewClient("bench-list")
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer c.Disconnect()

	time.Sleep(100 * time.Millisecond)
	registry := c.GetRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.ListAll()
	}
}

func BenchmarkRegistryCountObjects(b *testing.B) {
	c, err := client.NewClient("bench-count")
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer c.Disconnect()

	time.Sleep(100 * time.Millisecond)
	registry := c.GetRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.CountObjects()
	}
}
