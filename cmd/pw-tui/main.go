// Package main - TUI Client
// cmd/pw-tui/main.go
// PipeWire audio graph TUI client using bubbletea

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vignemail1/pipewire-go/client"
	"github.com/vignemail1/pipewire-go/verbose"
)

// Model represents the TUI application state
type Model struct {
	client      *client.Client
	logger      *verbose.Logger
	activeView  ViewType
	graphModel  *GraphModel
	routingMode bool
	error       error
	quit        bool
	status      string
	showHelp    bool
	needsRefresh bool

	// Screen dimensions
	width  int
	height int

	// Input/selection state
	selectedNodeID  uint32
	selectedPortID  uint32
	selectedLinkID  uint32
	selectedIndex   int
}

// ViewType represents the current view
type ViewType int

const (
	ViewTypeGraph ViewType = iota
	ViewTypeRouting
	ViewTypeDevices
	ViewTypeConnections
	ViewTypeProperties
	ViewTypeHelp
)

// GraphModel represents the audio graph state
type GraphModel struct {
	nodes        map[uint32]*client.Node
	ports        map[uint32]*client.Port
	links        map[uint32]*client.Link
	lastUpdate   int64
	selectedNode uint32
	selectedPort uint32
	selectedLink uint32
}

// NewGraphModel creates a new graph model
func NewGraphModel() *GraphModel {
	return &GraphModel{
		nodes:        make(map[uint32]*client.Node),
		ports:        make(map[uint32]*client.Port),
		links:        make(map[uint32]*client.Link),
		selectedNode: 0,
		selectedPort: 0,
		selectedLink: 0,
	}
}

// Update refreshes the graph from client
func (gm *GraphModel) Update(c *client.Client) {
	if c == nil {
		return
	}

	gm.nodes = make(map[uint32]*client.Node)
	gm.ports = make(map[uint32]*client.Port)
	gm.links = make(map[uint32]*client.Link)

	for _, node := range c.GetNodes() {
		gm.nodes[node.ID] = node
		for _, port := range node.GetPorts() {
			gm.ports[port.ID] = port
		}
	}

	for _, link := range c.GetLinks() {
		gm.links[link.ID] = link
	}

	gm.lastUpdate = time.Now().UnixNano()
}

// NewModel initializes the TUI model
func NewModel(socketPath string, logger *verbose.Logger) (*Model, error) {
	// Create client
	c, err := client.NewClient(socketPath, logger)
	if err != nil {
		return nil, err
	}

	// Wait for client to be ready
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.WaitUntilReady(ctx); err != nil {
		c.Close()
		return nil, fmt.Errorf("failed to connect to PipeWire: %w", err)
	}

	model := &Model{
		client:      c,
		logger:      logger,
		activeView:  ViewTypeGraph,
		graphModel:  NewGraphModel(),
		routingMode: false,
		selectedIndex: 0,
		width:       120,
		height:      40,
		status:      "Ready",
		showHelp:    false,
	}

	// Update graph
	model.graphModel.Update(c)
	model.logger.Info("TUI initialized successfully")

	return model, nil
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	// Start periodic refresh ticker
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return GraphUpdateMsg{}
	})
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case GraphUpdateMsg:
		// Periodic refresh
		m.graphModel.Update(m.client)
		m.status = fmt.Sprintf("Updated at %s", time.Now().Format("15:04:05"))
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return GraphUpdateMsg{}
		})

	case ErrorMsg:
		m.error = msg.err
		m.status = fmt.Sprintf("Error: %v", msg.err)
		m.logger.Error(msg.err.Error())
		return m, nil

	case DeleteLinkMsg:
		return m.handleDeleteLink(msg.linkID)

	case CreateLinkMsg:
		return m.handleCreateLink(msg.outputID, msg.inputID)

	case HelpMsg:
		m.showHelp = !m.showHelp
		if m.showHelp {
			m.activeView = ViewTypeHelp
		} else {
			m.activeView = ViewTypeGraph
		}
		return m, nil
	}

	return m, nil
}

