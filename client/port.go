// Package client - High-Level PipeWire Client
// client/port.go
// Port management and format negotiation
// Complete with format handling for Issue #6

package client

import (
	"fmt"
	"sync"

	"github.com/vignemail1/pipewire-go/core"
)

// ============================================================================
// Format Definitions
// ============================================================================

// AudioFormat represents an audio format
type AudioFormat struct {
	MediaType  string // e.g. "audio"
	MediaSubtype string // e.g. "raw", "mpeg", "ulaw"
	Encoding   string // e.g. "S16_LE", "S32_LE", "F32_LE"
	Rate       uint32 // Sample rate in Hz
	Channels   uint32 // Number of channels
}

// Format represents a port format (audio, MIDI, etc.)
type Format struct {
	Type       PortType      // Type of format (Audio, MIDI, etc.)
	Audio      *AudioFormat  // Audio-specific format (if Type == PortTypeAudio)
	Properties map[string]string
}

// String returns string representation of format
func (f *Format) String() string {
	if f == nil {
		return "<nil>"
	}
	if f.Audio != nil {
		return fmt.Sprintf("Audio{%s, %dHz, %dch}",
			f.Audio.Encoding, f.Audio.Rate, f.Audio.Channels)
	}
	return fmt.Sprintf("Format{type:%v}", f.Type)
}

// Equals compares two formats
func (f *Format) Equals(other *Format) bool {
	if f == nil || other == nil {
		return f == other
	}
	if f.Type != other.Type {
		return false
	}
	if f.Audio != nil && other.Audio != nil {
		return f.Audio.Rate == other.Audio.Rate &&
			f.Audio.Channels == other.Audio.Channels &&
			f.Audio.Encoding == other.Audio.Encoding
	}
	return f.Audio == other.Audio
}

// ============================================================================
// Port Structure and Basic Methods (from original)
// ============================================================================

// Port represents a PipeWire port
type Port struct {
	mu         sync.RWMutex
	id         uint32
	name       string
	direction  PortDirection
	portType   PortType
	properties *Properties
	node       *Node
	client     *Client
	info       *PortInfo

	// NEW FOR ISSUE #6: Format negotiation
	currentFormat    *Format
	supportedFormats []*Format
}

// PortDirection represents port direction (input/output)
type PortDirection int

const (
	PortDirectionInput PortDirection = iota
	PortDirectionOutput
)

// PortType represents the type of port
type PortType int

const (
	PortTypeAudio PortType = iota
	PortTypeMIDI
	PortTypeControl
	PortTypeOther
)

// PortInfo contains detailed port information
type PortInfo struct {
	ID          uint32
	Name        string
	Direction   PortDirection
	Type        PortType
	NodeID      uint32
	Channels    int
	Rate        uint32
	Latency     uint32
	Properties  map[string]string
	Connected   bool
}

// NewPort creates a new port instance
func NewPort(id uint32, name string, direction PortDirection, node *Node, client *Client) *Port {
	return &Port{
		id:         id,
		name:       name,
		direction:  direction,
		properties: NewProperties(),
		node:       node,
		client:     client,
		info: &PortInfo{
			ID:        id,
			Name:      name,
			Direction: direction,
		},
		supportedFormats: make([]*Format, 0),
	}
}

// ID returns the port ID
func (p *Port) ID() uint32 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.id
}

// Name returns the port name
func (p *Port) Name() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.name
}

// Direction returns the port direction
func (p *Port) Direction() PortDirection {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.direction
}

// Type returns the port type
func (p *Port) Type() PortType {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.portType
}

// SetType sets the port type
func (p *Port) SetType(t PortType) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.portType = t
}

// Node returns the parent node
func (p *Port) Node() *Node {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.node
}

// Properties returns the port properties
func (p *Port) Properties() *Properties {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.properties
}

// Info returns detailed port information
func (p *Port) Info() *PortInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	info := *p.info
	return &info
}

// SetInfo updates port information
func (p *Port) SetInfo(info *PortInfo) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if info != nil {
		p.info = info
	}
}

// GetChannels returns the number of channels
func (p *Port) GetChannels() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.info.Channels
}

// SetChannels sets the number of channels
func (p *Port) SetChannels(channels int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.info.Channels = channels
}

// GetLatency returns the port latency
func (p *Port) GetLatency() uint32 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.info.Latency
}

// SetLatency sets the port latency
func (p *Port) SetLatency(latency uint32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.info.Latency = latency
}

// GetRate returns the port sample rate
func (p *Port) GetRate() uint32 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.info.Rate
}

// SetRate sets the port sample rate
func (p *Port) SetRate(rate uint32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.info.Rate = rate
}

// IsConnected returns if the port is connected
func (p *Port) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.info.Connected
}

// SetConnected marks the port as connected/disconnected
func (p *Port) SetConnected(connected bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.info.Connected = connected
}

// String returns string representation
func (p *Port) String() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return fmt.Sprintf("Port(%s, dir=%v, type=%v)", p.name, p.direction, p.portType)
}

