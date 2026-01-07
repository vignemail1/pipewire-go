// Package client - node.go
// Node proxy implementation for audio/video nodes

package client

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/vignemail1/pipewire-go/core"
	"github.com/vignemail1/pipewire-go/verbose"
)

// Node represents a PipeWire audio/video node
type Node struct {
	ID       uint32
	Type     string
	Version  uint32
	Props    map[string]string
	propMut  sync.RWMutex

	conn   *core.Connection
	logger *verbose.Logger

	// Cached info
	info *NodeInfo

	// Ports owned by this node
	ports   map[uint32]*Port
	portMut sync.RWMutex
}

// newNode creates a new Node proxy
func newNode(id uint32, objType string, version uint32, props map[string]string, conn *core.Connection, logger *verbose.Logger) *Node {
	node := &Node{
		ID:       id,
		Type:     objType,
		Version:  version,
		Props:    make(map[string]string),
		conn:     conn,
		logger:   logger,
		ports:    make(map[uint32]*Port),
		info: &NodeInfo{
			ID:         id,
			Type:       objType,
			Version:    version,
			Properties: make(map[string]string),
		},
	}

	// Copy properties
	for k, v := range props {
		node.Props[k] = v
		node.info.Properties[k] = v
	}

	node.parseProperties()
	return node
}

// parseProperties parses Node properties into info struct
func (n *Node) parseProperties() {
	n.propMut.RLock()
	defer n.propMut.RUnlock()

	n.info.Name = n.Props["node.name"]
	n.info.Description = n.Props["node.description"]

	if class, ok := n.Props["media.class"]; ok {
		n.info.MediaClass = MediaClass(class)
	}

	if dir, ok := n.Props["node.direction"]; ok {
		n.info.Direction = NodeDirection(dir)
	}

	n.info.State = NodeState(n.Props["node.state"])

	if sr, ok := n.Props["audio.rate"]; ok {
		n.info.SampleRate, _ = strconv.ParseUint(sr, 10, 32)
	}

	if ch, ok := n.Props["audio.channels"]; ok {
		n.info.Channels, _ = strconv.ParseUint(ch, 10, 32)
	}
}

// Name returns the node's display name
func (n *Node) Name() string {
	n.propMut.RLock()
	defer n.propMut.RUnlock()
	if name, ok := n.Props["node.name"]; ok {
		return name
	}
	return fmt.Sprintf("Node[%d]", n.ID)
}

// Description returns the node's description
func (n *Node) Description() string {
	n.propMut.RLock()
	defer n.propMut.RUnlock()
	return n.Props["node.description"]
}

// GetDirection returns whether this is playback, capture, or duplex
func (n *Node) GetDirection() NodeDirection {
	n.propMut.RLock()
	defer n.propMut.RUnlock()
	if dir, ok := n.Props["node.direction"]; ok {
		return NodeDirection(dir)
	}
	return NodeDirectionDuplex
}

// GetState returns the current node state
func (n *Node) GetState() NodeState {
	n.propMut.RLock()
	defer n.propMut.RUnlock()
	state := NodeState(n.Props["node.state"])
	if state == "" {
		return NodeStateIdle
	}
	return state
}

// GetSampleRate returns the audio sample rate
func (n *Node) GetSampleRate() uint32 {
	n.propMut.RLock()
	defer n.propMut.RUnlock()
	if sr, ok := n.Props["audio.rate"]; ok {
		rate, _ := strconv.ParseUint(sr, 10, 32)
		return uint32(rate)
	}
	return 0
}

// GetChannels returns the number of audio channels
func (n *Node) GetChannels() uint32 {
	n.propMut.RLock()
	defer n.propMut.RUnlock()
	if ch, ok := n.Props["audio.channels"]; ok {
		channels, _ := strconv.ParseUint(ch, 10, 32)
		return uint32(channels)
	}
	return 0
}

// GetProperty retrieves a node property
func (n *Node) GetProperty(key string) (string, bool) {
	n.propMut.RLock()
	defer n.propMut.RUnlock()
	val, ok := n.Props[key]
	return val, ok
}

// GetProperties returns all node properties
func (n *Node) GetProperties() map[string]string {
	n.propMut.RLock()
	defer n.propMut.RUnlock()

	props := make(map[string]string)
	for k, v := range n.Props {
		props[k] = v
	}
	return props
}

// GetInfo returns complete node information
func (n *Node) GetInfo() *NodeInfo {
	n.propMut.RLock()
	defer n.propMut.RUnlock()

	info := *n.info
	info.Properties = make(map[string]string)
	for k, v := range n.Props {
		info.Properties[k] = v
	}
	return &info
}

// ============================================================================
// NEW METHODS FOR ISSUE #6: Node Parameter Handling
// ============================================================================

// ParamID represents a PipeWire node parameter identifier
type ParamID uint32

// Node parameter IDs (as defined in PipeWire spec)
const (
	ParamIDEnumFormat ParamID = iota
	ParamIDPropInfo
	ParamIDProps
	ParamIDFormat
	ParamIDBuffers
	ParamIDMeta
	ParamIDIO
	ParamIDEnumProfile
	ParamIDProfile
	ParamIDEnumPortConfig
	ParamIDPortConfig
	ParamIDProcessLatency
)

