package client

import (
	"sync"
)

// Registry holds all PipeWire objects (nodes, ports, links)
type Registry struct {
	mu    sync.RWMutex      // Protects all maps
	nodes map[uint32]*Node
	ports map[uint32]*Port
	links map[uint32]*Link
}

// NewRegistry creates a new empty registry
func NewRegistry() *Registry {
	return &Registry{
		nodes: make(map[uint32]*Node),
		ports: make(map[uint32]*Port),
		links: make(map[uint32]*Link),
	}
}

// Node operations

// AddNode adds a node to the registry (thread-safe)
func (r *Registry) AddNode(node *Node) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodes[node.ID] = node
}

// GetNode retrieves a node by ID (thread-safe)
func (r *Registry) GetNode(id uint32) *Node {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.nodes[id]
}

// GetNodes returns all nodes (thread-safe)
func (r *Registry) GetNodes() []*Node {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	nodes := make([]*Node, 0, len(r.nodes))
	for _, node := range r.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// RemoveNode removes a node from the registry (thread-safe)
func (r *Registry) RemoveNode(id uint32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.nodes, id)
}

// Port operations

// AddPort adds a port to the registry (thread-safe)
func (r *Registry) AddPort(port *Port) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ports[port.ID] = port
}

// GetPort retrieves a port by ID (thread-safe)
func (r *Registry) GetPort(id uint32) *Port {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ports[id]
}

// GetPorts returns all ports (thread-safe)
func (r *Registry) GetPorts() []*Port {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	ports := make([]*Port, 0, len(r.ports))
	for _, port := range r.ports {
		ports = append(ports, port)
	}
	return ports
}

// RemovePort removes a port from the registry (thread-safe)
func (r *Registry) RemovePort(id uint32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.ports, id)
}

// Link operations

// AddLink adds a link to the registry (thread-safe)
func (r *Registry) AddLink(link *Link) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.links[link.ID] = link
}

// GetLink retrieves a link by ID (thread-safe)
func (r *Registry) GetLink(id uint32) *Link {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.links[id]
}

// GetLinks returns all links (thread-safe)
func (r *Registry) GetLinks() []*Link {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	links := make([]*Link, 0, len(r.links))
	for _, link := range r.links {
		links = append(links, link)
	}
	return links
}

// RemoveLink removes a link from the registry (thread-safe)
func (r *Registry) RemoveLink(id uint32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.links, id)
}

// Clear removes all entries from the registry (thread-safe)
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.nodes = make(map[uint32]*Node)
	r.ports = make(map[uint32]*Port)
	r.links = make(map[uint32]*Link)
}
