// Package core - PipeWire Protocol Core
// core/connection.go
// Socket-based connection handling for PipeWire protocol
// Phase 1 - Core connection infrastructure
// NOTE: Error types (ConnectionError, TimeoutError, ProtocolError) are defined in errors.go

package core

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/vignemail1/pipewire-go/verbose"
)

// ProtocolVersion defines the protocol version
const ProtocolVersion = 3

// DefaultSocketPath is the default PipeWire socket path
const DefaultSocketPath = "/run/pipewire-0"

// Connection represents a connection to PipeWire daemon via unix socket
type Connection struct {
	socket     net.Conn
	logger     *verbose.Logger
	buffer     *bytes.Buffer
	timeout    time.Duration
	connected  bool
	readBuf    []byte
	writeBuf   []byte
	syncID     uint32
}

// Dial establishes a connection to the PipeWire daemon
// It takes a socket path (typically "/run/pipewire-0") and creates a Unix socket
// connection, performing the initial handshake if needed
func Dial(socketPath string, logger *verbose.Logger) (*Connection, error) {
	if socketPath == "" {
		socketPath = DefaultSocketPath
	}

	if logger == nil {
		logger = verbose.NewLogger(verbose.LogLevelInfo, false)
	}

	logger.Infof("Dialing PipeWire daemon at %s", socketPath)

	// Create the connection
	conn, err := NewConnection(socketPath, logger)
	if err != nil {
		return nil, err
	}

	// Set a reasonable default timeout
	conn.SetTimeout(5 * time.Second)

	logger.Infof("Successfully connected to PipeWire daemon")
	return conn, nil
}

// NewConnection creates a new connection to PipeWire
func NewConnection(socketPath string, logger *verbose.Logger) (*Connection, error) {
	if logger == nil {
		logger = verbose.NewLogger(verbose.LogLevelInfo, false)
	}

	// Create unix socket connection
	socket, err := net.Dial("unix", socketPath)
	if err != nil {
		logger.Error("Failed to connect to PipeWire socket", "path", socketPath, "error", err)
		return nil, NewConnectionError(fmt.Sprintf("failed to connect to %s: %v", socketPath, err))
	}

	conn := &Connection{
		socket:    socket,
		logger:    logger,
		buffer:    new(bytes.Buffer),
		timeout:   5 * time.Second,
		connected: true,
		readBuf:   make([]byte, 4096),
		writeBuf:  make([]byte, 4096),
		syncID:    0,
	}

	logger.Debug("Connected to PipeWire", "path", socketPath)
	return conn, nil
}

// IsConnected returns true if connection is active
func (c *Connection) IsConnected() bool {
	return c.connected
}

// SetTimeout sets the read/write timeout
func (c *Connection) SetTimeout(duration time.Duration) {
	c.timeout = duration
}

// Write sends data to PipeWire
func (c *Connection) Write(data []byte) (int, error) {
	if !c.connected {
		return 0, NewConnectionError("connection is closed")
	}

	// Set write deadline
	if c.timeout > 0 {
		c.socket.SetWriteDeadline(time.Now().Add(c.timeout))
	}

	n, err := c.socket.Write(data)
	if err != nil {
		c.logger.Error("Write error", "error", err)
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return n, NewTimeoutError("write timeout")
		}
		return n, NewProtocolError(fmt.Sprintf("write error: %v", err))
	}

	c.logger.Debug("Wrote bytes", "count", n)
	return n, nil
}

// Read reads data from PipeWire
func (c *Connection) Read(p []byte) (int, error) {
	if !c.connected {
		return 0, NewConnectionError("connection is closed")
	}

	// Set read deadline
	if c.timeout > 0 {
		c.socket.SetReadDeadline(time.Now().Add(c.timeout))
	}

	n, err := c.socket.Read(p)
	if err != nil {
		if err == io.EOF {
			c.logger.Debug("Connection closed by remote")
			return 0, err
		}
		c.logger.Error("Read error", "error", err)
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return n, NewTimeoutError("read timeout")
		}
		return n, NewProtocolError(fmt.Sprintf("read error: %v", err))
	}

	c.logger.Debug("Read bytes", "count", n)
	return n, nil
}

// WriteMessage sends a protocol message
func (c *Connection) WriteMessage(msg []byte) error {
	if !c.connected {
		return NewConnectionError("connection is closed")
	}

	// Prefix message with length (4 bytes, little-endian)
	length := uint32(len(msg))
	lengthBytes := make([]byte, 4)
	for i := 0; i < 4; i++ {
		lengthBytes[i] = byte((length >> (uint(i) * 8)) & 0xFF)
	}

	// Write length then message
	if _, err := c.Write(lengthBytes); err != nil {
		return err
	}
	if _, err := c.Write(msg); err != nil {
		return err
	}

	c.logger.Debug("Message sent", "length", length)
	return nil
}

// ReadMessage reads a protocol message
func (c *Connection) ReadMessage() ([]byte, error) {
	if !c.connected {
		return nil, NewConnectionError("connection is closed")
	}

	// Read length header (4 bytes)
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(c, lengthBuf); err != nil {
		return nil, err
	}

	// Parse length
	length := uint32(0)
	for i := 0; i < 4; i++ {
		length |= uint32(lengthBuf[i]) << (uint(i) * 8)
	}

	if length == 0 {
		return nil, NewProtocolError("message length is 0")
	}
	if length > 1024*1024 { // 1MB limit
		return nil, NewProtocolError(fmt.Sprintf("message too large: %d bytes", length))
	}

	// Read message data
	msgBuf := make([]byte, length)
	if _, err := io.ReadFull(c, msgBuf); err != nil {
		return nil, err
	}

	c.logger.Debug("Message received", "length", length)
	return msgBuf, nil
}

// Flush flushes pending writes
func (c *Connection) Flush() error {
	if !c.connected {
		return NewConnectionError("connection is closed")
	}

	if c.buffer.Len() > 0 {
		data := c.buffer.Bytes()
		_, err := c.Write(data)
		if err != nil {
			return err
		}
		c.buffer.Reset()
	}
	return nil
}

// Close closes the connection
func (c *Connection) Close() error {
	if !c.connected {
		return nil
	}

	c.connected = false
	if c.socket != nil {
		c.logger.Debug("Closing connection")
		return c.socket.Close()
	}
	return nil
}

// LocalAddr returns the local address
func (c *Connection) LocalAddr() net.Addr {
	if c.socket != nil {
		return c.socket.LocalAddr()
	}
	return nil
}

// RemoteAddr returns the remote address
func (c *Connection) RemoteAddr() net.Addr {
	if c.socket != nil {
		return c.socket.RemoteAddr()
	}
	return nil
}

// GetSyncID returns and increments the sync ID
func (c *Connection) GetSyncID() uint32 {
	c.syncID++
	return c.syncID
}

// SetSyncID sets the sync ID
func (c *Connection) SetSyncID(id uint32) {
	c.syncID = id
}

// CurrentSyncID returns the current sync ID without incrementing
func (c *Connection) CurrentSyncID() uint32 {
	return c.syncID
}

// Verify implements io.Reader interface
var _ io.Reader = (*Connection)(nil)

// Verify implements io.Writer interface
var _ io.Writer = (*Connection)(nil)
