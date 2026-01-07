// Package pipewireio - Error definitions
package pipewireio

import (
	"fmt"
)

// ConnectionError represents a connection failure
type ConnectionError struct {
	Reason  string
	Address string
	Err     error
}

func (e *ConnectionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("connection failed to %q: %s: %v", e.Address, e.Reason, e.Err)
	}
	return fmt.Sprintf("connection failed to %q: %s", e.Address, e.Reason)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// ProtocolError represents a protocol violation
type ProtocolError struct {
	Message string
	Opcode  uint32
	Data    []byte
}

func (e *ProtocolError) Error() string {
	if len(e.Data) > 0 {
		return fmt.Sprintf("protocol error (opcode=%d): %s [data=%d bytes]", e.Opcode, e.Message, len(e.Data))
	}
	return fmt.Sprintf("protocol error (opcode=%d): %s", e.Opcode, e.Message)
}

// ValidationError represents invalid input
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("validation error in field %q (value=%v): %s", e.Field, e.Value, e.Message)
	}
	return fmt.Sprintf("validation error in field %q: %s", e.Field, e.Message)
}

// TimeoutError represents a timeout waiting for response
type TimeoutError struct {
	Operation string
	Duration  string
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("timeout waiting for %s after %s", e.Operation, e.Duration)
}

// FormatNegotiationError represents a format negotiation failure
type FormatNegotiationError struct {
	PortID         uint32
	RequestedFormat string
	AvailableFormats []string
}

func (e *FormatNegotiationError) Error() string {
	return fmt.Sprintf("format negotiation failed for port %d: requested %q, available formats: %v",
		e.PortID, e.RequestedFormat, e.AvailableFormats)
}

// ParameterError represents a parameter operation failure
type ParameterError struct {
	ObjectID      uint32
	ParameterID   uint32
	ParameterName string
	Reason        string
}

func (e *ParameterError) Error() string {
	return fmt.Sprintf("parameter error for object %d, param %s (id=%d): %s",
		e.ObjectID, e.ParameterName, e.ParameterID, e.Reason)
}

// StateError represents an invalid state transition
type StateError struct {
	ObjectID     uint32
	CurrentState string
	RequestedOp  string
	Message      string
}

func (e *StateError) Error() string {
	return fmt.Sprintf("state error for object %d in state %q: cannot %s: %s",
		e.ObjectID, e.CurrentState, e.RequestedOp, e.Message)
}

// NotFoundError represents a resource not found
type NotFoundError struct {
	ResourceType string
	ResourceID   uint32
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID %d not found", e.ResourceType, e.ResourceID)
}

// PermissionError represents a permission denial
type PermissionError struct {
	Operation string
	Resource  string
	Reason    string
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf("permission denied for %s on %s: %s", e.Operation, e.Resource, e.Reason)
}

// Common error instances
var (
	ErrNotConnected    = &ConnectionError{Reason: "not connected", Address: ""}
	ErrAlreadyConnected = &ConnectionError{Reason: "already connected", Address: ""}
	ErrConnectionClosed = &ConnectionError{Reason: "connection closed", Address: ""}
	ErrInvalidAddress  = &ValidationError{Field: "address", Message: "invalid address"}
	ErrNilPointer      = &ValidationError{Field: "pointer", Message: "nil pointer"}
)
