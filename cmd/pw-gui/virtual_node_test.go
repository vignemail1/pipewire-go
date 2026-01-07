package main

import (
	"testing"

	core "github.com/vignemail1/pipewire-go/core"
)

func TestGetVirtualNodePresets(t *testing.T) {
	presets := getVirtualNodePresets()
	if len(presets) != 6 {
		t.Errorf("getVirtualNodePresets() returned %d presets, want 6", len(presets))
	}

	// Check for expected presets
	expected := map[string]bool{
		"default": false,
		"null-sink": false,
		"null-source": false,
		"loopback": false,
		"recording": false,
		"monitoring": false,
	}

	for _, p := range presets {
		if _, ok := expected[p.ID]; ok {
			expected[p.ID] = true
		}
	}

	for preset, found := range expected {
		if !found {
			t.Errorf("Expected preset %q not found", preset)
		}
	}
}

func TestFormatPresetLabel(t *testing.T) {
	tests := []struct {
		name     string
		preset   VirtualNodePreset
		wantText bool
	}{
		{
			name: "with description",
			preset: VirtualNodePreset{
				ID:          "test",
				Name:        "Test Preset",
				Description: "A test preset",
			},
			wantText: true,
		},
		{
			name: "without description",
			preset: VirtualNodePreset{
				ID:   "test",
				Name: "Test Preset",
			},
			wantText: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPresetLabel(tt.preset)
			if tt.wantText {
				if got != "Test Preset - A test preset" {
					t.Errorf("formatPresetLabel() = %q, want 'Test Preset - A test preset'", got)
				}
			} else {
				if got != "Test Preset" {
					t.Errorf("formatPresetLabel() = %q, want 'Test Preset'", got)
				}
			}
		})
	}
}

func TestVirtualNodeDialogValidation(t *testing.T) {
	// Test that the dialog properly validates inputs
	// This would require mocking GTK4 which is complex
	// So we test the core logic instead

	client := &core.Client{
		connected:        true,
		nextVirtualNodeID: 0,
	}

	// Valid config
	config := core.VirtualNodeConfig{
		Name:       "Test Node",
		Type:       core.VirtualNode_Sink,
		Factory:    core.Factory_NullAudioSink,
		Channels:   2,
		SampleRate: 48000,
	}

	if err := config.Validate(); err != nil {
		t.Errorf("Valid config failed validation: %v", err)
	}

	// Invalid config (empty name)
	invalidConfig := core.VirtualNodeConfig{
		Name:       "",
		Type:       core.VirtualNode_Sink,
		Factory:    core.Factory_NullAudioSink,
		Channels:   2,
		SampleRate: 48000,
	}

	if err := invalidConfig.Validate(); err == nil {
		t.Error("Invalid config (empty name) should fail validation")
	}
}
