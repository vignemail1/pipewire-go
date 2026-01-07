// Package pipewireio provides comprehensive bindings for the PipeWire audio server.
//
// PipeWire is a modern audio and video server that aims to be a drop-in replacement
// for both PulseAudio and JACK. This library provides Go bindings for interacting with
// PipeWire's powerful audio graph capabilities.
//
// # Quick Start
//
// Connect to PipeWire and enumerate the audio graph:
//
//	c, err := client.NewClient("my-app")
//	if err != nil {
//		log.Fatalf("Failed to connect: %v", err)
//	}
//	defer c.Disconnect()
//
//	registry := c.GetRegistry()
//	nodes := registry.ListNodes()
//	
//	for _, node := range nodes {
//		fmt.Printf("Node: %s (ID: %d)\n", node.Name, node.ID)
//	}
//
// # Core Concepts
//
// PipeWire's audio system is organized around a few key concepts:
//
// - **Nodes**: Audio processing entities (e.g., applications, hardware devices)
// - **Ports**: Input/output endpoints on nodes (e.g., speaker outputs, microphone inputs)
// - **Links**: Connections between ports (audio routing)
// - **Registry**: Central repository listing all objects in the system
// - **Properties**: Metadata associated with nodes, ports, and links
//
// # Working with the Registry
//
// The Registry gives you access to all PipeWire objects:
//
//	registry := client.GetRegistry()
//	
//	// List all nodes
//	nodes := registry.ListNodes()
//	
//	// List all ports
//	ports := registry.ListPorts()
//	
//	// List all links
//	links := registry.ListLinks()
//	
//	// Get specific object
//	node := registry.GetObject(nodeID)
//
// # Event Handling
//
// Monitor changes to the audio graph with event listeners:
//
//	registry.OnGlobalAdded(func(obj *client.GlobalObject) {
//		fmt.Printf("Object added: %s (ID: %d)\n", obj.Type, obj.ID)
//	})
//
// # Creating Connections
//
// Connect audio ports to route audio:
//
//	link := &client.Link{
//		OutputPort: sourcePortID,
//		InputPort:  destPortID,
//	}
//	
//	err := client.CreateLink(link)
//	if err != nil {
//		log.Fatalf("Failed to create link: %v", err)
//	}
//
// # Error Handling
//
// The library provides typed errors for better error handling:
//
//	err := client.Connect()
//	if errors.Is(err, pipewireio.ErrNotConnected) {
//		log.Fatal("PipeWire daemon not running")
//	}
//	
//	var valErr *pipewireio.ValidationError
//	if errors.As(err, &valErr) {
//		log.Printf("Validation failed in field %s: %s", valErr.Field, valErr.Message)
//	}
//
// # Thread Safety
//
// All public APIs are thread-safe. You can safely use the Client and Registry
// from multiple goroutines concurrently. The library uses sync.RWMutex to protect
// concurrent access to shared data structures.
//
// # Performance Considerations
//
// - ListAll() performs a full copy of the registry - use sparingly for large graphs
// - Event listeners are called synchronously - keep handlers fast
// - Registry lookups are O(1) - GetObject(id) is always fast
// - Use filters when possible to reduce data copying
//
// # Best Practices
//
// 1. Always defer Disconnect() after creating a client
// 2. Handle errors appropriately - use typed errors for context
// 3. Keep event handlers fast to avoid blocking the event loop
// 4. Cache frequently accessed objects to reduce lookups
// 5. Use read locks (RLock) for queries, write locks (Lock) for modifications
//
// # Examples
//
// See the examples/ directory for complete working examples:
// - error_handling - Error handling patterns
// - basic_client - Basic client usage
// - graph_enumeration - Walking the audio graph
// - event_monitoring - Monitoring graph changes
//
// # External Resources
//
// - PipeWire Wiki: https://pipewire.org/
// - PipeWire Documentation: https://docs.pipewire.org/
// - PipeWire Source: https://gitlab.freedesktop.org/pipewire/pipewire
package pipewireio
