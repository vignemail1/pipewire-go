// Package graph - TUI Graph Renderer
// cmd/pw-tui/graph.go
// Audio graph visualization and interaction

package main

import (
	"fmt"
	"strings"

	"github.com/yourusername/pipewire-go/client"
)

// GraphRenderer handles rendering of the audio graph
type GraphRenderer struct {
	width   int
	height  int
	padding int
}

// NewGraphRenderer creates a new graph renderer
func NewGraphRenderer(width, height int) *GraphRenderer {
	return &GraphRenderer{
		width:   width,
		height:  height,
		padding: 2,
	}
}

// RenderAscii renders an ASCII representation of the audio graph
func (gr *GraphRenderer) RenderAscii(nodes map[uint32]*client.Node, links map[uint32]*client.Link) string {
	lines := make([]string, 0)

	// Title
	lines = append(lines, "╔"+repeatString("═", gr.width-2)+"╗")
	lines = append(lines, "║ Audio Graph"+repeatString(" ", gr.width-13)+"║")
	lines = append(lines, "╠"+repeatString("═", gr.width-2)+"╣")

	// Group nodes
	inputs := make([]*client.Node, 0)
	outputs := make([]*client.Node, 0)

	for _, node := range nodes {
		if node.GetDirection() == client.NodeDirectionCapture {
			inputs = append(inputs, node)
		} else {
			outputs = append(outputs, node)
		}
	}

	// Render input nodes
	if len(inputs) > 0 {
		lines = append(lines, "║ Inputs:"+repeatString(" ", gr.width-9)+"║")
		for _, node := range inputs {
			line := fmt.Sprintf("║  ├─ [%2d] %-40s", node.ID, node.Name())
			line += repeatString(" ", gr.width-len(line)-1) + "║"
			lines = append(lines, line)

			// Show ports
			for _, port := range node.GetPorts() {
				portType := "↓"
				if port.Direction == client.PortDirectionInput {
					portType = "↑"
				}
				portLine := fmt.Sprintf("║  │   %s %s", portType, port.Name)
				portLine += repeatString(" ", gr.width-len(portLine)-1) + "║"
				lines = append(lines, portLine)
			}
		}
		lines = append(lines, "║"+repeatString(" ", gr.width-2)+"║")
	}

	// Render connections
	if len(links) > 0 {
		lines = append(lines, "║ Connections:"+repeatString(" ", gr.width-14)+"║")
		for _, link := range links {
			status := "●"
			if !link.IsActive() {
				status = "○"
			}

			outputName := "?"
			inputName := "?"

			if link.Output != nil {
				outputName = link.Output.Name
			}
			if link.Input != nil {
				inputName = link.Input.Name
			}

			connLine := fmt.Sprintf("║  %s %s → %s", status, outputName, inputName)
			connLine += repeatString(" ", gr.width-len(connLine)-1) + "║"
			lines = append(lines, connLine)
		}
		lines = append(lines, "║"+repeatString(" ", gr.width-2)+"║")
	}

	// Render output nodes
	if len(outputs) > 0 {
		lines = append(lines, "║ Outputs:"+repeatString(" ", gr.width-10)+"║")
		for _, node := range outputs {
			line := fmt.Sprintf("║  ├─ [%2d] %-40s", node.ID, node.Name())
			line += repeatString(" ", gr.width-len(line)-1) + "║"
			lines = append(lines, line)

			// Show ports
			for _, port := range node.GetPorts() {
				portType := "↓"
				if port.Direction == client.PortDirectionInput {
					portType = "↑"
				}
				portLine := fmt.Sprintf("║  │   %s %s", portType, port.Name)
				portLine += repeatString(" ", gr.width-len(portLine)-1) + "║"
				lines = append(lines, portLine)
			}
		}
	}

	// Footer
	lines = append(lines, "╚"+repeatString("═", gr.width-2)+"╝")

	return strings.Join(lines, "\n")
}

