// Package client - client.go
// Main PipeWire Client API

package client

import (
	"context"
	"fmt"
	"sync"

	"github.com/vignemail1/pipewire-go/core"
	"github.com/vignemail1/pipewire-go/verbose"
)

// Client is the main API for interacting with PipeWire
type Client struct {
	conn   *core.Connection
	logger *verbose.Logger
	ctx    context.Context
	cancel context.CancelFunc

	// PipeWire objects
	core     *Core
	registry *Registry

	// Object caches
	nodes map[uint32]*Node
	ports map[uint32]*Port
	links map[uint32]*Link

	// Synchronization
	mu     sync.RWMutex
	done   chan struct{}
	errors chan error

	// Event handling
	eventChan chan Event
	listeners map[EventType][]EventListener
	listMut   sync.RWMutex
}

// Event represents an event from PipeWire
type Event struct {
	Type    EventType
	Object  interface{}
	Changes map[string]string
}

// EventListener is a callback for events
type EventListener func(*Event)

// NewClient creates a new PipeWire client
func NewClient(socketPath string, logger *verbose.Logger) (*Client, error) {
	if logger == nil {
		logger = verbose.NewLogger(verbose.LogLevelInfo, false)
	}

	logger.Infof("Client: Connecting to %s", socketPath)

	// Connect to PipeWire daemon
	conn, err := core.Dial(socketPath, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		conn:      conn,
		logger:    logger,
		ctx:       ctx,
		cancel:    cancel,
		nodes:     make(map[uint32]*Node),
		ports:     make(map[uint32]*Port),
		links:     make(map[uint32]*Link),
		done:      make(chan struct{}),
		errors:    make(chan error, 10),
		eventChan: make(chan Event, 100),
		listeners: make(map[EventType][]EventListener),
	}

	// Create Core proxy (id=0)
	client.core = newCore(0, conn, logger)

	// Create Registry proxy (id=1)
	client.registry = newRegistry(1, conn, logger)

	logger.Infof("Client: Connected to PipeWire daemon")

	// Start event loop
	go client.eventLoop()

	return client, nil
}

// NewClientWithDefaults creates a client with default socket path
func NewClientWithDefaults() (*Client, error) {
	logger := verbose.NewLogger(verbose.LogLevelInfo, false)
	return NewClient("/run/pipewire-0", logger)
}

// eventLoop handles events from the daemon
func (c *Client) eventLoop() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.done:
			return
		case event := <-c.eventChan:
			c.handleEvent(&event)
		}
	}
}

// handleEvent processes an event
func (c *Client) handleEvent(event *Event) {
	c.listMut.RLock()
	listeners := c.listeners[event.Type]
	c.listMut.RUnlock()

	for _, listener := range listeners {
		go listener(event)
	}
}

// On registers an event listener
func (c *Client) On(eventType EventType, listener EventListener) {
	c.listMut.Lock()
	defer c.listMut.Unlock()
	c.listeners[eventType] = append(c.listeners[eventType], listener)
}

// Close closes the client connection
func (c *Client) Close() error {
	c.logger.Infof("Client: Closing")
	c.cancel()
	close(c.done)

	if c.conn != nil {
		c.conn.Close()
	}

	return nil
}

// GetCore returns the Core object
func (c *Client) GetCore() *Core {
	return c.core
}

// GetRegistry returns the Registry object
func (c *Client) GetRegistry() *Registry {
	return c.registry
}

// GetNode retrieves a node by ID
func (c *Client) GetNode(id uint32) *Node {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.nodes[id]
}

