// Package core - Connection event loop additions
// core/connection_eventloop.go - ADDITIONS FOR Issue #6
// This file contains the methods to ADD to core/connection.go

package core

import (
	"context"
	"fmt"
	"time"
)

// StartEventLoop begins reading messages from the daemon
// This runs in a goroutine and continuously processes incoming messages
func (c *Connection) StartEventLoop(ctx context.Context) error {
	if c == nil {
		return fmt.Errorf("connection is nil")
	}

	c.logger.Infof("Connection: Starting event loop")

	// Initialize state machine
	stateMachine := NewProtocolStateMachine()

	// Transition to connected
	if err := stateMachine.TransitionTo(StateConnected); err != nil {
		c.logger.Errorf("Connection: Failed to transition to connected: %v", err)
		return err
	}

	// Perform Hello handshake
	if err := c.performHelloHandshake(stateMachine); err != nil {
		c.logger.Errorf("Connection: Hello handshake failed: %v", err)
		stateMachine.SetError(err)
		return err
	}

	// Transition to ready
	if err := stateMachine.TransitionTo(StateReady); err != nil {
		c.logger.Errorf("Connection: Failed to transition to ready: %v", err)
		return err
	}

	c.logger.Infof("Connection: Event loop ready, state=%s", stateMachine.GetState())

	// Message buffer for assembling frames
	buffer := NewMessageBuffer(1024 * 1024) // 1MB max

	// Main event loop
	for {
		select {
		case <-ctx.Done():
			c.logger.Infof("Connection: Event loop shutting down")
			return ctx.Err()

		default:
			// Read next message from daemon
			frame, err := c.readNextMessage(buffer)
			if err != nil {
				if err == context.DeadlineExceeded {
					// Timeout, try again
					continue
				}
				c.logger.Errorf("Connection: Read error: %v", err)
				stateMachine.SetError(err)
				return err
			}

			if frame == nil {
				// No complete frame yet, wait a bit
				time.Sleep(10 * time.Millisecond)
				continue
			}

			// Process the frame
			if err := c.processMessage(frame, stateMachine); err != nil {
				c.logger.Errorf("Connection: Message processing error: %v", err)
				// Don't stop loop on processing error, just log it
			}
		}
	}
}

// readNextMessage reads the next message from the socket
func (c *Connection) readNextMessage(buffer *MessageBuffer) (*Frame, error) {
	if c == nil || c.conn == nil {
		return nil, fmt.Errorf("connection not established")
	}

	// Set read timeout
	c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Read up to 4096 bytes
	data := make([]byte, 4096)
	n, err := c.conn.Read(data)
	if err != nil {
		return nil, err
	}

	if n > 0 {
		// Append to buffer
		if err := buffer.Append(data[:n]); err != nil {
			return nil, fmt.Errorf("buffer error: %w", err)
		}
	}

	// Try to extract a complete frame
	frame, err := buffer.ReadFrame()
	if err != nil {
		return nil, err
	}

	return frame, nil
}

// processMessage processes an incoming message frame
func (c *Connection) processMessage(frame *Frame, stateMachine *ProtocolStateMachine) error {
	if frame == nil {
		return fmt.Errorf("frame is nil")
	}

	// Dispatch to event handler if registered
	if c.eventHandler != nil {
		// Create an event from the frame
		// This is simplified - real implementation would parse frame contents
		event := &Event{
			Type: EventTypeMessage,
			Data: frame.Data,
		}

		// Dispatch asynchronously to avoid blocking
		go func() {
			if err := c.eventHandler.Dispatch(frame.Header[0:4], event); err != nil {
				c.logger.Warnf("Connection: Event dispatch error: %v", err)
			}
		}()
	}

	return nil
}

// performHelloHandshake performs the PipeWire Hello handshake
func (c *Connection) performHelloHandshake(stateMachine *ProtocolStateMachine) error {
	if c == nil {
		return fmt.Errorf("connection is nil")
	}

	c.logger.Debugf("Connection: Performing Hello handshake")

	// Create Hello request
	// ObjectID=0 (core), MethodID=0 (Hello)
	helloFrame := &MessageFrame{
		ObjectID: 0,
		MethodID: 0,
		Sequence: 1,
		Data:     nil, // No data for Hello
	}

	// Marshal and send
	data, err := helloFrame.Marshal()
	if err != nil {
		return fmt.Errorf("hello marshal error: %w", err)
	}

	if err := c.writeMessage(data); err != nil {
		return fmt.Errorf("hello write error: %w", err)
	}

	// Transition to HelloSent
	if err := stateMachine.TransitionTo(StateHelloSent); err != nil {
		return err
	}

	// Wait for Hello response (with timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Read Hello response
	buffer := NewMessageBuffer(1024 * 1024)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("hello handshake timeout")
		default:
			frame, err := c.readNextMessage(buffer)
			if err != nil {
				return fmt.Errorf("hello response read error: %w", err)
			}

			if frame != nil {
				// Transition to HelloReceived
				if err := stateMachine.TransitionTo(StateHelloReceived); err != nil {
					return err
				}

				// Parse version from response if available
				// This is simplified - real implementation would parse POD data
				stateMachine.SetVersion(3, 0)

				c.logger.Debugf("Connection: Hello handshake complete, version=%s", 
					stateMachine.GetVersion())
				return nil
			}

			time.Sleep(10 * time.Millisecond)
		}
	}
}

// writeMessage writes a message to the daemon
func (c *Connection) writeMessage(data []byte) error {
	if c == nil || c.conn == nil {
		return fmt.Errorf("connection not established")
	}

	// Set write timeout
	c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	n, err := c.conn.Write(data)
	if err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	if n != len(data) {
		return fmt.Errorf("partial write: %d of %d bytes", n, len(data))
	}

	return nil
}

// WaitUntilReady waits for the connection to be ready
func (c *Connection) WaitUntilReady(ctx context.Context) error {
	if c == nil {
		return fmt.Errorf("connection is nil")
	}

	// In a real implementation, this would check the state machine
	// For now, we'll just wait a bit and return
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}

// Shutdown gracefully closes the connection
func (c *Connection) Shutdown(ctx context.Context) error {
	if c == nil || c.conn == nil {
		return nil
	}

	c.logger.Infof("Connection: Shutting down")

	// Close the underlying connection
	return c.conn.Close()
}

// GetState returns the current connection state
func (c *Connection) GetState() State {
	// In a real implementation, this would return the state machine state
	// For now, return a default
	return StateReady
}

// Event represents an incoming event from the daemon
type Event struct {
	Type EventType
	Data []byte
}

// EventType represents the type of event
type EventType int

const (
	EventTypeMessage EventType = iota
	EventTypePort
	EventTypeNode
	EventTypeLink
)

// EventDispatchFunc is a function that handles events
type EventDispatchFunc func(event *Event) error