// RenderTree renders the graph as a tree structure
func (gr *GraphRenderer) RenderTree(nodes map[uint32]*client.Node, links map[uint32]*client.Link) string {
	output := ""

	// Build adjacency map
	nodeLinks := make(map[uint32][]*client.Link)
	for _, link := range links {
		if link.Output != nil && link.Output.NodeID > 0 {
			nodeLinks[link.Output.NodeID] = append(nodeLinks[link.Output.NodeID], link)
		}
	}

	// Render each node
	for _, node := range nodes {
		output += fmt.Sprintf("[%d] %s\n", node.ID, node.Name())

		// Show connected ports
		if links, ok := nodeLinks[node.ID]; ok {
			for i, link := range links {
				prefix := "  ├─"
				if i == len(links)-1 {
					prefix = "  └─"
				}

				if link.Output != nil && link.Input != nil {
					output += fmt.Sprintf("%s %s → %s\n", prefix, link.Output.Name, link.Input.Name)
				}
			}
		}
	}

	return output
}

// PortRenderer handles port-specific rendering
type PortRenderer struct {
	width int
}

// NewPortRenderer creates a new port renderer
func NewPortRenderer(width int) *PortRenderer {
	return &PortRenderer{width: width}
}

// RenderPorts renders a list of ports with details
func (pr *PortRenderer) RenderPorts(node *client.Node) string {
	output := fmt.Sprintf("Node [%d]: %s\n", node.ID, node.Name())
	output += repeatString("─", pr.width) + "\n\n"

	ports := node.GetPorts()
	if len(ports) == 0 {
		output += "No ports\n"
		return output
	}

	for _, port := range ports {
		direction := "output"
		if port.Direction == client.PortDirectionInput {
			direction = "input"
		}

		output += fmt.Sprintf("Port [%d]: %s (%s)\n", port.ID, port.Name, direction)

		// Show properties
		output += fmt.Sprintf("  Type: %s\n", port.Type)
		if port.IsConnected() {
			output += "  Status: Connected\n"
		} else {
			output += "  Status: Disconnected\n"
		}

		output += "\n"
	}

	return output
}

// LinkRenderer handles link-specific rendering
type LinkRenderer struct {
	width int
}

// NewLinkRenderer creates a new link renderer
func NewLinkRenderer(width int) *LinkRenderer {
	return &LinkRenderer{width: width}
}

// RenderLinks renders a list of connections/links
func (lr *LinkRenderer) RenderLinks(links map[uint32]*client.Link) string {
	output := "Audio Connections\n"
	output += repeatString("═", lr.width) + "\n\n"

	if len(links) == 0 {
		output += "No connections\n"
		return output
	}

	for _, link := range links {
		status := "ACTIVE"
		if !link.IsActive() {
			status = "INACTIVE"
		}

		outputName := "?"
		inputName := "?"

		if link.Output != nil {
			outputName = link.Output.Name
		}
		if link.Input != nil {
			inputName = link.Input.Name
		}

		output += fmt.Sprintf("[%d] %s: %s → %s\n", link.ID, status, outputName, inputName)

		// Show link properties
		output += fmt.Sprintf("  Output Node: %d, Input Node: %d\n", 
			link.OutputNodeID, link.InputNodeID)
		output += "\n"
	}

	return output
}

// StatisticsRenderer handles statistics visualization
type StatisticsRenderer struct {
	width int
}

// NewStatisticsRenderer creates a new statistics renderer
func NewStatisticsRenderer(width int) *StatisticsRenderer {
	return &StatisticsRenderer{width: width}
}

