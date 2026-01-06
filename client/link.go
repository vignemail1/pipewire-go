// Package client - High-Level PipeWire Client
// client/link.go
// Link management and port connections
// Phase 2 - Link handling

package client

import (
	"fmt"
	"sync"
	"time"
)

// Link represents a connection between two ports
type Link struct {
	mu         sync.RWMutex
	id         uint32
	inputPort  *Port
	outputPort *Port
	properties *Properties
	client     *Client
	info       *LinkInfo
	state      LinkState
	createdAt  time.Time
}

// LinkState represents the state of a link
type LinkState int

const (
	LinkStateUnknown LinkState = iota
	LinkStateActive
	LinkStateInactive
	LinkStateError
)

// LinkInfo contains detailed link information
type LinkInfo struct {
	ID         uint32
	InputPort  uint32
	OutputPort uint32
	Input      *Port
	Output     *Port
	Properties map[string]string
	State      LinkState
	Created    time.Time
}

// NewLink creates a new link between two ports
func NewLink(id uint32, inputPort, outputPort *Port, client *Client) *Link {
	return &Link{
		id:         id,
		inputPort:  inputPort,
		outputPort: outputPort,
		properties: NewProperties(),
		client:     client,
		state:      LinkStateActive,
		createdAt:  time.Now(),
		info: &LinkInfo{
			ID:         id,
			InputPort:  inputPort.ID(),
			OutputPort: outputPort.ID(),
			Input:      inputPort,
			Output:     outputPort,
			State:      LinkStateActive,
			Created:    time.Now(),
		},
	}
}

// ID returns the link ID
func (l *Link) ID() uint32 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.id
}

// InputPort returns the input port
func (l *Link) InputPort() *Port {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.inputPort
}

// OutputPort returns the output port
func (l *Link) OutputPort() *Port {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.outputPort
}

// Properties returns the link properties
func (l *Link) Properties() *Properties {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.properties
}

// Info returns detailed link information
func (l *Link) Info() *LinkInfo {
	l.mu.RLock()
	defer l.mu.RUnlock()
	info := *l.info
	return &info
}

// SetInfo updates link information
func (l *Link) SetInfo(info *LinkInfo) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if info != nil {
		l.info = info
	}
}

// State returns the link state
func (l *Link) State() LinkState {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.state
}

// SetState sets the link state
func (l *Link) SetState(state LinkState) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.state = state
	l.info.State = state
}

// IsActive returns true if the link is active
func (l *Link) IsActive() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.state == LinkStateActive
}

// CreatedAt returns the link creation time
func (l *Link) CreatedAt() time.Time {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.createdAt
}

// Duration returns the time since link creation
func (l *Link) Duration() time.Duration {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return time.Since(l.createdAt)
}

// Disconnect removes this link
func (l *Link) Disconnect() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.client == nil {
		return fmt.Errorf("link not associated with client")
	}
	if l.outputPort == nil || l.inputPort == nil {
		return fmt.Errorf("link ports are not set")
	}

	// Call the client's DisconnectPorts method
	return l.client.DisconnectPorts(l.outputPort.ID(), l.inputPort.ID())
}

// IsValid checks if the link is valid
func (l *Link) IsValid() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.inputPort == nil || l.outputPort == nil {
		return false
	}

	return l.inputPort.ID() != 0 && l.outputPort.ID() != 0
}

// String returns string representation
func (l *Link) String() string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	inputName := "nil"
	outputName := "nil"

	if l.inputPort != nil {
		inputName = l.inputPort.Name()
	}
	if l.outputPort != nil {
		outputName = l.outputPort.Name()
	}

	return fmt.Sprintf("Link(%s -> %s, state=%v)", outputName, inputName, l.state)
}

// Reverse returns a new link with reversed ports
// (useful for checking backward connections)
func (l *Link) Reverse() *Link {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return &Link{
		id:         l.id,
		inputPort:  l.outputPort,
		outputPort: l.inputPort,
		properties: l.properties.Copy(),
		client:     l.client,
		state:      l.state,
		createdAt:  l.createdAt,
		info: &LinkInfo{
			ID:         l.info.ID,
			InputPort:  l.info.OutputPort,
			OutputPort: l.info.InputPort,
			Input:      l.info.Output,
			Output:     l.info.Input,
			Properties: l.info.Properties,
			State:      l.info.State,
			Created:    l.info.Created,
		},
	}
}

// LinkManager manages collections of links
type LinkManager struct {
	mu    sync.RWMutex
	links map[uint32]*Link
}

// NewLinkManager creates a new link manager
func NewLinkManager() *LinkManager {
	return &LinkManager{
		links: make(map[uint32]*Link),
	}
}

// Add adds a link to the manager
func (lm *LinkManager) Add(link *Link) error {
	if link == nil {
		return fmt.Errorf("link is nil")
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.links[link.ID()] = link
	return nil
}

// Remove removes a link from the manager
func (lm *LinkManager) Remove(id uint32) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	delete(lm.links, id)
}

// Get retrieves a link by ID
func (lm *LinkManager) Get(id uint32) (*Link, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	link, ok := lm.links[id]
	return link, ok
}

// All returns all links
func (lm *LinkManager) All() []*Link {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	links := make([]*Link, 0, len(lm.links))
	for _, link := range lm.links {
		links = append(links, link)
	}
	return links
}

// Count returns the number of links
func (lm *LinkManager) Count() int {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return len(lm.links)
}

// Clear removes all links
func (lm *LinkManager) Clear() {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.links = make(map[uint32]*Link)
}

// FilterByPort returns all links connected to a port
func (lm *LinkManager) FilterByPort(portID uint32) []*Link {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	var result []*Link
	for _, link := range lm.links {
		if (link.inputPort != nil && link.inputPort.ID() == portID) ||
			(link.outputPort != nil && link.outputPort.ID() == portID) {
			result = append(result, link)
		}
	}
	return result
}

// FilterByNode returns all links from/to a node's ports
func (lm *LinkManager) FilterByNode(nodeID uint32) []*Link {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	var result []*Link
	for _, link := range lm.links {
		if (link.inputPort != nil && link.inputPort.Node() != nil && link.inputPort.Node().ID() == nodeID) ||
			(link.outputPort != nil && link.outputPort.Node() != nil && link.outputPort.Node().ID() == nodeID) {
			result = append(result, link)
		}
	}
	return result
}

// FilterByState returns all links in a specific state
func (lm *LinkManager) FilterByState(state LinkState) []*Link {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	var result []*Link
	for _, link := range lm.links {
		if link.State() == state {
			result = append(result, link)
		}
	}
	return result
}

// FindLink finds a link between two specific ports
func (lm *LinkManager) FindLink(outputPortID, inputPortID uint32) *Link {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	for _, link := range lm.links {
		if link.outputPort != nil && link.inputPort != nil &&
			link.outputPort.ID() == outputPortID &&
			link.inputPort.ID() == inputPortID {
			return link
		}
	}
	return nil
}

// StateString returns string representation of link state
func (ls LinkState) String() string {
	switch ls {
	case LinkStateActive:
		return "active"
	case LinkStateInactive:
		return "inactive"
	case LinkStateError:
		return "error"
	case LinkStateUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}