// handleDeleteLink processes link deletion
func (m *Model) handleDeleteLink(linkID uint32) (tea.Model, tea.Cmd) {
	if linkID == 0 {
		m.error = fmt.Errorf("invalid link ID")
		return m, nil
	}

	m.status = "Deleting link..."
	err := m.client.DestroyLink(linkID)
	if err != nil {
		m.error = fmt.Errorf("failed to delete link: %w", err)
		m.status = "Delete failed"
		m.logger.Error(m.error.Error())
		return m, nil
	}

	// Refresh graph
	m.graphModel.Update(m.client)
	m.status = fmt.Sprintf("Link %d deleted successfully", linkID)
	m.logger.Info(m.status)
	m.selectedLinkID = 0

	return m, nil
}

// handleCreateLink processes link creation
func (m *Model) handleCreateLink(sourcePortID, sinkPortID uint32) (tea.Model, tea.Cmd) {
	if sourcePortID == 0 || sinkPortID == 0 {
		m.error = fmt.Errorf("invalid port selection")
		m.status = "Invalid port selection"
		return m, nil
	}

	sourcePort := m.graphModel.ports[sourcePortID]
	sinkPort := m.graphModel.ports[sinkPortID]

	if sourcePort == nil || sinkPort == nil {
		m.error = fmt.Errorf("port not found")
		m.status = "Port not found"
		return m, nil
	}

	m.status = "Creating link..."
	link, err := m.client.CreateLink(sourcePort, sinkPort, nil)
	if err != nil {
		m.error = fmt.Errorf("failed to create link: %w", err)
		m.status = "Link creation failed"
		m.logger.Error(m.error.Error())
		return m, nil
	}

	// Refresh graph
	m.graphModel.Update(m.client)
	m.status = fmt.Sprintf("Link %d created successfully", link.ID())
	m.logger.Info(m.status)
	m.selectedPortID = 0
	m.routingMode = false

	return m, nil
}

// handleKeyPress processes keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If showing help, only allow quit or exit help
	if m.showHelp {
		switch msg.String() {
		case "q", "ctrl+c", "?", "esc":
			m.showHelp = false
			m.activeView = ViewTypeGraph
			return m, nil
		}
		return m, nil
	}

	switch msg.String() {
	case "q", "ctrl+c":
		m.quit = true
		m.logger.Info("Exiting TUI")
		return m, tea.Quit

	case "?":
		m.showHelp = true
		m.activeView = ViewTypeHelp
		return m, nil

	case "tab":
		// Switch view
		if m.activeView < ViewTypeHelp {
			m.activeView = (m.activeView + 1) % ViewTypeHelp
		}
		m.selectedIndex = 0
		return m, nil

	case "shift+tab":
		// Switch view backwards
		if m.activeView == 0 {
			m.activeView = ViewTypeHelp - 1
		} else {
			m.activeView--
		}
		m.selectedIndex = 0
		return m, nil

	case "up":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}
		return m, nil

	case "down":
		maxIndex := m.getMaxIndex()
		if m.selectedIndex < maxIndex {
			m.selectedIndex++
		}
		return m, nil

	case "enter":
		return m.handleSelect()

	case "r":
		// Toggle routing mode
		m.routingMode = !m.routingMode
		if m.routingMode {
			m.status = "Routing mode ON - select ports to connect"
			m.logger.Info("Routing mode enabled")
		} else {
			m.status = "Routing mode OFF"
			m.logger.Info("Routing mode disabled")
		}
		return m, nil

	case "d":
		// Delete selected link
		if m.activeView == ViewTypeConnections && m.selectedLinkID > 0 {
			m.status = fmt.Sprintf("Deleting link %d...", m.selectedLinkID)
			return m, deleteLink(m, m.selectedLinkID)
		}

	case "c":
		// Create link (in routing mode)
		if m.routingMode && m.selectedPortID > 0 {
			// This would need a second selection - for now just show status
			m.status = fmt.Sprintf("Source port selected: %d - select sink port", m.selectedPortID)
		}

	case "ctrl+r":
		// Refresh graph
		m.graphModel.Update(m.client)
		m.status = "Graph refreshed"
		return m, nil

	case "1":
		m.activeView = ViewTypeGraph
		m.selectedIndex = 0
		return m, nil

	case "2":
		m.activeView = ViewTypeDevices
		m.selectedIndex = 0
		return m, nil

	case "3":
		m.activeView = ViewTypeConnections
		m.selectedIndex = 0
		return m, nil
	}

	return m, nil
}

