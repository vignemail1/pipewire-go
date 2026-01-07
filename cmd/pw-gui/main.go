// Package main - GUI Client
// cmd/pw-gui/main.go
// PipeWire audio graph GUI client using GTK4

package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/graphene"
	"github.com/vignemail1/pipewire-go/client"
	"github.com/vignemail1/pipewire-go/verbose"
)

// GraphVisualizer renders the audio graph
type GraphVisualizer struct {
	client       *client.Client
	drawingArea  *gtk.DrawingArea
	zoomLevel    float64
	panX, panY   float64
	nodes        map[uint32]nodeLayout
	mutex        sync.RWMutex
}

// nodeLayout stores node position information
type nodeLayout struct {
	ID   uint32
	x, y float64
	w, h float64
}

// NewGraphVisualizer creates a new graph visualizer
func NewGraphVisualizer(pwClient *client.Client) *GraphVisualizer {
	return &GraphVisualizer{
		client:    pwClient,
		zoomLevel: 1.0,
		panX:      0,
		panY:      0,
		nodes:     make(map[uint32]nodeLayout),
	}
}

// SetDrawingArea sets the drawing area for rendering
func (gv *GraphVisualizer) SetDrawingArea(da *gtk.DrawingArea) {
	gv.drawingArea = da
}

// layoutNodes calculates node positions
func (gv *GraphVisualizer) layoutNodes(width, height int) {
	gv.mutex.Lock()
	defer gv.mutex.Unlock()

	nodes := gv.client.GetNodes()
	if len(nodes) == 0 {
		return
	}

	// Simple grid layout
	colsPerRow := 4
	nodeWidth := 120.0
	nodeHeight := 60.0
	startX := 50.0
	startY := 50.0
	xSpacing := 150.0
	ySpacing := 100.0

	for i, node := range nodes {
		col := i % colsPerRow
		row := i / colsPerRow
		gv.nodes[node.ID] = nodeLayout{
			ID: node.ID,
			x:  startX + float64(col)*xSpacing,
			y:  startY + float64(row)*ySpacing,
			w:  nodeWidth,
			h:  nodeHeight,
		}
	}
}

// Draw renders the audio graph
func (gv *GraphVisualizer) Draw(cairo *gtk.DrawingAreaDrawFuncContext, width, height int, nodes []*client.Node, links []*client.Link) {
	ctx := cairo.Context()

	// Apply transformations
	ctx.Translate(gv.panX, gv.panY)
	ctx.Scale(gv.zoomLevel, gv.zoomLevel)

	// Layout nodes
	gv.layoutNodes(width, height)

	// Draw links first (so they appear behind nodes)
	for _, link := range links {
		gv.drawLink(ctx, link)
	}

	// Draw nodes
	for _, node := range nodes {
		gv.drawNode(ctx, node)
	}
}

// drawNode draws a single node
func (gv *GraphVisualizer) drawNode(ctx *gtk.DrawingAreaDrawFuncContext, node *client.Node) {
	gv.mutex.RLock()
	layout, exists := gv.nodes[node.ID]
	gv.mutex.RUnlock()

	if !exists {
		return
	}

	// Determine color based on node state
	var r, g, b float64
	state := node.GetState()
	switch state {
	case "running":
		r, g, b = 0.2, 0.8, 0.2
	case "idle":
		r, g, b = 0.2, 0.2, 0.8
	default:
		r, g, b = 0.6, 0.6, 0.6
	}

	// Draw node rectangle
	ctx.SetSourceRGB(r, g, b)
	ctx.Rectangle(layout.x, layout.y, layout.w, layout.h)
	ctx.Fill()

	// Draw border
	ctx.SetSourceRGB(1, 1, 1)
	ctx.SetLineWidth(2)
	ctx.Rectangle(layout.x, layout.y, layout.w, layout.h)
	ctx.Stroke()

	// Draw node label
	ctx.SetSourceRGB(1, 1, 1)
	ctx.MoveTo(layout.x+5, layout.y+25)
	fmt.Fprintf(os.Stderr, "[%d] %s\\n", node.ID, node.Name())
}

