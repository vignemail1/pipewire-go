// Package client - High-Level PipeWire Client
// client/core.go
// Object registry and discovery
// Phase 2 - Core object management

package client

import (
	"fmt"
	"sync"
)

// ObjectRegistry manages and caches PipeWire objects
type ObjectRegistry struct {
	mu         sync.RWMutex
	nodes      map[uint32]*Node
	ports      map[uint32]*Port
	links      map[uint32]*Link
	endpoints  map[uint32]*Endpoint
	watchers   map[string][]ObjectWatcher
	idCounter  uint32
}

// ObjectWatcher is called when objects change
type ObjectWatcher func(event ObjectEvent)

// ObjectEvent represents a change to an object
type ObjectEvent struct {
	Type   ObjectEventType
	Object interface{}
}

// ObjectEventType represents the type of change
type ObjectEventType int

const (
	ObjectEventTypeAdded ObjectEventType = iota
	ObjectEventTypeRemoved
	ObjectEventTypeUpdated
)

// NewObjectRegistry creates a new object registry
func NewObjectRegistry() *ObjectRegistry {
	return &ObjectRegistry{
		nodes:     make(map[uint32]*Node),
		ports:     make(map[uint32]*Port),
		links:     make(map[uint32]*Link),
		endpoints: make(map[uint32]*Endpoint),
		watchers:  make(map[string][]ObjectWatcher),
		idCounter: 0,
	}
}

// RegisterNode adds or updates a node in the registry
func (r *ObjectRegistry) RegisterNode(node *Node) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if node.ID() == 0 {
		r.idCounter++
		// Note: In real implementation, would set node.id
	}

	r.nodes[node.ID()] = node
	r.notifyWatchers("node", ObjectEvent{
		Type:   ObjectEventTypeAdded,
		Object: node,
	})
}

// UnregisterNode removes a node from the registry
func (r *ObjectRegistry) UnregisterNode(id uint32) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if node, exists := r.nodes[id]; exists {
		delete(r.nodes, id)
		r.notifyWatchers("node", ObjectEvent{
			Type:   ObjectEventTypeRemoved,
			Object: node,
		})
	}
}

// GetNode retrieves a node by ID
func (r *ObjectRegistry) GetNode(id uint32) (*Node, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	node, exists := r.nodes[id]
	return node, exists
}

// GetNodes returns all nodes
func (r *ObjectRegistry) GetNodes() []*Node {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]*Node, 0, len(r.nodes))
	for _, node := range r.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// RegisterPort adds or updates a port in the registry
func (r *ObjectRegistry) RegisterPort(port *Port) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if port.ID() == 0 {
		r.idCounter++
	}

	r.ports[port.ID()] = port
	r.notifyWatchers("port", ObjectEvent{
		Type:   ObjectEventTypeAdded,
		Object: port,
	})
}

// UnregisterPort removes a port from the registry
func (r *ObjectRegistry) UnregisterPort(id uint32) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if port, exists := r.ports[id]; exists {
		delete(r.ports, id)
		r.notifyWatchers("port", ObjectEvent{
			Type:   ObjectEventTypeRemoved,
			Object: port,
		})
	}
}

// GetPort retrieves a port by ID
func (r *ObjectRegistry) GetPort(id uint32) (*Port, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	port, exists := r.ports[id]
	return port, exists
}

// GetPorts returns all ports
func (r *ObjectRegistry) GetPorts() []*Port {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ports := make([]*Port, 0, len(r.ports))
	for _, port := range r.ports {
		ports = append(ports, port)
	}
	return ports
}

// RegisterLink adds or updates a link in the registry
func (r *ObjectRegistry) RegisterLink(link *Link) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if link.ID() == 0 {
		r.idCounter++
	}

	r.links[link.ID()] = link
	r.notifyWatchers("link", ObjectEvent{
		Type:   ObjectEventTypeAdded,
		Object: link,
	})
}

// UnregisterLink removes a link from the registry
func (r *ObjectRegistry) UnregisterLink(id uint32) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if link, exists := r.links[id]; exists {
		delete(r.links, id)
		r.notifyWatchers("link", ObjectEvent{
			Type:   ObjectEventTypeRemoved,
			Object: link,
		})
	}
}

// GetLink retrieves a link by ID
func (r *ObjectRegistry) GetLink(id uint32) (*Link, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	link, exists := r.links[id]
	return link, exists
}

