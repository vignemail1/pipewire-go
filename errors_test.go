package pipewireio

import (
	"errors"
	"testing"
)

func TestConnectionError(t *testing.T) {
	inner := errors.New("socket error")
	err := &ConnectionError{
		Reason: "failed to connect",
		Err:    inner,
	}

	msg := err.Error()
	if msg != "connection failed: failed to connect (socket error)" {
		t.Errorf("Error message = %q, want 'connection failed: failed to connect (socket error)'", msg)
	}

	if errors.Unwrap(err) != inner {
		t.Error("Unwrap failed")
	}
}

func TestProtocolError(t *testing.T) {
	err := &ProtocolError{
		Message: "invalid header",
		Data:    []byte{0x01, 0x02, 0x03},
	}

	msg := err.Error()
	if msg != "protocol error: invalid header (data: 3 bytes)" {
		t.Errorf("Error message = %q", msg)
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "name",
		Message: "cannot be empty",
	}

	msg := err.Error()
	if msg != `validation error in field "name": cannot be empty` {
		t.Errorf("Error message = %q", msg)
	}
}

func TestTimeoutError(t *testing.T) {
	err := &TimeoutError{
		Operation: "connect",
		Duration:  "5s",
	}

	msg := err.Error()
	if msg != "timeout during connect (after 5s)" {
		t.Errorf("Error message = %q", msg)
	}
}

func TestNotFoundError(t *testing.T) {
	err := &NotFoundError{
		ResourceType: "Node",
		ResourceID:   uint32(42),
	}

	msg := err.Error()
	if msg != "Node not found: 42" {
		t.Errorf("Error message = %q", msg)
	}
}

func TestPermissionError(t *testing.T) {
	err := &PermissionError{
		Operation: "create_link",
		Reason:    "insufficient privileges",
	}

	msg := err.Error()
	if msg != "permission denied for create_link: insufficient privileges" {
		t.Errorf("Error message = %q", msg)
	}
}

func TestConfigurationError(t *testing.T) {
	err := &ConfigurationError{
		Config: "audio_buffer_size",
		Reason: "must be a power of 2",
	}

	msg := err.Error()
	if msg != `configuration error in "audio_buffer_size": must be a power of 2` {
		t.Errorf("Error message = %q", msg)
	}
}
