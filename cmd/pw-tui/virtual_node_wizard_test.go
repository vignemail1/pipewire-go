package main

import (
	"testing"

	core "github.com/vignemail1/pipewire-go/core"
)

func TestNewVirtualNodeWizardState(t *testing.T) {
	state := NewVirtualNodeWizardState()

	if state.Step != 1 {
		t.Errorf("NewVirtualNodeWizardState() Step = %d, want 1", state.Step)
	}

	if state.Type != core.VirtualNode_Sink {
		t.Errorf("NewVirtualNodeWizardState() Type = %v, want Sink", state.Type)
	}

	if state.Channels != 2 {
		t.Errorf("NewVirtualNodeWizardState() Channels = %d, want 2", state.Channels)
	}

	if state.SampleRate != 48000 {
		t.Errorf("NewVirtualNodeWizardState() SampleRate = %d, want 48000", state.SampleRate)
	}
}

func TestBuildVirtualNodeConfig(t *testing.T) {
	state := NewVirtualNodeWizardState()
	state.Name = "Test Node"

	config, err := state.BuildVirtualNodeConfig()
	if err != nil {
		t.Errorf("BuildVirtualNodeConfig() error = %v, want nil", err)
	}

	if config.Name != "Test Node" {
		t.Errorf("BuildVirtualNodeConfig() Name = %q, want %q", config.Name, "Test Node")
	}
}

func TestBuildVirtualNodeConfigValidation(t *testing.T) {
	state := NewVirtualNodeWizardState()
	// Don't set name

	_, err := state.BuildVirtualNodeConfig()
	if err == nil {
		t.Errorf("BuildVirtualNodeConfig() with empty name error = nil, want error")
	}
}

func TestApplyPreset(t *testing.T) {
	state := NewVirtualNodeWizardState()
	err := state.ApplyPreset("recording")

	if err != nil {
		t.Errorf("ApplyPreset(\"recording\") error = %v, want nil", err)
	}

	if state.Passive != true {
		t.Errorf("ApplyPreset(\"recording\") Passive = %v, want true", state.Passive)
	}
}

func TestNextStep(t *testing.T) {
	state := NewVirtualNodeWizardState()

	if state.Step != 1 {
		t.Fatalf("Initial Step = %d, want 1", state.Step)
	}

	state.NextStep()
	if state.Step != 2 {
		t.Errorf("After NextStep() Step = %d, want 2", state.Step)
	}

	state.NextStep()
	if state.Step != 3 {
		t.Errorf("After NextStep() Step = %d, want 3", state.Step)
	}

	state.NextStep()
	if state.Step != 3 {
		t.Errorf("After NextStep() at max Step = %d, want 3 (no advance)", state.Step)
	}
}

func TestPrevStep(t *testing.T) {
	state := NewVirtualNodeWizardState()
	state.Step = 3

	state.PrevStep()
	if state.Step != 2 {
		t.Errorf("After PrevStep() Step = %d, want 2", state.Step)
	}

	state.PrevStep()
	if state.Step != 1 {
		t.Errorf("After PrevStep() Step = %d, want 1", state.Step)
	}

	state.PrevStep()
	if state.Step != 1 {
		t.Errorf("After PrevStep() at min Step = %d, want 1 (no retreat)", state.Step)
	}
}

func TestSetErrorClearError(t *testing.T) {
	state := NewVirtualNodeWizardState()

	if state.Error != "" {
		t.Errorf("Initial Error = %q, want empty", state.Error)
	}

	state.SetError("Test error")
	if state.Error != "Test error" {
		t.Errorf("After SetError() Error = %q, want 'Test error'", state.Error)
	}

	state.ClearError()
	if state.Error != "" {
		t.Errorf("After ClearError() Error = %q, want empty", state.Error)
	}
}

func TestGetStep1Fields(t *testing.T) {
	state := NewVirtualNodeWizardState()
	fields := state.GetStep1Fields()

	expected := []string{"Node Type", "Name", "Description", "Preset Template"}
	if len(fields) != len(expected) {
		t.Errorf("GetStep1Fields() returned %d fields, want %d", len(fields), len(expected))
	}
}

func TestGetCurrentStepFields(t *testing.T) {
	state := NewVirtualNodeWizardState()

	// Step 1
	fields := state.GetCurrentStepFields()
	if len(fields) != 4 {
		t.Errorf("GetCurrentStepFields() at step 1 returned %d fields, want 4", len(fields))
	}

	// Step 2
	state.NextStep()
	fields = state.GetCurrentStepFields()
	if len(fields) != 8 {
		t.Errorf("GetCurrentStepFields() at step 2 returned %d fields, want 8", len(fields))
	}

	// Step 3
	state.NextStep()
	fields = state.GetCurrentStepFields()
	if len(fields) != 1 {
		t.Errorf("GetCurrentStepFields() at step 3 returned %d fields, want 1", len(fields))
	}
}

func TestGetSummary(t *testing.T) {
	state := NewVirtualNodeWizardState()
	state.Name = "Test Node"
	state.Desc = "Test Description"

	summary := state.GetSummary()
	if summary == "" {
		t.Error("GetSummary() returned empty string")
	}

	if !contains(summary, "Test Node") {
		t.Error("GetSummary() does not contain node name")
	}
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
