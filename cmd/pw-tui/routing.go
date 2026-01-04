// Package routing - Audio Routing Manager
// cmd/pw-tui/routing.go
// Handles creation, modification and deletion of audio links

package main

import (
	"fmt"
	"sync"

	"github.com/vignemail1/pipewire-go/client"
	"github.com/vignemail1/pipewire-go/core"
)

// RoutingManager handles audio routing operations
type RoutingManager struct {
	client     *client.Client
	mu         sync.RWMutex
	lastError  error
	history    []*RoutingOperation
	maxHistory int
}

// RoutingOperation represents a single routing operation
type RoutingOperation struct {
	Type      OperationType
	LinkID    uint32
	OutputID  uint32
	InputID   uint32
	Timestamp int64
	Success   bool
	Error     error
}

// OperationType represents the type of routing operation
type OperationType int

const (
	OpTypeCreate OperationType = iota
	OpTypeDelete
	OpTypeModify
	OpTypeConnect
	OpTypeDisconnect
)

// NewRoutingManager creates a new routing manager
func NewRoutingManager(c *client.Client) *RoutingManager {
	return &RoutingManager{
		client:     c,
		history:    make([]*RoutingOperation, 0),
		maxHistory: 100,
	}
}

// CreateLink creates a new audio link between output and input ports
func (rm *RoutingManager) CreateLink(outputID, inputID uint32) (uint32, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.client == nil {
		return 0, fmt.Errorf("client not initialized")
	}

	// Validate ports exist
	output := rm.client.GetPort(outputID)
	if output == nil {
		return 0, core.NewError(core.ErrResourceNotFound, "output port not found")
	}

	input := rm.client.GetPort(inputID)
	if input == nil {
		return 0, core.NewError(core.ErrResourceNotFound, "input port not found")
	}

	// Validate port directions
	if output.Direction == client.PortDirectionInput {
		return 0, core.NewError(core.ErrInvalidArgument, "output port must be an output port")
	}

	if input.Direction == client.PortDirectionOutput {
		return 0, core.NewError(core.ErrInvalidArgument, "input port must be an input port")
	}

	// Check for existing link
	for _, link := range rm.client.GetLinks() {
		if link.Output.ID == outputID && link.Input.ID == inputID {
			return link.ID, core.NewError(core.ErrResourceExists, "link already exists")
		}
	}

	// Create the link
	linkID, err := rm.client.CreateLink(outputID, inputID)
	if err != nil {
		rm.recordOperation(&RoutingOperation{
			Type:     OpTypeCreate,
			OutputID: outputID,
			InputID:  inputID,
			Success:  false,
			Error:    err,
		})
		return 0, err
	}

	rm.recordOperation(&RoutingOperation{
		Type:     OpTypeCreate,
		LinkID:   linkID,
		OutputID: outputID,
		InputID:  inputID,
		Success:  true,
	})

	return linkID, nil
}

// DeleteLink removes an audio link
func (rm *RoutingManager) DeleteLink(linkID uint32) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	link := rm.client.GetLink(linkID)
	if link == nil {
		return core.NewError(core.ErrResourceNotFound, "link not found")
	}

	err := rm.client.RemoveLink(linkID)
	if err != nil {
		rm.recordOperation(&RoutingOperation{
			Type:    OpTypeDelete,
			LinkID:  linkID,
			Success: false,
			Error:   err,
		})
		return err
	}

	rm.recordOperation(&RoutingOperation{
		Type:    OpTypeDelete,
		LinkID:  linkID,
		Success: true,
	})

	return nil
}

