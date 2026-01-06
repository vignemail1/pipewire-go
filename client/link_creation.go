// Package client - link_creation.go
// Link creation and removal helper types
// Phase 3 - Link Management

package client

import (
	"fmt"
	"sync"
)

// ============================================================================
// LINK PARAMETERS & POD STRUCTURES
// ============================================================================

// LinkCreateParams contains parameters for creating a new link
type LinkCreateParams struct {
	OutputNodeID  uint32
	OutputPortID  uint32
	InputNodeID   uint32
	InputPortID   uint32
	Properties    map[string]string
	PassiveLink   bool
	PhysicalLinks bool
	MonitorLinks  bool
}

// Validate checks if link creation parameters are valid
func (lcp *LinkCreateParams) Validate() error {
	if lcp.OutputPortID == 0 {
		return fmt.Errorf("output port ID cannot be 0")
	}
	if lcp.InputPortID == 0 {
		return fmt.Errorf("input port ID cannot be 0")
	}
	if lcp.OutputPortID == lcp.InputPortID {
		return fmt.Errorf("cannot create link: output and input ports are the same")
	}
	return nil
}

// PODLinkCreate represents a POD structure for link.create method
// This is sent to the core object to create a new link
// Reference: pw_core_proxy_create (id/object_type/properties)
type PODLinkCreate struct {
	// PipeWire object type for this structure
	ObjectType string // "Link" or "pw.Link"

	// Version of the Link interface
	Version uint32

	// Input port ID
	InputPort uint32

	// Output port ID
	OutputPort uint32

	// Link properties as POD dict
	// Common properties:
	// - "passive": "true"/"false" - if true, link doesn't drive execution
	// - "link.passive": "true"/"false" - alias
	// - "object.loglevel": "0"-"5" - logging level
	Properties map[string]string
}

// PODLinkDestroy represents a POD structure for link.destroy method
// Sent to destroy a link object
type PODLinkDestroy struct {
	// Link object ID to destroy
	LinkID uint32
}

// LinkRemovalContext tracks a link removal request
type LinkRemovalContext struct {
	mu         sync.RWMutex
	LinkID     uint32
	OutputPort uint32
	InputPort  uint32
	State      LinkRemovalState
	Error      error
}

// LinkRemovalState represents the state of a link removal
type LinkRemovalState int

const (
	LinkRemovalStatePending LinkRemovalState = iota
	LinkRemovalStateWaitingForDestroy
	LinkRemovalStateDestroyed
	LinkRemovalStateError
)

// String returns a human-readable state
func (lrs LinkRemovalState) String() string {
	switch lrs {
	case LinkRemovalStatePending:
		return "pending"
	case LinkRemovalStateWaitingForDestroy:
		return "waiting_for_destroy"
	case LinkRemovalStateDestroyed:
		return "destroyed"
	case LinkRemovalStateError:
		return "error"
	default:
		return "unknown"
	}
}

// LinkCreationContext tracks a link creation request
type LinkCreationContext struct {
	mu            sync.RWMutex
	Params        *LinkCreateParams
	State         LinkCreationState
	CreatedLinkID uint32
	Error         error
}

// LinkCreationState represents the state of a link creation
type LinkCreationState int

const (
	LinkCreationStatePending LinkCreationState = iota
	LinkCreationStateWaitingForRegistry
	LinkCreationStateCreated
	LinkCreationStateError
)

// String returns a human-readable state
func (lcs LinkCreationState) String() string {
	switch lcs {
	case LinkCreationStatePending:
		return "pending"
	case LinkCreationStateWaitingForRegistry:
		return "waiting_for_registry"
	case LinkCreationStateCreated:
		return "created"
	case LinkCreationStateError:
		return "error"
	default:
		return "unknown"
	}
}

// GetState safely returns the current state
func (lcc *LinkCreationContext) GetState() LinkCreationState {
	lcc.mu.RLock()
	defer lcc.mu.RUnlock()
	return lcc.State
}

// SetState safely sets the state
func (lcc *LinkCreationContext) SetState(state LinkCreationState) {
	lcc.mu.Lock()
	defer lcc.mu.Unlock()
	lcc.State = state
}

// SetCreated marks the link as created
func (lcc *LinkCreationContext) SetCreated(linkID uint32) {
	lcc.mu.Lock()
	defer lcc.mu.Unlock()
	lcc.State = LinkCreationStateCreated
	lcc.CreatedLinkID = linkID
}

// SetError marks the creation as failed
func (lcc *LinkCreationContext) SetError(err error) {
	lcc.mu.Lock()
	defer lcc.mu.Unlock()
	lcc.State = LinkCreationStateError
	lcc.Error = err
}

// GetError safely returns the error
func (lcc *LinkCreationContext) GetError() error {
	lcc.mu.RLock()
	defer lcc.mu.RUnlock()
	return lcc.Error
}

// LinkValidator provides validation utilities for link operations
type LinkValidator struct {
	client *Client
}

// NewLinkValidator creates a new link validator
func NewLinkValidator(client *Client) *LinkValidator {
	return &LinkValidator{client: client}
}

// CanCreateLink checks if a link can be created between two ports
// Returns (canCreate, reason)
func (lv *LinkValidator) CanCreateLink(outputPort, inputPort *Port) (bool, string) {
	if outputPort == nil {
		return false, "output port is nil"
	}
	if inputPort == nil {
		return false, "input port is nil"
	}

	// Check port directions
	if outputPort.info.Direction != PortDirectionOutput {
		return false, fmt.Sprintf("output port has wrong direction: %v", outputPort.info.Direction)
	}
	if inputPort.info.Direction != PortDirectionInput {
		return false, fmt.Sprintf("input port has wrong direction: %v", inputPort.info.Direction)
	}

	// Check if ports are already connected
	existingLink := lv.findExistingLink(outputPort.ID(), inputPort.ID())
	if existingLink != nil {
		return false, fmt.Sprintf("link already exists: %d -> %d", outputPort.ID(), inputPort.ID())
	}

	return true, ""
}

// findExistingLink checks if a link already exists between two ports
func (lv *LinkValidator) findExistingLink(outputPortID, inputPortID uint32) *Link {
	links := lv.client.GetLinks()
	for _, link := range links {
		if link.OutputPort() != nil && link.InputPort() != nil {
			if link.OutputPort().ID() == outputPortID && link.InputPort().ID() == inputPortID {
				return link
			}
		}
	}
	return nil
}
