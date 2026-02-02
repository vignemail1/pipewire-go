package core

import (
	"fmt"
	"sync"
	"time"
)

// VirtualNodeType defines the type of virtual node to create
type VirtualNodeType string

const (
	// VirtualNode_Sink represents an audio output node
	VirtualNode_Sink VirtualNodeType = "sink"

	// VirtualNode_Source represents an audio input node
	VirtualNode_Source VirtualNodeType = "source"

	// VirtualNode_Filter represents an audio filter/processor node
	VirtualNode_Filter VirtualNodeType = "filter"

	// VirtualNode_Loopback represents a virtual loopback pair (source + sink)
	VirtualNode_Loopback VirtualNodeType = "loopback"
)

// VirtualNodeFactory specifies which spa-node-factory to use for creation
type VirtualNodeFactory string

const (
	// Factory_NullAudioSink creates a null audio sink that discards audio
	Factory_NullAudioSink VirtualNodeFactory = "support.null-audio-sink"

	// Factory_NullAudioSource creates a null audio source that generates silence
	Factory_NullAudioSource VirtualNodeFactory = "support.null-audio-source"

	// Factory_Adapter wraps other audio APIs (ALSA, etc.)
	Factory_Adapter VirtualNodeFactory = "adapter"

	// Factory_Loopback creates a virtual loopback pair
	Factory_Loopback VirtualNodeFactory = "libpipewire-module-loopback"

	// Factory_FilterChain creates a filter chain node
	Factory_FilterChain VirtualNodeFactory = "filter-chain"
)

// VirtualNodeConfig holds the configuration for creating a virtual node
type VirtualNodeConfig struct {
	// Basic properties
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=1000"`
	Type        VirtualNodeType `json:"type" validate:"required,oneof=sink source filter loopback"`
	Factory     VirtualNodeFactory `json:"factory" validate:"required"`

	// Audio properties
	Channels      uint32 `json:"channels" validate:"required,min=1,max=8"`
	SampleRate    uint32 `json:"sample_rate" validate:"required,oneof=44100 48000 88200 96000 176400 192000"`
	BitDepth      uint32 `json:"bit_depth" validate:"oneof=16 24 32"`
	ChannelLayout string `json:"channel_layout" validate:"max=256"` // e.g. "FL FR" for stereo

	// Behavior properties
	Passive         bool `json:"passive"` // Don't hold graph playing
	Virtual         bool `json:"virtual"` // Mark as virtual
	Exclusive       bool `json:"exclusive"` // Exclusive access
	DontReconnect   bool `json:"dont_reconnect"` // Don't auto-reconnect

	// Advanced properties
	Latency  string `json:"latency" validate:"max=64"` // e.g. "1024/48000"
	Priority int    `json:"priority"` // Priority level

	// Custom properties
	CustomProps map[string]interface{} `json:"custom_props"`
}

// VirtualNode represents a virtual node in the PipeWire graph
type VirtualNode struct {
	// Immutable properties
	ID        uint32
	Config    VirtualNodeConfig
	CreatedAt time.Time

	// Mutable properties
	Ports     []*Port
	UpdatedAt time.Time
	Properties map[string]interface{}

	// Internal state
	mu        sync.RWMutex
	client    *Client
}

// VirtualNodeError represents an error during virtual node operations
type VirtualNodeError struct {
	Reason string
	Details string
}

func (e *VirtualNodeError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("virtual node error: %s (%s)", e.Reason, e.Details)
	}
	return fmt.Sprintf("virtual node error: %s", e.Reason)
}

// VirtualNodeNotFoundError represents when a virtual node is not found
type VirtualNodeNotFoundError struct {
	NodeID uint32
}

func (e *VirtualNodeNotFoundError) Error() string {
	return fmt.Sprintf("virtual node not found: ID=%d", e.NodeID)
}

// VirtualNodePropertyError represents an error modifying a property
type VirtualNodePropertyError struct {
	Property string
	Value    interface{}
	Message  string
}

func (e *VirtualNodePropertyError) Error() string {
	return fmt.Sprintf("virtual node property error: %s=%v: %s", e.Property, e.Value, e.Message)
}

// Preset Configurations