// ConnectPorts attempts to connect two ports intelligently
func (rm *RoutingManager) ConnectPorts(port1ID, port2ID uint32) (uint32, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	port1 := rm.client.GetPort(port1ID)
	port2 := rm.client.GetPort(port2ID)

	if port1 == nil || port2 == nil {
		return 0, core.NewError(core.ErrResourceNotFound, "one or both ports not found")
	}

	// Determine direction
	var outputID, inputID uint32
	if port1.Direction == client.PortDirectionOutput && port2.Direction == client.PortDirectionInput {
		outputID = port1ID
		inputID = port2ID
	} else if port1.Direction == client.PortDirectionInput && port2.Direction == client.PortDirectionOutput {
		outputID = port2ID
		inputID = port1ID
	} else {
		return 0, core.NewError(core.ErrInvalidArgument, "cannot connect ports with same direction")
	}

	return rm.createLinkInternal(outputID, inputID)
}

// createLinkInternal is the internal link creation method
func (rm *RoutingManager) createLinkInternal(outputID, inputID uint32) (uint32, error) {
	linkID, err := rm.client.CreateLink(outputID, inputID)
	if err != nil {
		return 0, err
	}
	return linkID, nil
}

// DisconnectPort disconnects all connections from a port
func (rm *RoutingManager) DisconnectPort(portID uint32) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	port := rm.client.GetPort(portID)
	if port == nil {
		return core.NewError(core.ErrResourceNotFound, "port not found")
	}

	var linksToDelete []*client.Link

	// Find all links connected to this port
	for _, link := range rm.client.GetLinks() {
		if (link.Output != nil && link.Output.ID == portID) ||
			(link.Input != nil && link.Input.ID == portID) {
			linksToDelete = append(linksToDelete, link)
		}
	}

	// Delete all links
	for _, link := range linksToDelete {
		if err := rm.client.RemoveLink(link.ID); err != nil {
			return err
		}
	}

	return nil
}

// GetConnectionCount returns the number of active connections
func (rm *RoutingManager) GetConnectionCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	count := 0
	for _, link := range rm.client.GetLinks() {
		if link.IsActive() {
			count++
		}
	}
	return count
}

// GetCompatiblePorts returns ports that can be connected to a given port
func (rm *RoutingManager) GetCompatiblePorts(portID uint32) []*client.Port {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	port := rm.client.GetPort(portID)
	if port == nil {
		return nil
	}

	compatible := make([]*client.Port, 0)

	// Get all ports with opposite direction
	allPorts := rm.client.GetPorts()
	for _, p := range allPorts {
		// Skip the same port
		if p.ID == portID {
			continue
		}

		// Check direction compatibility
		if port.Direction == client.PortDirectionOutput && p.Direction == client.PortDirectionInput {
			compatible = append(compatible, p)
		} else if port.Direction == client.PortDirectionInput && p.Direction == client.PortDirectionOutput {
			compatible = append(compatible, p)
		}
	}

	return compatible
}

// RoutingPreset represents a saved routing configuration
type RoutingPreset struct {
	Name        string
	Description string
	Links       []*LinkPreset
}

// LinkPreset represents a link in a preset
type LinkPreset struct {
	OutputNodeID uint32
	OutputPortName string
	InputNodeID  uint32
	InputPortName  string
}

// ApplyPreset applies a routing preset
func (rm *RoutingManager) ApplyPreset(preset *RoutingPreset) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	successCount := 0
	var lastError error

	for _, linkPreset := range preset.Links {
		// Find matching ports
		var outputPort, inputPort *client.Port

		for _, port := range rm.client.GetPorts() {
			if port.Name == linkPreset.OutputPortName && port.NodeID == linkPreset.OutputNodeID {
				outputPort = port
			}
			if port.Name == linkPreset.InputPortName && port.NodeID == linkPreset.InputNodeID {
				inputPort = port
			}
		}

		if outputPort == nil || inputPort == nil {
			lastError = core.NewError(core.ErrResourceNotFound, 
				fmt.Sprintf("preset port not found: %s or %s", 
					linkPreset.OutputPortName, linkPreset.InputPortName))
			continue
		}

		// Create link
		_, err := rm.createLinkInternal(outputPort.ID, inputPort.ID)
		if err != nil {
			lastError = err
			continue
		}

		successCount++
	}

	if successCount == 0 && lastError != nil {
		return lastError
	}

	return nil
}

