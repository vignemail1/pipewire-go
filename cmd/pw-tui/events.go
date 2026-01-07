package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// EventHandler processes TUI events
type EventHandler interface {
	HandleEvent(event tcell.Event) bool
}

// EventRouter routes events to handlers
type EventRouter struct {
	handlers []EventHandler
}

// NewEventRouter creates a new event router
func NewEventRouter() *EventRouter {
	return &EventRouter{
		handlers: make([]EventHandler, 0),
	}
}

// Register adds an event handler
func (er *EventRouter) Register(handler EventHandler) {
	er.handlers = append(er.handlers, handler)
}

// RouteEvent dispatches event to handlers
func (er *EventRouter) RouteEvent(event tcell.Event) bool {
	for _, handler := range er.handlers {
		if handler.HandleEvent(event) {
			return true // Event consumed
		}
	}
	return false // Event not consumed
}

// KeyboardHandler handles keyboard events
type KeyboardHandler struct {
	QuitChan chan bool
	Callbacks map[tcell.Key]func()
}

// NewKeyboardHandler creates a new keyboard handler
func NewKeyboardHandler() *KeyboardHandler {
	return &KeyboardHandler{
		QuitChan: make(chan bool, 1),
		Callbacks: make(map[tcell.Key]func()),
	}
}

// RegisterKey registers a callback for a key
func (kh *KeyboardHandler) RegisterKey(key tcell.Key, callback func()) {
	kh.Callbacks[key] = callback
}

// HandleEvent processes keyboard events
func (kh *KeyboardHandler) HandleEvent(event tcell.Event) bool {
	switch ev := event.(type) {
	case *tcell.EventKey:
		// Handle common keys
		switch ev.Key() {
		case tcell.KeyCtrlC, tcell.KeyEsc:
			select {
			case kh.QuitChan <- true:
			default:
			}
			return true
		}

		// Check registered callbacks
		if callback, exists := kh.Callbacks[ev.Key()]; exists {
			callback()
			return true
		}

		return false
	default:
		return false
	}
}

// MouseHandler handles mouse events
type MouseHandler struct {
	OnClick   func(x, y int)
	OnScroll  func(x, y int, direction int)
	OnMove    func(x, y int)
}

// NewMouseHandler creates a new mouse handler
func NewMouseHandler() *MouseHandler {
	return &MouseHandler{}
}

// HandleEvent processes mouse events
func (mh *MouseHandler) HandleEvent(event tcell.Event) bool {
	switch ev := event.(type) {
	case *tcell.EventMouse:
		x, y := ev.Position()
		button := ev.Buttons()

		if button&tcell.Button1 != 0 && mh.OnClick != nil {
			mh.OnClick(x, y)
			return true
		}

		if button&tcell.WheelUp != 0 && mh.OnScroll != nil {
			mh.OnScroll(x, y, -1)
			return true
		}

		if button&tcell.WheelDown != 0 && mh.OnScroll != nil {
			mh.OnScroll(x, y, 1)
			return true
		}

		if mh.OnMove != nil {
			mh.OnMove(x, y)
		}

		return false
	default:
		return false
	}
}

// ResizeHandler handles terminal resize events
type ResizeHandler struct {
	OnResize func(width, height int)
}

// NewResizeHandler creates a new resize handler
func NewResizeHandler() *ResizeHandler {
	return &ResizeHandler{}
}

// HandleEvent processes resize events
func (rh *ResizeHandler) HandleEvent(event tcell.Event) bool {
	switch ev := event.(type) {
	case *tcell.EventResize:
		if rh.OnResize != nil {
			rh.OnResize(ev.Size())
		}
		return true
	default:
		return false
	}
}
