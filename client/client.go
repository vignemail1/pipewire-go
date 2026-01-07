// Package client - client.go
// Main PipeWire Client API

package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vignemail1/pipewire-go/core"
	"github.com/vignemail1/pipewire-go/verbose"
)

func NewClient(socketPath string, logger *verbose.Logger) (*Client, error) {
	if logger == nil {
		logger = verbose.NewLogger(verbose.LogLevelInfo, false)
	}

	logger.Infof("Client: Connecting to %s", socketPath)

	// Connect to PipeWire daemon
	connection, err := core.Dial(socketPath, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// ========================================================================
	// Initialize protocol components and client structures
	// ========================================================================
	eventHandler := core.NewEventHandler()

	client := &Client{
		logger:       logger,
		ctx:          ctx,
		cancel:       cancel,
		nodes:        make(map[uint32]*Node),
		ports:        make(map[uint32]*Port),
		links:        make(map[uint32]*Link),
		done:         make(chan struct{}),
		errors:       make(chan error, 10),
		eventChan:    make(chan Event, 100),
		listeners:    make(map[EventType][]EventListener),
		mu:           sync.RWMutex{},
		connection:   connection,      // Unix socket connection to daemon
		registryID:   1,               // Registry object ID
		coreID:       0,               // Core object ID
		eventHandler: eventHandler,    // Event handler for protocol events
		lastSequence: 0,               // Request sequence counter
		dispatcher:   NewEventDispatcher(), // Application event dispatcher
	}

	// Create Core proxy (id=0)
	client.core = newCore(0, connection, logger)

	// Create Registry proxy (id=1)
	client.registry = newRegistry(1, connection, logger)

	logger.Infof("Client: Connected to PipeWire daemon")

	// FIXED: Start protocol event loop with correct variable name
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Fixed: Use 'client' instead of undefined 'c'
		if err := client.connection.StartEventLoop(ctx) {
			logger.Errorf("Protocol event loop error: %v", err)
		}
	}()

	// Wait for connection to be ready
	ctxReady, cancelReady := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelReady()

	// Fixed: Use 'client' instead of undefined 'c'
	if err := client.connection.WaitUntilReady(ctxReady); err != nil {
		return nil, fmt.Errorf("connection not ready: %w", err)
	}

	// Start application event loop
	go client.eventLoop()

	return client, nil
}

// ============================================================================
// Client STRUCT - CONSOLIDATED AND COMPLETE
// ============================================================================

type Client struct {
	// Core logging and lifecycle
	logger *verbose.Logger
	ctx    context.Context
	cancel context.CancelFunc

	// Data structures for storing PipeWire objects
	nodes map[uint32]*Node
	ports map[uint32]*Port
	links map[uint32]*Link

	// Event channels
	done      chan struct{}
	errors    chan error
	eventChan chan Event
	listeners map[EventType][]EventListener

	// Proxy objects for Core and Registry
	core     *coreProxy
	registry *registryProxy

	// Protocol communication and synchronization
	mu           sync.RWMutex       // Protects all fields below
	connection   *core.Connection   // Unix socket connection to daemon (consolidated from 'conn')
	registryID   uint32             // Registry object ID (1)
	coreID       uint32             // Core object ID (0)
	eventHandler *core.EventHandler // Protocol-level event handler
	lastSequence uint32             // Sequence counter for protocol requests
	dispatcher   *EventDispatcher   // Application-level event dispatcher
}

// ============================================================================
// FIXED AND COMPLETE: eventLoop() METHOD - WAS MISSING
// ============================================================================

// eventLoop handles incoming events from PipeWire daemon
func (c *Client) eventLoop() {
	defer func() {
		c.done <- struct{}{}
	}()

	for {
		select {
		case <-c.ctx.Done():
			c.logger.Debugf("Event loop shutting down")
			return

		case event, ok := <-c.eventChan:
			if !ok {
				c.logger.Debugf("Event channel closed")
				return
			}

			// Dispatch event to registered listeners
			c.mu.RLock()
			listeners := c.listeners[event.Type]
			c.mu.RUnlock()

			for _, listener := range listeners {
				// Call listeners asynchronously to avoid blocking
				go func(l EventListener, e Event) {
					if err := l(e); err != nil {
						c.logger.Warnf("Event listener error: %v", err)
					}
				}(listener, event)
			}

		case err := <-c.errors:
			c.logger.Errorf("Client error: %v", err)
		}
	}
}

// ============================================================================
// COMPLETE Close() METHOD - Properly cleanup all resources
// ============================================================================

func (c *Client) Close() error {
	if c == nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Stop event dispatcher
	if c.dispatcher != nil {
		if err := c.dispatcher.Stop(); err != nil {
			c.logger.Warnf("Error stopping dispatcher: %v", err)
		}
	}

	// Shutdown protocol connection
	if c.connection != nil {
		if err := c.connection.Shutdown(context.Background()); err != nil {
			c.logger.Warnf("Error shutting down connection: %v", err)
		}
	}

	// Signal event loop to stop
	c.cancel()

	// Wait for event loop to finish
	<-c.done

	// Close the connection
	if c.connection != nil {
		return c.connection.Close()
	}
	return nil
}

// ============================================================================
// CONSOLIDATED HELPER METHODS - Access consolidated 'connection' field
// ============================================================================