// drawLink draws a connection between two nodes
func (gv *GraphVisualizer) drawLink(ctx *gtk.DrawingAreaDrawFuncContext, link *client.Link) {
	if link.Output == nil || link.Input == nil {
		return
	}

	gv.mutex.RLock()
	_, outputExists := gv.nodes[link.Output.NodeID]
	_, inputExists := gv.nodes[link.Input.NodeID]
	gv.mutex.RUnlock()

	if !outputExists || !inputExists {
		return
	}

	// Draw connection line
	ctx.SetSourceRGB(0.5, 0.5, 0.5)
	ctx.SetLineWidth(2)

	gv.mutex.RLock()
	outNode := gv.nodes[link.Output.NodeID]
	inNode := gv.nodes[link.Input.NodeID]
	gv.mutex.RUnlock()

	// Draw Bezier curve
	x1 := outNode.x + outNode.w
	y1 := outNode.y + outNode.h/2
	x2 := inNode.x
	y2 := inNode.y + inNode.h/2

	ctx.MoveTo(x1, y1)
	ctxControl := (x1 + x2) / 2
	ctx.CubicTo(ctxControl, y1, ctxControl, y2, x2, y2)
	ctx.Stroke()
}

// ZoomIn increases zoom level
func (gv *GraphVisualizer) ZoomIn() {
	gv.zoomLevel *= 1.2
	if gv.drawingArea != nil {
		gv.drawingArea.QueueDraw()
	}
}

// ZoomOut decreases zoom level
func (gv *GraphVisualizer) ZoomOut() {
	gv.zoomLevel /= 1.2
	if gv.drawingArea != nil {
		gv.drawingArea.QueueDraw()
	}
}

// FitToWindow fits the graph to window
func (gv *GraphVisualizer) FitToWindow() {
	gv.zoomLevel = 1.0
	gv.panX = 0
	gv.panY = 0
	if gv.drawingArea != nil {
		gv.drawingArea.QueueDraw()
	}
}

// RoutingEngine manages audio routing
type RoutingEngine struct {
	client        *client.Client
	sourcePorts   []uint32
	sinkPorts     []uint32
	connections   map[uint32]*client.Link
	mutex         sync.RWMutex
}

// NewRoutingEngine creates a new routing engine
func NewRoutingEngine(pwClient *client.Client) *RoutingEngine {
	return &RoutingEngine{
		client:      pwClient,
		sourcePorts: []uint32{},
		sinkPorts:   []uint32{},
		connections: make(map[uint32]*client.Link),
	}
}

// GetSourcePorts returns available source ports
func (re *RoutingEngine) GetSourcePorts() []*client.Port {
	var ports []*client.Port
	for _, node := range re.client.GetNodes() {
		for _, port := range node.GetPorts() {
			if port.Direction == client.PortDirectionOutput {
				ports = append(ports, port)
			}
		}
	}
	return ports
}

// GetSinkPorts returns available sink ports
func (re *RoutingEngine) GetSinkPorts() []*client.Port {
	var ports []*client.Port
	for _, node := range re.client.GetNodes() {
		for _, port := range node.GetPorts() {
			if port.Direction == client.PortDirectionInput {
				ports = append(ports, port)
			}
		}
	}
	return ports
}

// CreateConnection creates a new audio connection
func (re *RoutingEngine) CreateConnection(sourceID, sinkID uint32) error {
	re.mutex.Lock()
	defer re.mutex.Unlock()

	var sourcePort, sinkPort *client.Port

	// Find ports
	for _, node := range re.client.GetNodes() {
		for _, port := range node.GetPorts() {
			if port.ID == sourceID {
				sourcePort = port
			}
			if port.ID == sinkID {
				sinkPort = port
			}
		}
	}

	if sourcePort == nil || sinkPort == nil {
		return fmt.Errorf("port not found")
	}

	link, err := re.client.CreateLink(sourcePort, sinkPort, nil)
	if err != nil {
		return err
	}

	re.connections[link.ID] = link
	return nil
}

// DeleteConnection removes an audio connection
func (re *RoutingEngine) DeleteConnection(linkID uint32) error {
	re.mutex.Lock()
	defer re.mutex.Unlock()

	err := re.client.DestroyLink(linkID)
	if err != nil {
		return err
	}

	delete(re.connections, linkID)
	return nil
}

