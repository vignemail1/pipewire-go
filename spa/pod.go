// Package spa - Simple Protocol Audio
// spa/pod.go
// POD (Plain Old Data) serialization and deserialization for PipeWire protocol
// Phase 1 - Core POD handling

package spa

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// PODType interface for all POD values
type PODType interface {
	Marshal() []byte
	Unmarshal([]byte) error
	String() string
}

// PODValue is the interface for all POD value types
// NOTE: Concrete types (PODBool, PODInt, etc.) are defined in types.go
// This file contains only parsing and marshaling infrastructure
type PODValue interface {
	Marshal() []byte
	Unmarshal([]byte) error
	String() string
	Type() string
}

// PODObjectBuilder helps build POD objects
type PODObjectBuilder struct {
	properties map[string]PODValue
}

// NewPODObjectBuilder creates a new object builder
func NewPODObjectBuilder() *PODObjectBuilder {
	return &PODObjectBuilder{
		properties: make(map[string]PODValue),
	}
}

// Add adds a property to the object
func (pb *PODObjectBuilder) Add(key string, value PODValue) *PODObjectBuilder {
	pb.properties[key] = value
	return pb
}

// Build builds the POD object
func (pb *PODObjectBuilder) Build() PODValue {
	// Return a PODObject representation
	// Actual implementation in types.go
	return nil
}

// PODArrayBuilder helps build POD arrays
type PODArrayBuilder struct {
	items []PODValue
}

// NewPODArrayBuilder creates a new array builder
func NewPODArrayBuilder() *PODArrayBuilder {
	return &PODArrayBuilder{
		items: make([]PODValue, 0),
	}
}

// Add adds an item to the array
func (ab *PODArrayBuilder) Add(value PODValue) *PODArrayBuilder {
	ab.items = append(ab.items, value)
	return ab
}

// Build builds the POD array
func (ab *PODArrayBuilder) Build() PODValue {
	// Return a PODArray representation
	// Actual implementation in types.go
	return nil
}

// PODParser parses POD binary format
type PODParser struct {
	reader io.Reader
	buffer *bytes.Buffer
}

// NewPODParser creates a new POD parser
func NewPODParser(r io.Reader) *PODParser {
	return &PODParser{
		reader: r,
		buffer: new(bytes.Buffer),
	}
}

// ParseValue parses a POD value from reader
func (pp *PODParser) ParseValue() (PODValue, error) {
	// Read type ID first (4 bytes)
	var typeID uint32
	if err := binary.Read(pp.reader, binary.LittleEndian, &typeID); err != nil {
		return nil, fmt.Errorf("failed to read type ID: %w", err)
	}

	// Route to appropriate type parser based on typeID
	// This would call type-specific parsing logic
	return nil, fmt.Errorf("type ID %d not yet implemented", typeID)
}

// PODWriter writes POD binary format
type PODWriter struct {
	writer io.Writer
	buffer *bytes.Buffer
}

// NewPODWriter creates a new POD writer
func NewPODWriter(w io.Writer) *PODWriter {
	return &PODWriter{
		writer: w,
		buffer: new(bytes.Buffer),
	}
}

// WriteValue writes a POD value to writer
func (pw *PODWriter) WriteValue(value PODValue) error {
	data := value.Marshal()
	_, err := pw.writer.Write(data)
	return err
}

// Flush flushes the buffer
func (pw *PODWriter) Flush() error {
	if pw.buffer.Len() > 0 {
		_, err := pw.writer.Write(pw.buffer.Bytes())
		if err != nil {
			return err
		}
		pw.buffer.Reset()
	}
	return nil
}

// Helper functions for marshaling common types

// MarshalUint32 marshals uint32 value
func MarshalUint32(value uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, value)
	return buf
}

// UnmarshalUint32 unmarshals uint32 value
func UnmarshalUint32(data []byte) (uint32, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("insufficient data for uint32")
	}
	return binary.LittleEndian.Uint32(data[:4]), nil
}

// MarshalInt32 marshals int32 value
func MarshalInt32(value int32) []byte {
	return MarshalUint32(uint32(value))
}

// UnmarshalInt32 unmarshals int32 value
func UnmarshalInt32(data []byte) (int32, error) {
	val, err := UnmarshalUint32(data)
	return int32(val), err
}

// MarshalString marshals string value
func MarshalString(value string) []byte {
	// Null-terminated string
	return append([]byte(value), 0)
}

// UnmarshalString unmarshals string value
func UnmarshalString(data []byte) (string, error) {
	// Find null terminator
	end := bytes.IndexByte(data, 0)
	if end == -1 {
		return "", fmt.Errorf("no null terminator in string")
	}
	return string(data[:end]), nil
}

// POD protocol constants
const (
	PODVersionMajor = 3
	PODVersionMinor = 0
)

// POD size constants
const (
	PODSizeNull   = 0
	PODSizeBool   = 4
	PODSizeInt    = 4
	PODSizeLong   = 8
	PODSizeFloat  = 4
	PODSizeDouble = 8
)
