// Package client - High-level protocol operations
// client/protocol_client.go
// Implements CreateLink, DestroyLink, and other protocol operations

package client

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vignemail1/pipewire-go/core"
)

// ProtocolClient provides high-level protocol operations
// for creating and managing links and other PipeWire objects
type ProtocolClient struct {
	mu              sync.RWMutex
	connection      *core.Connection
	eventHandler    *core.EventHandler
	registryID      uint32
	coreID          uint32
	logger          *core.VerboseLogger
	lastSequence    uint32
	requestTimeout  time.Duration
}

// NewProtocolClient creates a new protocol client
func NewProtocolClient(conn *core.Connection, registryID, coreID uint32, logger *core.VerboseLogger) *ProtocolClient {
	if logger == nil {
		logger = core.NewVerboseLogger(core.LogLevelInfo, false)
	}

	return &ProtocolClient{
		connection:     conn,
		eventHandler:   core.NewEventHandler(),
		registryID:     registryID,
		coreID:         coreID,
		logger:         logger,
		lastSequence:   0,
		requestTimeout: 5 * time.Second,
	}
}

// nextSequence generates the next sequence number
func (pc *ProtocolClient) nextSequence() uint32 {
	return atomic.AddUint32(&pc.lastSequence, 1)
}

// SetRequestTimeout configures the timeout for protocol requests
func (pc *ProtocolClient) SetRequestTimeout(timeout time.Duration) {
	if pc != nil && timeout > 0 {
		pc.mu.Lock()
		defer pc.mu.Unlock()
		pc.requestTimeout = timeout
		pc.eventHandler.SetRequestTimeout(timeout)
	}
}

// RegisterEventHandler registers a handler for events from an object
func (pc *ProtocolClient) RegisterEventHandler(objectID uint32, handler core.HandlerFunc) error {
	if pc == nil {
		return fmt.Errorf("ProtocolClient is nil")
	}
	if pc.eventHandler == nil {
		return fmt.Errorf("EventHandler is nil")
	}

	return pc.eventHandler.RegisterHandler(objectID, handler)
}

// UnregisterEventHandler removes handlers for an object
func (pc *ProtocolClient) UnregisterEventHandler(objectID uint32) {
	if pc != nil && pc.eventHandler != nil {
		pc.eventHandler.UnregisterHandler(objectID)
	}
}

// CreateLink creates a connection between two ports
// Returns the link ID assigned by the daemon, or error
func (pc *ProtocolClient) CreateLink(outputPortID, inputPortID uint32, properties map[string]string) (uint32, error) {
	if pc == nil {
		return 0, fmt.Errorf("ProtocolClient is nil")
	}

	pc.mu.Lock()
	sequence := pc.nextSequence()
	timeout := pc.requestTimeout
	pc.mu.Unlock()

	pc.logger.Logf(core.LogLevelDebug, "CreateLink: output=%d input=%d seq=%d", outputPortID, inputPortID, sequence)

	// Build request
	req := &core.LinkCreateRequest{
		OutputPortID: outputPortID,
		InputPortID:  inputPortID,
		Properties:   properties,
		Passive:      false,
	}

	// Convert to POD
	podObj, err := req.ToPOD()
	if err != nil {
		return 0, fmt.Errorf("failed to convert to POD: %w", err)
	}

	// Create message frame
	frame := core.NewMessageBuilder(pc.registryID, 4). // Method 4 = registry.bind()
		WithSequence(sequence).
		WithPOD(podObj).
		Build()

	if frame == nil {
		return 0, fmt.Errorf("failed to build message frame")
	}

	// Create pending request context
	ctx := pc.eventHandler.CreatePendingRequest(sequence)
	if ctx == nil {
		return 0, fmt.Errorf("failed to create pending request")
	}

	// Marshal and send message
	data, err := frame.Marshal()
	if err != nil {
		return 0, fmt.Errorf("failed to marshal frame: %w", err)
	}

	if pc.connection == nil {
		return 0, fmt.Errorf("connection is nil")
	}

	if err := pc.connection.Write(data); err != nil {
		return 0, fmt.Errorf("failed to send message: %w", err)
	}

	pc.logger.Logf(core.LogLevelDebug, "CreateLink: sent message, waiting for response")

	// Wait for response
	result, err := pc.eventHandler.WaitForRequest(ctx, timeout)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}

	// Extract link ID from result
	if linkID, ok := result.(uint32); ok {
		pc.logger.Logf(core.LogLevelDebug, "CreateLink: received link ID %d", linkID)
		return linkID, nil
	}

	return 0, fmt.Errorf("unexpected response type: %T", result)
}

