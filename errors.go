package pipewireio

import (
	"fmt"
)

// ConnectionError represents a connection failure
type ConnectionError struct {
	Reason string
	Err    error
}

func (e *ConnectionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("connection failed: %s (%v)", e.Reason, e.Err)
	}
	return fmt.Sprintf("connection failed: %s", e.Reason)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// ProtocolError represents a protocol violation
type ProtocolError struct {
	Message string
	Data    []byte
}

func (e *ProtocolError) Error() string {
	if len(e.Data) > 0 {
		return fmt.Sprintf("protocol error: %s (data: %v bytes)", e.Message, len(e.Data))
	}
	return fmt.Sprintf("protocol error: %s", e.Message)
}

// ValidationError represents invalid input
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in field %q: %s", e.Field, e.Message)
}

// TimeoutError represents a timeout during operation
type TimeoutError struct {
	Operation string
	Duration  string
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("timeout during %s (after %s)", e.Operation, e.Duration)
}

// NotFoundError represents a resource not found
type NotFoundError struct {
	ResourceType string
	ResourceID   interface{}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %v", e.ResourceType, e.ResourceID)
}

// PermissionError represents insufficient permissions
type PermissionError struct {
	Operation string
	Reason    string
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf("permission denied for %s: %s", e.Operation, e.Reason)
}

// ConfigurationError represents invalid configuration
type ConfigurationError struct {
	Config  string
	Reason  string
}

func (e *ConfigurationError) Error() string {
	return fmt.Sprintf("configuration error in %q: %s", e.Config, e.Reason)
}
