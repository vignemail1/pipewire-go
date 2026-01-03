// Package core - types.go
// Core data types and structures for PipeWire protocol

package core

import "fmt"

// ProcessState represents the processing state of a node
type ProcessState uint32

const (
	ProcessStateIdle    ProcessState = 0
	ProcessStateRunning ProcessState = 1
	ProcessStateError   ProcessState = 2
)

// String returns a human-readable process state
func (ps ProcessState) String() string {
	switch ps {
	case ProcessStateIdle:
		return "Idle"
	case ProcessStateRunning:
		return "Running"
	case ProcessStateError:
		return "Error"
	default:
		return fmt.Sprintf("Unknown(%d)", ps)
	}
}

// ChannelPosition represents audio channel positions
type ChannelPosition uint32

const (
	ChannelPositionMono ChannelPosition = iota
	ChannelPositionFL   // Front Left
	ChannelPositionFR   // Front Right
	ChannelPositionFC   // Front Center
	ChannelPositionLFE  // Low-Frequency Effects
	ChannelPositionBL   // Back Left
	ChannelPositionBR   // Back Right
	ChannelPositionFLC  // Front Left Center
	ChannelPositionFRC  // Front Right Center
	ChannelPositionBC   // Back Center
	ChannelPositionSL   // Side Left
	ChannelPositionSR   // Side Right
	ChannelPositionTC   // Top Center
	ChannelPositionTFL  // Top Front Left
	ChannelPositionTFR  // Top Front Right
	ChannelPositionTFC  // Top Front Center
	ChannelPositionTBL  // Top Back Left
	ChannelPositionTBR  // Top Back Right
	ChannelPositionTBC  // Top Back Center
	ChannelPositionBLC  // Back Left Center
	ChannelPositionBRC  // Back Right Center
	ChannelPositionLW   // Left Wide
	ChannelPositionRW   // Right Wide
)

// String returns a human-readable channel position
func (cp ChannelPosition) String() string {
	switch cp {
	case ChannelPositionMono:
		return "Mono"
	case ChannelPositionFL:
		return "FL"
	case ChannelPositionFR:
		return "FR"
	case ChannelPositionFC:
		return "FC"
	case ChannelPositionLFE:
		return "LFE"
	case ChannelPositionBL:
		return "BL"
	case ChannelPositionBR:
		return "BR"
	default:
		return fmt.Sprintf("Unknown(%d)", cp)
	}
}

// AudioFormat represents audio format specifications
type AudioFormat struct {
	SampleRate  uint32
	Channels    uint32
	Format      string // "S16LE", "S32LE", "F32LE", "F64LE"
	ChannelMask string // "FL,FR" for stereo, etc.
}

// String returns a human-readable audio format
func (af *AudioFormat) String() string {
	return fmt.Sprintf("%s@%dHz %dch (%s)", af.Format, af.SampleRate, af.Channels, af.ChannelMask)
}

// NodeFlags represents node capabilities and options
type NodeFlags uint32

const (
	NodeFlagNonTerminal NodeFlags = 1 << 0 // Not a terminal node
	NodeFlagPassthrough NodeFlags = 1 << 1 // Passthrough operation
	NodeFlagAsync       NodeFlags = 1 << 2 // Async processing
)

// PortFlags represents port capabilities and options
type PortFlags uint32

const (
	PortFlagPhysical   PortFlags = 1 << 0 // Physical port
	PortFlagTerminal   PortFlags = 1 << 1 // Terminal port
	PortFlagMonitor    PortFlags = 1 << 2 // Monitor port
	PortFlagDontConnect PortFlags = 1 << 3 // Don't auto-connect
)

// LinkState represents the state of an audio link
type LinkState uint32

const (
	LinkStateError      LinkState = 0
	LinkStateUnlinked   LinkState = 1
	LinkStateLinked     LinkState = 2
	LinkStateLinkedOld  LinkState = 3
)

// String returns a human-readable link state
func (ls LinkState) String() string {
	switch ls {
	case LinkStateError:
		return "Error"
	case LinkStateUnlinked:
		return "Unlinked"
	case LinkStateLinked:
		return "Linked"
	case LinkStateLinkedOld:
		return "LinkedOld"
	default:
		return fmt.Sprintf("Unknown(%d)", ls)
	}
}

// ClientInfo contains information about a client
type ClientInfo struct {
	Version    uint32
	ID         uint32
	Properties map[string]string
}

// ModuleInfo contains information about a module
type ModuleInfo struct {
	Version    uint32
	ID         uint32
	Name       string
	Filename   string
	Args       string
	Properties map[string]string
}

// FactoryInfo contains information about an object factory
type FactoryInfo struct {
	Version    uint32
	ID         uint32
	Name       string
	ObjectType string
	Properties map[string]string
}

// PoolInfo contains information about a memory pool
type PoolInfo struct {
	ID         uint32
	Direction  ParamDirection
	Flags      uint32
	Size       uint32
	Stride     uint32
	Buffers    uint32
	DataType   uint32
}

// MemoryInfo contains information about shared memory
type MemoryInfo struct {
	ID        uint32
	Type      MemType
	FD        int32
	Flags     uint32
	Size      uint32
	Offset    uint32
}

// ProfileInfo contains information about audio profiles
type ProfileInfo struct {
	Index  uint32
	ID     string
	Name   string
	Info   string
	Format *AudioFormat
}

// RouteInfo contains information about audio routes
type RouteInfo struct {
	Index   uint32
	ID      string
	Name    string
	Info    string
	Profile string
}

// EnumFormat represents enumeration of supported formats
type EnumFormat struct {
	Index     uint32
	SampleRate uint32
	Channels  uint32
	Format    string
}

// ProcessLatency represents processing latency information
type ProcessLatency struct {
	Version    uint32
	Direction  ParamDirection
	MinQuantum uint32
	MaxQuantum uint32
	NumSamples uint32
}

// IOConf represents input/output configuration
type IOConf struct {
	Version    uint32
	Direction  ParamDirection
	BufferSize uint32
	Format     *AudioFormat
}

// PortConfiguration represents a port's configuration
type PortConfiguration struct {
	Index      uint32
	Direction  ParamDirection
	PortID     uint32
	Format     *AudioFormat
	Buffers    uint32
	BufferSize uint32
}

// NodeLatency represents node latency information
type NodeLatency struct {
	Direction    ParamDirection
	ProcessLatency uint32
	Quantum      uint32
	Rate         uint32
}

// MetaData represents metadata information
type MetaData struct {
	Type  string
	Size  uint32
	Data  []byte
}

// ObjectProperties represents an object's properties as key-value pairs
type ObjectProperties struct {
	ID    uint32
	Props map[string]string
}

// GetProperty retrieves a property value
func (op *ObjectProperties) GetProperty(key string) (string, bool) {
	val, ok := op.Props[key]
	return val, ok
}

// SetProperty sets a property value
func (op *ObjectProperties) SetProperty(key string, value string) {
	op.Props[key] = value
}

// DeleteProperty removes a property
func (op *ObjectProperties) DeleteProperty(key string) {
	delete(op.Props, key)
}

// HasProperty checks if a property exists
func (op *ObjectProperties) HasProperty(key string) bool {
	_, ok := op.Props[key]
	return ok
}

// ListProperties returns all property keys
func (op *ObjectProperties) ListProperties() []string {
	keys := make([]string, 0, len(op.Props))
	for k := range op.Props {
		keys = append(keys, k)
	}
	return keys
}