// getMaxIndex returns the maximum selectable index for current view
func (m *Model) getMaxIndex() int {
	switch m.activeView {
	case ViewTypeGraph:
		return len(m.graphModel.nodes) - 1
	case ViewTypeDevices:
		return len(m.graphModel.nodes) - 1
	case ViewTypeConnections:
		return len(m.graphModel.links) - 1
	case ViewTypeProperties:
		return 10 // Placeholder
	default:
		return 0
	}
}

// handleSelect processes selection
func (m *Model) handleSelect() (tea.Model, tea.Cmd) {
	switch m.activeView {
	case ViewTypeGraph, ViewTypeDevices:
		// Select node at index
		idx := 0
		for _, node := range m.graphModel.nodes {
			if idx == m.selectedIndex {
				m.selectedNodeID = node.ID
				m.status = fmt.Sprintf("Selected node: %s (ID: %d)", node.Name(), node.ID)
				break
			}
			idx++
		}

	case ViewTypeConnections:
		// Select link at index
		idx := 0
		for _, link := range m.graphModel.links {
			if idx == m.selectedIndex {
				m.selectedLinkID = link.ID
				m.status = fmt.Sprintf("Selected link: %d", link.ID)
				break
			}
			idx++
		}
	}

	return m, nil
}

// View renders the UI
func (m *Model) View() string {
	if m.quit {
		return ""
	}

	// Build header
	header := m.renderHeader()

	// Build content based on view
	var content string
	switch m.activeView {
	case ViewTypeGraph:
		content = m.renderGraphView()
	case ViewTypeDevices:
		content = m.renderDevicesView()
	case ViewTypeConnections:
		content = m.renderConnectionsView()
	case ViewTypeProperties:
		content = m.renderPropertiesView()
	case ViewTypeHelp:
		content = m.renderHelpView()
	default:
		content = "Unknown view"
	}

	// Build footer
	footer := m.renderFooter()

	return header + "\n" + content + "\n" + footer
}

// renderHeader renders the header section
func (m *Model) renderHeader() string {
	status := "Ready"
	if m.routingMode {
		status = "🔄 Routing Mode"
	}

	viewNames := []string{"Graph", "Devices", "Connections", "Properties", "Help"}
	viewName := "Unknown"
	if m.activeView < ViewType(len(viewNames)) {
		viewName = viewNames[m.activeView]
	}

	return fmt.Sprintf(
		"PipeWire Audio Graph TUI  [%s]  %s",
		viewName,
		status,
	)
}

