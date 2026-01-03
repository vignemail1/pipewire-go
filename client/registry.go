// Package client - core.go
// Core proxy implementation

package client

import (
	"fmt"
	"sync"

	"github.com/yourusername/pipewire-go/core"
	"github.com/yourusername/pipewire-go/verbose"
)

// Core represents the PipeWire Core object (id=0)
type Core struct {
	id      uint32
	conn    *core.Connection
	logger  *verbose.Logger
	props   map[string]string
	propMut sync.RWMutex
	
	// Cached info
	info *CoreInfo
}

// newCore creates a new Core proxy
func newCore(id uint32, conn *core.Connection, logger *verbose.Logger) *Core {
	return &Core{
		id:     id,
		conn:   conn,
		logger: logger,
		props:  make(map[string]string),
		info: &CoreInfo{
			ID:      id,
			Version: 0,
			Props:   make(map[string]string),
		},
	}
}

// ID returns the Core object ID (always 0)
func (c *Core) ID() uint32 {
	return c.id
}

// GetInfo returns the Core information
func (c *Core) GetInfo() *CoreInfo {
	c.propMut.RLock()
	defer c.propMut.RUnlock()
	
	info := *c.info
	info.Props = make(map[string]string)
	for k, v := range c.props {
		info.Props[k] = v
	}
	return &info
}

// GetProperty retrieves a Core property
func (c *Core) GetProperty(key string) (string, bool) {
	c.propMut.RLock()
	defer c.propMut.RUnlock()
	val, ok := c.props[key]
	return val, ok
}

// GetProperties returns all Core properties
func (c *Core) GetProperties() map[string]string {
	c.propMut.RLock()
	defer c.propMut.RUnlock()
	
	props := make(map[string]string)
	for k, v := range c.props {
		props[k] = v
	}
	return props
}

// Ping sends a ping request to the server
func (c *Core) Ping() error {
	c.logger.Debugf("Core.Ping()")
	
	// Send ping message to Core (object_id=0, method_id=0)
	return c.conn.SendMessage(c.id, 0, nil)
}

// GetRegistry fetches the registry object ID
// In real implementation, this would bind the registry interface
func (c *Core) GetRegistry() (uint32, error) {
	c.logger.Debugf("Core.GetRegistry()")
	
	// In a real implementation, this would:
	// 1. Send registry_bind method to Core
	// 2. Parse response to get registry ID
	// 3. Return registry ID
	
	// For now, registry is typically id=1
	return 1, nil
}

// Sync sends a sync request and waits for response
func (c *Core) Sync() error {
	c.logger.Debugf("Core.Sync()")
	
	// Send sync message
	return c.conn.SendMessage(c.id, 1, nil)
}

// UpdateProperties updates Core properties
func (c *Core) UpdateProperties(props map[string]string) error {
	c.logger.Debugf("Core.UpdateProperties() with %d properties", len(props))
	
	c.propMut.Lock()
	for k, v := range props {
		c.props[k] = v
	}
	c.propMut.Unlock()
	
	// Send update_properties message
	// Details depend on protocol implementation
	
	return nil
}

// setProperty is internal method to update a property
func (c *Core) setProperty(key string, value string) {
	c.propMut.Lock()
	defer c.propMut.Unlock()
	c.props[key] = value
}

// parseInfo parses Core properties into CoreInfo struct
func (c *Core) parseInfo() {
	c.propMut.Lock()
	defer c.propMut.Unlock()
	
	if ver, ok := c.props["core.version"]; ok {
		c.info.VersionStr = ver
	}
	if name, ok := c.props["core.name"]; ok {
		c.info.Name = name
	}
}

// Registry represents the PipeWire Registry object
type Registry struct {
	id         uint32
	conn       *core.Connection
	logger     *verbose.Logger
	
	// Object cache
	objects map[uint32]*GlobalObject
	objMut  sync.RWMutex
	
	// Event listeners
	listeners []RegistryListener
	listMut   sync.RWMutex
}

// newRegistry creates a new Registry proxy
func newRegistry(id uint32, conn *core.Connection, logger *verbose.Logger) *Registry {
	return &Registry{
		id:        id,
		conn:      conn,
		logger:    logger,
		objects:   make(map[uint32]*GlobalObject),
		listeners: []RegistryListener{},
	}
}

// ID returns the Registry object ID
func (r *Registry) ID() uint32 {
	return r.id
}

// ListAll returns all objects in the registry
func (r *Registry) ListAll() []*GlobalObject {
	r.objMut.RLock()
	defer r.objMut.RUnlock()
	
	objects := make([]*GlobalObject, 0, len(r.objects))
	for _, obj := range r.objects {
		objects = append(objects, obj)
	}
	return objects
}

// GetObject retrieves an object by ID
func (r *Registry) GetObject(id uint32) *GlobalObject {
	r.objMut.RLock()
	defer r.objMut.RUnlock()
	return r.objects[id]
}

// ListByType returns all objects of a specific type
func (r *Registry) ListByType(objType string) []*GlobalObject {
	r.objMut.RLock()
	defer r.objMut.RUnlock()
	
	var result []*GlobalObject
	for _, obj := range r.objects {
		if obj.Type == objType {
			result = append(result, obj)
		}
	}
	return result
}

// ListNodes returns all Node objects
func (r *Registry) ListNodes() []*GlobalObject {
	return r.ListByType("Node")
}

// ListPorts returns all Port objects
func (r *Registry) ListPorts() []*GlobalObject {
	return r.ListByType("Port")
}

// ListLinks returns all Link objects
func (r *Registry) ListLinks() []*GlobalObject {
	return r.ListByType("Link")
}

// OnGlobalAdded registers a listener for object additions
func (r *Registry) OnGlobalAdded(listener RegistryListener) {
	r.listMut.Lock()
	defer r.listMut.Unlock()
	r.listeners = append(r.listeners, listener)
}

// Bind binds to an object in the registry
func (r *Registry) Bind(id uint32, iface string) error {
	r.logger.Debugf("Registry.Bind(id=%d, iface=%s)", id, iface)
	
	// Send bind method to registry
	// Details depend on protocol implementation
	
	return nil
}

// addObject adds an object to the registry
func (r *Registry) addObject(obj *GlobalObject) {
	r.objMut.Lock()
	r.objects[obj.ID] = obj
	r.objMut.Unlock()
	
	// Notify listeners
	r.listMut.RLock()
	listeners := make([]RegistryListener, len(r.listeners))
	copy(listeners, r.listeners)
	r.listMut.RUnlock()
	
	for _, listener := range listeners {
		listener(obj)
	}
	
	r.logger.Infof("Registry: Object added: id=%d type=%s", obj.ID, obj.Type)
}

// removeObject removes an object from the registry
func (r *Registry) removeObject(id uint32) {
	r.objMut.Lock()
	delete(r.objects, id)
	r.objMut.Unlock()
	
	r.logger.Infof("Registry: Object removed: id=%d", id)
}

// CountObjects returns the number of objects in the registry
func (r *Registry) CountObjects() int {
	r.objMut.RLock()
	defer r.objMut.RUnlock()
	return len(r.objects)
}

// Count returns count of objects by type
func (r *Registry) Count(objType string) int {
	r.objMut.RLock()
	defer r.objMut.RUnlock()
	
	count := 0
	for _, obj := range r.objects {
		if obj.Type == objType {
			count++
		}
	}
	return count
}

// String returns a human-readable representation
func (r *Registry) String() string {
	return fmt.Sprintf("Registry(id=%d, objects=%d)", r.id, r.CountObjects())
}