// SettingsPanel manages application settings
type SettingsPanel struct {
	themeDark    bool
	autoRefresh  bool
	refreshRate  int // in milliseconds
	scaleFactor  float64
}

// NewSettingsPanel creates a new settings panel
func NewSettingsPanel() *SettingsPanel {
	return &SettingsPanel{
		themeDark:   true,
		autoRefresh: true,
		refreshRate: 1000,
		scaleFactor: 1.0,
	}
}

// SetThemeDark sets dark theme preference
func (sp *SettingsPanel) SetThemeDark(dark bool) {
	sp.themeDark = dark
}

// SetAutoRefresh sets auto refresh preference
func (sp *SettingsPanel) SetAutoRefresh(enabled bool) {
	sp.autoRefresh = enabled
}

// SetRefreshRate sets refresh rate
func (sp *SettingsPanel) SetRefreshRate(rate int) {
	sp.refreshRate = rate
}

// GetRefreshRate returns refresh rate in milliseconds
func (sp *SettingsPanel) GetRefreshRate() int {
	return sp.refreshRate
}

// StatusBar displays application status
type StatusBar struct {
	status  string
	mode    string
	label   *gtk.Label
	modeLabel *gtk.Label
}

// NewStatusBar creates a new status bar
func NewStatusBar() *StatusBar {
	statusLabel := gtk.NewLabel("Ready")
	statusLabel.SetXAlign(0)

	modeLabel := gtk.NewLabel("Normal")
	modeLabel.SetXAlign(1)

	return &StatusBar{
		status:    "Ready",
		mode:      "Normal",
		label:     statusLabel,
		modeLabel: modeLabel,
	}
}

// GetWidget returns the GTK widget for the status bar
func (sb *StatusBar) GetWidget() gtk.Widgetter {
	box := gtk.NewBoxOrientation(gtk.OrientationHorizontal, 10)
	box.SetMarginTop(5)
	box.SetMarginBottom(5)
	box.SetMarginStart(10)
	box.SetMarginEnd(10)

	// Status label
	separator := gtk.NewSeparator(gtk.OrientationVertical)

	box.Append(sb.label)
	box.SetCenterWidget(separator)
	box.Append(sb.modeLabel)

	return box
}

// SetStatus updates the status message
func (sb *StatusBar) SetStatus(status string) {
	sb.status = status
	if sb.label != nil {
		sb.label.SetText(status)
	}
}

// SetMode updates the mode display
func (sb *StatusBar) SetMode(mode string) {
	sb.mode = mode
	if sb.modeLabel != nil {
		sb.modeLabel.SetText(mode)
	}
}

// GuiApp represents the main GUI application
type GuiApp struct {
	app            *gtk.Application
	window         *gtk.ApplicationWindow
	client         *client.Client
	logger         *verbose.Logger
	graphRenderer  *GraphVisualizer
	routingEngine  *RoutingEngine
	settingsPanel  *SettingsPanel
	statusBar      *StatusBar
	notebook       *gtk.Notebook
	updateID       uint

	// State
	selectedNode   uint32
	selectedPort   uint32
	selectedLink   uint32
	routingMode    bool
	isDarkMode     bool
}

// NewGuiApp creates a new GUI application
func NewGuiApp(socketPath string, logger *verbose.Logger) (*GuiApp, error) {
	// Create GTK application
	app := gtk.NewApplication("com.github.pipewire-go.gui", 0)

	// Create PipeWire client
	pwClient, err := client.NewClient(socketPath, logger)
	if err != nil {
		return nil, err
	}

	gui := &GuiApp{
		app:            app,
		client:         pwClient,
		logger:         logger,
		routingMode:    false,
		isDarkMode:     true,
		graphRenderer:  NewGraphVisualizer(pwClient),
		routingEngine:  NewRoutingEngine(pwClient),
		statusBar:      NewStatusBar(),
		settingsPanel:  NewSettingsPanel(),
	}

	// Setup application signals
	gui.setupSignals()

	return gui, nil
}

