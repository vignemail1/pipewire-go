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
	
	conn     *core.Connection
	logger   *verbose.Logger
	
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