// GetNodes returns all nodes
func (c *Client) GetNodes() []*Node {
	c.mu.RLock()
	defer c.mu.RUnlock()

	nodes := make([]*Node, 0, len(c.nodes))
	for _, node := range c.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetNodesByType returns nodes filtered by type/class
func (c *Client) GetNodesByType(mediaClass MediaClass) []*Node {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*Node
	for _, node := range c.nodes {
		if node.info.MediaClass == mediaClass {
			result = append(result, node)
		}
	}
	return result
}

// GetPort retrieves a port by ID
func (c *Client) GetPort(id uint32) *Port {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ports[id]
}

// GetPorts returns all ports
func (c *Client) GetPorts() []*Port {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ports := make([]*Port, 0, len(c.ports))
	for _, port := range c.ports {
		ports = append(ports, port)
	}
	return ports
}

// GetLink retrieves a link by ID
func (c *Client) GetLink(id uint32) *Link {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.links[id]
}

// GetLinks returns all links
func (c *Client) GetLinks() []*Link {
	c.mu.RLock()
	defer c.mu.RUnlock()

	links := make([]*Link, 0, len(c.links))
	for _, link := range c.links {
		links = append(links, link)
	}
	return links
}

// CreateLink creates a new audio link between two ports
// Returns the created Link object or an error
func (c *Client) CreateLink(output, input *Port) (*Link, error) {
	if output == nil || input == nil {
		return nil, fmt.Errorf("output and input ports cannot be nil")
	}

	// Validate ports
	validator := NewLinkValidator(c)
	canCreate, reason := validator.CanCreateLink(output, input)
	if !canCreate {
		return nil, fmt.Errorf("cannot create link: %s", reason)
	}

	c.logger.Infof("Client: Creating link %s → %s", output.Name(), input.Name())

	// Create parameters for the link
	params := &LinkCreateParams{
		OutputNodeID:  output.NodeID(),
		OutputPortID:  output.ID(),
		InputNodeID:   input.NodeID(),
		InputPortID:   input.ID(),
		Properties:    make(map[string]string),
		PassiveLink:   false,
		PhysicalLinks: false,
		MonitorLinks:  false,
	}

	// Validate parameters
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid link parameters: %w", err)
	}

	// Create POD structure for link.create
	podLinkCreate := &PODLinkCreate{
		ObjectType: "Link",
		Version:    3,
		InputPort:  params.InputPortID,
		OutputPort: params.OutputPortID,
		Properties: params.Properties,
	}

	// Create context to track this creation
	creationCtx := &LinkCreationContext{
		Params: params,
		State:  LinkCreationStatePending,
	}

	c.logger.Debugf("Client: Link creation context created: %+v", podLinkCreate)

	// TODO (Issue #5): Protocol Message Handling
	// In a real implementation, we would:
	// 1. Marshal the POD structure (podLinkCreate)
	// 2. Send it to the core object via registry.bind() method
	//    - Method: registry.bind(id=registry_id, method="create_object")
	//    - Params: {"object_type": "Link", "version": 3, "properties": {...}}
	// 3. Wait for registry.global event with new Link object
	// 4. Extract link ID from the event
	// 5. Call creationCtx.SetCreated(linkID)
	// 6. Retrieve the link from cache and return it

	// For now, we simulate with a temporary link (will be removed in issue #5)
	// This allows testing the API structure
	tempLink := NewLink(
		0, // Will be assigned by daemon
		input,
		output,
		c,
	)

	c.logger.Warnf("Client: Link creation stubbed - actual protocol implementation deferred to Issue #5")
	return tempLink, nil
}


// RemoveLink removes an audio link from the graph
// Sends a destroy method to the link's proxy object
func (c *Client) RemoveLink(link *Link) error {
	if link == nil {
		return fmt.Errorf("link cannot be nil")
	}

	linkID := link.ID()
	c.logger.Infof("Client: Removing link id=%d", linkID)

	// Validate link exists in cache
	if _, exists := c.GetLink(linkID); !exists && linkID != 0 {
		c.logger.Warnf("Client: Link %d not in cache, attempting removal anyway", linkID)
	}

	// Create removal context
	removalCtx := &LinkRemovalContext{
		LinkID:     linkID,
		OutputPort: link.OutputPort().ID(),
		InputPort:  link.InputPort().ID(),
		State:      LinkRemovalStatePending,
	}

	c.logger.Debugf("Client: Link removal context created for link %d", linkID)

	// TODO (Issue #5): Protocol Message Handling
	// In a real implementation, we would:
	// 1. Send a destroy method to the link's proxy object
	//    - Object ID: link.ID()
	//    - Method: "destroy"
	// 2. Wait for registry.global_remove event with matching link ID
	// 3. Update removalCtx.State to LinkRemovalStateWaitingForDestroy
	// 4. Verify the link was removed from registry
	// 5. Remove from local cache
	// 6. Update removalCtx.State to LinkRemovalStateDestroyed
	// 7. Return success or error

	// For now, remove from cache immediately (will be coordinated in issue #5)
	c.removeLink(linkID)

	c.logger.Warnf("Client: Link removal stubbed - actual protocol implementation deferred to Issue #5")
	return nil
}