// setupSignals sets up GTK application signals
func (ga *GuiApp) setupSignals() {
	ga.app.ConnectActivate(func() {
		ga.onActivate()
	})
}

// onActivate is called when the application is activated
func (ga *GuiApp) onActivate() {
	// Create main window
	ga.window = gtk.NewApplicationWindow(ga.app)
	ga.window.SetTitle("PipeWire Audio Graph")
	ga.window.SetDefaultSize(1200, 800)
	ga.window.SetIconName("audio-card")

	// Create main layout
	mainBox := gtk.NewBoxOrientation(gtk.OrientationVertical, 0)
	mainBox.SetMarginTop(10)
	mainBox.SetMarginBottom(10)
	mainBox.SetMarginStart(10)
	mainBox.SetMarginEnd(10)

	// Create menu bar
	menuBar := ga.createMenuBar()
	mainBox.Append(menuBar)

	// Create notebook for tabs
	ga.notebook = gtk.NewNotebook()
	ga.notebook.SetCanFocus(false)

	// Add tabs
	ga.addGraphTab()
	ga.addDevicesTab()
	ga.addConnectionsTab()
	ga.addPropertiesTab()

	mainBox.Append(ga.notebook)

	// Add status bar
	statusWidget := ga.statusBar.GetWidget()
	mainBox.Append(statusWidget)

	// Set main content
	ga.window.SetChild(mainBox)

	// Apply theme
	if ga.isDarkMode {
		ga.applyDarkTheme()
	}

	// Update initial status
	ga.statusBar.SetStatus("Ready")
	ga.statusBar.SetMode("Normal")

	// Show window
	ga.window.Show()

	// Start update loop
	ga.startUpdateLoop()
}

// createMenuBar creates the application menu bar
func (ga *GuiApp) createMenuBar() *gtk.Box {
	menuBox := gtk.NewBoxOrientation(gtk.OrientationHorizontal, 10)

	// File menu button
	fileBtn := gtk.NewButtonWithLabel("File")
	fileBtn.ConnectClicked(func() {
		ga.showFileMenu()
	})

	// Edit menu button
	editBtn := gtk.NewButtonWithLabel("Edit")
	editBtn.ConnectClicked(func() {
		ga.showEditMenu()
	})

	// View menu button
	viewBtn := gtk.NewButtonWithLabel("View")
	viewBtn.ConnectClicked(func() {
		ga.showViewMenu()
	})

	// Audio menu button
	audioBtn := gtk.NewButtonWithLabel("Audio")
	audioBtn.ConnectClicked(func() {
		ga.showAudioMenu()
	})

	// Help menu button
	helpBtn := gtk.NewButtonWithLabel("Help")
	helpBtn.ConnectClicked(func() {
		ga.showHelpMenu()
	})

	menuBox.Append(fileBtn)
	menuBox.Append(editBtn)
	menuBox.Append(viewBtn)
	menuBox.Append(audioBtn)
	menuBox.Append(helpBtn)

	return menuBox
}

// addGraphTab adds the graph visualization tab
func (ga *GuiApp) addGraphTab() {
	// Create graph container
	graphBox := gtk.NewBoxOrientation(gtk.OrientationVertical, 5)
	graphBox.SetMarginTop(5)
	graphBox.SetMarginBottom(5)
	graphBox.SetMarginStart(5)
	graphBox.SetMarginEnd(5)

	// Create drawing area for graph
	drawingArea := gtk.NewDrawingArea()
	drawingArea.SetContentWidth(800)
	drawingArea.SetContentHeight(600)

	// Setup draw handler
	drawingArea.SetDrawFunc(func(_ *gtk.DrawingArea, cr *gtk.DrawingAreaDrawFuncContext, width, height int) {
		ga.drawAudioGraph(cr, width, height)
	})

	// Queue redraw periodically
	ga.graphRenderer.SetDrawingArea(drawingArea)

	graphBox.Append(drawingArea)

	// Add toolbar
	toolbar := ga.createGraphToolbar()
	graphBox.Prepend(toolbar)

	// Add to notebook
	label := gtk.NewLabel("Graph")
	ga.notebook.AppendPage(graphBox, label)
}

