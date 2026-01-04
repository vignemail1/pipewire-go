// Package main - PipeWire GUI Application
// cmd/pw-gui/app.go
// Main GUI application with event handling
// Phase 3 - GUI Application Loop

package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/vignemail1/pipewire-go/client"
	"github.com/vignemail1/pipewire-go/core"
	"github.com/vignemail1/pipewire-go/verbose"
)

// App represents the GUI application
type App struct {
	mu         sync.RWMutex
	client     *client.Client
	dispatcher *core.EventDispatcher
	logger     *verbose.Logger
	registry   *client.ObjectRegistry
	graph      *GraphView
	state      AppState
	running    bool
	done       chan bool
}

// AppState represents the current application state
type AppState struct {
	SelectedNodeID  uint32
	SelectedPortID  uint32
	SelectedLinkID  uint32
	SelectedMode    SelectionMode
	ViewMode        ViewMode
	AutoRefresh     bool
	RefreshInterval time.Duration
}

// SelectionMode represents what's currently selected
type SelectionMode int

const (
	SelectionModeNone SelectionMode = iota
	SelectionModeNode
	SelectionModePort
	SelectionModeLink
)

// ViewMode represents the current view
type ViewMode int

const (
	ViewModeGraph ViewMode = iota
	ViewModeList
	ViewModeProperties
)

// NewApp creates a new GUI application
func NewApp(logger *verbose.Logger) *App {
	if logger == nil {
		logger = verbose.NewLogger(verbose.LogLevelInfo, false)
	}

	return &App{
		client:      nil,
		dispatcher:  core.NewEventDispatcher(4),
		logger:      logger,
		registry:    client.NewObjectRegistry(),
		graph:       NewGraphView(),
		done:        make(chan bool),
		running:     false,
		state: AppState{
			ViewMode:        ViewModeGraph,
			AutoRefresh:     true,
			RefreshInterval: 500 * time.Millisecond,
		},
	}
}

// Connect establishes connection to PipeWire
func (a *App) Connect(socketPath string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.logger.Info("Connecting to PipeWire", "socket", socketPath)

	// Create and connect client
	c := client.NewClient(a.logger)
	if err := c.Connect(socketPath); err != nil {
		a.logger.Error("Failed to connect", "error", err)
		return fmt.Errorf("connection failed: %w", err)
	}

	a.client = c

	// Initialize registry with existing objects
	if err := a.populateRegistry(); err != nil {
		a.logger.Error("Failed to populate registry", "error", err)
		c.Disconnect()
		return err
	}

	a.logger.Info("Connected to PipeWire successfully")
	return nil
}

// populateRegistry loads all current objects into registry
func (a *App) populateRegistry() error {
	if a.client == nil {
		return fmt.Errorf("client not connected")
	}

	// Load nodes
	nodes, err := a.client.GetNodes()
	if err != nil {
		return err
	}
	for _, node := range nodes {
		a.registry.RegisterNode(node)
		a.logger.Debug("Registered node", "id", node.ID(), "name", node.Name())
	}

	// Load ports
	ports, err := a.client.GetPorts()
	if err != nil {
		return err
	}
	for _, port := range ports {
		a.registry.RegisterPort(port)
		a.logger.Debug("Registered port", "id", port.ID(), "name", port.Name())
	}

	// Load links
	links, err := a.client.GetLinks()
	if err != nil {
		return err
	}
	for _, link := range links {
		a.registry.RegisterLink(link)
		a.logger.Debug("Registered link", "id", link.ID())
	}

	return nil
}

// Start starts the application main loop
func (a *App) Start() error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return fmt.Errorf("app already running")
	}
	a.running = true
	a.mu.Unlock()

	a.logger.Info("Starting GUI application")

	// Start event dispatcher
	if err := a.dispatcher.Start(); err != nil {
		a.logger.Error("Failed to start dispatcher", "error", err)
		return err
	}

	// Register event handlers
	a.registerEventHandlers()

	// Start main loop
	go a.mainLoop()

	return nil
}

