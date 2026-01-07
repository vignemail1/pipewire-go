// Package core - Message frame tests
// core/message_test.go
// Comprehensive test suite for protocol message handling

package core

import (
	"testing"
	"time"

	"github.com/vignemail1/pipewire-go/spa"
)

// TestMessageFrameMarshal tests basic message frame marshalling
func TestMessageFrameMarshal(t *testing.T) {
	tests := []struct {
		name    string
		frame   *MessageFrame
		wantErr bool
	}{
		{
			name: "valid frame without POD",
			frame: &MessageFrame{
				ObjectID: 1,
				MethodID: 4,
				Sequence: 1,
				PODData:  nil,
			},
			wantErr: false,
		},
		{
			name: "nil frame",
			frame: nil,
			wantErr: true,
		},
		{
			name: "frame with large sequence",
			frame: &MessageFrame{
				ObjectID: 999,
				MethodID: 255,
				Sequence: 0xFFFFFFFF,
				PODData:  nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.frame.Marshal()
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && len(data) < 12 {
				t.Errorf("Marshal() returned data too short: %d bytes (expected >= 12)", len(data))
			}
		})
	}
}

// TestMessageFrameUnmarshal tests message frame unmarshalling
func TestMessageFrameUnmarshal(t *testing.T) {
	// Create a valid frame and marshal it
	original := &MessageFrame{
		ObjectID: 42,
		MethodID: 3,
		Sequence: 5,
		PODData:  nil,
	}

	data, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal() failed: %v", err)
	}

	// Unmarshal it
	restored := &MessageFrame{}
	if err := restored.Unmarshal(data); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	// Verify fields match
	if restored.ObjectID != original.ObjectID {
		t.Errorf("ObjectID mismatch: got %d, want %d", restored.ObjectID, original.ObjectID)
	}
	if restored.MethodID != original.MethodID {
		t.Errorf("MethodID mismatch: got %d, want %d", restored.MethodID, original.MethodID)
	}
	if restored.Sequence != original.Sequence {
		t.Errorf("Sequence mismatch: got %d, want %d", restored.Sequence, original.Sequence)
	}
}

// TestMessageFrameTooShort tests error handling for short data
func TestMessageFrameTooShort(t *testing.T) {
	frame := &MessageFrame{}
	err := frame.Unmarshal([]byte{1, 2, 3})
	if err == nil {
		t.Error("Expected error for short data, got nil")
	}
}

// TestMessageBuilder tests the fluent builder API
func TestMessageBuilder(t *testing.T) {
	builder := NewMessageBuilder(10, 2)
	if builder == nil {
		t.Fatal("NewMessageBuilder returned nil")
	}

	frame := builder.
		WithSequence(99).
		Build()

	if frame == nil {
		t.Fatal("Build() returned nil")
	}

	if frame.ObjectID != 10 || frame.MethodID != 2 || frame.Sequence != 99 {
		t.Errorf("Builder result incorrect: obj=%d method=%d seq=%d", frame.ObjectID, frame.MethodID, frame.Sequence)
	}
}

// TestEventHandlerRegister tests handler registration
func TestEventHandlerRegister(t *testing.T) {
	eh := NewEventHandler()
	if eh == nil {
		t.Fatal("NewEventHandler returned nil")
	}

	called := false
	handler := func(frame *MessageFrame) error {
		called = true
		return nil
	}

	err := eh.RegisterHandler(42, handler)
	if err != nil {
		t.Fatalf("RegisterHandler failed: %v", err)
	}

	if eh.HandlerCount(42) != 1 {
		t.Errorf("HandlerCount: got %d, want 1", eh.HandlerCount(42))
	}
}

// TestEventHandlerDispatch tests event dispatching to handlers
func TestEventHandlerDispatch(t *testing.T) {
	eh := NewEventHandler()

	received := &MessageFrame{}
	handler := func(frame *MessageFrame) error {
		received = frame
		return nil
	}

	eh.RegisterHandler(42, handler)

	frame := &MessageFrame{
		ObjectID: 42,
		MethodID: 1,
		Sequence: 10,
	}

	err := eh.Dispatch(frame)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if received.Sequence != 10 {
		t.Errorf("Handler received wrong frame: seq=%d", received.Sequence)
	}
}

// TestEventHandlerMultipleHandlers tests multiple handlers on same object
func TestEventHandlerMultipleHandlers(t *testing.T) {
	eh := NewEventHandler()

	count := 0
	handler1 := func(frame *MessageFrame) error {
		count++
		return nil
	}
	handler2 := func(frame *MessageFrame) error {
		count++
		return nil
	}

	eh.RegisterHandler(42, handler1)
	eh.RegisterHandler(42, handler2)

	frame := &MessageFrame{ObjectID: 42}
	eh.Dispatch(frame)

	if count != 2 {
		t.Errorf("Expected 2 handlers called, got %d", count)
	}
}

// TestPendingRequest tests request/response matching
func TestPendingRequest(t *testing.T) {
	eh := NewEventHandler()

	ctx := eh.CreatePendingRequest(1)
	if ctx == nil {
		t.Fatal("CreatePendingRequest returned nil")
	}

	// Resolve the request
	go func() {
		time.Sleep(10 * time.Millisecond)
		eh.ResolvePendingRequest(1, uint32(42))
	}()

	result, err := eh.WaitForRequest(ctx, 1*time.Second)
	if err != nil {
		t.Fatalf("WaitForRequest failed: %v", err)
	}

	if linkID, ok := result.(uint32); !ok || linkID != 42 {
		t.Errorf("Got wrong result: %v", result)
	}
}

// TestPendingRequestTimeout tests request timeout
func TestPendingRequestTimeout(t *testing.T) {
	eh := NewEventHandler()

	ctx := eh.CreatePendingRequest(1)
	if ctx == nil {
		t.Fatal("CreatePendingRequest returned nil")
	}

	// Don't resolve the request - it should timeout
	_, err := eh.WaitForRequest(ctx, 100*time.Millisecond)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

// TestLinkCreateRequestToPOD tests request to POD conversion
func TestLinkCreateRequestToPOD(t *testing.T) {
	req := &LinkCreateRequest{
		OutputPortID: 10,
		InputPortID:  20,
		Properties: map[string]string{
			"passive": "false",
		},
		Passive: false,
	}

	pod, err := req.ToPOD()
	if err != nil {
		t.Fatalf("ToPOD failed: %v", err)
	}

	if pod == nil {
		t.Fatal("ToPOD returned nil")
	}

	if len(pod.Values) < 4 {
		t.Errorf("POD object too small: %d values", len(pod.Values))
	}
}

// BenchmarkMessageMarshal benchmarks message marshalling
func BenchmarkMessageMarshal(b *testing.B) {
	frame := &MessageFrame{
		ObjectID: 1,
		MethodID: 4,
		Sequence: 1,
		PODData:  nil,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = frame.Marshal()
	}
}

// BenchmarkMessageUnmarshal benchmarks message unmarshalling
func BenchmarkMessageUnmarshal(b *testing.B) {
	frame := &MessageFrame{
		ObjectID: 1,
		MethodID: 4,
		Sequence: 1,
	}

	data, _ := frame.Marshal()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := &MessageFrame{}
		_ = f.Unmarshal(data)
	}
}

// BenchmarkEventDispatch benchmarks event dispatch
func BenchmarkEventDispatch(b *testing.B) {
	eh := NewEventHandler()
	handler := func(frame *MessageFrame) error {
		return nil
	}
	eh.RegisterHandler(42, handler)

	frame := &MessageFrame{
		ObjectID: 42,
		MethodID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eh.Dispatch(frame)
	}
}
