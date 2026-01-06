// Package core - PipeWire Protocol Core
// core/proxy.go
// Base proxy object for PipeWire protocol communication

package core

import (
	"fmt"
	"sync"
	"time"
)

// Proxy represents a proxy to a remote PipeWire object
type Proxy struct {
	mu       sync.RWMutex
	id       uint32
	connType string // "core", "registry", "node", etc.
	conn     *Connection

timeout  time.Duration
	listeners map[string][]EventListener
}

// EventListener is a callback for proxy events
type EventListener func(data []byte) error

// NewProxy creates a new proxy
func NewProxy(id uint32, connType string, conn *Connection) *Proxy {
	return &Proxy{
		id:        id,
		connType: connType,
		conn:      conn,
		timeout:   5 * time.Second,
		listeners: make(map[string][]EventListener),
	}
}

// ID returns the proxy object ID
func (p *Proxy) ID() uint32 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.id
}

// Type returns the proxy type
func (p *Proxy) Type() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.connType
}

// SetTimeout sets the operation timeout
func (p *Proxy) SetTimeout(d time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.timeout = d
}

// GetTimeout returns the operation timeout
func (p *Proxy) GetTimeout() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.timeout
}

// AddEventListener registers an event listener
func (p *Proxy) AddEventListener(eventName string, listener EventListener) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.listeners[eventName] = append(p.listeners[eventName], listener)
}

// RemoveEventListener removes an event listener
func (p *Proxy) RemoveEventListener(eventName string, listener EventListener) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if listeners, exists := p.listeners[eventName]; exists {
		for i, l := range listeners {
			// Compare pointers (not ideal but sufficient for function references)
			if fmt.Sprintf("%p", l) == fmt.Sprintf("%p", listener) {
				p.listeners[eventName] = append(listeners[:i], listeners[i+1:]...)
				break
			}
		}
	}
}

// FireEvent triggers all listeners for an event
func (p *Proxy) FireEvent(eventName string, data []byte) error {
	p.mu.RLock()
	listeners := p.listeners[eventName]
	p.mu.RUnlock()

	for _, listener := range listeners {
		if err := listener(data); err != nil {
			return err
		}
	}
	return nil
}

// GetListeners returns count of listeners for an event
func (p *Proxy) GetListeners(eventName string) int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.listeners[eventName])
}

// String returns a string representation
func (p *Proxy) String() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return fmt.Sprintf("Proxy{id=%d, type=%s}", p.id, p.connType)
}
