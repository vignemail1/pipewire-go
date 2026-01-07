// Package core - Event handler and request/response dispatch
// core/event_handler.go
// Handles event routing and request/response matching

package core

import (
	"fmt"
	"sync"
	"time"
)

// HandlerFunc is the signature for event handler callbacks
type HandlerFunc func(*MessageFrame) error

// RequestContext tracks a pending request and its response
type RequestContext struct {
	Sequence uint32
	Result   chan interface{}
	Error    chan error
	Timeout  time.Time
}

// EventHandler routes incoming events to registered handlers
// and matches responses to pending requests
type EventHandler struct {
	mu               sync.RWMutex
	handlers         map[uint32][]HandlerFunc           // ObjectID -> handlers
	pendingRequests  map[uint32]*RequestContext         // Sequence -> context
	requestTimeout   time.Duration
}

// NewEventHandler creates a new event handler with default 5 second timeout
func NewEventHandler() *EventHandler {
	return &EventHandler{
		handlers:        make(map[uint32][]HandlerFunc),
		pendingRequests: make(map[uint32]*RequestContext),
		requestTimeout:  5 * time.Second,
	}
}

// SetRequestTimeout configures the timeout for request/response matching
func (eh *EventHandler) SetRequestTimeout(timeout time.Duration) {
	if eh != nil {
		eh.mu.Lock()
		defer eh.mu.Unlock()
		eh.requestTimeout = timeout
	}
}

// RegisterHandler registers a handler for events from a specific object
// Multiple handlers can be registered for the same object
func (eh *EventHandler) RegisterHandler(objectID uint32, handler HandlerFunc) error {
	if eh == nil {
		return fmt.Errorf("EventHandler is nil")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	eh.mu.Lock()
	defer eh.mu.Unlock()

	eh.handlers[objectID] = append(eh.handlers[objectID], handler)
	return nil
}

// UnregisterHandler removes all handlers for a specific object
func (eh *EventHandler) UnregisterHandler(objectID uint32) {
	if eh == nil {
		return
	}

	eh.mu.Lock()
	defer eh.mu.Unlock()

	delete(eh.handlers, objectID)
}

// Dispatch sends a message frame to all registered handlers for its object
// Returns error if any handler returns error
func (eh *EventHandler) Dispatch(frame *MessageFrame) error {
	if eh == nil {
		return fmt.Errorf("EventHandler is nil")
	}
	if frame == nil {
		return fmt.Errorf("MessageFrame is nil")
	}

	// Get handlers for this object (with read lock)
	eh.mu.RLock()
	handlers, ok := eh.handlers[frame.ObjectID]
	eh.mu.RUnlock()

	if !ok {
		return fmt.Errorf("no handlers registered for object %d", frame.ObjectID)
	}

	// Call handlers outside lock to avoid deadlocks
	for _, handler := range handlers {
		if err := handler(frame); err != nil {
			return fmt.Errorf("handler error: %w", err)
		}
	}

	return nil
}

// DispatchToAll sends a frame to all handlers that exist, ignoring if no handlers
func (eh *EventHandler) DispatchToAll(frame *MessageFrame) {
	if eh == nil || frame == nil {
		return
	}

	eh.mu.RLock()
	handlers, ok := eh.handlers[frame.ObjectID]
	eh.mu.RUnlock()

	if !ok {
		return
	}

	for _, handler := range handlers {
		_ = handler(frame)
	}
}

// CreatePendingRequest creates a new request context and registers it
func (eh *EventHandler) CreatePendingRequest(sequence uint32) *RequestContext {
	if eh == nil {
		return nil
	}

	ctx := &RequestContext{
		Sequence: sequence,
		Result:   make(chan interface{}, 1),
		Error:    make(chan error, 1),
		Timeout:  time.Now().Add(eh.requestTimeout),
	}

	eh.mu.Lock()
	eh.pendingRequests[sequence] = ctx
	eh.mu.Unlock()

	return ctx
}

// WaitForRequest waits for a response to a pending request
func (eh *EventHandler) WaitForRequest(ctx *RequestContext, timeout time.Duration) (interface{}, error) {
	if eh == nil {
		return nil, fmt.Errorf("EventHandler is nil")
	}
	if ctx == nil {
		return nil, fmt.Errorf("RequestContext is nil")
	}

	actualTimeout := timeout
	if actualTimeout == 0 {
		actualTimeout = eh.requestTimeout
	}

	select {
	case result := <-ctx.Result:
		eh.mu.Lock()
		delete(eh.pendingRequests, ctx.Sequence)
		eh.mu.Unlock()
		return result, nil

	case err := <-ctx.Error:
		eh.mu.Lock()
		delete(eh.pendingRequests, ctx.Sequence)
		eh.mu.Unlock()
		return nil, err

	case <-time.After(actualTimeout):
		eh.mu.Lock()
		delete(eh.pendingRequests, ctx.Sequence)
		eh.mu.Unlock()
		return nil, fmt.Errorf("request %d timeout after %v", ctx.Sequence, actualTimeout)
	}
}

// ResolvePendingRequest completes a pending request with a result
func (eh *EventHandler) ResolvePendingRequest(sequence uint32, result interface{}) error {
	if eh == nil {
		return fmt.Errorf("EventHandler is nil")
	}

	eh.mu.RLock()
	ctx, ok := eh.pendingRequests[sequence]
	eh.mu.RUnlock()

	if !ok {
		return fmt.Errorf("no pending request for sequence %d", sequence)
	}

	select {
	case ctx.Result <- result:
		return nil
	case <-time.After(100 * time.Millisecond):
		return fmt.Errorf("failed to send result for sequence %d", sequence)
	}
}

// RejectPendingRequest completes a pending request with an error
func (eh *EventHandler) RejectPendingRequest(sequence uint32, err error) error {
	if eh == nil {
		return fmt.Errorf("EventHandler is nil")
	}

	eh.mu.RLock()
	ctx, ok := eh.pendingRequests[sequence]
	eh.mu.RUnlock()

	if !ok {
		return fmt.Errorf("no pending request for sequence %d", sequence)
	}

	select {
	case ctx.Error <- err:
		return nil
	case <-time.After(100 * time.Millisecond):
		return fmt.Errorf("failed to send error for sequence %d", sequence)
	}
}

// GetPendingRequest retrieves a pending request context
func (eh *EventHandler) GetPendingRequest(sequence uint32) (*RequestContext, bool) {
	if eh == nil {
		return nil, false
	}

	eh.mu.RLock()
	defer eh.mu.RUnlock()

	ctx, ok := eh.pendingRequests[sequence]
	return ctx, ok
}

// ClearPendingRequests removes all pending requests (for cleanup)
func (eh *EventHandler) ClearPendingRequests() {
	if eh == nil {
		return
	}

	eh.mu.Lock()
	defer eh.mu.Unlock()

	for _, ctx := range eh.pendingRequests {
		close(ctx.Result)
		close(ctx.Error)
	}
	eh.pendingRequests = make(map[uint32]*RequestContext)
}

// PendingRequestCount returns the number of pending requests
func (eh *EventHandler) PendingRequestCount() int {
	if eh == nil {
		return 0
	}

	eh.mu.RLock()
	defer eh.mu.RUnlock()

	return len(eh.pendingRequests)
}

// HandlerCount returns the number of handlers for a specific object
func (eh *EventHandler) HandlerCount(objectID uint32) int {
	if eh == nil {
		return 0
	}

	eh.mu.RLock()
	defer eh.mu.RUnlock()

	return len(eh.handlers[objectID])
}
