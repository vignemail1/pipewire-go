// Package client - types.go
// Common types, enums, and constants for PipeWire client API

package client

import "fmt"

// NodeState represents the current state of a node
type NodeState string

const (
	NodeStateError     NodeState = "error"
	NodeStateSuspended NodeState = "suspended"
	NodeStateIdle      NodeState = "idle"
	NodeStateRunning   NodeState = "running"
)

// NodeDirection indicates whether a node is for playback, capture, or both
type NodeDirection string

const (
	NodeDirectionPlayback NodeDirection = "playback"
	NodeDirectionCapture  NodeDirection = "capture"
	NodeDirectionDuplex   NodeDirection = "duplex"
)

// PortDirection indicates whether a port is input or output
type PortDirection uint32

const (
	PortDirectionInput PortDirection = iota
	PortDirectionOutput
)

func (pd PortDirection) String() string {
	switch pd {
	case PortDirectionInput:
		return "Input"
	case PortDirectionOutput:
		return "Output"
	default:
		return "Unknown"
	}
}

// PortType indicates the type of data flowing through a port
type PortType uint32

const (
	PortTypeAudio PortType = iota
	PortTypeMidi
	PortTypeVideo
	PortTypeControl
)

func (pt PortType) String() string {
	switch pt {
	case PortTypeAudio:
		return "Audio"
	case PortTypeMidi:
		return "Midi"
	case PortTypeVideo:
		return "Video"
	case PortTypeControl:
		return "Control"
	default:
		return "Unknown"
	}
}

// AudioFormat describes an audio stream format
type AudioFormat struct {
	Format      string // "S16LE", "S32LE", "F32LE", etc.
	SampleRate  uint32 // 44100, 48000, 96000, etc.
	Channels    uint32 // 1, 2, 6, 8, etc.
	ChannelMask string // "FL,FR" for stereo, "FL,FR,FC,LFE,SL,SR" for 5.1
}

// String returns a human-readable format description
func (af *AudioFormat) String() string {
	return fmt.Sprintf("%s@%dHz %dch (%s)", af.Format, af.SampleRate, af.Channels, af.ChannelMask)
}

// MediaClass describes the class/category of a node
type MediaClass string

const (
	MediaClassAudio            MediaClass = "Audio"
	MediaClassAudioSource      MediaClass = "Audio/Source"
	MediaClassAudioSink        MediaClass = "Audio/Sink"
	MediaClassAudioDuplex      MediaClass = "Audio/Duplex"
	MediaClassAudioRaw         MediaClass = "Audio/Raw"
	MediaClassAudioBridge      MediaClass = "Audio/Bridge"
	MediaClassVideo            MediaClass = "Video"
	MediaClassVideoSource      MediaClass = "Video/Source"
	MediaClassVideoSink        MediaClass = "Video/Sink"
	MediaClassMidi             MediaClass = "Midi"
	MediaClassStream           MediaClass = "Stream"
	MediaClassStreamAudio      MediaClass = "Stream/Audio"
	MediaClassStreamAudioPlayback MediaClass = "Stream/Audio/Playback"
	MediaClassStreamAudioCapture  MediaClass = "Stream/Audio/Capture"
)

// RegistryListener is a callback for registry events
type RegistryListener func(*GlobalObject)

// GlobalObject represents an object in the PipeWire registry
type GlobalObject struct {
	ID         uint32
	Type       string // "Node", "Port", "Link", "Endpoint", etc.
	Version    uint32
	Properties map[string]string
}

// GetProperty retrieves a property value with a default fallback
func (go *GlobalObject) GetProperty(key string, defaultVal string) string {
	if val, ok := go.Properties[key]; ok {
		return val
	}
	return defaultVal
}

// IsNode checks if this object is a Node
func (go *GlobalObject) IsNode() bool {
	return go.Type == "Node" || go.Type == "pw.Node"
}

// IsPort checks if this object is a Port
func (go *GlobalObject) IsPort() bool {
	return go.Type == "Port" || go.Type == "pw.Port"
}

// IsLink checks if this object is a Link
func (go *GlobalObject) IsLink() bool {
	return go.Type == "Link" || go.Type == "pw.Link"
}

// NodeInfo contains detailed information about a node
type NodeInfo struct {
	ID         uint32
	Type       string
	Version    uint32
	Properties map[string]string
	
	// Parsed properties
	Name        string
	Description string
	MediaClass  MediaClass
	Direction   NodeDirection
	State       NodeState
	SampleRate  uint32
	Channels    uint32
}

// PortInfo contains detailed information about a port
type PortInfo struct {
	ID         uint32
	NodeID     uint32
	Type       string
	Version    uint32
	Direction  PortDirection
	Properties map[string]string
	
	// Parsed properties
	Name   string
	Format PortType
}

// LinkInfo contains detailed information about a link
type LinkInfo struct {
	ID         uint32
	OutputNode uint32
	OutputPort uint32
	InputNode  uint32
	InputPort  uint32
	Type       string
	Version    uint32
	Properties map[string]string
	
	// Parsed properties
	Format    *AudioFormat
	IsPassive bool
}

// EventType represents different event types in PipeWire
type EventType uint32

const (
	EventNodeAdded EventType = iota
	EventNodeRemoved
	EventPortAdded
	EventPortRemoved
	EventLinkAdded
	EventLinkRemoved
	EventNodeUpdated
	EventPortUpdated
	EventLinkUpdated
)

// CoreInfo contains information about the Core object
type CoreInfo struct {
	ID      uint32
	Version uint32
	Props   map[string]string
	
	// Parsed properties
	VersionStr string
	Name       string
}

// ClientProxy represents the client object on the server
type ClientProxy struct {
	ID    uint32
	Props map[string]string
}

// PortChange describes what changed in a port update
type PortChange struct {
	PortID    uint32
	Direction PortDirection
	Property  string
	OldValue  string
	NewValue  string
}

// NodeChange describes what changed in a node update
type NodeChange struct {
	NodeID   uint32
	Property string
	OldValue string
	NewValue string
}

// LinkChange describes what changed in a link update
type LinkChange struct {
	LinkID   uint32
	Property string
	OldValue string
	NewValue string
}
