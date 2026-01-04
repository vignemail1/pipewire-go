// Package main - Graph Visualization
// cmd/pw-gui/graph.go
// GTK-based graph rendering and visualization

package main

import (
	"fmt"
	"math"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/vignemail1/pipewire-go/client"
)

// GraphVisualizer handles audio graph visualization
type GraphVisualizer struct {
	client       *client.Client
	drawingArea  *gtk.DrawingArea
	zoomLevel    float64
	panX         float64
	panY         float64
	nodeRadius   float64
	selectedNode uint32
}

// NewGraphVisualizer creates a new graph visualizer
func NewGraphVisualizer(c *client.Client) *GraphVisualizer {
	return &GraphVisualizer{
		client:      c,
		zoomLevel:   1.0,
		panX:        0,
		panY:        0,
		nodeRadius:  30,
	}
}

// SetDrawingArea sets the drawing area
func (gv *GraphVisualizer) SetDrawingArea(da *gtk.DrawingArea) {
	gv.drawingArea = da
}

// Draw renders the audio graph
func (gv *GraphVisualizer) Draw(cairo interface{}, width, height int, nodes map[uint32]*client.Node, links map[uint32]*client.Link) {
	// Layout nodes
	nodePositions := gv.layoutNodes(nodes, width, height)

	// Draw connections first
	gv.drawConnections(cairo, nodePositions, links, width, height)

	// Draw nodes
	gv.drawNodes(cairo, nodePositions, nodes, width, height)
}

// layoutNodes calculates positions for nodes
func (gv *GraphVisualizer) layoutNodes(nodes map[uint32]*client.Node, width, height int) map[uint32][2]float64 {
	positions := make(map[uint32][2]float64)

	// Separate inputs and outputs
	var inputs, outputs []*client.Node
	for _, node := range nodes {
		if node.GetDirection() == client.NodeDirectionCapture {
			inputs = append(inputs, node)
		} else {
			outputs = append(outputs, node)
		}
	}

	// Layout inputs on left
	inputY := float64(height) / float64(len(inputs)+1)
	for i, node := range inputs {
		x := 50.0
		y := inputY * float64(i+1)
		positions[node.ID] = [2]float64{x, y}
	}

	// Layout outputs on right
	outputY := float64(height) / float64(len(outputs)+1)
	for i, node := range outputs {
		x := float64(width) - 50.0
		y := outputY * float64(i+1)
		positions[node.ID] = [2]float64{x, y}
	}

	return positions
}

// drawNodes draws the node circles
func (gv *GraphVisualizer) drawNodes(cairo interface{}, positions map[uint32][2]float64, nodes map[uint32]*client.Node, width, height int) {
	// Cast cairo context
	cr := cairo.(interface{ /* cairo methods */ })

	for nodeID, node := range nodes {
		if pos, exists := positions[nodeID]; exists {
			x := pos[0]
			y := pos[1]

			// Draw circle
			cr.(*interface{})./* SetSourceRGB */(0.3, 0.6, 0.9)
			cr.(*interface{})./* Arc */(x, y, gv.nodeRadius, 0, 2*math.Pi)
			cr.(*interface{})./* Fill */()

			// Draw border
			cr.(*interface{})./* SetSourceRGB */(1.0, 1.0, 1.0)
			cr.(*interface{})./* SetLineWidth */(2)
			cr.(*interface{})./* Arc */(x, y, gv.nodeRadius, 0, 2*math.Pi)
			cr.(*interface{})./* Stroke */()

			// Draw text
			cr.(*interface{})./* SetSourceRGB */(1.0, 1.0, 1.0)
			cr.(*interface{})./* MoveTo */(x-20, y+5)
			cr.(*interface{})./* ShowText */(node.Name())
		}
	}
}

