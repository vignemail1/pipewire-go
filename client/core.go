// Package client - High-Level PipeWire Client
// client/core.go
// Core and Registry proxy implementation
// Phase 2 - Core object management

package client

import (
	"fmt"
	"sync"

	"github.com/vignemail1/pipewire-go/core"
	"github.com/vignemail1/pipewire-go/verbose"
)

// ============================================================================
// OBJECT REGISTRY (Gestion du cache local)
// ============================================================================

// ObjectRegistry manages and caches PipeWire objects
type ObjectRegistry struct {
	mu        sync.RWMutex
	nodes     map[uint32]*Node
	ports     map[uint32]*Port
	links     map[uint32]*Link
	endpoints map[uint32]*Endpoint
	watchers  map[string][]ObjectWatcher
	idCounter uint32
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

// ============================================================================
// CORE PROXY (Object ID=0)
// ============================================================================

// Core is the PipeWire Core object proxy (object id=0)
// It represents the root object for the daemon connection
type Core struct {
	proxy      *core.Proxy
	conn       *core.Connection
	logger     *verbose.Logger
	mu         sync.RWMutex
	version    uint32
	name       string
	properties map[string]string
}

// newCore creates a new Core proxy
func newCore(id uint32, conn *core.Connection, logger *verbose.Logger) *Core {
	return &Core{
		proxy:      core.NewProxy(id, "core", conn),
		conn:       conn,
		logger:     logger,
		version:    0,
		name:       "Core",
		properties: make(map[string]string),
	}
}

// Ping sends a ping message to verify the connection is alive
func (c *Core) Ping() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.conn.IsConnected() {
		return core.NewConnectionError("connection is closed")
	}

	// Get a sequence ID for this request
	seq := c.conn.GetSyncID()

	c.logger.Debugf("Core: Sending ping (seq=%d)", seq)

	// In a real implementation, we would:
	// 1. Marshal a POD with the ping request
	// 2. Send via connection
	// 3. Wait for pong response

	// For now, we just log and return success
	// TODO: Implement actual ping/pong message exchange
	return nil
}

// Sync sends a sync request and waits for the daemon to acknowledge
// This is used as a synchronization point in the protocol
func (c *Core) Sync() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.conn.IsConnected() {
		return core.NewConnectionError("connection is closed")
	}

	// Get a sequence ID for this sync request
	seq := c.conn.GetSyncID()

	c.logger.Debugf("Core: Sending sync (seq=%d)", seq)

	// In a real implementation, we would:
	// 1. Send a sync method call with this sequence
	// 2. Wait for a done event with matching sequence
	// 3. Return once received

	// For now, we just log and return success
	// TODO: Implement actual sync message exchange and waiting
	return nil
}

// Version returns the core protocol version
func (c *Core) Version() uint32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.version
}

// SetVersion sets the core protocol version
func (c *Core) SetVersion(v uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.version = v
}

// Name returns the core name
func (c *Core) Name() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.name
}

// Properties returns the core properties
func (c *Core) Properties() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	props := make(map[string]string)
	for k, v := range c.properties {
		props[k] = v
	}
	return props
}

// SetProperties sets the core properties
func (c *Core) SetProperties(props map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.properties = props
}

// String returns a human-readable representation
func (c *Core) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return fmt.Sprintf("Core{version=%d, name=%s}", c.version, c.name)
}

// ============================================================================
// REGISTRY PROXY (Object ID=1)
// ============================================================================

// GlobalObject represents a PipeWire object advertised by the registry
type GlobalObject struct {
	ID         uint32
	Type       string // e.g. "Node", "Port", "Link"
	Version    uint32
	Properties map[string]string
}

// RegistryListener is called when objects are added/removed
type RegistryListener func(event RegistryEvent)

// RegistryEvent represents a global object event
type RegistryEvent struct {
	Type     RegistryEventType
	Object   *GlobalObject
	ObjectID uint32
}

// RegistryEventType represents the type of registry event
type RegistryEventType int

