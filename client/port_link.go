// Package client - port.go
// Port proxy implementation

package client

import (
	"fmt"
	"sync"

	"github.com/yourusername/pipewire-go/core"
	"github.com/yourusername/pipewire-go/verbose"
)

// Port represents a PipeWire port
type Port struct {
	ID        uint32
	NodeID    uint32
	Name      string
	Direction PortDirection
	Type      PortType
	Version   uint32
	Props     map[string]string
	propMut   sync.RWMutex
	
	conn      *core.Connection
	logger    *verbose.Logger
	
	// Parent node reference
	ParentNode *Node
	
	// Connected links
	links   map[uint32]*Link
	linkMut sync.RWMutex
}

// newPort creates a new Port proxy
func newPort(id uint32, nodeID uint32, objType string, version uint32, props map[string]string, conn *core.Connection, logger *verbose.Logger) *Port {
	port := &Port{
		ID:        id,
		NodeID:    nodeID,
		Name:      props["port.name"],
		Direction: parsePortDirection(props["port.direction"]),
		Type:      parsePortType(props["port.type"]),
		Version:   version,
		Props:     make(map[string]string),
		conn:      conn,
		logger:    logger,
		links:     make(map[uint32]*Link),
	}
	
	for k, v := range props {
		port.Props[k] = v
	}
	
	return port
}

// parsePortDirection parses port direction from string
func parsePortDirection(dirStr string) PortDirection {
	switch dirStr {
	case "Input", "in":
		return PortDirectionInput
	case "Output", "out":
		return PortDirectionOutput
	default:
		return PortDirectionInput
	}
}

// parsePortType parses port type from string
func parsePortType(typeStr string) PortType {
	switch typeStr {
	case "Audio":
		return PortTypeAudio
	case "Midi":
		return PortTypeMidi
	case "Video":
		return PortTypeVideo
	case "Control":
		return PortTypeControl
	default:
		return PortTypeAudio
	}
}

// IsInput checks if this is an input port
func (p *Port) IsInput() bool {
	return p.Direction == PortDirectionInput
}

// IsOutput checks if this is an output port
func (p *Port) IsOutput() bool {
	return p.Direction == PortDirectionOutput
}

// IsConnected checks if this port has any connections
func (p *Port) IsConnected() bool {
	p.linkMut.RLock()
	defer p.linkMut.RUnlock()
	return len(p.links) > 0
}

// GetLinks returns all links connected to this port
func (p *Port) GetLinks() []*Link {
	p.linkMut.RLock()
	defer p.linkMut.RUnlock()
	
	links := make([]*Link, 0, len(p.links))
	for _, link := range p.links {
		links = append(links, link)
	}
	return links
}

// GetConnectedPorts returns all ports connected to this one
func (p *Port) GetConnectedPorts() []*Port {
	links := p.GetLinks()
	var ports []*Port
	
	for _, link := range links {
		if link.Output.ID == p.ID && link.Input != nil {
			ports = append(ports, link.Input)
		} else if link.Input.ID == p.ID && link.Output != nil {
			ports = append(ports, link.Output)
		}
	}
	
	return ports
}

// AddLink associates a link with this port
func (p *Port) AddLink(link *Link) {
	p.linkMut.Lock()
	defer p.linkMut.Unlock()
	p.links[link.ID] = link
}

// RemoveLink disassociates a link from this port
func (p *Port) RemoveLink(linkID uint32) {
	p.linkMut.Lock()
	defer p.linkMut.Unlock()
	delete(p.links, linkID)
}

// GetProperty retrieves a port property
func (p *Port) GetProperty(key string) (string, bool) {
	p.propMut.RLock()
	defer p.propMut.RUnlock()
	val, ok := p.Props[key]
	return val, ok
}

// GetProperties returns all port properties
func (p *Port) GetProperties() map[string]string {
	p.propMut.RLock()
	defer p.propMut.RUnlock()
	
	props := make(map[string]string)
	for k, v := range p.Props {
		props[k] = v
	}
	return props
}

// String returns a human-readable port description
func (p *Port) String() string {
	var dirStr string
	if p.Direction == PortDirectionInput {
		dirStr = "→"
	} else {
		dirStr = "←"
	}
	return fmt.Sprintf("Port{ID:%d Name:%q %s %s}", p.ID, p.Name, p.Type, dirStr)
}

// Link represents a connection between two ports
type Link struct {
	ID         uint32
	OutputPort uint32
	InputPort  uint32
	Type       string
	Version    uint32
	Props      map[string]string
	propMut    sync.RWMutex
	
	conn   *core.Connection
	logger *verbose.Logger
	
	// Port references
	Output *Port
	Input  *Port
}

// newLink creates a new Link proxy
func newLink(id uint32, outPort uint32, inPort uint32, objType string, version uint32, props map[string]string, conn *core.Connection, logger *verbose.Logger) *Link {
	link := &Link{
		ID:         id,
		OutputPort: outPort,
		InputPort:  inPort,
		Type:       objType,
		Version:    version,
		Props:      make(map[string]string),
		conn:       conn,
		logger:     logger,
	}
	
	for k, v := range props {
		link.Props[k] = v
	}
	
	return link
}

// IsActive checks if the link is active
func (l *Link) IsActive() bool {
	l.propMut.RLock()
	defer l.propMut.RUnlock()
	
	// Check properties to determine if link is active
	// This depends on PipeWire implementation details
	if state, ok := l.Props["link.state"]; ok {
		return state == "active" || state == "running"
	}
	return true // Default to active if not specified
}

// GetProperty retrieves a link property
func (l *Link) GetProperty(key string) (string, bool) {
	l.propMut.RLock()
	defer l.propMut.RUnlock()
	val, ok := l.Props[key]
	return val, ok
}

// GetProperties returns all link properties
func (l *Link) GetProperties() map[string]string {
	l.propMut.RLock()
	defer l.propMut.RUnlock()
	
	props := make(map[string]string)
	for k, v := range l.Props {
		props[k] = v
	}
	return props
}

// UpdateProperties updates link properties
func (l *Link) UpdateProperties(props map[string]string) error {
	l.propMut.Lock()
	for k, v := range props {
		l.Props[k] = v
	}
	l.propMut.Unlock()
	
	l.logger.Debugf("Link %d: UpdateProperties", l.ID)
	
	// Send update_properties message to server
	// Details depend on protocol implementation
	
	return nil
}

// Remove removes this link
func (l *Link) Remove() error {
	l.logger.Debugf("Link %d: Remove", l.ID)
	
	// Send remove message to server
	// Details depend on protocol implementation
	
	return nil
}

// String returns a human-readable link description
func (l *Link) String() string {
	if l.Output != nil && l.Input != nil {
		return fmt.Sprintf("Link{ID:%d %s→%s}", l.ID, l.Output.Name, l.Input.Name)
	}
	return fmt.Sprintf("Link{ID:%d %d→%d}", l.ID, l.OutputPort, l.InputPort)
}