// drawConnections draws the connection lines
func (gv *GraphVisualizer) drawConnections(cairo interface{}, positions map[uint32][2]float64, links map[uint32]*client.Link, width, height int) {
	for _, link := range links {
		if link.Output != nil && link.Input != nil {
			outputPos, outputExists := positions[link.Output.NodeID]
			inputPos, inputExists := positions[link.Input.NodeID]

			if outputExists && inputExists {
				// Draw connection line
				cr := cairo.(interface{})
				if link.IsActive() {
					cr.(*interface{})./* SetSourceRGB */(0.0, 1.0, 0.0)
				} else {
					cr.(*interface{})./* SetSourceRGB */(0.5, 0.5, 0.5)
				}

				cr.(*interface{})./* SetLineWidth */(2)
				cr.(*interface{})./* MoveTo */(outputPos[0]+gv.nodeRadius, outputPos[1])
				cr.(*interface{})./* LineTo */(inputPos[0]-gv.nodeRadius, inputPos[1])
				cr.(*interface{})./* Stroke */()
			}
		}
	}
}

// ZoomIn increases zoom level
func (gv *GraphVisualizer) ZoomIn() {
	gv.zoomLevel *= 1.2
	gv.redraw()
}

// ZoomOut decreases zoom level
func (gv *GraphVisualizer) ZoomOut() {
	gv.zoomLevel /= 1.2
	if gv.zoomLevel < 0.1 {
		gv.zoomLevel = 0.1
	}
	gv.redraw()
}

// FitToWindow adjusts zoom to fit all nodes
func (gv *GraphVisualizer) FitToWindow() {
	gv.zoomLevel = 1.0
	gv.panX = 0
	gv.panY = 0
	gv.redraw()
}

// Pan moves the view by offset
func (gv *GraphVisualizer) Pan(dx, dy float64) {
	gv.panX += dx
	gv.panY += dy
	gv.redraw()
}

// SelectNode selects a node
func (gv *GraphVisualizer) SelectNode(nodeID uint32) {
	gv.selectedNode = nodeID
	gv.redraw()
}

// redraw triggers a redraw
func (gv *GraphVisualizer) redraw() {
	if gv.drawingArea != nil {
		gv.drawingArea.QueueDraw()
	}
}

// RoutingEngine handles audio routing operations
type RoutingEngine struct {
	client           *client.Client
	selectedOutput   uint32
	selectedInput    uint32
	previewLink      *PreviewLink
}

// PreviewLink represents a preview of a link before creation
type PreviewLink struct {
	OutputID uint32
	InputID  uint32
	Valid    bool
}

// NewRoutingEngine creates a new routing engine
func NewRoutingEngine(c *client.Client) *RoutingEngine {
	return &RoutingEngine{
		client: c,
	}
}

// SelectOutput selects an output port
func (re *RoutingEngine) SelectOutput(portID uint32) error {
	port := re.client.GetPort(portID)
	if port == nil {
		return fmt.Errorf("port not found")
	}

	if port.Direction != client.PortDirectionOutput {
		return fmt.Errorf("port is not an output")
	}

	re.selectedOutput = portID
	re.updatePreview()
	return nil
}

// SelectInput selects an input port
func (re *RoutingEngine) SelectInput(portID uint32) error {
	port := re.client.GetPort(portID)
	if port == nil {
		return fmt.Errorf("port not found")
	}

	if port.Direction != client.PortDirectionInput {
		return fmt.Errorf("port is not an input")
	}

	re.selectedInput = portID
	re.updatePreview()
	return nil
}

// updatePreview updates the link preview
func (re *RoutingEngine) updatePreview() {
	if re.selectedOutput > 0 && re.selectedInput > 0 {
		re.previewLink = &PreviewLink{
			OutputID: re.selectedOutput,
			InputID:  re.selectedInput,
			Valid:    true,
		}
	}
}

// CreateLink creates the previewed link
func (re *RoutingEngine) CreateLink() (uint32, error) {
	if re.previewLink == nil || !re.previewLink.Valid {
		return 0, fmt.Errorf("no valid preview link")
	}

	linkID, err := re.client.CreateLink(re.previewLink.OutputID, re.previewLink.InputID)
	if err != nil {
		return 0, err
	}

	// Reset selection
	re.selectedOutput = 0
	re.selectedInput = 0
	re.previewLink = nil

	return linkID, nil
}

