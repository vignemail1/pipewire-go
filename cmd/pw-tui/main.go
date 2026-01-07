// pw-tui - Terminal UI for PipeWire
package main

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/vignemail1/pipewire-go/client"
)

func main() {
	// Create TUI app
	app, err := NewApp()
	if err != nil {
		log.Fatalf("Failed to create TUI: %v", err)
	}
	defer app.Stop()

	// Connect to PipeWire
	c, err := client.NewClient("pw-tui")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Disconnect()

	// Setup keyboard handlers
	app.RegisterKeyHandler(tcell.KeyF1, func() {
		displayHelp(app)
	})

	app.RegisterKeyHandler(tcell.KeyCtrlR, func() {
		// Refresh
	})

	// Setup mouse handler
	app.mouseHandler.OnClick = func(x, y int) {
		// Handle click
	}

	// Setup resize handler
	app.resizeHandler.OnResize = func(width, height int) {
		// Handle resize
	}

	// Draw initial screen
	drawScreen(app, c)

	// Run event loop
	if err := app.Run(); err != nil {
		log.Fatalf("App error: %v", err)
	}
}

func drawScreen(app *App, c *client.Client) {
	s := app.GetScreen()
	s.Clear()

	// Draw title
	drawString(s, 0, 0, tcell.StyleDefault.Bold(true), "PipeWire TUI")

	// Draw info
	registry := c.GetRegistry()
	info := fmt.Sprintf("Objects: %d | Nodes: %d | Ports: %d | Links: %d",
		registry.CountObjects(),
		registry.Count("Node"),
		registry.Count("Port"),
		registry.Count("Link"))

	drawString(s, 0, 2, tcell.StyleDefault, info)

	// Draw help
	drawString(s, 0, 4, tcell.StyleDefault, "Press F1 for help, Ctrl+C to quit")

	s.Show()
}

func displayHelp(app *App) {
	s := app.GetScreen()
	s.Clear()

	drawString(s, 0, 0, tcell.StyleDefault.Bold(true), "Help")
	drawString(s, 0, 2, tcell.StyleDefault, "F1 - Show this help")
	drawString(s, 0, 3, tcell.StyleDefault, "Ctrl+R - Refresh")
	drawString(s, 0, 4, tcell.StyleDefault, "Ctrl+C - Quit")
	drawString(s, 0, 6, tcell.StyleDefault, "Press any key to continue...")

	s.Show()

	// Wait for key
	s.PollEvent()
}

func drawString(s tcell.Screen, x, y int, style tcell.Style, text string) {
	for i, r := range text {
		s.SetContent(x+i, y, r, nil, style)
	}
}
