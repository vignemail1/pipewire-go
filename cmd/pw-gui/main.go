// Package main - GUI Client
// cmd/pw-gui/main.go
// PipeWire audio graph GUI client using GTK4

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/yourusername/pipewire-go/client"
	"github.com/yourusername/pipewire-go/verbose"
)

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
		app:           app,
		client:        pwClient,
		logger:        logger,
		routingMode:   false,
		isDarkMode:    true,
		graphRenderer: NewGraphVisualizer(pwClient),
		routingEngine: NewRoutingEngine(pwClient),
		statusBar:     NewStatusBar(),
		settingsPanel: NewSettingsPanel(),
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
		box.Append(button)
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
	ga.graphRenderer.Draw(cairo, width, height, ga.client.GetNodes(), ga.client.GetLinks())
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
	ga.logger.Info("File menu opened")
}

// showEditMenu shows the edit menu
func (ga *GuiApp) showEditMenu() {
	ga.logger.Info("Edit menu opened")
}

// showViewMenu shows the view menu
func (ga *GuiApp) showViewMenu() {
	ga.logger.Info("View menu opened")
}

// showAudioMenu shows the audio menu
func (ga *GuiApp) showAudioMenu() {
	ga.logger.Info("Audio menu opened")
}

// showHelpMenu shows the help menu
func (ga *GuiApp) showHelpMenu() {
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
		if responseID == gtk.ResponseOK {
			ga.logger.Info("Creating connection...")
		}
		dialog.Close()
	})

	dialog.Show()
}

// refreshGraph refreshes the audio graph
func (ga *GuiApp) refreshGraph() {
	ga.logger.Info("Refreshing graph...")
	ga.statusBar.SetStatus("Refreshing...")
	// Trigger redraw
}

// applyDarkTheme applies dark theme to the application
func (ga *GuiApp) applyDarkTheme() {
	settings := gtk.SettingsGetDefault()
	settings.SetPropertyBool("gtk-application-prefer-dark-theme", true)
}

// startUpdateLoop starts the periodic update loop
func (ga *GuiApp) startUpdateLoop() {
	// In a real implementation, this would use glib.TimeoutAdd
	// to refresh the UI periodically
	ga.logger.Debug("Update loop started")
}

// Run starts the application
func (ga *GuiApp) Run() int {
	return ga.app.Run(os.Args)
}

// Close closes the application
func (ga *GuiApp) Close() {
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
		socketPath = "/run/pipewire-0"
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