// CancelRouting cancels routing mode
func (re *RoutingEngine) CancelRouting() {
	re.selectedOutput = 0
	re.selectedInput = 0
	re.previewLink = nil
}

// SettingsPanel manages application settings
type SettingsPanel struct {
	theme        string
	autoRefresh  bool
	refreshRate  int
	showMetadata bool
	compactMode  bool
}

// NewSettingsPanel creates a new settings panel
func NewSettingsPanel() *SettingsPanel {
	return &SettingsPanel{
		theme:        "dark",
		autoRefresh:  true,
		refreshRate:  500,
		showMetadata: true,
		compactMode:  false,
	}
}

// SetTheme sets the UI theme
func (sp *SettingsPanel) SetTheme(theme string) {
	sp.theme = theme
}

// SetAutoRefresh sets auto-refresh
func (sp *SettingsPanel) SetAutoRefresh(enabled bool) {
	sp.autoRefresh = enabled
}

// SetRefreshRate sets the refresh rate
func (sp *SettingsPanel) SetRefreshRate(rate int) {
	sp.refreshRate = rate
}

// SetCompactMode sets compact mode
func (sp *SettingsPanel) SetCompactMode(compact bool) {
	sp.compactMode = compact
}

// StatusBar displays application status
type StatusBar struct {
	status   string
	mode     string
	time     string
	message  string
	progress int
}

// NewStatusBar creates a new status bar
func NewStatusBar() *StatusBar {
	return &StatusBar{
		status:   "Ready",
		mode:     "Normal",
		progress: 0,
	}
}

// GetWidget returns the status bar widget
func (sb *StatusBar) GetWidget() gtk.Widgetter {
	box := gtk.NewBoxOrientation(gtk.OrientationHorizontal, 10)
	statusLabel := gtk.NewLabel(sb.status)
	modeLabel := gtk.NewLabel(sb.mode)
	box.Append(statusLabel)
	box.Append(modeLabel)
	return box
}

// SetStatus sets the status message
func (sb *StatusBar) SetStatus(status string) {
	sb.status = status
}

// SetMode sets the current mode
func (sb *StatusBar) SetMode(mode string) {
	sb.mode = mode
}

// SetMessage sets a temporary message
func (sb *StatusBar) SetMessage(message string) {
	sb.message = message
}

// SetProgress sets the progress value (0-100)
func (sb *StatusBar) SetProgress(progress int) {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	sb.progress = progress
}

// ConnectionDialog handles connection creation
type ConnectionDialog struct {
	dialog         *gtk.Dialog
	outputCombo    *gtk.ComboBox
	inputCombo     *gtk.ComboBox
	outputModel    *gtk.StringList
	inputModel     *gtk.StringList
}

// NewConnectionDialog creates a new connection dialog
func NewConnectionDialog(nodes map[uint32]*client.Node) *ConnectionDialog {
	dialog := gtk.NewDialog()
	dialog.SetTitle("Create Audio Connection")

	// Create lists
	outputModel := gtk.NewStringList([]string{})
	inputModel := gtk.NewStringList([]string{})

	// Populate lists
	for _, node := range nodes {
		for _, port := range node.GetPorts() {
			if port.Direction == client.PortDirectionOutput {
				outputModel.Append(fmt.Sprintf("[%d] %s", port.ID, port.Name))
			} else {
				inputModel.Append(fmt.Sprintf("[%d] %s", port.ID, port.Name))
			}
		}
	}

	outputCombo := gtk.NewComboBox()
	inputCombo := gtk.NewComboBox()

	return &ConnectionDialog{
		dialog:      dialog,
		outputCombo: outputCombo,
		inputCombo:  inputCombo,
		outputModel: outputModel,
		inputModel:  inputModel,
	}
}

// Show displays the dialog
func (cd *ConnectionDialog) Show() bool {
	response := cd.dialog.Run()
	return response == gtk.ResponseOK
}

// GetSelectedPorts returns selected output and input ports
func (cd *ConnectionDialog) GetSelectedPorts() (uint32, uint32) {
	// Extract IDs from combo selections
	return 0, 0
}