// addDevicesTab adds the devices list tab
func (ga *GuiApp) addDevicesTab() {
	devicesBox := gtk.NewBoxOrientation(gtk.OrientationVertical, 5)
	devicesBox.SetMarginTop(5)
	devicesBox.SetMarginBottom(5)
	devicesBox.SetMarginStart(5)
	devicesBox.SetMarginEnd(5)

	// Create devices list
	listModel := gtk.NewStringList([]string{})

	// Create list view
	factory := gtk.NewSignalListItemFactory()
	factory.ConnectSetup(func(_ *gtk.SignalListItemFactory, obj *gtk.ListItem) {
		label := gtk.NewLabel("")
		obj.SetChild(label)
	})

	factory.ConnectBind(func(_ *gtk.SignalListItemFactory, obj *gtk.ListItem) {
		label := obj.Child().(*gtk.Label)
		item := obj.Item().(*gtk.StringObject)
		label.SetText(item.String())
	})

	listView := gtk.NewListView(gtk.NewMultiSelection(listModel), factory)
	listView.SetCanFocus(true)

	// Populate devices
	ga.updateDevicesList(listModel)

	// Add scroll view
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(listView)
	scrolled.SetPropagateNaturalWidth(true)
	scrolled.SetPropagateNaturalHeight(true)

	devicesBox.Append(scrolled)

	// Add to notebook
	label := gtk.NewLabel("Devices")
	ga.notebook.AppendPage(devicesBox, label)
}

// addConnectionsTab adds the connections list tab
func (ga *GuiApp) addConnectionsTab() {
	connectionsBox := gtk.NewBoxOrientation(gtk.OrientationVertical, 5)
	connectionsBox.SetMarginTop(5)
	connectionsBox.SetMarginBottom(5)
	connectionsBox.SetMarginStart(5)
	connectionsBox.SetMarginEnd(5)

	// Create connections list
	listModel := gtk.NewStringList([]string{})

	factory := gtk.NewSignalListItemFactory()
	factory.ConnectSetup(func(_ *gtk.SignalListItemFactory, obj *gtk.ListItem) {
		box := gtk.NewBoxOrientation(gtk.OrientationHorizontal, 10)
		label := gtk.NewLabel("")
		button := gtk.NewButtonWithLabel("Delete")
		box.Append(label)
		box.SetCenterWidget(button)
		obj.SetChild(box)
	})

	factory.ConnectBind(func(_ *gtk.SignalListItemFactory, obj *gtk.ListItem) {
		box := obj.Child().(*gtk.Box)
		label := box.FirstChild().(*gtk.Label)
		item := obj.Item().(*gtk.StringObject)
		label.SetText(item.String())
	})

	listView := gtk.NewListView(gtk.NewMultiSelection(listModel), factory)
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(listView)
	scrolled.SetPropagateNaturalWidth(true)
	scrolled.SetPropagateNaturalHeight(true)

	connectionsBox.Append(scrolled)

	// Add button toolbar
	toolbar := gtk.NewBoxOrientation(gtk.OrientationHorizontal, 5)
	createBtn := gtk.NewButtonWithLabel("Create Connection")
	createBtn.ConnectClicked(func() {
		ga.showCreateConnectionDialog()
	})
	toolbar.Append(createBtn)
	connectionsBox.Prepend(toolbar)

	// Add to notebook
	label := gtk.NewLabel("Connections")
	ga.notebook.AppendPage(connectionsBox, label)
}

// addPropertiesTab adds the properties panel tab
func (ga *GuiApp) addPropertiesTab() {
	propsBox := gtk.NewBoxOrientation(gtk.OrientationVertical, 5)
	propsBox.SetMarginTop(5)
	propsBox.SetMarginBottom(5)
	propsBox.SetMarginStart(5)
	propsBox.SetMarginEnd(5)

	// Create properties display
	propsLabel := gtk.NewLabel("Select a device to view properties")
	propsLabel.SetWrapMode(gtk.WrapWord)

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(propsLabel)
	scrolled.SetPropagateNaturalWidth(true)
	scrolled.SetPropagateNaturalHeight(true)

	propsBox.Append(scrolled)

	// Add to notebook
	label := gtk.NewLabel("Properties")
	ga.notebook.AppendPage(propsBox, label)
}