// Stop stops the application
func (a *App) Stop() error {
	a.mu.Lock()
	if !a.running {
		a.mu.Unlock()
		return fmt.Errorf("app not running")
	}
	a.running = false
	a.mu.Unlock()

	a.logger.Info("Stopping GUI application")

	// Stop dispatcher
	if err := a.dispatcher.Stop(); err != nil {
		a.logger.Error("Failed to stop dispatcher", "error", err)
	}

	// Disconnect client
	if a.client != nil {
		a.client.Disconnect()
	}

	return nil
}

// registerEventHandlers registers handlers for PipeWire events
func (a *App) registerEventHandlers() {
	// Node events
	a.dispatcher.RegisterHandler(core.EventTypeNodeAdded, &NodeAddedHandler{app: a})
	a.dispatcher.RegisterHandler(core.EventTypeNodeRemoved, &NodeRemovedHandler{app: a})

	// Port events
	a.dispatcher.RegisterHandler(core.EventTypePortAdded, &PortAddedHandler{app: a})
	a.dispatcher.RegisterHandler(core.EventTypePortRemoved, &PortRemovedHandler{app: a})

	// Link events
	a.dispatcher.RegisterHandler(core.EventTypeLinkAdded, &LinkAddedHandler{app: a})
	a.dispatcher.RegisterHandler(core.EventTypeLinkRemoved, &LinkRemovedHandler{app: a})

	// Error handling
	a.dispatcher.SetErrorHandler(func(err error) {
		a.logger.Error("Event handler error", "error", err)
	})
}

// mainLoop runs the main application loop
func (a *App) mainLoop() {
	defer func() { a.done <- true }()

	ticker := time.NewTicker(a.state.RefreshInterval)
	defer ticker.Stop()

	for a.running {
		select {
		case <-ticker.C:
			if a.state.AutoRefresh {
				a.refresh()
			}
		}
	}
}

// refresh refreshes the UI state
func (a *App) refresh() {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.running || a.client == nil {
		return
	}

	// Update graph with current registry state
	nodes := a.registry.GetNodes()
	ports := a.registry.GetPorts()
	links := a.registry.GetLinks()

	a.graph.Update(nodes, ports, links)

	a.logger.Debug("UI refresh",
		"nodes", len(nodes),
		"ports", len(ports),
		"links", len(links),
	)
}

// SelectNode selects a node
func (a *App) SelectNode(nodeID uint32) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	node, exists := a.registry.GetNode(nodeID)
	if !exists {
		return fmt.Errorf("node %d not found", nodeID)
	}

	a.state.SelectedNodeID = nodeID
	a.state.SelectedMode = SelectionModeNode
	a.logger.Debug("Selected node", "id", nodeID, "name", node.Name())

	return nil
}

// SelectPort selects a port
func (a *App) SelectPort(portID uint32) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	port, exists := a.registry.GetPort(portID)
	if !exists {
		return fmt.Errorf("port %d not found", portID)
	}

	a.state.SelectedPortID = portID
	a.state.SelectedMode = SelectionModePort
	a.logger.Debug("Selected port", "id", portID, "name", port.Name())

	return nil
}

// SelectLink selects a link
func (a *App) SelectLink(linkID uint32) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	link, exists := a.registry.GetLink(linkID)
	if !exists {
		return fmt.Errorf("link %d not found", linkID)
	}

	a.state.SelectedLinkID = linkID
	a.state.SelectedMode = SelectionModeLink
	a.logger.Debug("Selected link", "id", linkID)

	return nil
}