// DisconnectPorts removes a link between two ports by their IDs
// This sends a destroy request to the link object
func (c *Client) DisconnectPorts(outputPortID, inputPortID uint32) error {
	if outputPortID == 0 || inputPortID == 0 {
		return fmt.Errorf("port IDs cannot be 0")
	}

	c.logger.Infof("Client: Disconnecting ports %d -> %d", outputPortID, inputPortID)

	// Find the link in cache
	links := c.GetLinks()
	var linkToRemove *Link
	for _, link := range links {
		if link.OutputPort() != nil && link.InputPort() != nil {
			if link.OutputPort().ID() == outputPortID && link.InputPort().ID() == inputPortID {
				linkToRemove = link
				break
			}
		}
	}

	if linkToRemove == nil {
		return fmt.Errorf("link not found: %d -> %d", outputPortID, inputPortID)
	}

	return c.RemoveLink(linkToRemove)
}

// Sync synchronizes with the daemon
func (c *Client) Sync() error {
	return c.core.Sync()
}

// Ping verifies the connection is alive
func (c *Client) Ping() error {
	return c.core.Ping()
}

// addNode adds a node to the cache
func (c *Client) addNode(node *Node) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nodes[node.ID] = node

	c.logger.Debugf("Client: Node added: id=%d name=%s", node.ID, node.Name())

	// Emit event
	select {
	case c.eventChan <- Event{Type: EventNodeAdded, Object: node}:
	default:
		c.logger.Warnf("Client: Event queue full")
	}
}

// addPort adds a port to the cache
func (c *Client) addPort(port *Port) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ports[port.ID] = port

	// Add to parent node
	if node, ok := c.nodes[port.NodeID]; ok {
		node.AddPort(port)
		port.ParentNode = node
	}

	c.logger.Debugf("Client: Port added: id=%d name=%s node=%d", port.ID, port.Name, port.NodeID)

	// Emit event
	select {
	case c.eventChan <- Event{Type: EventPortAdded, Object: port}:
	default:
		c.logger.Warnf("Client: Event queue full")
	}
}

// addLink adds a link to the cache
func (c *Client) addLink(link *Link) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.links[link.ID] = link

	// Reference ports
	if outPort, ok := c.ports[link.OutputPort]; ok {
		link.Output = outPort
		outPort.AddLink(link)
	}
	if inPort, ok := c.ports[link.InputPort]; ok {
		link.Input = inPort
		inPort.AddLink(link)
	}

	c.logger.Debugf("Client: Link added: id=%d %d→%d", link.ID, link.OutputPort, link.InputPort)

	// Emit event
	select {
	case c.eventChan <- Event{Type: EventLinkAdded, Object: link}:
	default:
		c.logger.Warnf("Client: Event queue full")
	}
}

// removeNode removes a node from cache
func (c *Client) removeNode(nodeID uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.nodes, nodeID)

	c.logger.Debugf("Client: Node removed: id=%d", nodeID)

	select {
	case c.eventChan <- Event{Type: EventNodeRemoved, Object: nodeID}:
	default:
	}
}

// removePort removes a port from cache
func (c *Client) removePort(portID uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.ports, portID)

	c.logger.Debugf("Client: Port removed: id=%d", portID)

	select {
	case c.eventChan <- Event{Type: EventPortRemoved, Object: portID}:
	default:
	}
}

// removeLink removes a link from cache
func (c *Client) removeLink(linkID uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.links, linkID)

	c.logger.Debugf("Client: Link removed: id=%d", linkID)

	select {
	case c.eventChan <- Event{Type: EventLinkRemoved, Object: linkID}:
	default:
	}
}

// GetStatistics returns current graph statistics
func (c *Client) GetStatistics() map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]int{
		"nodes": len(c.nodes),
		"ports": len(c.ports),
		"links": len(c.links),
	}
}

// String returns a human-readable client description
func (c *Client) String() string {
	stats := c.GetStatistics()
	return fmt.Sprintf("Client{Nodes:%d Ports:%d Links:%d}",
		stats["nodes"], stats["ports"], stats["links"])
}
