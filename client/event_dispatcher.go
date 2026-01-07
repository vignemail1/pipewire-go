// Package client - Event dispatcher for application-level events
// client/event_dispatcher.go
// Issue #6 - Event Loop & Automatic Dispatch

package client

import (
	"fmt"
	"sync"

	"github.com/vignemail1/pipewire-go/core"
)

// EventType represents the type of application-level event
type EventType int

const (
	EventTypePort EventType = iota
	EventTypeNode
	EventTypeLink
	EventTypeRegistry
	EventTypeUnknown
)

// String returns the string representation of EventType
func (et EventType) String() string {
	switch et {
	case EventTypePort:
		return "Port"
	case EventTypeNode:
		return "Node"
	case EventTypeLink:
		return "Link"
	case EventTypeRegistry:
		return "Registry"
	default:
		return "Unknown"
	}
}

// ApplicationEvent represents an application-level event from PipeWire
type ApplicationEvent struct {
	Type       EventType
	ObjectID   uint32
	ObjectType string
	Data       map[string]interface{}
}

// EventListener is called when an event is dispatched
type EventListener interface {
	OnEvent(event *ApplicationEvent) error
}

// EventListenerFunc is a function that implements EventListener
type EventListenerFunc func(event *ApplicationEvent) error

// OnEvent calls the function
func (f EventListenerFunc) OnEvent(event *ApplicationEvent) error {
	return f(event)
}

// EventDispatcher manages event listeners and dispatch
type EventDispatcher struct {
	mu              sync.RWMutex
	listeners       map[EventType][]EventListener
	asyncDispatch   bool
	maxQueueSize    int
	eventQueue      chan *ApplicationEvent
	running         bool
	stopChan        chan struct{}
}

// NewEventDispatcher creates a new event dispatcher
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		listeners:    make(map[EventType][]EventListener),
		asyncDispatch: true,
		maxQueueSize: 1000,
		eventQueue:   make(chan *ApplicationEvent, 1000),
		stopChan:     make(chan struct{}),
	}
}

// RegisterListener registers an event listener for a specific event type
func (ed *EventDispatcher) RegisterListener(eventType EventType, listener EventListener) error {
	if listener == nil {
		return fmt.Errorf("listener cannot be nil")
	}

	ed.mu.Lock()
	defer ed.mu.Unlock()

	if ed.listeners[eventType] == nil {
		ed.listeners[eventType] = make([]EventListener, 0)
	}

	ed.listeners[eventType] = append(ed.listeners[eventType], listener)
	return nil
}

// UnregisterListener removes all listeners for an event type
func (ed *EventDispatcher) UnregisterListener(eventType EventType) {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	delete(ed.listeners, eventType)
}

// RemoveListener removes a specific listener
func (ed *EventDispatcher) RemoveListener(eventType EventType, listener EventListener) error {
	if listener == nil {
		return fmt.Errorf("listener cannot be nil")
	}

	ed.mu.Lock()
	defer ed.mu.Unlock()

	listeners, ok := ed.listeners[eventType]
	if !ok {
		return fmt.Errorf("no listeners for event type %s", eventType)
	}

	// Find and remove the listener
	for i, l := range listeners {
		if l == listener {
			// Remove by swapping with last and truncating
			ed.listeners[eventType] = append(
				listeners[:i],
				listeners[i+1:]...,
			)
			return nil
		}
	}

	return fmt.Errorf("listener not found")
}

// Dispatch sends an event to all registered listeners
func (ed *EventDispatcher) Dispatch(event *ApplicationEvent) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	if ed.asyncDispatch {
		// Queue event for async dispatch
		select {
		case ed.eventQueue <- event:
			return nil
		case <-ed.stopChan:
			return fmt.Errorf("dispatcher stopped")
		default:
			return fmt.Errorf("event queue full")
		}
	}

	// Synchronous dispatch
	return ed.dispatchSync(event)
}

// dispatchSync dispatches synchronously to all listeners
func (ed *EventDispatcher) dispatchSync(event *ApplicationEvent) error {
	ed.mu.RLock()
	listeners, ok := ed.listeners[event.Type]
	ed.mu.RUnlock()

	if !ok || len(listeners) == 0 {
		return nil // No listeners for this event type
	}

	var errs []error

	for _, listener := range listeners {
		if err := listener.OnEvent(event); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("dispatch errors: %v", errs)
	}

	return nil
}

// Start begins the async event dispatcher
func (ed *EventDispatcher) Start() error {
	ed.mu.Lock()
	if ed.running {
		ed.mu.Unlock()
		return fmt.Errorf("dispatcher already running")
	}
	ed.running = true
	ed.mu.Unlock()

	// Start event loop goroutine
	go ed.eventLoop()

	return nil
}

// Stop stops the async event dispatcher
func (ed *EventDispatcher) Stop() error {
	ed.mu.Lock()
	if !ed.running {
		ed.mu.Unlock()
		return fmt.Errorf("dispatcher not running")
	}
	ed.running = false
	ed.mu.Unlock()

	// Signal stop and wait for goroutine
	close(ed.stopChan)

	return nil
}

// eventLoop runs the async event dispatching
func (ed *EventDispatcher) eventLoop() {
	for {
		select {
		case event := <-ed.eventQueue:
			_ = ed.dispatchSync(event)

		case <-ed.stopChan:
			return
		}
	}
}

// GetListenersCount returns the number of listeners for an event type
func (ed *EventDispatcher) GetListenersCount(eventType EventType) int {
	ed.mu.RLock()
	defer ed.mu.RUnlock()

	listeners, ok := ed.listeners[eventType]
	if !ok {
		return 0
	}

	return len(listeners)
}

// GetTotalListenersCount returns the total number of registered listeners
func (ed *EventDispatcher) GetTotalListenersCount() int {
	ed.mu.RLock()
	defer ed.mu.RUnlock()

	total := 0
	for _, listeners := range ed.listeners {
		total += len(listeners)
	}

	return total
}

// ClearAllListeners removes all registered listeners
func (ed *EventDispatcher) ClearAllListeners() {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	ed.listeners = make(map[EventType][]EventListener)
}

// ConvertCoreEventToApplicationEvent converts a core event to application event
func ConvertCoreEventToApplicationEvent(coreEvent *core.Event) *ApplicationEvent {
	if coreEvent == nil {
		return nil
	}

	// Map core event types to application event types
	eventType := EventTypeUnknown
	
	// This is simplified - in reality would parse POD data to determine type
	// For now, default to Unknown
	
	return &ApplicationEvent{
		Type:       eventType,
		ObjectID:   0,
		ObjectType: "Unknown",
		Data:       make(map[string]interface{}),
	}
}
