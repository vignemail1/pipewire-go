package main

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
)

// App represents the TUI application
type App struct {
	screen tcell.Screen
	router *EventRouter
	keyboardHandler *KeyboardHandler
	mouseHandler *MouseHandler
	resizeHandler *ResizeHandler
	running bool
	mu sync.RWMutex
	done chan bool
}

// NewApp creates a new TUI application
func NewApp() (*App, error) {
	// Initialize tcell screen
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("failed to create screen: %w", err)
	}

	if err := s.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize screen: %w", err)
	}

	// Create app
	app := &App{
		screen: s,
		router: NewEventRouter(),
		keyboardHandler: NewKeyboardHandler(),
		mouseHandler: NewMouseHandler(),
		resizeHandler: NewResizeHandler(),
		running: true,
		done: make(chan bool, 1),
	}

	// Register handlers with router
	app.router.Register(app.keyboardHandler)
	app.router.Register(app.mouseHandler)
	app.router.Register(app.resizeHandler)

	return app, nil
}

// Run starts the event loop
func (a *App) Run() error {
	a.mu.Lock()
	a.running = true
	a.mu.Unlock()

	for a.isRunning() {
		event := a.screen.PollEvent()
		if event == nil {
			continue
		}

		// Route event to handlers
		a.router.RouteEvent(event)

		// Check if quit was requested
		select {
		case <-a.keyboardHandler.QuitChan:
			return a.Stop()
		default:
		}
	}

	return nil
}

// Stop terminates the application
func (a *App) Stop() error {
	a.mu.Lock()
	a.running = false
	a.mu.Unlock()

	// Cleanup
	a.screen.Fini()

	return nil
}

// isRunning checks if app is running
func (a *App) isRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}

// RegisterKeyHandler registers a keyboard callback
func (a *App) RegisterKeyHandler(key tcell.Key, callback func()) {
	a.keyboardHandler.RegisterKey(key, callback)
}

// RegisterMouseHandler sets the mouse handler
func (a *App) RegisterMouseHandler(handler *MouseHandler) {
	a.mh = handler
}

// RegisterResizeHandler sets the resize handler
func (a *App) RegisterResizeHandler(handler *ResizeHandler) {
	a.resizeHandler = handler
}

// GetScreen returns the tcell screen
func (a *App) GetScreen() tcell.Screen {
	return a.screen
}
