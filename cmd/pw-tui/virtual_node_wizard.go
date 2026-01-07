package main

import (
	"fmt"

	core "github.com/vignemail1/pipewire-go/core"
)

// VirtualNodeWizardState holds the state of the TUI wizard for virtual node creation
type VirtualNodeWizardState struct {
	Step           int
	Type           core.VirtualNodeType
	Name           string
	Desc           string
	PresetID       string
	Channels       int
	SampleRate     int
	BitDepth       int
	Layout         string
	Passive        bool
	Virtual        bool
	Exclusive      bool
	DontReconnect  bool
	Error          string
	CurrentFocusIdx int
}

// NewVirtualNodeWizardState creates a new wizard state with defaults
func NewVirtualNodeWizardState() *VirtualNodeWizardState {
	return &VirtualNodeWizardState{
		Step:           1,
		Type:           core.VirtualNode_Sink,
		Channels:       2,
		SampleRate:     48000,
		BitDepth:       32,
		Layout:         "FL,FR",
		PresetID:       "default",
		CurrentFocusIdx: 0,
	}
}

// BuildVirtualNodeConfig creates a VirtualNodeConfig from the wizard state
func (s *VirtualNodeWizardState) BuildVirtualNodeConfig() (core.VirtualNodeConfig, error) {
	if s.Name == "" {
		return core.VirtualNodeConfig{}, fmt.Errorf("node name cannot be empty")
	}

	config := core.VirtualNodeConfig{
		Name:          s.Name,
		Description:   s.Desc,
		Type:          s.Type,
		Factory:       s.getFactory(),
		Channels:      s.Channels,
		SampleRate:    s.SampleRate,
		BitDepth:      s.BitDepth,
		Layout:        s.Layout,
		Passive:       s.Passive,
		Virtual:       s.Virtual,
		Exclusive:     s.Exclusive,
		DontReconnect: s.DontReconnect,
	}

	if err := config.Validate(); err != nil {
		return core.VirtualNodeConfig{}, err
	}

	return config, nil
}

func (s *VirtualNodeWizardState) getFactory() core.VirtualNodeFactory {
	switch s.Type {
	case core.VirtualNode_Source:
		return core.Factory_NullAudioSource
	case core.VirtualNode_Loopback:
		return core.Factory_Loopback
	case core.VirtualNode_Filter:
		return core.Factory_FilterChain
	default:
		return core.Factory_NullAudioSink
	}
}

// ApplyPreset applies preset values to the wizard state
func (s *VirtualNodeWizardState) ApplyPreset(presetID string) error {
	config := core.GetVirtualNodePreset(presetID)

	s.PresetID = presetID
	s.Type = config.Type
	s.Channels = config.Channels
	s.SampleRate = config.SampleRate
	s.BitDepth = config.BitDepth
	s.Layout = config.Layout
	s.Passive = config.Passive
	s.Virtual = config.Virtual
	s.Exclusive = config.Exclusive
	s.DontReconnect = config.DontReconnect

	return nil
}

// NextStep advances to the next wizard step
func (s *VirtualNodeWizardState) NextStep() {
	if s.Step < 3 {
		s.Step++
		s.CurrentFocusIdx = 0
		s.Error = ""
	}
}

// PrevStep goes back to the previous wizard step
func (s *VirtualNodeWizardState) PrevStep() {
	if s.Step > 1 {
		s.Step--
		s.CurrentFocusIdx = 0
		s.Error = ""
	}
}

// SetError sets an error message on the wizard state
func (s *VirtualNodeWizardState) SetError(msg string) {
	s.Error = msg
}

// ClearError clears any error message
func (s *VirtualNodeWizardState) ClearError() {
	s.Error = ""
}

// GetStep1Fields returns the fields for step 1 (type and naming)
func (s *VirtualNodeWizardState) GetStep1Fields() []string {
	return []string{"Node Type", "Name", "Description", "Preset Template"}
}

// GetStep2Fields returns the fields for step 2 (audio configuration)
func (s *VirtualNodeWizardState) GetStep2Fields() []string {
	return []string{"Channels", "Sample Rate", "Bit Depth", "Channel Layout", "Passive", "Virtual", "Exclusive", "Don't Reconnect"}
}

// GetStep3Fields returns the fields for step 3 (review)
func (s *VirtualNodeWizardState) GetStep3Fields() []string {
	return []string{"Review Configuration"}
}

// GetCurrentStepFields returns the fields for the current step
func (s *VirtualNodeWizardState) GetCurrentStepFields() []string {
	switch s.Step {
	case 2:
		return s.GetStep2Fields()
	case 3:
		return s.GetStep3Fields()
	default:
		return s.GetStep1Fields()
	}
}

// GetSummary returns a string summary of the configuration
func (s *VirtualNodeWizardState) GetSummary() string {
	return fmt.Sprintf(
		`Virtual Node Configuration
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Name:              %s
Description:      %s
Type:             %v
Channels:         %d
Sample Rate:      %d Hz
Bit Depth:        %d-bit
Channel Layout:   %s
Passive:          %v
Virtual:          %v
Exclusive:        %v
Don't Reconnect:  %v
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`,
		s.Name,
		s.Desc,
		s.Type,
		s.Channels,
		s.SampleRate,
		s.BitDepth,
		s.Layout,
		s.Passive,
		s.Virtual,
		s.Exclusive,
		s.DontReconnect,
	)
}