// GetHistory returns the routing operation history
func (rm *RoutingManager) GetHistory() []*RoutingOperation {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	copy := make([]*RoutingOperation, len(rm.history))
	copy = append(copy, rm.history...)
	return copy
}

// recordOperation records a routing operation in history
func (rm *RoutingManager) recordOperation(op *RoutingOperation) {
	if rm.history == nil {
		rm.history = make([]*RoutingOperation, 0)
	}

	rm.history = append(rm.history, op)

	// Keep history size under control
	if len(rm.history) > rm.maxHistory {
		rm.history = rm.history[1:]
	}

	rm.lastError = op.Error
}

// RoutingValidator validates routing operations
type RoutingValidator struct{}

// ValidateLink checks if a link can be created
func (rv *RoutingValidator) ValidateLink(output, input *client.Port) error {
	if output == nil || input == nil {
		return core.NewError(core.ErrInvalidArgument, "ports cannot be nil")
	}

	// Check directions
	if output.Direction != client.PortDirectionOutput {
		return core.NewError(core.ErrInvalidArgument, "output port must have output direction")
	}

	if input.Direction != client.PortDirectionInput {
		return core.NewError(core.ErrInvalidArgument, "input port must have input direction")
	}

	// Check if already connected
	if output.IsConnected() && input.IsConnected() {
		// Some ports support multiple connections
		return nil
	}

	return nil
}

// ValidateRouting validates a complete routing preset
func (rv *RoutingValidator) ValidateRouting(links []*client.Link) error {
	if len(links) == 0 {
		return core.NewError(core.ErrInvalidArgument, "no links to validate")
	}

	// Check for duplicate links
	linkMap := make(map[string]bool)
	for _, link := range links {
		if link.Output == nil || link.Input == nil {
			return core.NewError(core.ErrMalformed, "link missing output or input")
		}

		key := fmt.Sprintf("%d->%d", link.Output.ID, link.Input.ID)
		if linkMap[key] {
			return core.NewError(core.ErrResourceExists, "duplicate link found")
		}
		linkMap[key] = true
	}

	return nil
}

// RoutingAnalyzer analyzes the routing topology
type RoutingAnalyzer struct{}

// DetectLoops detects routing loops that could cause issues
func (ra *RoutingAnalyzer) DetectLoops(nodes map[uint32]*client.Node, links map[uint32]*client.Link) bool {
	// Build adjacency list
	adjacency := make(map[uint32][]uint32)
	for _, link := range links {
		if link.Output != nil && link.Input != nil {
			adjacency[link.Output.NodeID] = append(adjacency[link.Output.NodeID], link.Input.NodeID)
		}
	}

	// Check for cycles using DFS
	visited := make(map[uint32]bool)
	recStack := make(map[uint32]bool)

	for nodeID := range nodes {
		if !visited[nodeID] {
			if ra.hasCycle(nodeID, adjacency, visited, recStack) {
				return true
			}
		}
	}

	return false
}

// hasCycle helper for cycle detection
func (ra *RoutingAnalyzer) hasCycle(node uint32, adj map[uint32][]uint32, 
	visited map[uint32]bool, recStack map[uint32]bool) bool {
	
	visited[node] = true
	recStack[node] = true

	for _, neighbor := range adj[node] {
		if !visited[neighbor] {
			if ra.hasCycle(neighbor, adj, visited, recStack) {
				return true
			}
		} else if recStack[neighbor] {
			return true
		}
	}

	recStack[node] = false
	return false
}

// AnalyzeLatency analyzes end-to-end latency
func (ra *RoutingAnalyzer) AnalyzeLatency(links []*client.Link) uint32 {
	totalLatency := uint32(0)
	for _, link := range links {
		if link != nil {
			// Each link adds minimal latency
			totalLatency += 1
		}
	}
	return totalLatency
}