// ConnectPorts creates a connection between ports
func (a *App) ConnectPorts(sourcePortID, destPortID uint32) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.client == nil {
		return fmt.Errorf("not connected")
	}

	a.logger.Info("Connecting ports", "source", sourcePortID, "dest", destPortID)

	link, err := a.client.ConnectPorts(sourcePortID, destPortID)
	if err != nil {
		a.logger.Error("Connection failed", "error", err)
		return err
	}

	a.registry.RegisterLink(link)
	a.logger.Info("Ports connected", "linkID", link.ID())

	return nil
}

// DisconnectPorts removes a connection
func (a *App) DisconnectPorts(sourcePortID, destPortID uint32) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.client == nil {
		return fmt.Errorf("not connected")
	}

	a.logger.Info("Disconnecting ports", "source", sourcePortID, "dest", destPortID)

	err := a.client.DisconnectPorts(sourcePortID, destPortID)
	if err != nil {
		a.logger.Error("Disconnection failed", "error", err)
		return err
	}

	a.logger.Info("Ports disconnected")
	return nil
}

// GetRegistry returns the object registry
func (a *App) GetRegistry() *client.ObjectRegistry {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.registry
}

// GetGraph returns the graph view
func (a *App) GetGraph() *GraphView {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.graph
}

// GetState returns the current app state
func (a *App) GetState() AppState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state
}

// Event Handlers

// NodeAddedHandler handles node added events
type NodeAddedHandler struct {
	app *App
}

func (h *NodeAddedHandler) Handle(event core.Event) error {
	if baseEvent, ok := event.(*core.BaseEvent); ok {
		if node, ok := baseEvent.Data().(*client.Node); ok {
			h.app.registry.RegisterNode(node)
			h.app.logger.Debug("Node added via event", "id", node.ID())
		}
	}
	return nil
}

// NodeRemovedHandler handles node removed events
type NodeRemovedHandler struct {
	app *App
}

func (h *NodeRemovedHandler) Handle(event core.Event) error {
	if baseEvent, ok := event.(*core.BaseEvent); ok {
		if node, ok := baseEvent.Data().(*client.Node); ok {
			h.app.registry.UnregisterNode(node.ID())
			h.app.logger.Debug("Node removed via event", "id", node.ID())
		}
	}
	return nil
}

// PortAddedHandler handles port added events
type PortAddedHandler struct {
	app *App
}

func (h *PortAddedHandler) Handle(event core.Event) error {
	if baseEvent, ok := event.(*core.BaseEvent); ok {
		if port, ok := baseEvent.Data().(*client.Port); ok {
			h.app.registry.RegisterPort(port)
			h.app.logger.Debug("Port added via event", "id", port.ID())
		}
	}
	return nil
}

// PortRemovedHandler handles port removed events
type PortRemovedHandler struct {
	app *App
}

func (h *PortRemovedHandler) Handle(event core.Event) error {
	if baseEvent, ok := event.(*core.BaseEvent); ok {
		if port, ok := baseEvent.Data().(*client.Port); ok {
			h.app.registry.UnregisterPort(port.ID())
			h.app.logger.Debug("Port removed via event", "id", port.ID())
		}
	}
	return nil
}

// LinkAddedHandler handles link added events
type LinkAddedHandler struct {
	app *App
}

func (h *LinkAddedHandler) Handle(event core.Event) error {
	if baseEvent, ok := event.(*core.BaseEvent); ok {
		if link, ok := baseEvent.Data().(*client.Link); ok {
			h.app.registry.RegisterLink(link)
			h.app.logger.Debug("Link added via event", "id", link.ID())
		}
	}
	return nil
}

// LinkRemovedHandler handles link removed events
type LinkRemovedHandler struct {
	app *App
}

func (h *LinkRemovedHandler) Handle(event core.Event) error {
	if baseEvent, ok := event.(*core.BaseEvent); ok {
		if link, ok := baseEvent.Data().(*client.Link); ok {
			h.app.registry.UnregisterLink(link.ID())
			h.app.logger.Debug("Link removed via event", "id", link.ID())
		}
	}
	return nil
}