const (
	RegistryEventTypeGlobal RegistryEventType = iota
	RegistryEventTypeGlobalRemove
)

// Registry is the PipeWire Registry object proxy (object id=1)
// It manages discovery and binding to global PipeWire objects
type Registry struct {
	proxy     *core.Proxy
	conn      *core.Connection
	logger    *verbose.Logger
	mu        sync.RWMutex
	objects   map[uint32]*GlobalObject // Discovered global objects
	listeners map[uint32][]RegistryListener
}

// newRegistry creates a new Registry proxy
func newRegistry(id uint32, conn *core.Connection, logger *verbose.Logger) *Registry {
	return &Registry{
		proxy:     core.NewProxy(id, "registry", conn),
		conn:      conn,
		logger:    logger,
		objects:   make(map[uint32]*GlobalObject),
		listeners: make(map[uint32][]RegistryListener),
	}
}

// AddListener registers a listener for all global object events
func (r *Registry) AddListener(listener RegistryListener) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Use ID 0 as the key for "all events"
	r.listeners[0] = append(r.listeners[0], listener)
}

// RemoveListener unregisters a listener
func (r *Registry) RemoveListener(listener RegistryListener) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if listeners, exists := r.listeners[0]; exists {
		for i, l := range listeners {
			if fmt.Sprintf("%p", l) == fmt.Sprintf("%p", listener) {
				r.listeners[0] = append(listeners[:i], listeners[i+1:]...)
				break
			}
		}
	}
}

// Objects returns all discovered global objects
func (r *Registry) Objects() []*GlobalObject {
	r.mu.RLock()
	defer r.mu.RUnlock()
	objects := make([]*GlobalObject, 0, len(r.objects))
	for _, obj := range r.objects {
		objects = append(objects, obj)
	}
	return objects
}

// GetObject retrieves a specific global object by ID
func (r *Registry) GetObject(id uint32) (*GlobalObject, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	obj, exists := r.objects[id]
	return obj, exists
}

// ObjectsByType returns all objects of a specific type
func (r *Registry) ObjectsByType(objType string) []*GlobalObject {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*GlobalObject
	for _, obj := range r.objects {
		if obj.Type == objType {
			result = append(result, obj)
		}
	}
	return result
}

// handleGlobal is called when a new global object is advertised
// This is typically called from the event loop when registry.global event is received
func (r *Registry) handleGlobal(id uint32, objType string, version uint32, props map[string]string) {
	r.mu.Lock()
	obj := &GlobalObject{
		ID:         id,
		Type:       objType,
		Version:    version,
		Properties: props,
	}
	r.objects[id] = obj
	r.mu.Unlock()

	r.logger.Debugf("Registry: Global object added: id=%d type=%s version=%d", id, objType, version)

	// Notify listeners
	r.mu.RLock()
	listeners := r.listeners[0]
	r.mu.RUnlock()

	event := RegistryEvent{
		Type:   RegistryEventTypeGlobal,
		Object: obj,
	}
	for _, listener := range listeners {
		go listener(event)
	}
}

// handleGlobalRemove is called when a global object is removed
// This is typically called from the event loop when registry.global_remove event is received
func (r *Registry) handleGlobalRemove(id uint32) {
	r.mu.Lock()
	obj, exists := r.objects[id]
	delete(r.objects, id)
	r.mu.Unlock()

	r.logger.Debugf("Registry: Global object removed: id=%d", id)

	if !exists {
		return
	}

	// Notify listeners
	r.mu.RLock()
	listeners := r.listeners[0]
	r.mu.RUnlock()

	event := RegistryEvent{
		Type:     RegistryEventTypeGlobalRemove,
		Object:   obj,
		ObjectID: id,
	}
	for _, listener := range listeners {
		go listener(event)
	}
}

// Count returns the number of global objects
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.objects)
}

// String returns a human-readable representation
func (r *Registry) String() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return fmt.Sprintf("Registry{objects=%d}", len(r.objects))
}