// createGraphToolbar creates toolbar for graph tab
func (ga *GuiApp) createGraphToolbar() *gtk.Box {
	toolbar := gtk.NewBoxOrientation(gtk.OrientationHorizontal, 5)

	// Refresh button
	refreshBtn := gtk.NewButtonWithLabel("Refresh")
	refreshBtn.ConnectClicked(func() {
		ga.refreshGraph()
	})

	// Zoom in button
	zoomInBtn := gtk.NewButtonWithLabel("Zoom +")
	zoomInBtn.ConnectClicked(func() {
		ga.graphRenderer.ZoomIn()
	})

	// Zoom out button
	zoomOutBtn := gtk.NewButtonWithLabel("Zoom -")
	zoomOutBtn.ConnectClicked(func() {
		ga.graphRenderer.ZoomOut()
	})

	// Fit to window button
	fitBtn := gtk.NewButtonWithLabel("Fit Window")
	fitBtn.ConnectClicked(func() {
		ga.graphRenderer.FitToWindow()
	})

	toolbar.Append(refreshBtn)
	toolbar.Append(zoomInBtn)
	toolbar.Append(zoomOutBtn)
	toolbar.Append(fitBtn)

	return toolbar
}

// drawAudioGraph draws the audio graph
func (ga *GuiApp) drawAudioGraph(cr *gtk.DrawingAreaDrawFuncContext, width, height int) {
	// Get cairo context
	cairo := cr.Context()

	// Draw background
	cairo.SetSourceRGB(0.1, 0.1, 0.1)
	cairo.Rectangle(0, 0, float64(width), float64(height))
	cairo.Fill()

	// Draw graph using renderer
	ga.graphRenderer.Draw(cr, width, height, ga.client.GetNodes(), ga.client.GetLinks())
}

// updateDevicesList updates the devices list
func (ga *GuiApp) updateDevicesList(listModel *gtk.StringList) {
	items := []string{}
	for _, node := range ga.client.GetNodes() {
		items = append(items, fmt.Sprintf("[%d] %s (%s)", node.ID, node.Name(), node.GetState()))
	}
	listModel.Splice(0, listModel.NItems(), items)
}

// showFileMenu shows the file menu
func (ga *GuiApp) showFileMenu() {
	dialog := gtk.NewMessageDialog(
		ga.window,
		gtk.DialogDestroyWithParent,
		gtk.MessageTypeInfo,
		gtk.ButtonsOK,
		"File Menu",
	)
	dialog.SetSecondaryText("File menu options:\n- Export configuration\n- Import configuration\n- Exit")
	dialog.ConnectResponse(func(_ *gtk.MessageDialog, _ int) {
		dialog.Close()
	})
	dialog.Show()
	ga.logger.Info("File menu opened")
}

// showEditMenu shows the edit menu
func (ga *GuiApp) showEditMenu() {
	dialog := gtk.NewMessageDialog(
		ga.window,
		gtk.DialogDestroyWithParent,
		gtk.MessageTypeInfo,
		gtk.ButtonsOK,
		"Edit Menu",
	)
	dialog.SetSecondaryText("Edit menu options:\n- Undo\n- Redo\n- Copy\n- Paste")
	dialog.ConnectResponse(func(_ *gtk.MessageDialog, _ int) {
		dialog.Close()
	})
	dialog.Show()
	ga.logger.Info("Edit menu opened")
}

// showViewMenu shows the view menu
func (ga *GuiApp) showViewMenu() {
	dialog := gtk.NewMessageDialog(
		ga.window,
		gtk.DialogDestroyWithParent,
		gtk.MessageTypeInfo,
		gtk.ButtonsOK,
		"View Menu",
	)
	dialog.SetSecondaryText("View menu options:\n- Toggle sidebar\n- Dark/Light theme\n- Full screen")
	dialog.ConnectResponse(func(_ *gtk.MessageDialog, _ int) {
		dialog.Close()
	})
	dialog.Show()
	ga.logger.Info("View menu opened")
}

