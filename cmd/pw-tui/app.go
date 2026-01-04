// Package main - PipeWire TUI Application - CORRECTED VERSION
// cmd/pw-tui/app.go
// Main TUI application with terminal event handling - COMPLETE IMPLEMENTATION
// Phase 3 - TUI Application Loop

package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/vignemail1/pipewire-go/client"
	"github.com/vignemail1/pipewire-go/core"
	"github.com/vignemail1/pipewire-go/verbose"
)

// TUIApp represents the terminal UI application
type TUIApp struct {
	mu         sync.RWMutex
	client     *client.Client
	dispatcher *core.EventDispatcher
	logger     *verbose.Logger
	registry   *client.ObjectRegistry
	state      TUIAppState
	running    bool
	done       chan bool
	commands   chan string
}

// TUIAppState represents the current TUI application state
type TUIAppState struct {
	SelectedNodeID    uint32
	SelectedPortID    uint32
	ViewMode          TUIViewMode
	CurrentLine       int
	ScrollOffset      int
	AutoRefresh       bool
	RefreshInterval   time.Duration
	CommandHistory    []string
	HistoryIndex      int
	GraphData         GraphData
	ShowHelp          bool
}

// TUIViewMode represents the current TUI view
type TUIViewMode int

const (
	TUIViewModeGraph TUIViewMode = iota
	TUIViewModeList
	TUIViewModeRouting
	TUIViewModeProperties
)

// GraphData holds current graph state for rendering
type GraphData struct {
	Nodes []*client.Node
	Ports []*client.Port
	Links []*client.Link
}

// NewTUIApp creates a new TUI application
func NewTUIApp(logger *verbose.Logger) *TUIApp {
	if logger == nil {
		logger = verbose.NewLogger(verbose.LogLevelInfo, false)
	}

	return &TUIApp{
		client:     nil,
		dispatcher: core.NewEventDispatcher(4),
		logger:     logger,
		registry:   client.NewObjectRegistry(),
		running:    false,
		done:       make(chan bool),
		commands:   make(chan string, 100),
		state: TUIAppState{
			ViewMode:        TUIViewModeGraph,
			AutoRefresh:     true,
			RefreshInterval: 1 * time.Second,
			CommandHistory:  make([]string, 0),
		},
	}
}

// Connect establishes connection to PipeWire
func (a *TUIApp) Connect(socketPath string) error {
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
func (a *TUIApp) populateRegistry() error {
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

// Start starts the TUI application main loop
func (a *TUIApp) Start() error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return fmt.Errorf("app already running")
	}
	a.running = true
	a.mu.Unlock()

	a.logger.Info("Starting TUI application")

	// Start event dispatcher
	if err := a.dispatcher.Start(); err != nil {
		a.logger.Error("Failed to start dispatcher", "error", err)
		return err
	}

	// Register event handlers
	a.registerEventHandlers()

	// Start main loop
	go a.mainLoop()

	// Start command processor
	go a.commandProcessor()

	return nil
}

// Stop stops the application
func (a *TUIApp) Stop() error {
	a.mu.Lock()
	if !a.running {
		a.mu.Unlock()
		return fmt.Errorf("app not running")
	}
	a.running = false
	a.mu.Unlock()

	a.logger.Info("Stopping TUI application")

	// Stop dispatcher
	if err := a.dispatcher.Stop(); err != nil {
		a.logger.Error("Failed to stop dispatcher", "error", err)
	}

	// Disconnect client
	if a.client != nil {
		a.client.Disconnect()
	}

	close(a.commands)
	return nil
}

// registerEventHandlers registers handlers for PipeWire events
func (a *TUIApp) registerEventHandlers() {
	// Node events
	a.dispatcher.RegisterHandler(core.EventTypeNodeAdded, &TUINodeAddedHandler{app: a})
	a.dispatcher.RegisterHandler(core.EventTypeNodeRemoved, &TUINodeRemovedHandler{app: a})

	// Port events
	a.dispatcher.RegisterHandler(core.EventTypePortAdded, &TUIPortAddedHandler{app: a})
	a.dispatcher.RegisterHandler(core.EventTypePortRemoved, &TUIPortRemovedHandler{app: a})

	// Link events
	a.dispatcher.RegisterHandler(core.EventTypeLinkAdded, &TUILinkAddedHandler{app: a})
	a.dispatcher.RegisterHandler(core.EventTypeLinkRemoved, &TUILinkRemovedHandler{app: a})

	// Error handling
	a.dispatcher.SetErrorHandler(func(err error) {
		a.logger.Error("Event handler error", "error", err)
	})
}

