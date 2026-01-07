// Package core - Message buffer for PipeWire protocol
// core/message_buffer.go
// Issue #6 - Event Loop & Automatic Dispatch

package core

import (
	"encoding/binary"
	"fmt"
)

// MessageBuffer buffers bytes from socket until complete frame received
type MessageBuffer struct {
	buffer  []byte
	pos     int
	maxSize int
}

// Frame represents a complete protocol frame
type Frame struct {
	Header    []byte // 12 bytes (ObjectID, MethodID, Sequence)
	Data      []byte // Variable POD data
	Complete  bool
	FrameSize int
}

// NewMessageBuffer creates a new message buffer
func NewMessageBuffer(maxSize int) *MessageBuffer {
	if maxSize <= 0 {
		maxSize = 1024 * 1024 // 1MB default
	}
	return &MessageBuffer{
		buffer:  make([]byte, 0, maxSize),
		pos:     0,
		maxSize: maxSize,
	}
}

// Append adds bytes to the buffer
func (m *MessageBuffer) Append(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	// Check if adding would exceed max size
	if len(m.buffer)+len(data) > m.maxSize {
		return fmt.Errorf("buffer overflow: %d + %d > %d", 
			len(m.buffer), len(data), m.maxSize)
	}

	m.buffer = append(m.buffer, data...)
	return nil
}

// ReadFrame extracts a complete frame from buffer
func (m *MessageBuffer) ReadFrame() (*Frame, error) {
	// Need at least 12 bytes for header
	if len(m.buffer) < 12 {
		return nil, nil // Not enough data yet
	}

	// Parse header (12 bytes little-endian)
	header := m.buffer[:12]
	objectID := binary.LittleEndian.Uint32(header[0:4])
	methodID := binary.LittleEndian.Uint32(header[4:8])
	sequence := binary.LittleEndian.Uint32(header[8:12])

	_ = objectID  // Used by caller
	_ = methodID  // Used by caller
	_ = sequence  // Used by caller

	// In PipeWire, after 12-byte header comes POD data
	// For now, we'll assume single-byte size prefix for data length
	// This is simplified; real implementation would use POD unmarshalling

	// Check if we have enough data
	// Minimum frame: 12 bytes header + 0 bytes data
	if len(m.buffer) < 12 {
		return nil, nil
	}

	// Extract the frame (assuming data is POD-encoded)
	// For this simplified version, we'll return what we have
	frame := &Frame{
		Header:    make([]byte, 12),
		Data:      make([]byte, 0),
		Complete:  true,
		FrameSize: 12,
	}

	copy(frame.Header, header)

	// Remove consumed frame from buffer
	m.buffer = m.buffer[12:]

	return frame, nil
}

// Reset clears the buffer
func (m *MessageBuffer) Reset() {
	m.buffer = m.buffer[:0]
	m.pos = 0
}

// Remaining returns number of bytes in buffer
func (m *MessageBuffer) Remaining() int {
	return len(m.buffer)
}

// Peek returns a copy of buffer without consuming
func (m *MessageBuffer) Peek(n int) ([]byte, error) {
	if n > len(m.buffer) {
		return nil, fmt.Errorf("not enough data: want %d, have %d", n, len(m.buffer))
	}
	result := make([]byte, n)
	copy(result, m.buffer[:n])
	return result, nil
}

// PeekAll returns a copy of entire buffer
func (m *MessageBuffer) PeekAll() []byte {
	result := make([]byte, len(m.buffer))
	copy(result, m.buffer)
	return result
}

// Consume removes n bytes from buffer
func (m *MessageBuffer) Consume(n int) error {
	if n > len(m.buffer) {
		return fmt.Errorf("consume overflow: want %d, have %d", n, len(m.buffer))
	}
	m.buffer = m.buffer[n:]
	return nil
}

// Cap returns the maximum capacity
func (m *MessageBuffer) Cap() int {
	return m.maxSize
}

// Size returns current size
func (m *MessageBuffer) Size() int {
	return len(m.buffer)
}

// IsEmpty returns true if buffer is empty
func (m *MessageBuffer) IsEmpty() bool {
	return len(m.buffer) == 0
}

// Copy returns a deep copy of the buffer
func (m *MessageBuffer) Copy() *MessageBuffer {
	newBuf := NewMessageBuffer(m.maxSize)
	newBuf.buffer = make([]byte, len(m.buffer))
	copy(newBuf.buffer, m.buffer)
	return newBuf
}

// String returns a string representation of buffer contents
func (m *MessageBuffer) String() string {
	return fmt.Sprintf("MessageBuffer{size=%d, cap=%d, data=%d bytes}", 
		len(m.buffer), m.maxSize, len(m.buffer))
}