// GetVirtualNodePreset returns a preconfigured VirtualNodeConfig for common use cases
func GetVirtualNodePreset(preset string) VirtualNodeConfig {
	presets := map[string]VirtualNodeConfig{
		"default": {
			Name:          "Default Sink",
			Description:   "Default stereo audio sink",
			Type:          VirtualNode_Sink,
			Factory:       Factory_NullAudioSink,
			Channels:      2,
			SampleRate:    48000,
			BitDepth:      32,
			ChannelLayout: "FL FR",
			Passive:       false,
		},
		"null-sink": {
			Name:          "Null Sink",
			Description:   "Null audio sink - discards all audio",
			Type:          VirtualNode_Sink,
			Factory:       Factory_NullAudioSink,
			Channels:      2,
			SampleRate:    48000,
			BitDepth:      32,
			ChannelLayout: "FL FR",
			Passive:       true,
		},
		"null-source": {
			Name:          "Null Source",
			Description:   "Null audio source - generates silence",
			Type:          VirtualNode_Source,
			Factory:       Factory_NullAudioSource,
			Channels:      2,
			SampleRate:    48000,
			BitDepth:      32,
			ChannelLayout: "FL FR",
			Passive:       false,
		},
		"loopback": {
			Name:          "Virtual Loopback",
			Description:   "Virtual audio loopback pair",
			Type:          VirtualNode_Loopback,
			Factory:       Factory_Loopback,
			Channels:      2,
			SampleRate:    48000,
			BitDepth:      32,
			ChannelLayout: "FL FR",
			Passive:       false,
		},
		"recording": {
			Name:          "Recording",
			Description:   "Virtual recording sink (passive)",
			Type:          VirtualNode_Sink,
			Factory:       Factory_NullAudioSink,
			Channels:      2,
			SampleRate:    48000,
			BitDepth:      32,
			ChannelLayout: "FL FR",
			Passive:       true,
		},
		"monitoring": {
			Name:          "Monitoring",
			Description:   "Virtual monitoring source",
			Type:          VirtualNode_Source,
			Factory:       Factory_NullAudioSource,
			Channels:      2,
			SampleRate:    48000,
			BitDepth:      32,
			ChannelLayout: "FL FR",
			Passive:       false,
		},
	}

	if config, exists := presets[preset]; exists {
		return config
	}

	// Return default if preset not found
	return presets["default"]
}

// GetVirtualNodePresetNames returns the list of available preset names
func GetVirtualNodePresetNames() []string {
	return []string{"default", "null-sink", "null-source", "loopback", "recording", "monitoring"}
}

// VirtualNodeMethods

// UpdateProperty updates a property of the virtual node
func (v *VirtualNode) UpdateProperty(key string, value interface{}) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.Properties == nil {
		v.Properties = make(map[string]interface{})
	}

	v.Properties[key] = value
	v.UpdatedAt = time.Now()

	return nil
}

// GetProperty retrieves a property of the virtual node
func (v *VirtualNode) GetProperty(key string) (interface{}, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.Properties == nil {
		return nil, fmt.Errorf("no properties set")
	}

	value, exists := v.Properties[key]
	if !exists {
		return nil, fmt.Errorf("property not found: %s", key)
	}

	return value, nil
}

// GetPorts returns all ports belonging to this virtual node
func (v *VirtualNode) GetPorts() []*Port {
	v.mu.RLock()
	defer v.mu.RUnlock()

	ports := make([]*Port, len(v.Ports))
	copy(ports, v.Ports)
	return ports
}

// AddPort adds a port to this virtual node
func (v *VirtualNode) AddPort(port *Port) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.Ports = append(v.Ports, port)
}

// Refresh syncs the node state with the daemon
func (v *VirtualNode) Refresh() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.client == nil {
		return fmt.Errorf("virtual node not connected to client")
	}

	// TODO: Implement actual refresh from daemon
	v.UpdatedAt = time.Now()
	return nil
}

// Delete removes this virtual node from the graph
func (v *VirtualNode) Delete() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.client == nil {
		return fmt.Errorf("virtual node not connected to client")
	}

	// TODO: Implement actual deletion via client
	return nil
}

// Validate validates the virtual node configuration
func (config *VirtualNodeConfig) Validate() error {
	if config.Name == "" {
		return &VirtualNodePropertyError{
			Property: "name",
			Value:    config.Name,
			Message:  "name cannot be empty",
		}
	}

	if len(config.Name) > 255 {
		return &VirtualNodePropertyError{
			Property: "name",
			Value:    config.Name,
			Message:  "name cannot exceed 255 characters",
		}
	}

	if config.Channels < 1 || config.Channels > 8 {
		return &VirtualNodePropertyError{
			Property: "channels",
			Value:    config.Channels,
			Message:  "channels must be between 1 and 8",
		}
	}

	validRates := map[uint32]bool{
		44100:  true,
		48000:  true,
		88200:  true,
		96000:  true,
		176400: true,
		192000: true,
	}
	if !validRates[config.SampleRate] {
		return &VirtualNodePropertyError{
			Property: "sample_rate",
			Value:    config.SampleRate,
			Message:  "unsupported sample rate",
		}
	}

	if config.BitDepth != 0 {
		validBitDepths := map[uint32]bool{
			16: true,
			24: true,
			32: true,
		}
		if !validBitDepths[config.BitDepth] {
			return &VirtualNodePropertyError{
				Property: "bit_depth",
				Value:    config.BitDepth,
				Message:  "unsupported bit depth (must be 16, 24, or 32)",
			}
		}
	}

	return nil
}
