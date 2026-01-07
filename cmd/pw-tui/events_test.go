package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestEventRouter(t *testing.T) {
	router := NewEventRouter()

	var handled bool
	handler := &testHandler{handleFunc: func(event tcell.Event) bool {
		handled = true
		return true
	}}

	router.Register(handler)

	event := &tcell.EventKey{}
	result := router.RouteEvent(event)

	if !handled {
		t.Error("Handler was not called")
	}
	if !result {
		t.Error("Event was not consumed")
	}
}

func TestKeyboardHandler(t *testing.T) {
	kh := NewKeyboardHandler()

	// Register callback
	called := false
	kh.RegisterKey(tcell.KeyF1, func() {
		called = true
	})

	// Create event
	event := tcell.NewEventKey(tcell.KeyF1, 0, tcell.ModNone)

	// Handle event
	result := kh.HandleEvent(event)

	if !result {
		t.Error("Event was not consumed")
	}
	if !called {
		t.Error("Callback was not called")
	}
}

func TestMouseHandler(t *testing.T) {
	mh := NewMouseHandler()

	// Register callback
	var clickX, clickY int
	mh.OnClick = func(x, y int) {
		clickX = x
		clickY = y
	}

	// Create event
	event := tcell.NewEventMouse(10, 20, tcell.Button1, tcell.ModNone)

	// Handle event
	result := mh.HandleEvent(event)

	if !result {
		t.Error("Event was not consumed")
	}
	if clickX != 10 || clickY != 20 {
		t.Errorf("Click position incorrect: got (%d, %d), want (10, 20)", clickX, clickY)
	}
}

func TestResizeHandler(t *testing.T) {
	rh := NewResizeHandler()

	// Register callback
	var width, height int
	rh.OnResize = func(w, h int) {
		width = w
		height = h
	}

	// Create event
	event := tcell.NewEventResize(80, 24)

	// Handle event
	result := rh.HandleEvent(event)

	if !result {
		t.Error("Event was not consumed")
	}
	if width != 80 || height != 24 {
		t.Errorf("Resize dimensions incorrect: got (%d, %d), want (80, 24)", width, height)
	}
}

// Test helper
type testHandler struct {
	handleFunc func(tcell.Event) bool
}

func (th *testHandler) HandleEvent(event tcell.Event) bool {
	return th.handleFunc(event)
}
