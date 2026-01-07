// pw-gui - GTK4 GUI for PipeWire
package main

import (
	"fmt"
	"log"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/vignemail1/pipewire-go/client"
)

func main() {
	// Initialize GTK
	app := gtk.NewApplication("", 0)

	// Connect signal for activation
	app.ConnectActivate(onActivate)

	// Run application
	status := app.Run(nil)
	if status != 0 {
		log.Fatalf("Application exit status: %d", status)
	}
}

func onActivate(app *gtk.Application) {
	// Connect to PipeWire
	c, err := client.NewClient("pw-gui")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Disconnect()

	// Create main window
	window := gtk.NewApplicationWindow(app)
	window.SetTitle("PipeWire")
	window.SetDefaultSize(1024, 768)

	// Create main container
	box := gtk.NewBox(gtk.OrientationVertical, 0)

	// Create header bar
	header := gtk.NewHeaderBar()
	window.SetTitlebar(header)

	// Create status label
	registry := c.GetRegistry()
	status := fmt.Sprintf("Objects: %d | Nodes: %d | Ports: %d | Links: %d",
		registry.CountObjects(),
		registry.Count("Node"),
		registry.Count("Port"),
		registry.Count("Link"))

	label := gtk.NewLabel(status)
	label.AddCSSClass("title-3")

	box.Append(label)

	// Create drawing area for graph
	drawingArea := gtk.NewDrawingArea()
	drawingArea.SetExpand(true)

	box.Append(drawingArea)

	// Set content
	window.SetChild(box)

	// Show window
	window.Show()
}
