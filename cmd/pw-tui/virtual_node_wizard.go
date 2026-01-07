package main

// VirtualNodeWizardState holds the state of the TUI wizard for virtual node creation
type VirtualNodeWizardState struct {
	Step       int
	Type       string
	Name       string
	Desc       string
	PresetID   string
	Channels   int
	SampleRate int
	BitDepth   int
	Layout     string
	Passive    bool
	Virtual    bool
	Exclusive  bool
	DontReconnect bool
}

// NewVirtualNodeWizardState creates a new wizard state with defaults
func NewVirtualNodeWizardState() *VirtualNodeWizardState {
	return &VirtualNodeWizardState{
		Step:       1,
		Type:       "sink",
		Channels:   2,
		SampleRate: 48000,
		BitDepth:   32,
		Layout:     "FL FR",
		PresetID:   "default",
	}
}