// GetLinks returns all links
func (r *ObjectRegistry) GetLinks() []*Link {
	r.mu.RLock()
	defer r.mu.RUnlock()

	links := make([]*Link, 0, len(r.links))
	for _, link := range r.links {
		links = append(links, link)
	}
	return links
}

// RegisterEndpoint adds or updates an endpoint in the registry
func (r *ObjectRegistry) RegisterEndpoint(endpoint *Endpoint) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if endpoint.ID() == 0 {
		r.idCounter++
	}

	r.endpoints[endpoint.ID()] = endpoint
	r.notifyWatchers("endpoint", ObjectEvent{
		Type:   ObjectEventTypeAdded,
		Object: endpoint,
	})
}

// UnregisterEndpoint removes an endpoint from the registry
func (r *ObjectRegistry) UnregisterEndpoint(id uint32) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if endpoint, exists := r.endpoints[id]; exists {
		delete(r.endpoints, id)
		r.notifyWatchers("endpoint", ObjectEvent{
			Type:   ObjectEventTypeRemoved,
			Object: endpoint,
		})
	}
}

// GetEndpoint retrieves an endpoint by ID
func (r *ObjectRegistry) GetEndpoint(id uint32) (*Endpoint, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	endpoint, exists := r.endpoints[id]
	return endpoint, exists
}

// GetEndpoints returns all endpoints
func (r *ObjectRegistry) GetEndpoints() []*Endpoint {
	r.mu.RLock()
	defer r.mu.RUnlock()

	endpoints := make([]*Endpoint, 0, len(r.endpoints))
	for _, endpoint := range r.endpoints {
		endpoints = append(endpoints, endpoint)
	}
	return endpoints
}

// Watch registers a watcher for object changes
func (r *ObjectRegistry) Watch(objectType string, watcher ObjectWatcher) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.watchers[objectType] = append(r.watchers[objectType], watcher)
}

// notifyWatchers notifies all watchers of an event
func (r *ObjectRegistry) notifyWatchers(objectType string, event ObjectEvent) {
	if watchers, exists := r.watchers[objectType]; exists {
		for _, watcher := range watchers {
			go watcher(event)
		}
	}
}

// Clear removes all objects from the registry
func (r *ObjectRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.nodes = make(map[uint32]*Node)
	r.ports = make(map[uint32]*Port)
	r.links = make(map[uint32]*Link)
	r.endpoints = make(map[uint32]*Endpoint)
}

// Count returns the count of each object type
func (r *ObjectRegistry) Count() (nodes, ports, links, endpoints int) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.nodes), len(r.ports), len(r.links), len(r.endpoints)
}

// Summary returns a string summary of registry contents
func (r *ObjectRegistry) Summary() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return fmt.Sprintf(
		"Registry(nodes=%d, ports=%d, links=%d, endpoints=%d)",
		len(r.nodes), len(r.ports), len(r.links), len(r.endpoints),
	)
}

// FindPortsByNode returns all ports belonging to a node
func (r *ObjectRegistry) FindPortsByNode(nodeID uint32) []*Port {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Port
	for _, port := range r.ports {
		if port.Node() != nil && port.Node().ID() == nodeID {
			result = append(result, port)
		}
	}
	return result
}

// FindLinksByPort returns all links connected to a port
func (r *ObjectRegistry) FindLinksByPort(portID uint32) []*Link {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Link
	for _, link := range r.links {
		if (link.InputPort() != nil && link.InputPort().ID() == portID) ||
			(link.OutputPort() != nil && link.OutputPort().ID() == portID) {
			result = append(result, link)
		}
	}
	return result
}

// FindNodeByName finds a node by its name
func (r *ObjectRegistry) FindNodeByName(name string) (*Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, node := range r.nodes {
		if node.Name() == name {
			return node, nil
		}
	}
	return nil, fmt.Errorf("node %q not found", name)
}

// FindPortByName finds a port by its name within a node
func (r *ObjectRegistry) FindPortByName(nodeName, portName string) (*Port, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, port := range r.ports {
		if port.Node() != nil && port.Node().Name() == nodeName && port.Name() == portName {
			return port, nil
		}
	}
	return nil, fmt.Errorf("port %q in node %q not found", portName, nodeName)
}