// renderGraphView renders the main graph view
func (m *Model) renderGraphView() string {
	output := "AUDIO GRAPH\n"
	output += "═" + repeatString("═", m.width-1) + "\n\n"

	if len(m.graphModel.nodes) == 0 {
		output += "No audio nodes connected\n"
		return output
	}

	// Group nodes by direction
	var inputs, outputs []*client.Node
	for _, node := range m.graphModel.nodes {
		state := node.GetState()
		if state == "running" {
			// Assume outputs are running nodes unless explicitly input
			outputs = append(outputs, node)
		} else {
			inputs = append(inputs, node)
		}
	}

	// Display input devices
	if len(inputs) > 0 {
		output += "INPUT DEVICES\n"
		output += "─" + repeatString("─", m.width-1) + "\n"
		for i, node := range inputs {
			selected := ""
			if m.selectedIndex == i && m.activeView == ViewTypeGraph {
				selected = "► "
			}
			output += fmt.Sprintf("%s[%2d] %s (%s)\n",
				selected, node.ID, node.Name(), node.GetState())
		}
		output += "\n"
	}

	// Display output devices
	if len(outputs) > 0 {
		output += "OUTPUT DEVICES\n"
		output += "─" + repeatString("─", m.width-1) + "\n"
		for i, node := range outputs {
			selected := ""
			if m.selectedIndex == (len(inputs) + i) && m.activeView == ViewTypeGraph {
				selected = "► "
			}
			output += fmt.Sprintf("%s[%2d] %s (%s)\n",
				selected, node.ID, node.Name(), node.GetState())
		}
	}

	return output
}

// renderDevicesView renders the devices view
func (m *Model) renderDevicesView() string {
	output := "AUDIO DEVICES\n"
	output += "═" + repeatString("═", m.width-1) + "\n\n"

	if len(m.graphModel.nodes) == 0 {
		output += "No audio devices\n"
		return output
	}

	idx := 0
	for _, node := range m.graphModel.nodes {
		selected := ""
		if idx == m.selectedIndex {
			selected = "► "
		}

		output += fmt.Sprintf("%s[%2d] %s\n", selected, node.ID, node.Name())

		// Use GetProperties() for device class instead of unexposed info field
		if props := node.GetProperties(); props != nil {
			if class, ok := props["media.class"]; ok {
				output += fmt.Sprintf("      Class: %s\n", class)
			}
		}

		output += fmt.Sprintf("      State: %s\n", node.GetState())

		if sr := node.GetSampleRate(); sr > 0 {
			output += fmt.Sprintf("      Rate: %d Hz / %d channels\n", sr, node.GetChannels())
		}

		ports := node.GetPorts()
		if len(ports) > 0 {
			output += fmt.Sprintf("      Ports: %d\n", len(ports))
			for _, port := range ports {
				dir := "→"
				if port.Direction == client.PortDirectionInput {
					dir = "←"
				}
				output += fmt.Sprintf("        %s %s\n", dir, port.Name)
			}
		}

		output += "\n"
		idx++
	}

	return output
}

// renderConnectionsView renders the connections view
func (m *Model) renderConnectionsView() string {
	output := "AUDIO CONNECTIONS\n"
	output += "═" + repeatString("═", m.width-1) + "\n\n"

	if len(m.graphModel.links) == 0 {
		output += "No connections\n"
		return output
	}

	idx := 0
	for _, link := range m.graphModel.links {
		selected := ""
		if idx == m.selectedIndex {
			selected = "► "
		}

		outputName := "?"
		inputName := "?"

		if link.Output != nil {
			outputName = link.Output.Name
		}
		if link.Input != nil {
			inputName = link.Input.Name
		}

		status := "●"
		if !link.IsActive() {
			status = "○"
		}

		output += fmt.Sprintf("%s[%2d] %s %s → %s\n",
			selected, link.ID, status, outputName, inputName)

		idx++
	}

	return output
}

// renderPropertiesView renders the properties view
func (m *Model) renderPropertiesView() string {
	output := "PROPERTIES\n"
	output += "═" + repeatString("═", m.width-1) + "\n\n"

	if m.selectedNodeID > 0 {
		node := m.graphModel.nodes[m.selectedNodeID]
		if node != nil {
			output += fmt.Sprintf("Node [%d]: %s\n\n", node.ID, node.Name())
			for k, v := range node.GetProperties() {
				output += fmt.Sprintf("  %s = %s\n", k, v)
			}
		} else {
			output += "Node not found\n"
		}
	} else if m.selectedLinkID > 0 {
		link := m.graphModel.links[m.selectedLinkID]
		if link != nil {
			output += fmt.Sprintf("Link [%d]\n\n", link.ID)
			output += fmt.Sprintf("  Status: %v\n", link.IsActive())
			if link.Output != nil {
				output += fmt.Sprintf("  Source: %s\n", link.Output.Name)
			}
			if link.Input != nil {
				output += fmt.Sprintf("  Sink: %s\n", link.Input.Name)
			}
		} else {
			output += "Link not found\n"
		}
	} else {
		output += "No item selected\n"
	}

	return output
}

