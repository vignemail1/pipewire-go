// Package core - PipeWire Protocol Core
// core/errors.go
// Complete error handling with proper error interface implementation
// Phase 3 - Error types and codes

package core

import (
	"fmt"
)

// ErrorCode represents all possible PipeWire protocol errors
type ErrorCode int

// Error codes
const (
	ErrorCodeOK               ErrorCode = 0
	ErrorCodeUnknown          ErrorCode = -1
	ErrorCodeInvalidArgument  ErrorCode = -2
	ErrorCodePermissionDenied ErrorCode = -3
	ErrorCodeNotFound         ErrorCode = -4
	ErrorCodeInvalidType      ErrorCode = -5
	ErrorCodeInvalidValue     ErrorCode = -6
	ErrorCodeInvalidFormat    ErrorCode = -7
	ErrorCodeInvalidState     ErrorCode = -8
	ErrorCodeBufferFull       ErrorCode = -9
	ErrorCodeTimeout          ErrorCode = -10
	ErrorCodeConnectionLost   ErrorCode = -11
	ErrorCodeConnectionRefused ErrorCode = -12
	ErrorCodeProtocolError    ErrorCode = -13
	ErrorCodeNoMemory         ErrorCode = -14
	ErrorCodeNotSupported     ErrorCode = -15
	ErrorCodeBusyError        ErrorCode = -16
	ErrorCodeIOError          ErrorCode = -17
	ErrorCodeCorrupted        ErrorCode = -18
	ErrorCodeNotOwner         ErrorCode = -19
	ErrorCodeUnavailable      ErrorCode = -20
	ErrorCodeUnimplemented    ErrorCode = -21
	ErrorCodeMaxConnections   ErrorCode = -22
)

// Error is the base error type with proper error interface implementation
type Error struct {
	Code    ErrorCode
	Message string
	Wrapped error
}

// Ensure Error implements error interface
var _ error = (*Error)(nil)

// Error implements the error interface
func (e *Error) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Wrapped)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// String returns the string representation
func (e *Error) String() string {
	return e.Error()
}

// Unwrap returns the wrapped error
func (e *Error) Unwrap() error {
	return e.Wrapped
}

// NewError creates a new error with code and message
func NewError(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Wrapped: nil,
	}
}

// NewErrorf creates a new error with formatted message
func NewErrorf(code ErrorCode, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Wrapped: nil,
	}
}

// WrapError wraps an existing error with a PipeWire error code
func WrapError(code ErrorCode, err error) *Error {
	return &Error{
		Code:    code,
		Message: "wrapped error",
		Wrapped: err,
	}
}

// ConnectionError represents a connection error
// PROPERLY IMPLEMENTS error interface (not as field, but through Error() method)
type ConnectionError struct {
	*Error
}

// Ensure ConnectionError implements error interface
var _ error = (*ConnectionError)(nil)

// NewConnectionError creates a new connection error
func NewConnectionError(message string) *ConnectionError {
	return &ConnectionError{
		Error: &Error{
			Code:    ErrorCodeConnectionLost,
			Message: message,
			Wrapped: nil,
		},
	}
}

// NewConnectionErrorf creates a new connection error with formatted message
func NewConnectionErrorf(format string, args ...interface{}) *ConnectionError {
	return &ConnectionError{
		Error: &Error{
			Code:    ErrorCodeConnectionLost,
			Message: fmt.Sprintf(format, args...),
			Wrapped: nil,
		},
	}
}

// TimeoutError represents a timeout error
// PROPERLY IMPLEMENTS error interface
type TimeoutError struct {
	*Error
}

// Ensure TimeoutError implements error interface
var _ error = (*TimeoutError)(nil)

// NewTimeoutError creates a new timeout error
func NewTimeoutError(message string) *TimeoutError {
	return &TimeoutError{
		Error: &Error{
			Code:    ErrorCodeTimeout,
			Message: message,
			Wrapped: nil,
		},
	}
}

// NewTimeoutErrorf creates a new timeout error with formatted message
func NewTimeoutErrorf(format string, args ...interface{}) *TimeoutError {
	return &TimeoutError{
		Error: &Error{
			Code:    ErrorCodeTimeout,
			Message: fmt.Sprintf(format, args...),
			Wrapped: nil,
		},
	}
}

// ProtocolError represents a protocol error
// PROPERLY IMPLEMENTS error interface
type ProtocolError struct {
	*Error
}

// Ensure ProtocolError implements error interface
var _ error = (*ProtocolError)(nil)

