package pipewireio

import (
	"errors"
	"testing"
)

func TestConnectionError(t *testing.T) {
	err := &ConnectionError{
		Reason:  "socket closed",
		Address: "pipewire-0",
	}

	if err.Error() == "" {
		t.Error("ConnectionError.Error() returned empty string")
	}

	expected := "connection failed to \"pipewire-0\": socket closed"
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestConnectionErrorWithWrap(t *testing.T) {
	baseErr := errors.New("base error")
	err := &ConnectionError{
		Reason:  "timeout",
		Address: "pipewire-0",
		Err:     baseErr,
	}

	unwrapped := errors.Unwrap(err)
	if unwrapped != baseErr {
		t.Error("Unwrap did not return base error")
	}

	if !errors.Is(err, baseErr) {
		t.Error("errors.Is failed to match wrapped error")
	}
}

func TestProtocolError(t *testing.T) {
	err := &ProtocolError{
		Message: "invalid opcode",
		Opcode:  999,
		Data:    []byte{1, 2, 3, 4},
	}

	if err.Error() == "" {
		t.Error("ProtocolError.Error() returned empty string")
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "port_id",
		Value:   -1,
		Message: "must be positive",
	}

	if err.Error() == "" {
		t.Error("ValidationError.Error() returned empty string")
	}
}

func TestTimeoutError(t *testing.T) {
	err := &TimeoutError{
		Operation: "sync",
		Duration:  "5s",
	}

	expected := "timeout waiting for sync after 5s"
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestFormatNegotiationError(t *testing.T) {
	err := &FormatNegotiationError{
		PortID:         123,
		RequestedFormat: "S32LE",
		AvailableFormats: []string{"S16LE", "F32LE"},
	}

	if err.Error() == "" {
		t.Error("FormatNegotiationError.Error() returned empty string")
	}
}

func TestParameterError(t *testing.T) {
	err := &ParameterError{
		ObjectID:      42,
		ParameterID:   1,
		ParameterName: "Props",
		Reason:        "read-only parameter",
	}

	if err.Error() == "" {
		t.Error("ParameterError.Error() returned empty string")
	}
}

func TestStateError(t *testing.T) {
	err := &StateError{
		ObjectID:     100,
		CurrentState: "SUSPENDED",
		RequestedOp:  "read data",
		Message:      "node is not running",
	}

	if err.Error() == "" {
		t.Error("StateError.Error() returned empty string")
	}
}

func TestNotFoundError(t *testing.T) {
	err := &NotFoundError{
		ResourceType: "Node",
		ResourceID:   999,
	}

	expected := "Node with ID 999 not found"
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestPermissionError(t *testing.T) {
	err := &PermissionError{
		Operation: "create link",
		Resource:  "node 42",
		Reason:    "insufficient privileges",
	}

	if err.Error() == "" {
		t.Error("PermissionError.Error() returned empty string")
	}
}

func TestErrorIsPattern(t *testing.T) {
	// Test with common error instances
	err := ErrNotConnected

	if !errors.Is(err, ErrNotConnected) {
		t.Error("errors.Is failed for ErrNotConnected")
	}
}

func TestErrorAsPattern(t *testing.T) {
	originalErr := &ValidationError{
		Field:   "test",
		Message: "test message",
	}

	var valErr *ValidationError
	if !errors.As(originalErr, &valErr) {
		t.Error("errors.As failed to match ValidationError")
	}

	if valErr.Field != "test" {
		t.Errorf("Expected field 'test', got %q", valErr.Field)
	}
}

func TestWrappedErrorChain(t *testing.T) {
	baseErr := errors.New("base")
	connErr := &ConnectionError{
		Reason:  "wrapped",
		Address: "test",
		Err:     baseErr,
	}

	// Test Unwrap chain
	unwrapped := errors.Unwrap(connErr)
	if unwrapped != baseErr {
		t.Error("Unwrap failed")
	}

	// Test Is chain
	if !errors.Is(connErr, baseErr) {
		t.Error("errors.Is failed for wrapped error")
	}
}
