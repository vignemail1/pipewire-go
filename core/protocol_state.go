// Package core - Protocol state machine for PipeWire connection
// core/protocol_state.go
// Issue #6 - Event Loop & Automatic Dispatch

package core

import (
	"fmt"
	"sync"
)

// State represents the current state of the protocol connection
type State int

const (
	StateDisconnected State = iota
	StateConnected
	StateHelloSent
	StateHelloReceived
	StateReady
	StateError
)

// String representation of states
func (s State) String() string {
	switch s {
	case StateDisconnected:
		return "Disconnected"
	case StateConnected:
		return "Connected"
	case StateHelloSent:
		return "HelloSent"
	case StateHelloReceived:
		return "HelloReceived"
	case StateReady:
		return "Ready"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

// ProtocolVersion represents the PipeWire protocol version
type ProtocolVersion struct {
	Major uint32
	Minor uint32
}

// String representation of version
func (v ProtocolVersion) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

// ProtocolStateMachine manages the protocol connection state
type ProtocolStateMachine struct {
	mu              sync.RWMutex
	state           State
	version         ProtocolVersion
	serverID        uint32
	serverCapabilities []string
	lastError       error
}

// NewProtocolStateMachine creates a new protocol state machine
func NewProtocolStateMachine() *ProtocolStateMachine {
	return &ProtocolStateMachine{
		state: StateDisconnected,
		version: ProtocolVersion{
			Major: 3,
			Minor: 0,
		},
	}
}

// TransitionTo attempts to transition to a new state
func (p *ProtocolStateMachine) TransitionTo(newState State) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Validate state transition
	if !p.isValidTransition(p.state, newState) {
		err := fmt.Errorf("invalid state transition: %s → %s", p.state, newState)
		p.lastError = err
		return err
	}

	p.state = newState
	return nil
}

// isValidTransition checks if a state transition is allowed
func (p *ProtocolStateMachine) isValidTransition(from, to State) bool {
	switch from {
	case StateDisconnected:
		return to == StateConnected || to == StateError
	case StateConnected:
		return to == StateHelloSent || to == StateError || to == StateDisconnected
	case StateHelloSent:
		return to == StateHelloReceived || to == StateError || to == StateDisconnected
	case StateHelloReceived:
		return to == StateReady || to == StateError || to == StateDisconnected
	case StateReady:
		return to == StateDisconnected || to == StateError
	case StateError:
		return to == StateDisconnected
	default:
		return false
	}
}

// GetState returns the current state
func (p *ProtocolStateMachine) GetState() State {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

// IsReady returns true if connection is ready for communication
func (p *ProtocolStateMachine) IsReady() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state == StateReady
}

// CanSendMessage returns true if messages can be sent
func (p *ProtocolStateMachine) CanSendMessage() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state == StateReady
}

// CanReceiveMessage returns true if messages can be received
func (p *ProtocolStateMachine) CanReceiveMessage() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state == StateReady || p.state == StateHelloReceived
}

// SetVersion sets the protocol version
func (p *ProtocolStateMachine) SetVersion(major, minor uint32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.version.Major = major
	p.version.Minor = minor
}

// GetVersion returns the protocol version
func (p *ProtocolStateMachine) GetVersion() ProtocolVersion {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.version
}

// SetServerID sets the server object ID
func (p *ProtocolStateMachine) SetServerID(id uint32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.serverID = id
}

// GetServerID returns the server object ID
func (p *ProtocolStateMachine) GetServerID() uint32 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.serverID
}

// SetServerCapabilities sets the server capabilities
func (p *ProtocolStateMachine) SetServerCapabilities(caps []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.serverCapabilities = make([]string, len(caps))
	copy(p.serverCapabilities, caps)
}

// GetServerCapabilities returns the server capabilities
func (p *ProtocolStateMachine) GetServerCapabilities() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	caps := make([]string, len(p.serverCapabilities))
	copy(caps, p.serverCapabilities)
	return caps
}

// HasCapability checks if server has a capability
func (p *ProtocolStateMachine) HasCapability(capability string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, cap := range p.serverCapabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

// GetLastError returns the last error
func (p *ProtocolStateMachine) GetLastError() error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lastError
}

// SetError sets an error state
func (p *ProtocolStateMachine) SetError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lastError = err
	p.state = StateError
}

// Reset resets the state machine to disconnected
func (p *ProtocolStateMachine) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = StateDisconnected
	p.lastError = nil
	p.serverID = 0
	p.serverCapabilities = nil
}