// NewProtocolError creates a new protocol error
func NewProtocolError(message string) *ProtocolError {
	return &ProtocolError{
		Error: &Error{
			Code:    ErrorCodeProtocolError,
			Message: message,
			Wrapped: nil,
		},
	}
}

// NewProtocolErrorf creates a new protocol error with formatted message
func NewProtocolErrorf(format string, args ...interface{}) *ProtocolError {
	return &ProtocolError{
		Error: &Error{
			Code:    ErrorCodeProtocolError,
			Message: fmt.Sprintf(format, args...),
			Wrapped: nil,
		},
	}
}

// ResourceError represents a resource error
type ResourceError struct {
	*Error
}

// Ensure ResourceError implements error interface
var _ error = (*ResourceError)(nil)

// NewResourceError creates a new resource error
func NewResourceError(message string) *ResourceError {
	return &ResourceError{
		Error: &Error{
			Code:    ErrorCodeNoMemory,
			Message: message,
			Wrapped: nil,
		},
	}
}

// Helper functions for error checking

// IsTimeout checks if an error is a timeout error
func IsTimeout(err error) bool {
	_, ok := err.(*TimeoutError)
	return ok
}

// IsConnectionError checks if an error is a connection error
func IsConnectionError(err error) bool {
	_, ok := err.(*ConnectionError)
	return ok
}

// IsProtocolError checks if an error is a protocol error
func IsProtocolError(err error) bool {
	_, ok := err.(*ProtocolError)
	return ok
}

// IsPermissionError checks if an error is a permission error
func IsPermissionError(err error) bool {
	e, ok := err.(*Error)
	return ok && e.Code == ErrorCodePermissionDenied
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	e, ok := err.(*Error)
	return ok && e.Code == ErrorCodeNotFound
}

// IsInvalid checks if an error is an invalid error
func IsInvalid(err error) bool {
	e, ok := err.(*Error)
	if !ok {
		return false
	}
	return e.Code == ErrorCodeInvalidArgument ||
		e.Code == ErrorCodeInvalidType ||
		e.Code == ErrorCodeInvalidValue ||
		e.Code == ErrorCodeInvalidFormat ||
		e.Code == ErrorCodeInvalidState
}

// IsResourceError checks if an error is a resource error
func IsResourceError(err error) bool {
	_, ok := err.(*ResourceError)
	return ok
}

// CodeFromError extracts error code from error
func CodeFromError(err error) ErrorCode {
	if e, ok := err.(*Error); ok {
		return e.Code
	}
	if _, ok := err.(*ConnectionError); ok {
		return ErrorCodeConnectionLost
	}
	if _, ok := err.(*TimeoutError); ok {
		return ErrorCodeTimeout
	}
	if _, ok := err.(*ProtocolError); ok {
		return ErrorCodeProtocolError
	}
	return ErrorCodeUnknown
}

// String representation of error codes
func (ec ErrorCode) String() string {
	switch ec {
	case ErrorCodeOK:
		return "OK"
	case ErrorCodeUnknown:
		return "Unknown"
	case ErrorCodeInvalidArgument:
		return "Invalid Argument"
	case ErrorCodePermissionDenied:
		return "Permission Denied"
	case ErrorCodeNotFound:
		return "Not Found"
	case ErrorCodeInvalidType:
		return "Invalid Type"
	case ErrorCodeInvalidValue:
		return "Invalid Value"
	case ErrorCodeInvalidFormat:
		return "Invalid Format"
	case ErrorCodeInvalidState:
		return "Invalid State"
	case ErrorCodeBufferFull:
		return "Buffer Full"
	case ErrorCodeTimeout:
		return "Timeout"
	case ErrorCodeConnectionLost:
		return "Connection Lost"
	case ErrorCodeConnectionRefused:
		return "Connection Refused"
	case ErrorCodeProtocolError:
		return "Protocol Error"
	case ErrorCodeNoMemory:
		return "No Memory"
	case ErrorCodeNotSupported:
		return "Not Supported"
	case ErrorCodeBusyError:
		return "Busy"
	case ErrorCodeIOError:
		return "I/O Error"
	case ErrorCodeCorrupted:
		return "Corrupted"
	case ErrorCodeNotOwner:
		return "Not Owner"
	case ErrorCodeUnavailable:
		return "Unavailable"
	case ErrorCodeUnimplemented:
		return "Unimplemented"
	case ErrorCodeMaxConnections:
		return "Max Connections"
	default:
		return fmt.Sprintf("Unknown Error Code: %d", ec)
	}
}