// RenderStatistics renders graph statistics
func (sr *StatisticsRenderer) RenderStatistics(nodes map[uint32]*client.Node, links map[uint32]*client.Link) string {
	output := "Graph Statistics\n"
	output += repeatString("═", sr.width) + "\n\n"

	// Count nodes by type
	inputs := 0
	outputs := 0
	for _, node := range nodes {
		if node.GetDirection() == client.NodeDirectionCapture {
			inputs++
		} else {
			outputs++
		}
	}

	activeLinks := 0
	inactiveLinks := 0
	for _, link := range links {
		if link.IsActive() {
			activeLinks++
		} else {
			inactiveLinks++
		}
	}

	output += fmt.Sprintf("Total Nodes: %d\n", len(nodes))
	output += fmt.Sprintf("  Input Devices: %d\n", inputs)
	output += fmt.Sprintf("  Output Devices: %d\n", outputs)
	output += fmt.Sprintf("\nTotal Connections: %d\n", len(links))
	output += fmt.Sprintf("  Active: %d\n", activeLinks)
	output += fmt.Sprintf("  Inactive: %d\n", inactiveLinks)

	// Count total ports
	totalPorts := 0
	for _, node := range nodes {
		totalPorts += len(node.GetPorts())
	}
	output += fmt.Sprintf("\nTotal Ports: %d\n", totalPorts)

	return output
}

// StateDisplayFormatter formats state information
type StateDisplayFormatter struct{}

// FormatNodeState formats node state for display
func (sdf *StateDisplayFormatter) FormatNodeState(node *client.Node) string {
	state := node.GetState()
	emoji := "●"

	switch state {
	case "running":
		emoji = "🟢"
	case "idle":
		emoji = "🟡"
	case "error":
		emoji = "🔴"
	}

	return fmt.Sprintf("%s %s (%s)", emoji, node.Name(), state)
}

// FormatPortState formats port state for display
func (sdf *StateDisplayFormatter) FormatPortState(port *client.Port, connected bool) string {
	icon := "○"
	if connected {
		icon = "●"
	}

	direction := "→"
	if port.Direction == client.PortDirectionInput {
		direction = "←"
	}

	return fmt.Sprintf("%s %s %s", icon, direction, port.Name)
}

// FormatLinkState formats link state for display
func (sdf *StateDisplayFormatter) FormatLinkState(link *client.Link) string {
	if link.IsActive() {
		return fmt.Sprintf("● %s → %s", link.Output.Name, link.Input.Name)
	}
	return fmt.Sprintf("○ %s ⇢ %s", link.Output.Name, link.Input.Name)
}

// SearchHelper helps search and filter nodes/ports/links
type SearchHelper struct {
	nodes map[uint32]*client.Node
	ports map[uint32]*client.Port
	links map[uint32]*client.Link
}

// NewSearchHelper creates a new search helper
func NewSearchHelper(nodes map[uint32]*client.Node, ports map[uint32]*client.Port, links map[uint32]*client.Link) *SearchHelper {
	return &SearchHelper{
		nodes: nodes,
		ports: ports,
		links: links,
	}
}

// FindNodesByName searches nodes by name pattern
func (sh *SearchHelper) FindNodesByName(pattern string) []*client.Node {
	results := make([]*client.Node, 0)
	pattern = strings.ToLower(pattern)

	for _, node := range sh.nodes {
		if strings.Contains(strings.ToLower(node.Name()), pattern) {
			results = append(results, node)
		}
	}

	return results
}

// FindPortsByName searches ports by name pattern
func (sh *SearchHelper) FindPortsByName(pattern string) []*client.Port {
	results := make([]*client.Port, 0)
	pattern = strings.ToLower(pattern)

	for _, port := range sh.ports {
		if strings.Contains(strings.ToLower(port.Name), pattern) {
			results = append(results, port)
		}
	}

	return results
}

// FindConnectedPorts finds all ports connected to a given port
func (sh *SearchHelper) FindConnectedPorts(portID uint32) []*client.Port {
	connected := make([]*client.Port, 0)

	for _, link := range sh.links {
		if link.Output != nil && link.Output.ID == portID {
			if input, ok := sh.ports[link.Input.ID]; ok {
				connected = append(connected, input)
			}
		}
		if link.Input != nil && link.Input.ID == portID {
			if output, ok := sh.ports[link.Output.ID]; ok {
				connected = append(connected, output)
			}
		}
	}

	return connected
}