// GetConnection returns the underlying connection to PipeWire daemon
func (c *Client) GetConnection() *core.Connection {
	if c != nil {
		c.mu.RLock()
		defer c.mu.RUnlock()
		return c.connection
	}
	return nil
}

// GetRegistryID returns the registry object ID
func (c *Client) GetRegistryID() uint32 {
	if c != nil {
		return c.registryID
	}
	return 0
}

// GetCoreID returns the core object ID
func (c *Client) GetCoreID() uint32 {
	if c != nil {
		return c.coreID
	}
	return 0
}

// GetLogger returns the logger instance
func (c *Client) GetLogger() *verbose.Logger {
	if c != nil {
		return c.logger
	}
	return nil
}

// GetEventHandler returns the event handler
func (c *Client) GetEventHandler() *core.EventHandler {
	if c != nil {
		c.mu.RLock()
		defer c.mu.RUnlock()
		return c.eventHandler
	}
	return nil
}

// GetNextSequence returns the next sequence number for protocol requests
func (c *Client) GetNextSequence() uint32 {
	if c != nil {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.lastSequence++
		return c.lastSequence
	}
	return 0
}

// IsReady returns true if the connection is ready for communication
func (c *Client) IsReady() bool {
	if c == nil || c.connection == nil {
		return false
	}
	return c.connection.GetState() == core.StateReady
}

// WaitUntilReady waits for the connection to be ready
func (c *Client) WaitUntilReady(ctx context.Context) error {
	if c == nil || c.connection == nil {
		return fmt.Errorf("client not initialized")
	}
	return c.connection.WaitUntilReady(ctx)
}

// GetConnectionState returns the current protocol connection state
func (c *Client) GetConnectionState() core.State {
	if c == nil || c.connection == nil {
		return core.StateDisconnected
	}
	return c.connection.GetState()
}

// RegisterEventListener registers an application-level event listener
func (c *Client) RegisterEventListener(eventType EventType, listener EventListener) error {
	if c.dispatcher == nil {
		return fmt.Errorf("event dispatcher not initialized")
	}
	return c.dispatcher.RegisterListener(eventType, listener)
}

// UnregisterEventListener removes event listeners for a type
func (c *Client) UnregisterEventListener(eventType EventType) {
	if c.dispatcher != nil {
		c.dispatcher.UnregisterListener(eventType)
	}
}

// ============================================================================
// COMPLETE GRAPH QUERY METHODS - ISSUE #5: New functionality
// ============================================================================

// GetNodes returns all loaded nodes from the audio graph
func (c *Client) GetNodes() []*Node {
	c.mu.RLock()
	defer c.mu.RUnlock()

	nodes := make([]*Node, 0, len(c.nodes))
	for _, node := range c.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetPorts returns all loaded ports from the audio graph
func (c *Client) GetPorts() []*Port {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ports := make([]*Port, 0, len(c.ports))
	for _, port := range c.ports {
		ports = append(ports, port)
	}
	return ports
}

// GetLinks returns all active audio links in the graph
func (c *Client) GetLinks() []*Link {
	c.mu.RLock()
	defer c.mu.RUnlock()

	links := make([]*Link, 0, len(c.links))
	for _, link := range c.links {
		links = append(links, link)
	}
	return links
}

// GetNodeByID finds a node by its ID
func (c *Client) GetNodeByID(id uint32) *Node {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.nodes[id]
}

// GetPortByID finds a port by its ID
func (c *Client) GetPortByID(id uint32) *Port {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ports[id]
}

// GetLinkByID finds a link by its ID
func (c *Client) GetLinkByID(id uint32) *Link {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.links[id]
}

// ============================================================================
// COMPLETE CreateLink() METHOD
// ============================================================================

func (c *Client) CreateLink(output, input *Port, params *LinkParams) (*Link, error) {
	if output == nil || input == nil {
		return nil, fmt.Errorf("ports cannot be nil")
	}

	// Use ProtocolClient to create link via daemon
	protoClient := NewProtocolClient(c.connection, c.registryID, c.coreID, c.logger)

	// Send CreateLink request and get link ID from daemon
	linkID, err := protoClient.CreateLink(output.ID(), input.ID(), params.Properties)
	if err != nil {
		return nil, fmt.Errorf("protocol error: %w", err)
	}

	// Create Link object with ID returned by daemon
	link := NewLink(linkID, input, output, c)
	c.mu.Lock()
	c.links[linkID] = link
	c.mu.Unlock()

	c.logger.Infof("Link created: %d (output=%d, input=%d)", linkID, output.ID(), input.ID())

	return link, nil
}

// ============================================================================
// COMPLETE RemoveLink() METHOD
// ============================================================================

func (c *Client) RemoveLink(link *Link) error {
	if link == nil || link.ID() == 0 {
		return fmt.Errorf("invalid link")
	}

	// Use ProtocolClient to destroy link via daemon
	protoClient := NewProtocolClient(c.connection, c.registryID, c.coreID, c.logger)

	// Send DestroyLink request to daemon
	if err := protoClient.DestroyLink(link.ID()); err != nil {
		return fmt.Errorf("protocol error: %w", err)
	}

	// Remove link from local registry
	c.mu.Lock()
	delete(c.links, link.ID())
	c.mu.Unlock()

	c.logger.Infof("Link removed: %d", link.ID())

	return nil
}