// showAudioMenu shows the audio menu
func (ga *GuiApp) showAudioMenu() {
	dialog := gtk.NewMessageDialog(
		ga.window,
		gtk.DialogDestroyWithParent,
		gtk.MessageTypeInfo,
		gtk.ButtonsOK,
		"Audio Menu",
	)
	dialog.SetSecondaryText("Audio menu options:\n- Create connection\n- Delete connection\n- Routing mode\n- Device settings")
	dialog.ConnectResponse(func(_ *gtk.MessageDialog, _ int) {
		dialog.Close()
	})
	dialog.Show()
	ga.logger.Info("Audio menu opened")
}

// showHelpMenu shows the help menu
func (ga *GuiApp) showHelpMenu() {
	dialog := gtk.NewMessageDialog(
		ga.window,
		gtk.DialogDestroyWithParent,
		gtk.MessageTypeInfo,
		gtk.ButtonsOK,
		"Help",
	)
	dialog.SetSecondaryText("PipeWire Audio Graph GUI v1.0\n\nFeatures:\n- Visual graph editing\n- Audio device management\n- Connection routing\n- Properties display\n\nDocs: https://github.com/vignemail1/pipewire-go")
	dialog.ConnectResponse(func(_ *gtk.MessageDialog, _ int) {
		dialog.Close()
	})
	dialog.Show()
	ga.logger.Info("Help menu opened")
}

// showCreateConnectionDialog shows the create connection dialog
func (ga *GuiApp) showCreateConnectionDialog() {
	dialog := gtk.NewMessageDialog(
		ga.window,
		gtk.DialogDestroyWithParent,
		gtk.MessageTypeQuestion,
		gtk.ButtonsOKCancel,
		"Create Connection",
	)
	dialog.SetSecondaryText("Select output and input ports to connect")

	dialog.ConnectResponse(func(_ *gtk.MessageDialog, responseID int) {
		if responseID == int(gtk.ResponseOK) {
			ga.logger.Info("Creating connection...")
			ga.statusBar.SetStatus("Creating connection...")
		}
		dialog.Close()
	})

	dialog.Show()
}

// refreshGraph refreshes the audio graph
func (ga *GuiApp) refreshGraph() {
	ga.logger.Info("Refreshing graph...")
	ga.statusBar.SetStatus("Refreshing...")
}

// applyDarkTheme applies dark theme to the application
func (ga *GuiApp) applyDarkTheme() {
	settings := gtk.SettingsGetDefault()
	settings.SetPropertyBool("gtk-application-prefer-dark-theme", true)
}

// startUpdateLoop starts the periodic update loop
func (ga *GuiApp) startUpdateLoop() {
	refreshRate := ga.settingsPanel.GetRefreshRate()
	ga.updateID = glib.TimeoutAdd(uint(refreshRate), func() bool {
		if !ga.settingsPanel.autoRefresh {
			return true
		}

		// Update graph
		ga.statusBar.SetStatus(fmt.Sprintf("Updated at %s", time.Now().Format("15:04:05")))
		ga.logger.Debug("Graph updated")

		return true
	})
}

// Run starts the application
func (ga *GuiApp) Run() int {
	return ga.app.Run(os.Args)
}

// Close closes the application
func (ga *GuiApp) Close() {
	if ga.updateID > 0 {
		glib.SourceRemove(ga.updateID)
	}
	if ga.client != nil {
		ga.client.Close()
	}
}

// Main entry point
func main() {
	var (
		socketPath = os.Getenv("PIPEWIRE_SOCKET")
		verbose_   = os.Getenv("VERBOSE") == "1"
	)

	if socketPath == "" {
		socketPath = "/run/user/1000/pipewire-0"
	}

	// Create logger
	level := verbose.LogLevelInfo
	if verbose_ {
		level = verbose.LogLevelDebug
	}
	logger := verbose.NewLogger(level, false)

	// Create GUI application
	gui, err := NewGuiApp(socketPath, logger)
	if err != nil {
		log.Fatalf("Failed to create GUI: %v", err)
	}

	// Run application
	code := gui.Run()

	// Cleanup
	gui.Close()

	os.Exit(code)
}
