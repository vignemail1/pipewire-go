// Package core - Tests for protocol state machine and message buffer
// core/connection_test.go
// Issue #6 - Event Loop & Automatic Dispatch

package core

import (
	"testing"
)

// ============================================================================
// PROTOCOL STATE MACHINE TESTS
// ============================================================================

func TestNewProtocolStateMachine(t *testing.T) {
	machine := NewProtocolStateMachine()
	if machine == nil {
		t.Fatal("NewProtocolStateMachine returned nil")
	}
	if machine.GetState() != StateDisconnected {
		t.Errorf("Initial state should be Disconnected, got %s", machine.GetState())
	}
}

func TestStateTransitionValid(t *testing.T) {
	machine := NewProtocolStateMachine()

	tests := []struct {
		from State
		to   State
		ok   bool
	}{
		{StateDisconnected, StateConnected, true},
		{StateConnected, StateHelloSent, true},
		{StateHelloSent, StateHelloReceived, true},
		{StateHelloReceived, StateReady, true},
		{StateReady, StateDisconnected, true},
		{StateDisconnected, StateReady, false},
		{StateConnected, StateReady, false},
		{StateReady, StateConnected, false},
	}

	for _, test := range tests {
		machine.Reset()
		machine.state = test.from

		err := machine.TransitionTo(test.to)
		if test.ok && err != nil {
			t.Errorf("Transition %s → %s should succeed, got error: %v",
				test.from, test.to, err)
		}
		if !test.ok && err == nil {
			t.Errorf("Transition %s → %s should fail, but succeeded",
				test.from, test.to)
		}
	}
}

func TestIsReady(t *testing.T) {
	machine := NewProtocolStateMachine()

	if machine.IsReady() {
		t.Error("NewMachine should not be ready")
	}

	machine.state = StateReady
	if !machine.IsReady() {
		t.Error("Machine in Ready state should be ready")
	}
}

func TestCanSendMessage(t *testing.T) {
	machine := NewProtocolStateMachine()

	if machine.CanSendMessage() {
		t.Error("NewMachine should not be able to send")
	}

	machine.state = StateReady
	if !machine.CanSendMessage() {
		t.Error("Machine in Ready state should be able to send")
	}
}

func TestCanReceiveMessage(t *testing.T) {
	machine := NewProtocolStateMachine()

	if machine.CanReceiveMessage() {
		t.Error("NewMachine should not be able to receive")
	}

	machine.state = StateHelloReceived
	if !machine.CanReceiveMessage() {
		t.Error("Machine in HelloReceived state should be able to receive")
	}

	machine.state = StateReady
	if !machine.CanReceiveMessage() {
		t.Error("Machine in Ready state should be able to receive")
	}
}

func TestVersionNegotiation(t *testing.T) {
	machine := NewProtocolStateMachine()

	machine.SetVersion(3, 0)
	version := machine.GetVersion()

	if version.Major != 3 || version.Minor != 0 {
		t.Errorf("Version should be 3.0, got %s", version)
	}
}

func TestServerCapabilities(t *testing.T) {
	machine := NewProtocolStateMachine()

	caps := []string{"core-api", "metadata", "links"}
	machine.SetServerCapabilities(caps)

	retrieved := machine.GetServerCapabilities()
	if len(retrieved) != len(caps) {
		t.Errorf("Should have %d capabilities, got %d", len(caps), len(retrieved))
	}

	if !machine.HasCapability("core-api") {
		t.Error("Should have core-api capability")
	}

	if machine.HasCapability("nonexistent") {
		t.Error("Should not have nonexistent capability")
	}
}

func TestReset(t *testing.T) {
	machine := NewProtocolStateMachine()
	machine.state = StateReady
	machine.SetVersion(4, 0)

	machine.Reset()

	if machine.GetState() != StateDisconnected {
		t.Errorf("After reset, state should be Disconnected, got %s", machine.GetState())
	}
}

// ============================================================================
// MESSAGE BUFFER TESTS
// ============================================================================

func TestNewMessageBuffer(t *testing.T) {
	buffer := NewMessageBuffer(1024)
	if buffer == nil {
		t.Fatal("NewMessageBuffer returned nil")
	}
	if buffer.Size() != 0 {
		t.Error("New buffer should be empty")
	}
}