// renderHelpView renders the help view
func (m *Model) renderHelpView() string {
	output := "PIPEWIRE AUDIO GRAPH TUI - HELP\n"
	output += "═" + repeatString("═", m.width-1) + "\n\n"

	output += "NAVIGATION:\n"
	output += "  ↑/↓              Navigate up/down\n"
	output += "  Tab/Shift+Tab    Switch views\n"
	output += "  1/2/3            Quick view (Graph/Devices/Connections)\n\n"

	output += "ACTIONS:\n"
	output += "  Enter            Select item\n"
	output += "  r                Toggle routing mode\n"
	output += "  c                Create link (select ports)\n"
	output += "  d                Delete selected link\n"
	output += "  Ctrl+R           Refresh graph\n\n"

	output += "GENERAL:\n"
	output += "  ?                Toggle help\n"
	output += "  q/Ctrl+C         Quit\n\n"

	output += "VIEWS:\n"
	output += "  Graph            Audio graph visualization\n"
	output += "  Devices          Audio devices and ports\n"
	output += "  Connections      Audio links/connections\n"
	output += "  Properties       Details of selected item\n\n"

	output += "ROUTING MODE:\n"
	output += "  Enable with 'r' key to switch between audio ports\n"
	output += "  Select source and sink ports to create connections\n\n"

	output += "Press ? or Esc to close help\n"

	return output
}

// renderFooter renders the footer with help text
func (m *Model) renderFooter() string {
	footer := "Keys: q-quit | ↑↓-navigate | Tab-view | ?: help | "
	if m.routingMode {
		footer += "[ROUTING]"
	} else {
		footer += "d-delete | c-create | Ctrl+R-refresh"
	}
	if m.error != nil {
		footer += " | ✗ " + m.error.Error()
	} else if m.status != "" {
		footer += " | ✓ " + m.status
	}
	return footer
}

// Message types for commands
type GraphUpdateMsg struct{}
type ErrorMsg struct {
	err error
}
type DeleteLinkMsg struct {
	linkID uint32
}
type CreateLinkMsg struct {
	outputID uint32
	inputID  uint32
}
type HelpMsg struct{}

// Commands
func showHelp() tea.Cmd {
	return func() tea.Msg {
		return HelpMsg{}
	}
}

func deleteLink(m *Model, linkID uint32) tea.Cmd {
	return func() tea.Msg {
		return DeleteLinkMsg{linkID: linkID}
	}
}

func createLink(m *Model, sourceID, sinkID uint32) tea.Cmd {
	return func() tea.Msg {
		return CreateLinkMsg{outputID: sourceID, inputID: sinkID}
	}
}

// Utility function
func repeatString(s string, count int) string {
	if count <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// Main entry point
func main() {
	var (
		socketPath = flag.String("socket", "/run/user/1000/pipewire-0", "PipeWire socket path")
		verbose_   = flag.Bool("v", false, "Verbose output")
	)
	flag.Parse()

	// Create logger
	level := verbose.LogLevelInfo
	if *verbose_ {
		level = verbose.LogLevelDebug
	}
	logger := verbose.NewLogger(level, false)

	// Create model
	model, err := NewModel(*socketPath, logger)
	if err != nil {
		log.Fatalf("Failed to create TUI: %v", err)
	}

	// Run TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("TUI error: %v", err)
	}

	// Cleanup
	model.client.Close()
	logger.Info("TUI closed")
}
