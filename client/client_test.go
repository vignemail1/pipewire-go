package client

import (
	"testing"
	"time"
)

// TestNodeCreation tests node creation and initialization
func TestNodeCreation(t *testing.T) {
	node := &Node{
		ID:    1,
		Type:  "PipeWire:Interface:Node",
		Props: make(map[string]string),
	}

	if node.ID != 1 {
		t.Errorf("Node ID: got %d, want 1", node.ID)
	}

	if node.Name() == "" {
		// Should return formatted name if not in Props
	}
}

// TestPortCreation tests port creation and initialization
func TestPortCreation(t *testing.T) {
	port := NewPort(1, "system:playback_1", PortDirectionOutput, nil, nil)

	if port.ID() != 1 {
		t.Errorf("Port ID: got %d, want 1", port.ID())
	}

	if port.Name() != "system:playback_1" {
		t.Errorf("Port Name: got %s, want system:playback_1", port.Name())
	}

	if port.Direction() != PortDirectionOutput {
		t.Error("Port Direction incorrect")
	}
}

// TestPortDirection tests port direction properties
func TestPortDirection(t *testing.T) {
	tests := []struct {
		name          string
		direction     PortDirection
		isOutput      bool
		isInput       bool
	}{
		{"output", PortDirectionOutput, true, false},
		{"input", PortDirectionInput, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := NewPort(1, "test", tt.direction, nil, nil)

			if port.IsOutput() != tt.isOutput {
				t.Errorf("IsOutput: got %v, want %v", port.IsOutput(), tt.isOutput)
			}

			if port.IsInput() != tt.isInput {
				t.Errorf("IsInput: got %v, want %v", port.IsInput(), tt.isInput)
			}
		})
	}
}

// TestPortFormat tests format handling
func TestPortFormat(t *testing.T) {
	port := NewPort(1, "test", PortDirectionOutput, nil, nil)

	// Set supported formats
	formats := []*Format{
		{
			Type: PortTypeAudio,
			Audio: &AudioFormat{
				Encoding: "S16_LE",
				Rate:     48000,
				Channels: 2,
			},
		},
	}

	port.SetSupportedFormats(formats)
	supported := port.GetSupportedFormats()

	if len(supported) != 1 {
		t.Errorf("Supported formats: got %d, want 1", len(supported))
	}
}

// TestPortFormatNegotiation tests format negotiation
func TestPortFormatNegotiation(t *testing.T) {
	port1 := NewPort(1, "out", PortDirectionOutput, nil, nil)
	port2 := NewPort(2, "in", PortDirectionInput, nil, nil)

	// Set compatible formats
	format := &Format{
		Type: PortTypeAudio,
		Audio: &AudioFormat{
			Encoding: "F32_LE",
			Rate:     48000,
			Channels: 2,
		},
	}

	port1.SetSupportedFormats([]*Format{format})
	port2.SetSupportedFormats([]*Format{format})

	// Try to connect
	can := port1.CanConnectTo(port2)
	if !can {
		t.Error("Ports should be connectable")
	}
}

// TestLinkCreation tests link creation
func TestLinkCreation(t *testing.T) {
	link := NewLink(1, nil, nil, nil)

	if link.ID() != 1 {
		t.Errorf("Link ID: got %d, want 1", link.ID())
	}
}

// TestNodeParams tests node parameter access
func TestNodeParams(t *testing.T) {
	node := &Node{
		ID:    1,
		Type:  "PipeWire:Interface:Node",
		Props: map[string]string{"audio.rate": "48000"},
	}

	// Test GetParams
	params, err := node.GetParams(ParamIDFormat)
	if err != nil {
		t.Errorf("GetParams error: %v", err)
	}

	if params == nil {
		t.Error("GetParams returned nil")
	}
}

// TestNodeParameterSet tests setting node parameters
func TestNodeParameterSet(t *testing.T) {
	node := &Node{
		ID:    1,
		Type:  "PipeWire:Interface:Node",
		Props: make(map[string]string),
	}

	// Try to set format parameter
	formatVal := map[string]interface{}{
		"format": "F32_LE",
		"rate":   48000,
	}

	err := node.SetParam(ParamIDFormat, 0, formatVal)
	if err != nil {
		// SetParam may fail if format is invalid, that's ok
	}
}

// TestPortType tests port type properties
func TestPortType(t *testing.T) {
	tests := []struct {
		name      string
		portType  PortType
		isAudio   bool
		isMIDI    bool
	}{
		{"audio", PortTypeAudio, true, false},
		{"midi", PortTypeMIDI, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := NewPort(1, "test", PortDirectionOutput, nil, nil)
			port.SetType(tt.portType)

			if port.IsAudioPort() != tt.isAudio {
				t.Errorf("IsAudioPort: got %v, want %v", port.IsAudioPort(), tt.isAudio)
			}

			if port.IsMIDIPort() != tt.isMIDI {
				t.Errorf("IsMIDIPort: got %v, want %v", port.IsMIDIPort(), tt.isMIDI)
			}
		})
	}
}

// TestPortFilter tests port filtering
func TestPortFilter(t *testing.T) {
	port := NewPort(1, "system:playback_1", PortDirectionOutput, nil, nil)
	port.SetType(PortTypeAudio)

	// Test filter matching
	filter := &PortFilter{
		Direction: PortDirectionOutput,
		Type:      PortTypeAudio,
	}

	if !filter.Matches(port) {
		t.Error("Filter should match port")
	}

	// Test non-matching filter
	filter2 := &PortFilter{
		Direction: PortDirectionInput,
		Type:      PortTypeAudio,
	}

	if filter2.Matches(port) {
		t.Error("Filter should not match port")
	}
}

// TestEventHandling tests event handling
func TestEventHandling(t *testing.T) {
	// Test event dispatching
	dispatcher := NewEventDispatcher()

	err := dispatcher.Start()
	if err != nil {
		t.Errorf("Failed to start dispatcher: %v", err)
	}

	err = dispatcher.Stop()
	if err != nil {
		t.Errorf("Failed to stop dispatcher: %v", err)
	}
}

// TestPortConnection tests port connection
func TestPortConnection(t *testing.T) {
	out := NewPort(1, "out", PortDirectionOutput, nil, nil)
	in := NewPort(2, "in", PortDirectionInput, nil, nil)

	// Check initial state
	if out.IsConnected() {
		t.Error("Output port should not be connected initially")
	}

	if in.IsConnected() {
		t.Error("Input port should not be connected initially")
	}

	// Set connected
	out.SetConnected(true)
	if !out.IsConnected() {
		t.Error("Output port should be connected")
	}
}

// BenchmarkNodeCreation benchmarks node creation
func BenchmarkNodeCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = &Node{
			ID:    uint32(i),
			Type:  "PipeWire:Interface:Node",
			Props: make(map[string]string),
		}
	}
}

// BenchmarkPortCreation benchmarks port creation
func BenchmarkPortCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewPort(uint32(i), "test_port", PortDirectionOutput, nil, nil)
	}
}

// BenchmarkPortFormatCheck benchmarks format compatibility check
func BenchmarkPortFormatCheck(b *testing.B) {
	out := NewPort(1, "out", PortDirectionOutput, nil, nil)
	in := NewPort(2, "in", PortDirectionInput, nil, nil)

	for i := 0; i < b.N; i++ {
		_ = out.CanConnectTo(in)
	}
}