// GetParams queries node parameters
// paramID specifies which parameter to query:
//   - ParamIDEnumFormat: Available audio formats
//   - ParamIDFormat: Current format
//   - ParamIDProcessLatency: Processing latency
//   - ParamIDProps: Node properties
func (n *Node) GetParams(paramID ParamID) (interface{}, error) {
	if n == nil || n.conn == nil {
		return nil, fmt.Errorf("node or connection not initialized")
	}

	n.logger.Debugf("Node %d: Getting parameter %d", n.ID, paramID)

	// In a full implementation, this would:
	// 1. Send a Get request to the node via the protocol
	// 2. Wait for the response
	// 3. Parse the POD structure response
	// 4. Return typed result

	// For now, return supported parameter info
	switch paramID {
	case ParamIDEnumFormat:
		return map[string]interface{}{
			"description": "Available audio formats",
			"values": []string{"s16", "s32", "f32", "f64"},
		}, nil

	case ParamIDFormat:
		return map[string]interface{}{
			"format": n.Props["audio.format"],
			"rate":   n.GetSampleRate(),
			"channels": n.GetChannels(),
		}, nil

	case ParamIDProcessLatency:
		return map[string]interface{}{
			"min_quantum": 32,
			"max_quantum": 8192,
			"latency":     64,
		}, nil

	case ParamIDProps:
		return n.GetProperties(), nil

	default:
		return nil, fmt.Errorf("unsupported parameter ID: %d", paramID)
	}
}

// SetParam modifies a node parameter
// paramID: which parameter to modify
// flags: modification flags
// value: new parameter value (type-specific)
func (n *Node) SetParam(paramID ParamID, flags uint32, value interface{}) error {
	if n == nil || n.conn == nil {
		return fmt.Errorf("node or connection not initialized")
	}

	n.logger.Debugf("Node %d: Setting parameter %d with flags %d", n.ID, paramID, flags)

	// In a full implementation, this would:
	// 1. Validate the parameter ID and value type
	// 2. Encode the value as a POD structure
	// 3. Send a Set request to the node via the protocol
	// 4. Wait for confirmation
	// 5. Update local cached properties if successful

	switch paramID {
	case ParamIDFormat:
		// Setting format - verify it's a proper format structure
		if formatMap, ok := value.(map[string]interface{}); ok {
			if format, hasFormat := formatMap["format"].(string); hasFormat {
				n.propMut.Lock()
				n.Props["audio.format"] = format
				n.propMut.Unlock()
				n.logger.Infof("Node %d: Format set to %s", n.ID, format)
				return nil
			}
		}
		return fmt.Errorf("invalid format structure")

	case ParamIDProps:
		// Setting properties
		if propsMap, ok := value.(map[string]string); ok {
			n.propMut.Lock()
			for k, v := range propsMap {
				n.Props[k] = v
			}
			n.propMut.Unlock()
			n.logger.Infof("Node %d: Properties updated", n.ID)
			return nil
		}
		return fmt.Errorf("invalid properties structure")

	default:
		return fmt.Errorf("parameter %d is read-only or unsupported", paramID)
	}
}

// ============================================================================
// Port Management Methods
// ============================================================================

// AddPort adds a port to this node
func (n *Node) AddPort(port *Port) {
	n.portMut.Lock()
	defer n.portMut.Unlock()
	n.ports[port.ID] = port
	n.logger.Debugf("Node %d: Port added: %s", n.ID, port.Name)
}

// GetPort retrieves a port by name
func (n *Node) GetPort(name string) *Port {
	n.portMut.RLock()
	defer n.portMut.RUnlock()

	for _, port := range n.ports {
		if port.Name == name {
			return port
		}
	}
	return nil
}

// GetPorts returns all ports of this node
func (n *Node) GetPorts() []*Port {
	n.portMut.RLock()
	defer n.portMut.RUnlock()

	ports := make([]*Port, 0, len(n.ports))
	for _, port := range n.ports {
		ports = append(ports, port)
	}
	return ports
}

// GetPortsByDirection returns ports filtered by direction
func (n *Node) GetPortsByDirection(dir PortDirection) []*Port {
	n.portMut.RLock()
	defer n.portMut.RUnlock()

	var result []*Port
	for _, port := range n.ports {
		if port.Direction == dir {
			result = append(result, port)
		}
	}
	return result
}

// GetPortsByType returns ports filtered by type
func (n *Node) GetPortsByType(portType PortType) []*Port {
	n.portMut.RLock()
	defer n.portMut.RUnlock()

	var result []*Port
	for _, port := range n.ports {
		if port.Type == portType {
			result = append(result, port)
		}
	}
	return result
}

// RemovePort removes a port from this node
func (n *Node) RemovePort(portID uint32) {
	n.portMut.Lock()
	defer n.portMut.Unlock()
	delete(n.ports, portID)
}

// String returns a human-readable node description
func (n *Node) String() string {
	return fmt.Sprintf("Node{ID:%d Name:%q Dir:%s State:%s Rate:%dHz}",
		n.ID, n.Name(), n.GetDirection(), n.GetState(), n.GetSampleRate())
}