// mainLoop runs the main application loop
func (a *TUIApp) mainLoop() {
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

// commandProcessor processes commands from the command queue
func (a *TUIApp) commandProcessor() {
	for cmd := range a.commands {
		a.processCommand(cmd)
	}
}

// refresh refreshes the UI state
func (a *TUIApp) refresh() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running || a.client == nil {
		return
	}

	// Update state with current registry data
	a.state.GraphData.Nodes = a.registry.GetNodes()
	a.state.GraphData.Ports = a.registry.GetPorts()
	a.state.GraphData.Links = a.registry.GetLinks()

	a.logger.Debug("TUI refresh",
		"nodes", len(a.state.GraphData.Nodes),
		"ports", len(a.state.GraphData.Ports),
		"links", len(a.state.GraphData.Links),
	)
}

// ExecuteCommand queues a command for processing
func (a *TUIApp) ExecuteCommand(cmd string) error {
	select {
	case a.commands <- cmd:
		return nil
	case <-time.After(time.Second):
		return fmt.Errorf("command queue full")
	}
}

// processCommand handles a command - COMPLETE IMPLEMENTATION
func (a *TUIApp) processCommand(cmd string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.logger.Debug("Processing command", "cmd", cmd)

	// Add to command history
	a.state.CommandHistory = append(a.state.CommandHistory, cmd)
	if len(a.state.CommandHistory) > 100 {
		a.state.CommandHistory = a.state.CommandHistory[1:]
	}
	a.state.HistoryIndex = len(a.state.CommandHistory)

	// Parse and execute command
	parts := parseCommand(cmd)
	if len(parts) == 0 {
		return
	}

	command := parts[0]
	args := parts[1:]

	switch command {
	case "connect":
		if len(args) >= 2 {
			sourceID := parseUint32(args[0])
			destID := parseUint32(args[1])
			if sourceID > 0 && destID > 0 {
				if err := a.ConnectPorts(sourceID, destID); err != nil {
					a.logger.Error("Connection failed", "error", err)
				}
			}
		}

	case "disconnect":
		if len(args) >= 2 {
			sourceID := parseUint32(args[0])
			destID := parseUint32(args[1])
			if sourceID > 0 && destID > 0 {
				if err := a.DisconnectPorts(sourceID, destID); err != nil {
					a.logger.Error("Disconnection failed", "error", err)
				}
			}
		}

	case "select-node":
		if len(args) >= 1 {
			nodeID := parseUint32(args[0])
			if nodeID > 0 {
				if err := a.SelectNode(nodeID); err != nil {
					a.logger.Error("Failed to select node", "error", err)
				}
			}
		}

	case "select-port":
		if len(args) >= 1 {
			portID := parseUint32(args[0])
			if portID > 0 {
				if err := a.SelectPort(portID); err != nil {
					a.logger.Error("Failed to select port", "error", err)
				}
			}
		}

	case "view":
		if len(args) >= 1 {
			mode := parseViewMode(args[0])
			a.SetViewMode(mode)
		}

	case "refresh":
		a.refresh()

	case "help":
		a.printHelp()

	case "quit", "exit":
		a.running = false

	default:
		a.logger.Warn("Unknown command", "cmd", command)
	}
}

// SelectNode selects a node
func (a *TUIApp) SelectNode(nodeID uint32) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	node, exists := a.registry.GetNode(nodeID)
	if !exists {
		return fmt.Errorf("node %d not found", nodeID)
	}

	a.state.SelectedNodeID = nodeID
	a.logger.Debug("Selected node", "id", nodeID, "name", node.Name())

	return nil
}

// SelectPort selects a port
func (a *TUIApp) SelectPort(portID uint32) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	port, exists := a.registry.GetPort(portID)
	if !exists {
		return fmt.Errorf("port %d not found", portID)
	}

	a.state.SelectedPortID = portID
	a.logger.Debug("Selected port", "id", portID, "name", port.Name())

	return nil
}

// ConnectPorts creates a connection between ports
func (a *TUIApp) ConnectPorts(sourcePortID, destPortID uint32) error {
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
func (a *TUIApp) DisconnectPorts(sourcePortID, destPortID uint32) error {
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
func (a *TUIApp) GetRegistry() *client.ObjectRegistry {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.registry
}

// GetState returns the current app state
func (a *TUIApp) GetState() TUIAppState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state
}

// SetViewMode sets the current view mode
func (a *TUIApp) SetViewMode(mode TUIViewMode) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state.ViewMode = mode
	a.logger.Debug("View mode changed", "mode", mode)
}

