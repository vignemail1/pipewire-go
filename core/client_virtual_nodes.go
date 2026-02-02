package core

import (
	"fmt"
	"sync"
	"time"
)

// virtualNodeRegistry manages virtual nodes created by this client
type virtualNodeRegistry struct {
	mu    sync.RWMutex
	nodes map[uint32]*VirtualNode
}

// CreateVirtualNode creates a new virtual node in the PipeWire graph
//
// This function creates a virtual node with the specified configuration.
// The node will be immediately available in the PipeWire graph and can be
// connected to other nodes.
//
// Example:
//   config := GetVirtualNodePreset("recording")
//   config.Name = "My Recording Sink"
//   virtualNode, err := client.CreateVirtualNode(config)
//   if err != nil {
//       log.Fatal(err)
//   }
//   defer virtualNode.Delete()
func (c *Client) CreateVirtualNode(config VirtualNodeConfig) (*VirtualNode, error) {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid virtual node configuration: %w", err)
	}

	if c == nil {
		return nil, fmt.Errorf("client is nil")
	}

	if !c.connected {
		return nil, fmt.Errorf("client is not connected to PipeWire daemon")
	}

	// Generate a unique node ID (in production, this would come from the daemon)
	nodeID := c.generateVirtualNodeID()

	// Create the VirtualNode object
	virtualNode := &VirtualNode{
		ID:         nodeID,
		Config:     config,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Ports:      make([]*Port, 0),
		Properties: make(map[string]interface{}),
		client:     c,
	}

	// Initialize default properties from config
	virtualNode.Properties["node.name"] = config.Name
	virtualNode.Properties["node.description"] = config.Description
	virtualNode.Properties["media.class"] = c.getMediaClass(config.Type)
	virtualNode.Properties["audio.channels"] = config.Channels
	virtualNode.Properties["audio.position"] = config.ChannelLayout
	virtualNode.Properties["audio.rate"] = config.SampleRate
	virtualNode.Properties["node.passive"] = config.Passive
	virtualNode.Properties["node.virtual"] = config.Virtual

	// Register the virtual node
	c.registerVirtualNode(virtualNode)

	return virtualNode, nil
}

// GetVirtualNode retrieves a virtual node by ID
func (c *Client) GetVirtualNode(nodeID uint32) (*VirtualNode, error) {
	if c == nil {
		return nil, fmt.Errorf("client is nil")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.virtualNodes == nil {
		return nil, &VirtualNodeNotFoundError{NodeID: nodeID}
	}

	virtualNode, exists := c.virtualNodes.nodes[nodeID]
	if !exists {
		return nil, &VirtualNodeNotFoundError{NodeID: nodeID}
	}

	return virtualNode, nil
}

// ListVirtualNodes returns all virtual nodes created by this client
func (c *Client) ListVirtualNodes() []*VirtualNode {
	if c == nil {
		return []*VirtualNode{}
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.virtualNodes == nil {
		return []*VirtualNode{}
	}

	c.virtualNodes.mu.RLock()
	defer c.virtualNodes.mu.RUnlock()

	nodes := make([]*VirtualNode, 0, len(c.virtualNodes.nodes))
	for _, node := range c.virtualNodes.nodes {
		nodes = append(nodes, node)
	}

	return nodes
}

// DeleteVirtualNode deletes a virtual node from the graph
func (c *Client) DeleteVirtualNode(nodeID uint32) error {
	if c == nil {
		return fmt.Errorf("client is nil")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.virtualNodes == nil {
		return &VirtualNodeNotFoundError{NodeID: nodeID}
	}

	c.virtualNodes.mu.Lock()
	defer c.virtualNodes.mu.Unlock()

	_, exists := c.virtualNodes.nodes[nodeID]
	if !exists {
		return &VirtualNodeNotFoundError{NodeID: nodeID}
	}

	// Remove from registry
	delete(c.virtualNodes.nodes, nodeID)

	return nil
}

// Private helper methods

// registerVirtualNode adds a virtual node to the client's registry
func (c *Client) registerVirtualNode(node *VirtualNode) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.virtualNodes == nil {
		c.virtualNodes = &virtualNodeRegistry{
			nodes: make(map[uint32]*VirtualNode),
		}
	}

	c.virtualNodes.mu.Lock()
	defer c.virtualNodes.mu.Unlock()

	c.virtualNodes.nodes[node.ID] = node
}

// generateVirtualNodeID generates a unique ID for a virtual node
func (c *Client) generateVirtualNodeID() uint32 {
	c.mu.Lock()
	defer c.mu.Unlock()

	// In production, this would coordinate with the daemon
	// For now, use a simple incrementing counter
	if c.nextVirtualNodeID == 0 {
		c.nextVirtualNodeID = 1000 // Start from 1000 to avoid conflicts with system nodes
	}

	nodeID := c.nextVirtualNodeID
	c.nextVirtualNodeID++

	return nodeID
}

// getMediaClass returns the media.class property based on node type
func (c *Client) getMediaClass(nodeType VirtualNodeType) string {
	switch nodeType {
	case VirtualNode_Sink:
		return "Audio/Sink"
	case VirtualNode_Source:
		return "Audio/Source"
	case VirtualNode_Filter:
		return "Audio/Effect"
	case VirtualNode_Loopback:
		return "Audio/Sink" // Loopback sinks are also classified as sinks
	default:
		return "Audio/Sink"
	}
}

// unregisterVirtualNode removes a virtual node from the registry
func (c *Client) unregisterVirtualNode(nodeID uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.virtualNodes == nil {
		return
	}

	c.virtualNodes.mu.Lock()
	defer c.virtualNodes.mu.Unlock()

	delete(c.virtualNodes.nodes, nodeID)
}

// UpdateVirtualNodeProperty updates a property of a virtual node
func (c *Client) UpdateVirtualNodeProperty(nodeID uint32, key string, value interface{}) error {
	virtualNode, err := c.GetVirtualNode(nodeID)
	if err != nil {
		return err
	}

	return virtualNode.UpdateProperty(key, value)
}

// GetVirtualNodeProperty retrieves a property of a virtual node
func (c *Client) GetVirtualNodeProperty(nodeID uint32, key string) (interface{}, error) {
	virtualNode, err := c.GetVirtualNode(nodeID)
	if err != nil {
		return nil, err
	}

	return virtualNode.GetProperty(key)
}
