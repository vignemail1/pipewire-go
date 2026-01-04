// Package client - High-Level PipeWire Client
// client/properties.go
// Object property management and manipulation
// Phase 2 - Property handling utilities

package client

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// Properties represents a collection of object properties
type Properties struct {
	mu    sync.RWMutex
	items map[string]string
}

// NewProperties creates a new empty properties collection
func NewProperties() *Properties {
	return &Properties{
		items: make(map[string]string),
	}
}

// NewPropertiesFromMap creates properties from a string map
func NewPropertiesFromMap(m map[string]string) *Properties {
	p := NewProperties()
	for k, v := range m {
		p.items[k] = v
	}
	return p
}

// Set sets a property value
func (p *Properties) Set(key, value string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.items[key] = value
}

// Get retrieves a property value
func (p *Properties) Get(key string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	val, ok := p.items[key]
	return val, ok
}

// GetOr retrieves a property value with default
func (p *Properties) GetOr(key, defaultValue string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if val, ok := p.items[key]; ok {
		return val
	}
	return defaultValue
}

// GetInt retrieves a property as integer
func (p *Properties) GetInt(key string) (int, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	val, ok := p.items[key]
	if !ok {
		return 0, fmt.Errorf("property %q not found", key)
	}
	return strconv.Atoi(val)
}

// GetIntOr retrieves a property as integer with default
func (p *Properties) GetIntOr(key string, defaultValue int) int {
	if val, err := p.GetInt(key); err == nil {
		return val
	}
	return defaultValue
}

// GetBool retrieves a property as boolean
func (p *Properties) GetBool(key string) (bool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	val, ok := p.items[key]
	if !ok {
		return false, fmt.Errorf("property %q not found", key)
	}
	return val == "true" || val == "1" || val == "yes", nil
}

// GetBoolOr retrieves a property as boolean with default
func (p *Properties) GetBoolOr(key string, defaultValue bool) bool {
	if val, err := p.GetBool(key); err == nil {
		return val
	}
	return defaultValue
}

// Delete removes a property
func (p *Properties) Delete(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.items, key)
}

// Has checks if a property exists
func (p *Properties) Has(key string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, ok := p.items[key]
	return ok
}

// Keys returns all property keys
func (p *Properties) Keys() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	keys := make([]string, 0, len(p.items))
	for k := range p.items {
		keys = append(keys, k)
	}
	return keys
}

// Values returns all property values
func (p *Properties) Values() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	values := make([]string, 0, len(p.items))
	for _, v := range p.items {
		values = append(values, v)
	}
	return values
}

// Items returns all property items as a map copy
func (p *Properties) Items() map[string]string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make(map[string]string)
	for k, v := range p.items {
		result[k] = v
	}
	return result
}

// Len returns the number of properties
func (p *Properties) Len() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.items)
}

// Clear removes all properties
func (p *Properties) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.items = make(map[string]string)
}

// Copy creates a deep copy of properties
func (p *Properties) Copy() *Properties {
	p.mu.RLock()
	defer p.mu.RUnlock()
	copy := NewProperties()
	for k, v := range p.items {
		copy.items[k] = v
	}
	return copy
}

// String returns string representation
func (p *Properties) String() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var parts []string
	for k, v := range p.items {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, " ")
}

// ToStringMap converts to string map
func (p *Properties) ToStringMap() map[string]string {
	return p.Items()
}

// FromStringMap sets properties from string map
func (p *Properties) FromStringMap(m map[string]string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for k, v := range m {
		p.items[k] = v
	}
}

// Filter returns a new Properties with filtered items
func (p *Properties) Filter(predicate func(key, value string) bool) *Properties {
	p.mu.RLock()
	defer p.mu.RUnlock()
	filtered := NewProperties()
	for k, v := range p.items {
		if predicate(k, v) {
			filtered.items[k] = v
		}
	}
	return filtered
}

// Map applies a function to all properties
func (p *Properties) Map(fn func(key, value string) (string, bool)) *Properties {
	p.mu.RLock()
	defer p.mu.RUnlock()
	mapped := NewProperties()
	for k, v := range p.items {
		if newV, ok := fn(k, v); ok {
			mapped.items[k] = newV
		}
	}
	return mapped
}

// Merge merges other properties into this one
func (p *Properties) Merge(other *Properties) {
	if other == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	other.mu.RLock()
	defer other.mu.RUnlock()
	for k, v := range other.items {
		p.items[k] = v
	}
}

// MergeMap merges a string map into properties
func (p *Properties) MergeMap(m map[string]string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for k, v := range m {
		p.items[k] = v
	}
}

// Diff returns the differences between this and another Properties
func (p *Properties) Diff(other *Properties) (added, removed, changed map[string]string) {
	added = make(map[string]string)
	removed = make(map[string]string)
	changed = make(map[string]string)

	p.mu.RLock()
	defer p.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	// Find added and changed
	for k, v := range p.items {
		if ov, ok := other.items[k]; !ok {
			added[k] = v
		} else if ov != v {
			changed[k] = v
		}
	}

	// Find removed
	for k, v := range other.items {
		if _, ok := p.items[k]; !ok {
			removed[k] = v
		}
	}

	return
}

// Equal checks if properties are equal
func (p *Properties) Equal(other *Properties) bool {
	if other == nil {
		return false
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	if len(p.items) != len(other.items) {
		return false
	}

	for k, v := range p.items {
		if ov, ok := other.items[k]; !ok || ov != v {
			return false
		}
	}

	return true
}

// CommonProperties returns properties that exist in both
func (p *Properties) CommonProperties(other *Properties) *Properties {
	if other == nil {
		return NewProperties()
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	common := NewProperties()
	for k, v := range p.items {
		if ov, ok := other.items[k]; ok && ov == v {
			common.items[k] = v
		}
	}
	return common
}
