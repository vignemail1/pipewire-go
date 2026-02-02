package core

import (
	"testing"
)

func TestGetVirtualNodePreset(t *testing.T) {
	tests := []struct {
		name   string
		preset string
		want   string
	}{
		{"default", "default", "Default Sink"},
		{"null-sink", "null-sink", "Null Sink"},
		{"null-source", "null-source", "Null Source"},
		{"loopback", "loopback", "Virtual Loopback"},
		{"recording", "recording", "Recording"},
		{"monitoring", "monitoring", "Monitoring"},
		{"unknown", "unknown", "Default Sink"}, // Should return default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GetVirtualNodePreset(tt.preset)
			if config.Name != tt.want {
				t.Errorf("GetVirtualNodePreset(%q) name = %q, want %q", tt.preset, config.Name, tt.want)
			}
		})
	}
}

func TestGetVirtualNodePresetNames(t *testing.T) {
	names := GetVirtualNodePresetNames()
	if len(names) != 6 {
		t.Errorf("GetVirtualNodePresetNames() returned %d presets, want 6", len(names))
	}

	expectedPresets := map[string]bool{
		"default":     true,
		"null-sink":   true,
		"null-source": true,
		"loopback":    true,
		"recording":   true,
		"monitoring":  true,
	}

	for _, name := range names {
		if !expectedPresets[name] {
			t.Errorf("GetVirtualNodePresetNames() returned unexpected preset: %s", name)
		}
	}
}

func TestVirtualNodeConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  VirtualNodeConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: VirtualNodeConfig{
				Name:       "Test Node",
				Type:       VirtualNode_Sink,
				Factory:    Factory_NullAudioSink,
				Channels:   2,
				SampleRate: 48000,
				BitDepth:   32,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			config: VirtualNodeConfig{
				Name:       "",
				Type:       VirtualNode_Sink,
				Factory:    Factory_NullAudioSink,
				Channels:   2,
				SampleRate: 48000,
			},
			wantErr: true,
		},
		{
			name: "invalid channels (0)",
			config: VirtualNodeConfig{
				Name:       "Test Node",
				Type:       VirtualNode_Sink,
				Factory:    Factory_NullAudioSink,
				Channels:   0,
				SampleRate: 48000,
			},
			wantErr: true,
		},
		{
			name: "invalid channels (9)",
			config: VirtualNodeConfig{
				Name:       "Test Node",
				Type:       VirtualNode_Sink,
				Factory:    Factory_NullAudioSink,
				Channels:   9,
				SampleRate: 48000,
			},
			wantErr: true,
		},
		{
			name: "invalid sample rate",
			config: VirtualNodeConfig{
				Name:       "Test Node",
				Type:       VirtualNode_Sink,
				Factory:    Factory_NullAudioSink,
				Channels:   2,
				SampleRate: 50000,
			},
			wantErr: true,
		},
		{
			name: "invalid bit depth",
			config: VirtualNodeConfig{
				Name:       "Test Node",
				Type:       VirtualNode_Sink,
				Factory:    Factory_NullAudioSink,
				Channels:   2,
				SampleRate: 48000,
				BitDepth:   20,
			},
			wantErr: true,
		},
		{
			name: "valid bit depth 24",
			config: VirtualNodeConfig{
				Name:       "Test Node",
				Type:       VirtualNode_Sink,
				Factory:    Factory_NullAudioSink,
				Channels:   2,
				SampleRate: 48000,
				BitDepth:   24,
			},
			wantErr: false,
		},
		{
			name: "max channels (8)",
			config: VirtualNodeConfig{
				Name:       "Test Node",
				Type:       VirtualNode_Sink,
				Factory:    Factory_NullAudioSink,
				Channels:   8,
				SampleRate: 48000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("VirtualNodeConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVirtualNodeUpdateProperty(t *testing.T) {
	node := &VirtualNode{
		ID:         1,
		Properties: make(map[string]interface{}),
	}

	// Update property
	err := node.UpdateProperty("test.key", "test.value")
	if err != nil {
		t.Errorf("UpdateProperty() error = %v, want nil", err)
	}

	// Verify property was updated
	if node.Properties["test.key"] != "test.value" {
		t.Errorf("UpdateProperty() did not set property correctly")
	}
}

func TestVirtualNodeGetProperty(t *testing.T) {
	node := &VirtualNode{
		ID:         1,
		Properties: make(map[string]interface{}),
	}

	node.Properties["test.key"] = "test.value"

	// Get existing property
	value, err := node.GetProperty("test.key")
	if err != nil {
		t.Errorf("GetProperty() error = %v, want nil", err)
	}
	if value != "test.value" {
		t.Errorf("GetProperty() returned %v, want 'test.value'", value)
	}

	// Get non-existent property
	_, err = node.GetProperty("nonexistent.key")
	if err == nil {
		t.Errorf("GetProperty() error = nil, want error for non-existent property")
	}
}

func TestVirtualNodeGetPorts(t *testing.T) {
	node := &VirtualNode{
		ID:    1,
		Ports: make([]*Port, 0),
	}

	// Empty ports
	ports := node.GetPorts()
	if len(ports) != 0 {
		t.Errorf("GetPorts() returned %d ports, want 0", len(ports))
	}

	// Add port
	port := &Port{ID: 1}
	node.AddPort(port)

	ports = node.GetPorts()
	if len(ports) != 1 {
		t.Errorf("GetPorts() returned %d ports, want 1", len(ports))
	}
	if ports[0].ID != 1 {
		t.Errorf("GetPorts() returned port with ID %d, want 1", ports[0].ID)
	}
}

func TestVirtualNodeError(t *testing.T) {
	err := &VirtualNodeError{Reason: "test reason", Details: "test details"}
	wantMsg := "virtual node error: test reason (test details)"
	if err.Error() != wantMsg {
		t.Errorf("VirtualNodeError.Error() = %q, want %q", err.Error(), wantMsg)
	}

	err2 := &VirtualNodeError{Reason: "test reason"}
	wantMsg2 := "virtual node error: test reason"
	if err2.Error() != wantMsg2 {
		t.Errorf("VirtualNodeError.Error() = %q, want %q", err2.Error(), wantMsg2)
	}
}

func TestVirtualNodeNotFoundError(t *testing.T) {
	err := &VirtualNodeNotFoundError{NodeID: 42}
	wantMsg := "virtual node not found: ID=42"
	if err.Error() != wantMsg {
		t.Errorf("VirtualNodeNotFoundError.Error() = %q, want %q", err.Error(), wantMsg)
	}
}

func TestVirtualNodePropertyError(t *testing.T) {
	err := &VirtualNodePropertyError{
		Property: "channels",
		Value:    9,
		Message:  "channels must be between 1 and 8",
	}
	wantMsg := "virtual node property error: channels=9: channels must be between 1 and 8"
	if err.Error() != wantMsg {
		t.Errorf("VirtualNodePropertyError.Error() = %q, want %q", err.Error(), wantMsg)
	}
}
