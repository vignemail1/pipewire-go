// Package core - Protocol message marshalling and unmarshalling
// core/message.go
// Implements PipeWire protocol message frame format with header + POD data

package core

import (
	"encoding/binary"
	"fmt"
	"github.com/vignemail1/pipewire-go/spa"
)

// MessageFrame represents a complete PipeWire protocol message
// Format: [ObjectID (4B)] [MethodID (4B)] [Sequence (4B)] [PODData (var)]
// This is the fundamental message unit for all client-daemon communication
type MessageFrame struct {
	ObjectID uint32       // ID of target object
	MethodID uint32       // Method ID or event ID
	Sequence uint32       // Sequence number for request/response matching
	PODData  spa.PODValue // POD-encoded arguments (usually PODObject)
}

// Marshal converts frame to bytes with little-endian encoding
// Returns: [objectID(4)] [methodID(4)] [sequence(4)] [podData]
func (m *MessageFrame) Marshal() ([]byte, error) {
	if m == nil {
		return nil, fmt.Errorf("MessageFrame is nil")
	}

	// Build 12-byte header
	header := make([]byte, 12)
	binary.LittleEndian.PutUint32(header[0:4], m.ObjectID)
	binary.LittleEndian.PutUint32(header[4:8], m.MethodID)
	binary.LittleEndian.PutUint32(header[8:12], m.Sequence)

	// Marshal POD data if present
	var podData []byte
	if m.PODData != nil {
		data, err := m.PODData.Marshal()
		if err != nil {
			return nil, fmt.Errorf("POD marshal failed: %w", err)
		}
		podData = data
	}

	// Combine header + POD
	result := append(header, podData...)
	return result, nil
}

// Unmarshal converts bytes to frame with little-endian decoding
func (m *MessageFrame) Unmarshal(data []byte) error {
	if m == nil {
		return fmt.Errorf("MessageFrame is nil")
	}

	if len(data) < 12 {
		return fmt.Errorf("message too short: %d bytes (need >= 12)", len(data))
	}

	// Parse header
	m.ObjectID = binary.LittleEndian.Uint32(data[0:4])
	m.MethodID = binary.LittleEndian.Uint32(data[4:8])
	m.Sequence = binary.LittleEndian.Uint32(data[8:12])

	// Parse POD data if present
	if len(data) > 12 {
		m.PODData = spa.NewPODBytes(data[12:])
	}

	return nil
}

// MessageBuilder provides fluent API for creating message frames
type MessageBuilder struct {
	frame *MessageFrame
}

// NewMessageBuilder creates builder with ObjectID and MethodID set
func NewMessageBuilder(objectID, methodID uint32) *MessageBuilder {
	return &MessageBuilder{
		frame: &MessageFrame{
			ObjectID: objectID,
			MethodID: methodID,
			Sequence: 0,
			PODData:  nil,
		},
	}
}

// WithSequence sets the sequence number for request/response tracking
func (b *MessageBuilder) WithSequence(seq uint32) *MessageBuilder {
	if b != nil && b.frame != nil {
		b.frame.Sequence = seq
	}
	return b
}

// WithPOD sets the POD data for message arguments
func (b *MessageBuilder) WithPOD(pod spa.PODValue) *MessageBuilder {
	if b != nil && b.frame != nil {
		b.frame.PODData = pod
	}
	return b
}

// Build returns the constructed MessageFrame
func (b *MessageBuilder) Build() *MessageFrame {
	if b == nil || b.frame == nil {
		return nil
	}
	return b.frame
}

// String returns human-readable representation
func (m *MessageFrame) String() string {
	if m == nil {
		return "MessageFrame{nil}"
	}
	podStr := "nil"
	if m.PODData != nil {
		podStr = fmt.Sprintf("%T", m.PODData)
	}
	return fmt.Sprintf("MessageFrame{obj:%d method:%d seq:%d pod:%s}",
		m.ObjectID, m.MethodID, m.Sequence, podStr)
}

// Helper functions for extracting POD values

// ExtractUint32FromPOD extracts a uint32 value from POD object
func ExtractUint32FromPOD(obj *spa.PODObject, key string) (uint32, error) {
	if obj == nil {
		return 0, fmt.Errorf("POD object is nil")
	}

	val, ok := obj.Get(key)
	if !ok {
		return 0, fmt.Errorf("key %q not found in POD", key)
	}

	if u32, ok := val.(*spa.PODUint32); ok {
		return u32.Value, nil
	}

	return 0, fmt.Errorf("value at %q is not uint32, got %T", key, val)
}

// ExtractStringFromPOD extracts a string value from POD object
func ExtractStringFromPOD(obj *spa.PODObject, key string) (string, error) {
	if obj == nil {
		return "", fmt.Errorf("POD object is nil")
	}

	val, ok := obj.Get(key)
	if !ok {
		return "", fmt.Errorf("key %q not found in POD", key)
	}

	if str, ok := val.(*spa.PODString); ok {
		return str.Value, nil
	}

	return "", fmt.Errorf("value at %q is not string, got %T", key, val)
}

// ExtractBoolFromPOD extracts a bool value from POD object
func ExtractBoolFromPOD(obj *spa.PODObject, key string) (bool, error) {
	if obj == nil {
		return false, fmt.Errorf("POD object is nil")
	}

	val, ok := obj.Get(key)
	if !ok {
		return false, fmt.Errorf("key %q not found in POD", key)
	}

	if b, ok := val.(*spa.PODBool); ok {
		return b.Value, nil
	}

	return false, fmt.Errorf("value at %q is not bool, got %T", key, val)
}