func TestAppendBytes(t *testing.T) {
	buffer := NewMessageBuffer(1024)

	data := []byte{1, 2, 3, 4, 5}
	err := buffer.Append(data)
	if err != nil {
		t.Errorf("Append failed: %v", err)
	}

	if buffer.Size() != len(data) {
		t.Errorf("Buffer size should be %d, got %d", len(data), buffer.Size())
	}
}

func TestBufferOverflow(t *testing.T) {
	buffer := NewMessageBuffer(10)

	data := make([]byte, 20)
	err := buffer.Append(data)
	if err == nil {
		t.Error("Should have error on buffer overflow")
	}
}

func TestPeekData(t *testing.T) {
	buffer := NewMessageBuffer(1024)
	data := []byte{1, 2, 3, 4, 5}
	buffer.Append(data)

	peeked, err := buffer.Peek(3)
	if err != nil {
		t.Errorf("Peek failed: %v", err)
	}

	if len(peeked) != 3 {
		t.Errorf("Peeked length should be 3, got %d", len(peeked))
	}

	// Buffer should still have all data after peek
	if buffer.Size() != len(data) {
		t.Error("Buffer size changed after peek")
	}
}

func TestConsume(t *testing.T) {
	buffer := NewMessageBuffer(1024)
	data := []byte{1, 2, 3, 4, 5}
	buffer.Append(data)

	err := buffer.Consume(2)
	if err != nil {
		t.Errorf("Consume failed: %v", err)
	}

	if buffer.Size() != 3 {
		t.Errorf("After consuming 2, buffer size should be 3, got %d", buffer.Size())
	}
}

func TestIsEmpty(t *testing.T) {
	buffer := NewMessageBuffer(1024)

	if !buffer.IsEmpty() {
		t.Error("New buffer should be empty")
	}

	buffer.Append([]byte{1})
	if buffer.IsEmpty() {
		t.Error("Buffer with data should not be empty")
	}

	buffer.Reset()
	if !buffer.IsEmpty() {
		t.Error("Buffer after reset should be empty")
	}
}

func TestCopyBuffer(t *testing.T) {
	buffer := NewMessageBuffer(1024)
	data := []byte{1, 2, 3, 4, 5}
	buffer.Append(data)

	copy := buffer.Copy()

	if copy.Size() != buffer.Size() {
		t.Error("Copied buffer should have same size")
	}

	// Modifying copy should not affect original
	copy.Reset()
	if buffer.Size() != len(data) {
		t.Error("Original buffer was affected by copy modification")
	}
}

func TestMultipleAppends(t *testing.T) {
	buffer := NewMessageBuffer(1024)

	for i := 0; i < 10; i++ {
		data := []byte{byte(i)}
		if err := buffer.Append(data); err != nil {
			t.Errorf("Append %d failed: %v", i, err)
		}
	}

	if buffer.Size() != 10 {
		t.Errorf("Buffer size should be 10, got %d", buffer.Size())
	}
}

func TestStateString(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateDisconnected, "Disconnected"},
		{StateConnected, "Connected"},
		{StateHelloSent, "HelloSent"},
		{StateHelloReceived, "HelloReceived"},
		{StateReady, "Ready"},
		{StateError, "Error"},
	}

	for _, test := range tests {
		if test.state.String() != test.want {
			t.Errorf("State %v String() = %q, want %q",
				test.state, test.state.String(), test.want)
		}
	}
}

func TestVersionString(t *testing.T) {
	version := ProtocolVersion{Major: 3, Minor: 0}
	expected := "3.0"
	if version.String() != expected {
		t.Errorf("Version.String() = %q, want %q", version.String(), expected)
	}
}

func BenchmarkStateTransition(b *testing.B) {
	machine := NewProtocolStateMachine()
	machine.state = StateConnected

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		machine.state = StateHelloSent
		machine.state = StateConnected
	}
}

func BenchmarkAppendToBuffer(b *testing.B) {
	buffer := NewMessageBuffer(1024 * 1024)
	data := []byte{1, 2, 3, 4, 5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer.Append(data)
		if buffer.Size() > 100000 {
			buffer.Reset()
		}
	}
}

func BenchmarkBufferPeek(b *testing.B) {
	buffer := NewMessageBuffer(1024)
	buffer.Append(make([]byte, 512))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer.Peek(100)
	}
}