// Helper Functions

// parseCommand splits command string into parts
func parseCommand(cmd string) []string {
	var parts []string
	var current string
	inQuotes := false

	for _, r := range cmd {
		switch r {
		case '"':
			inQuotes = !inQuotes
		case ' ', '\t':
			if !inQuotes && current != "" {
				parts = append(parts, current)
				current = ""
			} else if inQuotes {
				current += string(r)
			}
		default:
			current += string(r)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// parseUint32 converts string to uint32
func parseUint32(s string) uint32 {
	var result uint32
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		result = result*10 + uint32(r-'0')
	}
	return result
}

// parseViewMode converts string to TUIViewMode
func parseViewMode(mode string) TUIViewMode {
	switch mode {
	case "graph":
		return TUIViewModeGraph
	case "list":
		return TUIViewModeList
	case "routing":
		return TUIViewModeRouting
	case "properties":
		return TUIViewModeProperties
	default:
		return TUIViewModeGraph
	}
}

// printHelp displays help information
func (a *TUIApp) printHelp() {
	help := `Available Commands:
  connect <source_id> <dest_id>    - Connect two ports
  disconnect <source_id> <dest_id> - Disconnect ports
  select-node <node_id>            - Select a node
  select-port <port_id>            - Select a port
  view <mode>                      - Change view (graph/list/routing/properties)
  refresh                          - Refresh current view
  help                             - Show this help
  quit, exit                       - Exit application`

	a.logger.Info(help)
}

// TUI Event Handlers

// TUINodeAddedHandler handles node added events
type TUINodeAddedHandler struct {
	app *TUIApp
}

func (h *TUINodeAddedHandler) Handle(event core.Event) error {
	if baseEvent, ok := event.(*core.BaseEvent); ok {
		if node, ok := baseEvent.Data().(*client.Node); ok {
			h.app.registry.RegisterNode(node)
			h.app.logger.Debug("Node added via event", "id", node.ID())
		}
	}
	return nil
}

// TUINodeRemovedHandler handles node removed events
type TUINodeRemovedHandler struct {
	app *TUIApp
}

func (h *TUINodeRemovedHandler) Handle(event core.Event) error {
	if baseEvent, ok := event.(*core.BaseEvent); ok {
		if node, ok := baseEvent.Data().(*client.Node); ok {
			h.app.registry.UnregisterNode(node.ID())
			h.app.logger.Debug("Node removed via event", "id", node.ID())
		}
	}
	return nil
}

// TUIPortAddedHandler handles port added events
type TUIPortAddedHandler struct {
	app *TUIApp
}

func (h *TUIPortAddedHandler) Handle(event core.Event) error {
	if baseEvent, ok := event.(*core.BaseEvent); ok {
		if port, ok := baseEvent.Data().(*client.Port); ok {
			h.app.registry.RegisterPort(port)
			h.app.logger.Debug("Port added via event", "id", port.ID())
		}
	}
	return nil
}

// TUIPortRemovedHandler handles port removed events
type TUIPortRemovedHandler struct {
	app *TUIApp
}

func (h *TUIPortRemovedHandler) Handle(event core.Event) error {
	if baseEvent, ok := event.(*core.BaseEvent); ok {
		if port, ok := baseEvent.Data().(*client.Port); ok {
			h.app.registry.UnregisterPort(port.ID())
			h.app.logger.Debug("Port removed via event", "id", port.ID())
		}
	}
	return nil
}

// TUILinkAddedHandler handles link added events
type TUILinkAddedHandler struct {
	app *TUIApp
}

func (h *TUILinkAddedHandler) Handle(event core.Event) error {
	if baseEvent, ok := event.(*core.BaseEvent); ok {
		if link, ok := baseEvent.Data().(*client.Link); ok {
			h.app.registry.RegisterLink(link)
			h.app.logger.Debug("Link added via event", "id", link.ID())
		}
	}
	return nil
}

// TUILinkRemovedHandler handles link removed events
type TUILinkRemovedHandler struct {
	app *TUIApp
}

func (h *TUILinkRemovedHandler) Handle(event core.Event) error {
	if baseEvent, ok := event.(*core.BaseEvent); ok {
		if link, ok := baseEvent.Data().(*client.Link); ok {
			h.app.registry.UnregisterLink(link.ID())
			h.app.logger.Debug("Link removed via event", "id", link.ID())
		}
	}
	return nil
}