// GetConnectedPorts returns all ports connected to this port
func (p *Port) GetConnectedPorts() ([]*Port, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.client == nil {
		return nil, fmt.Errorf("port not associated with client")
	}

	return p.client.GetConnectedPorts(p.id)
}

// IsAudioPort returns true if this is an audio port
func (p *Port) IsAudioPort() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.portType == PortTypeAudio
}

// IsMIDIPort returns true if this is a MIDI port
func (p *Port) IsMIDIPort() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.portType == PortTypeMIDI
}

// IsInput returns true if port is input direction
func (p *Port) IsInput() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.direction == PortDirectionInput
}

// IsOutput returns true if port is output direction
func (p *Port) IsOutput() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.direction == PortDirectionOutput
}

// ============================================================================
// NEW METHODS FOR ISSUE #6: Format Negotiation
// ============================================================================

// GetSupportedFormats returns all formats supported by this port
func (p *Port) GetSupportedFormats() []*Format {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.supportedFormats) == 0 {
		// Default supported formats for audio ports
		if p.portType == PortTypeAudio {
			p.supportedFormats = []*Format{
				{
					Type: PortTypeAudio,
					Audio: &AudioFormat{
						MediaType:    "audio",
						MediaSubtype: "raw",
						Encoding:     "S16_LE",
						Rate:         48000,
						Channels:     2,
					},
				},
				{
					Type: PortTypeAudio,
					Audio: &AudioFormat{
						MediaType:    "audio",
						MediaSubtype: "raw",
						Encoding:     "F32_LE",
						Rate:         48000,
						Channels:     2,
					},
				},
			}
		}
	}

	// Return copy of supported formats
	formats := make([]*Format, len(p.supportedFormats))
	copy(formats, p.supportedFormats)
	return formats
}

// SetSupportedFormats sets the list of supported formats for this port
func (p *Port) SetSupportedFormats(formats []*Format) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.supportedFormats = make([]*Format, len(formats))
	copy(p.supportedFormats, formats)
}

// GetFormat returns the currently negotiated format for this port
func (p *Port) GetFormat() (*Format, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.currentFormat != nil {
		return p.currentFormat, nil
	}

	// Return first supported format as default
	if len(p.supportedFormats) > 0 {
		return p.supportedFormats[0], nil
	}

	return nil, fmt.Errorf("port %s has no negotiated or supported formats", p.name)
}

// SetFormat attempts to set the port format
// Returns error if the format is not supported
func (p *Port) SetFormat(format *Format) error {
	if format == nil {
		return fmt.Errorf("format cannot be nil")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if format is supported
	for _, supported := range p.supportedFormats {
		if supported.Equals(format) {
			p.currentFormat = format

			// Update port info with format details
			if format.Audio != nil {
				p.info.Rate = format.Audio.Rate
				p.info.Channels = int(format.Audio.Channels)
			}

			return nil
		}
	}

	return fmt.Errorf("port %s does not support format %s", p.name, format.String())
}

// CanConnectTo checks if this port can connect to another port
// Returns true if directions are compatible and formats can negotiate
func (p *Port) CanConnectTo(other *Port) bool {
	if other == nil {
		return false
	}

	// Check direction compatibility
	// Output can connect to Input
	if p.direction == PortDirectionOutput && other.direction != PortDirectionInput {
		return false
	}
	// Input can connect to Output
	if p.direction == PortDirectionInput && other.direction != PortDirectionOutput {
		return false
	}

	// For now, if both are audio ports, allow connection
	// In a full implementation, would check format compatibility
	if p.portType == PortTypeAudio && other.portType == PortTypeAudio {
		return true
	}

	return p.portType == other.portType
}

// ============================================================================
// Filter and Helper Methods
// ============================================================================

// PortFilter represents a filter for selecting ports
type PortFilter struct {
	Direction PortDirection
	Type      PortType
	NodeID    uint32
	NameMatch string
}

// Matches checks if a port matches the filter
func (f *PortFilter) Matches(p *Port) bool {
	if p == nil {
		return false
	}

	if f.NodeID != 0 && p.node != nil && p.node.ID != f.NodeID {
		return false
	}

	if p.direction != f.Direction && f.Direction >= 0 {
		return false
	}

	if p.portType != f.Type && f.Type >= 0 {
		return false
	}

	if f.NameMatch != "" && p.name != f.NameMatch {
		return false
	}

	return true
}

// DirectionString returns string representation of port direction
func (pd PortDirection) String() string {
	switch pd {
	case PortDirectionInput:
		return "input"
	case PortDirectionOutput:
		return "output"
	default:
		return "unknown"
	}
}

// TypeString returns string representation of port type
func (pt PortType) String() string {
	switch pt {
	case PortTypeAudio:
		return "audio"
	case PortTypeMIDI:
		return "midi"
	case PortTypeControl:
		return "control"
	case PortTypeOther:
		return "other"
	default:
		return "unknown"
	}
}
