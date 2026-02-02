package client

import (
	"sync"
)

// Handler is a function called when an event is dispatched
type Handler func(event Event)

// EventDispatcher manages event subscriptions and dispatching
type EventDispatcher struct {
	mu        sync.RWMutex                      // Protects listeners map
	listeners map[string][]Handler
}

// NewEventDispatcher creates a new event dispatcher
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		listeners: make(map[string][]Handler),
	}
}

// Subscribe adds a handler for a specific event type (thread-safe)
func (ed *EventDispatcher) Subscribe(eventType string, handler Handler) {
	ed.mu.Lock()
	defer ed.mu.Unlock()
	ed.listeners[eventType] = append(ed.listeners[eventType], handler)
}

// Unsubscribe removes all handlers for a specific event type (thread-safe)
func (ed *EventDispatcher) Unsubscribe(eventType string) {
	ed.mu.Lock()
	defer ed.mu.Unlock()
	delete(ed.listeners, eventType)
}

// Dispatch calls all registered handlers for an event type (thread-safe)
func (ed *EventDispatcher) Dispatch(event Event) {
	// Get handlers under lock
	ed.mu.RLock()
	handlers := make([]Handler, len(ed.listeners[event.Type]))
	copy(handlers, ed.listeners[event.Type])
	ed.mu.RUnlock()
	
	// Call handlers outside lock to prevent deadlocks
	for _, handler := range handlers {
		handler(event)
	}
}

// ClearAll removes all event subscriptions (thread-safe)
func (ed *EventDispatcher) ClearAll() {
	ed.mu.Lock()
	defer ed.mu.Unlock()
	ed.listeners = make(map[string][]Handler)
}

// GetSubscriberCount returns the number of subscribers for an event type (thread-safe)
func (ed *EventDispatcher) GetSubscriberCount(eventType string) int {
	ed.mu.RLock()
	defer ed.mu.RUnlock()
	return len(ed.listeners[eventType])
}
