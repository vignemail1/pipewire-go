// Package core - PipeWire Protocol Core
// core/handler.go
// Event handling and routing
// Phase 1 - Event system infrastructure

package core

import (
	"fmt"
	"sync"
	"time"
)

// EventHandler handles events from PipeWire
type EventHandler interface {
	Handle(event Event) error
}

// Event represents a PipeWire event
type Event interface {
	Type() EventType
	Timestamp() time.Time
}

// EventType represents the type of event
type EventType int

const (
	EventTypeUnknown EventType = iota
	EventTypeConnected
	EventTypeDisconnected
	EventTypeNodeAdded
	EventTypeNodeRemoved
	EventTypePortAdded
	EventTypePortRemoved
	EventTypeLinkAdded
	EventTypeLinkRemoved
	EventTypeObjectUpdated
	EventTypeError
	EventTypeInfo
)

// BaseEvent provides common event functionality
type BaseEvent struct {
	eventType EventType
	timestamp time.Time
	data      interface{}
}

// Type returns the event type
func (e *BaseEvent) Type() EventType {
	return e.eventType
}

// Timestamp returns the event timestamp
func (e *BaseEvent) Timestamp() time.Time {
	return e.timestamp
}

// Data returns event-specific data
func (e *BaseEvent) Data() interface{} {
	return e.data
}

// NewBaseEvent creates a new base event
func NewBaseEvent(eventType EventType, data interface{}) *BaseEvent {
	return &BaseEvent{
		eventType: eventType,
		timestamp: time.Now(),
		data:      data,
	}
}

// EventDispatcher routes events to handlers
type EventDispatcher struct {
	mu           sync.RWMutex
	handlers     map[EventType][]EventHandler
	errorHandler ErrorEventHandler
	queue        chan Event
	running      bool
	workers      int
	done         chan bool
}

// ErrorEventHandler handles error events
type ErrorEventHandler func(err error)

// NewEventDispatcher creates a new event dispatcher
func NewEventDispatcher(workers int) *EventDispatcher {
	if workers < 1 {
		workers = 1
	}
	if workers > 100 {
		workers = 100
	}

	return &EventDispatcher{
		handlers:     make(map[EventType][]EventHandler),
		queue:        make(chan Event, 1000),
		running:      false,
		workers:      workers,
		done:         make(chan bool),
	}
}

// RegisterHandler registers a handler for an event type
func (d *EventDispatcher) RegisterHandler(eventType EventType, handler EventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if handler == nil {
		return
	}

	if _, exists := d.handlers[eventType]; !exists {
		d.handlers[eventType] = make([]EventHandler, 0)
	}

	d.handlers[eventType] = append(d.handlers[eventType], handler)
}

// UnregisterHandler removes a handler
func (d *EventDispatcher) UnregisterHandler(eventType EventType, handler EventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if handlers, exists := d.handlers[eventType]; exists {
		for i, h := range handlers {
			if h == handler {
				d.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
	}
}

// SetErrorHandler sets the error handler
func (d *EventDispatcher) SetErrorHandler(handler ErrorEventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.errorHandler = handler
}

// Dispatch queues an event for processing
func (d *EventDispatcher) Dispatch(event Event) error {
	if !d.running {
		return fmt.Errorf("dispatcher not running")
	}

	select {
	case d.queue <- event:
		return nil
	case <-time.After(time.Second):
		return fmt.Errorf("event queue full")
	}
}

// Start starts the event dispatcher
func (d *EventDispatcher) Start() error {
	d.mu.Lock()
	if d.running {
		d.mu.Unlock()
		return fmt.Errorf("dispatcher already running")
	}
	d.running = true
	d.mu.Unlock()

	// Start worker goroutines
	for i := 0; i < d.workers; i++ {
		go d.worker()
	}

	return nil
}

// Stop stops the event dispatcher
func (d *EventDispatcher) Stop() error {
	d.mu.Lock()
	if !d.running {
		d.mu.Unlock()
		return fmt.Errorf("dispatcher not running")
	}
	d.running = false
	d.mu.Unlock()

	// Wait for workers to finish
	close(d.queue)
	for i := 0; i < d.workers; i++ {
		<-d.done
	}

	return nil
}

// worker processes events from the queue
func (d *EventDispatcher) worker() {
	defer func() {
		d.done <- true
	}()

	for event := range d.queue {
		d.handleEvent(event)
	}
}

// handleEvent dispatches an event to registered handlers
func (d *EventDispatcher) handleEvent(event Event) {
	d.mu.RLock()
	handlers, exists := d.handlers[event.Type()]
	d.mu.RUnlock()

	if !exists {
		return
	}

	for _, handler := range handlers {
		if err := handler.Handle(event); err != nil {
			d.reportError(err)
		}
	}
}

// reportError reports an error
func (d *EventDispatcher) reportError(err error) {
	d.mu.RLock()
	handler := d.errorHandler
	d.mu.RUnlock()

	if handler != nil {
		handler(err)
	}
}

// EventFilter filters events based on criteria
type EventFilter struct {
	EventTypes map[EventType]bool
	Before     time.Time
	After      time.Time
}

// Matches checks if an event matches the filter
func (f *EventFilter) Matches(event Event) bool {
	if f == nil {
		return true
	}

	// Check event type
	if len(f.EventTypes) > 0 {
		if !f.EventTypes[event.Type()] {
			return false
		}
	}

	// Check time range
	ts := event.Timestamp()
	if !f.Before.IsZero() && ts.After(f.Before) {
		return false
	}
	if !f.After.IsZero() && ts.Before(f.After) {
		return false
	}

	return true
}

// EventBuffer buffers events
type EventBuffer struct {
	mu     sync.RWMutex
	events []Event
	maxLen int
}

// NewEventBuffer creates a new event buffer
func NewEventBuffer(maxLen int) *EventBuffer {
	return &EventBuffer{
		events: make([]Event, 0, maxLen),
		maxLen: maxLen,
	}
}

// Add adds an event to the buffer
func (b *EventBuffer) Add(event Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.events = append(b.events, event)
	if len(b.events) > b.maxLen {
		b.events = b.events[1:]
	}
}

// Get returns all buffered events
func (b *EventBuffer) Get() []Event {
	b.mu.RLock()
	defer b.mu.RUnlock()

	events := make([]Event, len(b.events))
	copy(events, b.events)
	return events
}

// Clear clears the buffer
func (b *EventBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.events = b.events[:0]
}

// Count returns the number of buffered events
func (b *EventBuffer) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.events)
}

// TypeString returns string representation of event type
func (et EventType) String() string {
	switch et {
	case EventTypeConnected:
		return "connected"
	case EventTypeDisconnected:
		return "disconnected"
	case EventTypeNodeAdded:
		return "node_added"
	case EventTypeNodeRemoved:
		return "node_removed"
	case EventTypePortAdded:
		return "port_added"
	case EventTypePortRemoved:
		return "port_removed"
	case EventTypeLinkAdded:
		return "link_added"
	case EventTypeLinkRemoved:
		return "link_removed"
	case EventTypeObjectUpdated:
		return "object_updated"
	case EventTypeError:
		return "error"
	case EventTypeInfo:
		return "info"
	default:
		return "unknown"
	}
}