// DestroyLink destroys an existing link
func (pc *ProtocolClient) DestroyLink(linkID uint32) error {
	if pc == nil {
		return fmt.Errorf("ProtocolClient is nil")
	}

	pc.mu.Lock()
	sequence := pc.nextSequence()
	timeout := pc.requestTimeout
	pc.mu.Unlock()

	pc.logger.Logf(core.LogLevelDebug, "DestroyLink: link=%d seq=%d", linkID, sequence)

	// Build request
	req := &core.LinkDestroyRequest{
		LinkID: linkID,
	}

	// Convert to POD
	podObj, err := req.ToPOD()
	if err != nil {
		return fmt.Errorf("failed to convert to POD: %w", err)
	}

	// Create message frame for destroy method (method 0)
	frame := core.NewMessageBuilder(linkID, 0).
		WithSequence(sequence).
		WithPOD(podObj).
		Build()

	if frame == nil {
		return fmt.Errorf("failed to build message frame")
	}

	// Create pending request context
	ctx := pc.eventHandler.CreatePendingRequest(sequence)
	if ctx == nil {
		return fmt.Errorf("failed to create pending request")
	}

	// Marshal and send message
	data, err := frame.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal frame: %w", err)
	}

	if pc.connection == nil {
		return fmt.Errorf("connection is nil")
	}

	if err := pc.connection.Write(data); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	pc.logger.Logf(core.LogLevelDebug, "DestroyLink: sent message, waiting for response")

	// Wait for response
	_, err = pc.eventHandler.WaitForRequest(ctx, timeout)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	pc.logger.Logf(core.LogLevelDebug, "DestroyLink: link destroyed successfully")
	return nil
}

// SetLinkActive sets the active state of a link
func (pc *ProtocolClient) SetLinkActive(linkID uint32, active bool) error {
	if pc == nil {
		return fmt.Errorf("ProtocolClient is nil")
	}

	pc.mu.Lock()
	sequence := pc.nextSequence()
	timeout := pc.requestTimeout
	pc.mu.Unlock()

	pc.logger.Logf(core.LogLevelDebug, "SetLinkActive: link=%d active=%v seq=%d", linkID, active, sequence)

	// Build request
	props := make(map[string]string)
	props["state"] = fmt.Sprintf("%v", active)

	req := &core.LinkCreateRequest{
		OutputPortID: 0, // Not used for state change
		InputPortID:  0, // Not used for state change
		Properties:   props,
		Passive:      !active,
	}

	// Convert to POD
	podObj, err := req.ToPOD()
	if err != nil {
		return fmt.Errorf("failed to convert to POD: %w", err)
	}

	// Create message frame (method 2 = set_param)
	frame := core.NewMessageBuilder(linkID, 2).
		WithSequence(sequence).
		WithPOD(podObj).
		Build()

	if frame == nil {
		return fmt.Errorf("failed to build message frame")
	}

	// Create pending request context
	ctx := pc.eventHandler.CreatePendingRequest(sequence)
	if ctx == nil {
		return fmt.Errorf("failed to create pending request")
	}

	// Marshal and send message
	data, err := frame.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal frame: %w", err)
	}

	if pc.connection == nil {
		return fmt.Errorf("connection is nil")
	}

	if err := pc.connection.Write(data); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	pc.logger.Logf(core.LogLevelDebug, "SetLinkActive: sent message, waiting for response")

	// Wait for response
	_, err = pc.eventHandler.WaitForRequest(ctx, timeout)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	pc.logger.Logf(core.LogLevelDebug, "SetLinkActive: link state updated successfully")
	return nil
}

// DispatchMessage sends a message frame to registered handlers
func (pc *ProtocolClient) DispatchMessage(frame *core.MessageFrame) error {
	if pc == nil {
		return fmt.Errorf("ProtocolClient is nil")
	}
	if pc.eventHandler == nil {
		return fmt.Errorf("EventHandler is nil")
	}

	return pc.eventHandler.Dispatch(frame)
}

// GetEventHandler returns the underlying event handler
func (pc *ProtocolClient) GetEventHandler() *core.EventHandler {
	if pc != nil {
		return pc.eventHandler
	}
	return nil
}

// GetConnection returns the underlying connection
func (pc *ProtocolClient) GetConnection() *core.Connection {
	if pc != nil {
		return pc.connection
	}
	return nil
}

// PendingRequestCount returns the number of pending requests
func (pc *ProtocolClient) PendingRequestCount() int {
	if pc != nil && pc.eventHandler != nil {
		return pc.eventHandler.PendingRequestCount()
	}
	return 0
}
