// Package main - TUI Client
// cmd/pw-tui/main.go
// PipeWire audio graph TUI client using bubbletea

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourusername/pipewire-go/client"
	"github.com/yourusername/pipewire-go/verbose"
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
)

// GraphModel represents the audio graph state
type GraphModel struct {
	nodes      map[uint32]*client.Node
	ports      map[uint32]*client.Port
	links      map[uint32]*client.Link
	lastUpdate int64
	selectedNode  uint32
	selectedPort  uint32
	selectedLink  uint32
}

// NewGraphModel creates a new graph model
func NewGraphModel() *GraphModel {
	return &GraphModel{
		nodes:         make(map[uint32]*client.Node),
		ports:         make(map[uint32]*client.Port),
		links:         make(map[uint32]*client.Link),
		selectedNode:  0,
		selectedPort:  0,
		selectedLink:  0,
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
}

// NewModel initializes the TUI model
func NewModel(socketPath string, logger *verbose.Logger) (*Model, error) {
	// Create client
	c, err := client.NewClient(socketPath, logger)
	if err != nil {
		return nil, err
	}

	model := &Model{
		client:        c,
		logger:        logger,
		activeView:    ViewTypeGraph,
		graphModel:    NewGraphModel(),
		routingMode:   false,
		selectedIndex: 0,
		width:         120,
		height:        40,
	}

	// Update graph
	model.graphModel.Update(c)

	return model, nil
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return nil
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
		m.graphModel.Update(m.client)
		return m, nil

	case ErrorMsg:
		m.error = msg.err
		return m, nil
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quit = true
		return m, tea.Quit

	case "?":
		return m, showHelp()

	case "tab":
		// Switch view
		m.activeView = (m.activeView + 1) % 5
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
		return m, nil

	case "d":
		// Delete selected link
		if m.activeView == ViewTypeConnections && m.selectedLinkID > 0 {
			return m, deleteLink(m.selectedLinkID)
		}

	case "c":
		// Create link (in routing mode)
		if m.routingMode && m.selectedPortID > 0 {
			return m, createLink()
		}

	case "ctrl+r":
		// Refresh graph
		m.graphModel.Update(m.client)
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
		status = "Routing Mode"
	}

	return fmt.Sprintf(
		"PipeWire Audio Graph TUI  [%s]  %s",
		[]string{"Graph", "Devices", "Connections", "Properties", "Unused"}[m.activeView],
		status,
	)
}

// renderGraphView renders the main graph view
func (m *Model) renderGraphView() string {
	output := "AUDIO GRAPH\n"
	output += "═" + repeatString("═", m.width-1) + "\n\n"

	// Group nodes by direction
	var inputs, outputs []*client.Node
	for _, node := range m.graphModel.nodes {
		if node.GetDirection() == client.NodeDirectionCapture {
			inputs = append(inputs, node)
		} else {
			outputs = append(outputs, node)
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

	idx := 0
	for _, node := range m.graphModel.nodes {
		selected := ""
		if idx == m.selectedIndex {
			selected = "► "
		}

		output += fmt.Sprintf("%s[%2d] %s\n", selected, node.ID, node.Name())
		output += fmt.Sprintf("      Class: %s\n", node.info.MediaClass)
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
		}
	}

	return output
}

// renderFooter renders the footer with help text
func (m *Model) renderFooter() string {
	footer := "Keys: q-quit | ↑↓-navigate | Enter-select | Tab-view | Ctrl+R-refresh | ?: help"
	if m.error != nil {
		footer += " | Error: " + m.error.Error()
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

// Commands
func showHelp() tea.Cmd {
	return nil
}

func deleteLink(linkID uint32) tea.Cmd {
	return func() tea.Msg {
		return DeleteLinkMsg{linkID: linkID}
	}
}

func createLink() tea.Cmd {
	return nil
}

// Utility function
func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// Main entry point
func main() {
	var (
		socketPath = flag.String("socket", "/run/pipewire-0", "PipeWire socket path")
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
}
