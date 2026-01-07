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

	// ========================================================================
	// NEW FOR ISSUE #5: Initialize protocol components
	// ========================================================================
	eventHandler := core.NewEventHandler()
	if err != nil {
		conn.Close()
		cancel()
		return nil, fmt.Errorf("failed to create event handler: %w", err)
	}

	client := &Client{
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

		mu:              sync.RWMutex,
		connection:      conn,           // Unix socket connection
		registryID:      1,              // Registry object ID
		coreID:          0,              // Core object ID
		eventHandler:    eventHandler,   // Event dispatcher
		lastSequence:    0,              // Request sequence counter
	}

	// Create Core proxy (id=0)
	client.core = newCore(0, conn, logger)

	// Create Registry proxy (id=1)
	client.registry = newRegistry(1, conn, logger)

	logger.Infof("Client: Connected to PipeWire daemon")

	// ========================================================================
	// NEW FOR ISSUE #5: Register event handlers with dispatcher
	// ========================================================================
	// These handlers will be called by EventHandler when events arrive
	eventHandler.RegisterEventHandler(client.registryID, func(event core.Event) error {
		logger.Debugf("Registry event received: %v", event)
		// Event will be dispatched to appropriate handlers
		return nil
	})

	eventHandler.RegisterEventHandler(client.coreID, func(event core.Event) error {
		logger.Debugf("Core event received: %v", event)
		// Event will be dispatched to appropriate handlers
		return nil
	})

	// Start event loop
	go client.eventLoop()

	return client, nil
}

// ============================================================================
// Client STRUCT - COMPLETE WITH ISSUE #5 FIELDS
// ============================================================================

type Client struct {
	// EXISTING FIELDS
	conn      *core.Connection
	logger    *verbose.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	nodes     map[uint32]*Node
	ports     map[uint32]*Port
	links     map[uint32]*Link
	done      chan struct{}
	errors    chan error
	eventChan chan Event
	listeners map[EventType][]EventListener

	core     *coreProxy
	registry *registryProxy

	// ========================================================================
	// NEW FIELDS FOR ISSUE #5
	// ========================================================================
	mu              sync.RWMutex                    // Synchronization mutex
	connection      *core.Connection                // Duplicate reference for protocol
	registryID      uint32                          // Registry object ID (1)
	coreID          uint32                          // Core object ID (0)
	eventHandler    *core.EventHandler              // Event dispatcher
	lastSequence    uint32                          // Last used sequence number
}

// ============================================================================
// COMPLETE Close() METHOD - ISSUE #5 INTEGRATED
// ============================================================================

func (c *Client) Close() error {
	if c == nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Signal event loop to stop
	c.cancel()

	// Wait for event loop to finish
	<-c.done

	// Close connection to daemon
	if c.connection != nil {
		return c.connection.Close()
	}

	return nil
}

// ============================================================================
// NEW HELPER METHODS FOR ISSUE #5
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

// ============================================================================
// COMPLETE CreateLink() METHOD - ISSUE #5
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
	c.links[linkID] = link

	c.logger.Infof("Link created: %d (output=%d, input=%d)", linkID, output.ID(), input.ID())

	return link, nil
}

// ============================================================================
// COMPLETE RemoveLink() METHOD - ISSUE #5
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
	delete(c.links, link.ID())

	c.logger.Infof("Link removed: %d", link.ID())

	return nil
}
