// Package client - High-Level PipeWire Client
// client/endpoint.go
// Endpoint management
// Phase 2 - Endpoint support

package client

import (
	"fmt"
	"sync"
)

// Endpoint represents a PipeWire endpoint (group of ports)
type Endpoint struct {
	mu         sync.RWMutex
	id         uint32
	name       string
	endpointType EndpointType
	properties *Properties
	ports      []*Port
	node       *Node
	client     *Client
	info       *EndpointInfo
}

// EndpointType represents the type of endpoint
type EndpointType int

const (
	EndpointTypeSource EndpointType = iota
	EndpointTypeSink
	EndpointTypeGeneric
)

// EndpointInfo contains detailed endpoint information
type EndpointInfo struct {
	ID         uint32
	Name       string
	Type       EndpointType
	NodeID     uint32
	Ports      []uint32
	Properties map[string]string
}

// NewEndpoint creates a new endpoint
func NewEndpoint(id uint32, name string, endpointType EndpointType, node *Node, client *Client) *Endpoint {
	return &Endpoint{
		id:           id,
		name:         name,
		endpointType: endpointType,
		properties:   NewProperties(),
		ports:        make([]*Port, 0),
		node:         node,
		client:       client,
		info: &EndpointInfo{
			ID:    id,
			Name:  name,
			Type:  endpointType,
			Ports: make([]uint32, 0),
		},
	}
}

// ID returns the endpoint ID
func (e *Endpoint) ID() uint32 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.id
}

// Name returns the endpoint name
func (e *Endpoint) Name() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.name
}

// Type returns the endpoint type
func (e *Endpoint) Type() EndpointType {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.endpointType
}

// SetType sets the endpoint type
func (e *Endpoint) SetType(t EndpointType) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.endpointType = t
	e.info.Type = t
}

// Node returns the parent node
func (e *Endpoint) Node() *Node {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.node
}

// Properties returns the endpoint properties
func (e *Endpoint) Properties() *Properties {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.properties
}

// Info returns detailed endpoint information
func (e *Endpoint) Info() *EndpointInfo {
	e.mu.RLock()
	defer e.mu.RUnlock()
	info := *e.info
	info.Ports = make([]uint32, len(e.info.Ports))
	copy(info.Ports, e.info.Ports)
	return &info
}

// SetInfo updates endpoint information
func (e *Endpoint) SetInfo(info *EndpointInfo) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if info != nil {
		e.info = info
	}
}

// AddPort adds a port to the endpoint
func (e *Endpoint) AddPort(port *Port) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if port == nil {
		return fmt.Errorf("port is nil")
	}

	// Check if port already exists
	for _, p := range e.ports {
		if p.ID() == port.ID() {
			return nil // Already added
		}
	}

	e.ports = append(e.ports, port)
	e.info.Ports = append(e.info.Ports, port.ID())
	return nil
}

// RemovePort removes a port from the endpoint
func (e *Endpoint) RemovePort(portID uint32) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Remove from ports slice
	for i, p := range e.ports {
		if p.ID() == portID {
			e.ports = append(e.ports[:i], e.ports[i+1:]...)
			break
		}
	}

	// Remove from info
	for i, id := range e.info.Ports {
		if id == portID {
			e.info.Ports = append(e.info.Ports[:i], e.info.Ports[i+1:]...)
			break
		}
	}
}

// GetPorts returns all ports in this endpoint
func (e *Endpoint) GetPorts() []*Port {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ports := make([]*Port, len(e.ports))
	copy(ports, e.ports)
	return ports
}

// GetPort returns a specific port by ID
func (e *Endpoint) GetPort(portID uint32) (*Port, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, p := range e.ports {
		if p.ID() == portID {
			return p, true
		}
	}
	return nil, false
}

// PortCount returns the number of ports
func (e *Endpoint) PortCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.ports)
}

// GetAudioPorts returns only audio ports
func (e *Endpoint) GetAudioPorts() []*Port {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var audio []*Port
	for _, p := range e.ports {
		if p.IsAudioPort() {
			audio = append(audio, p)
		}
	}
	return audio
}

// GetMIDIPorts returns only MIDI ports
func (e *Endpoint) GetMIDIPorts() []*Port {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var midi []*Port
	for _, p := range e.ports {
		if p.IsMIDIPort() {
			midi = append(midi, p)
		}
	}
	return midi
}

// IsSource returns true if this is a source endpoint
func (e *Endpoint) IsSource() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.endpointType == EndpointTypeSource
}

// IsSink returns true if this is a sink endpoint
func (e *Endpoint) IsSink() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.endpointType == EndpointTypeSink
}

// String returns string representation
func (e *Endpoint) String() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return fmt.Sprintf("Endpoint(%s, type=%v, ports=%d)", e.name, e.endpointType, len(e.ports))
}

// GetConnectedEndpoints returns all endpoints connected to this one
func (e *Endpoint) GetConnectedEndpoints() ([]*Endpoint, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.client == nil {
		return nil, fmt.Errorf("endpoint not associated with client")
	}

	return e.client.GetConnectedEndpoints(e.id)
}

// TypeString returns string representation of endpoint type
func (et EndpointType) String() string {
	switch et {
	case EndpointTypeSource:
		return "source"
	case EndpointTypeSink:
		return "sink"
	case EndpointTypeGeneric:
		return "generic"
	default:
		return "unknown"
	}
}
