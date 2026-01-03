// Package verbose provides logging and debugging utilities for PipeWire
// including verbose output of all messages, POD structures, and binary data
package verbose

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

// LogLevel defines the verbosity level
type LogLevel int

const (
	LogLevelSilent LogLevel = iota
	LogLevelError
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

// Logger provides verbose logging for PipeWire operations
type Logger struct {
	level             LogLevel
	includeTimestamps bool
	includeCaller     bool
	mutex             sync.Mutex

	// Callbacks for external logging
	onSendCallback func(objID uint32, methodID uint32, data []byte)
	onRecvCallback func(objID uint32, eventID uint32, data []byte)
	onErrorCallback func(err error, context string)
}

// DefaultLogger creates a logger with default settings
func DefaultLogger() *Logger {
	return &Logger{
		level:             LogLevelInfo,
		includeTimestamps: true,
		includeCaller:     false,
	}
}

// NewLogger creates a new logger with specified settings
func NewLogger(level LogLevel, includeTimestamps bool) *Logger {
	return &Logger{
		level:             level,
		includeTimestamps: includeTimestamps,
		includeCaller:     false,
	}
}

// SetLevel changes the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.level = level
}

// IsVerbose checks if a certain level is active
func (l *Logger) IsVerbose(level LogLevel) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l.level >= level
}

// Errorf logs an error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.level >= LogLevelError {
		l.log("[ERROR]", format, args...)
	}
}

// Warnf logs a warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	if l.level >= LogLevelWarn {
		l.log("[WARN]", format, args...)
	}
}

// Infof logs an info message
func (l *Logger) Infof(format string, args ...interface{}) {
	if l.level >= LogLevelInfo {
		l.log("[INFO]", format, args...)
	}
}

// Debugf logs a debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.level >= LogLevelDebug {
		l.log("[DEBUG]", format, args...)
	}
}

// log is the internal logging function
func (l *Logger) log(prefix, format string, args ...interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	msg := fmt.Sprintf(format, args...)

	if l.includeTimestamps {
		ts := time.Now().Format("15:04:05.000")
		fmt.Printf("[%s] %s %s\n", ts, prefix, msg)
	} else {
		fmt.Printf("%s %s\n", prefix, msg)
	}
}

// DumpBinary logs raw binary data in hex format
func (l *Logger) DumpBinary(label string, data []byte) {
	if !l.IsVerbose(LogLevelDebug) {
		return
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.dumpBinaryInternal(label, data, 0, len(data))
}

// DumpBinaryRange logs a portion of binary data
func (l *Logger) DumpBinaryRange(label string, data []byte, offset, length int) {
	if !l.IsVerbose(LogLevelDebug) {
		return
	}

	if offset+length > len(data) {
		length = len(data) - offset
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.dumpBinaryInternal(label, data, offset, length)
}

// dumpBinaryInternal performs the actual binary dump
func (l *Logger) dumpBinaryInternal(label string, data []byte, offset, length int) {
	if len(data) == 0 {
		fmt.Printf("[DEBUG] %s (empty)\n", label)
		return
	}

	fmt.Printf("[DEBUG] %s (%d bytes)\n", label, length)
	fmt.Printf("[DEBUG] %s\n", strings.Repeat("─", 60))

	// Print hex dump with ASCII
	const bytesPerLine = 16
	for i := 0; i < length; i += bytesPerLine {
		// Address
		fmt.Printf("[DEBUG] %08x: ", offset+i)

		// Hex values
		hexPart := ""
		asciiPart := ""

		for j := 0; j < bytesPerLine && i+j < length; j++ {
			b := data[offset+i+j]
			hexPart += fmt.Sprintf("%02x ", b)

			// ASCII representation
			if b >= 32 && b < 127 {
				asciiPart += string(b)
			} else {
				asciiPart += "."
			}
		}

		// Pad hex
		for j := (length - i); j < bytesPerLine; j++ {
			hexPart += "   "
		}

		fmt.Printf("%-*s %s\n", bytesPerLine*3, hexPart, asciiPart)
	}

	fmt.Printf("[DEBUG] %s\n", strings.Repeat("─", 60))
}

// DumpPOD logs a POD structure in a human-readable format
func (l *Logger) DumpPOD(label string, podType uint32, podData []byte) {
	if !l.IsVerbose(LogLevelDebug) {
		return
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	fmt.Printf("[DEBUG] POD: %s\n", label)
	fmt.Printf("[DEBUG]   Type: 0x%02x (%s)\n", podType, typeNameString(podType))
	fmt.Printf("[DEBUG]   Size: %d bytes\n", len(podData))
	fmt.Printf("[DEBUG]   Data: %s\n", hex.EncodeToString(podData[:min(len(podData), 64)]))
}

// OnSend registers a callback for outgoing messages
func (l *Logger) OnSend(callback func(objID uint32, methodID uint32, data []byte)) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.onSendCallback = callback
}

// OnReceive registers a callback for incoming messages
func (l *Logger) OnReceive(callback func(objID uint32, eventID uint32, data []byte)) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.onRecvCallback = callback
}

// OnError registers a callback for errors
func (l *Logger) OnError(callback func(err error, context string)) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.onErrorCallback = callback
}

// ReportSend reports an outgoing message
func (l *Logger) ReportSend(objID uint32, methodID uint32, data []byte) {
	l.mutex.Lock()
	callback := l.onSendCallback
	l.mutex.Unlock()

	if callback != nil {
		callback(objID, methodID, data)
	}
}

// ReportReceive reports an incoming message
func (l *Logger) ReportReceive(objID uint32, eventID uint32, data []byte) {
	l.mutex.Lock()
	callback := l.onRecvCallback
	l.mutex.Unlock()

	if callback != nil {
		callback(objID, eventID, data)
	}
}

// ReportError reports an error
func (l *Logger) ReportError(err error, context string) {
	l.mutex.Lock()
	callback := l.onErrorCallback
	l.mutex.Unlock()

	if callback != nil {
		callback(err, context)
	}
}

// typeNameString returns human-readable POD type name
func typeNameString(podType uint32) string {
	typeNames := map[uint32]string{
		0x00: "None",
		0x01: "Bool",
		0x02: "Id",
		0x03: "Int",
		0x04: "Long",
		0x05: "Float",
		0x06: "Double",
		0x07: "String",
		0x08: "Bytes",
		0x09: "Rectangle",
		0x0a: "Fraction",
		0x0b: "Bitmap",
		0x0c: "Array",
		0x0d: "Struct",
		0x0e: "Object",
		0x0f: "Choice",
		0x10: "Pointer",
		0x11: "Fd",
		0x12: "Sequence",
	}

	if name, ok := typeNames[podType]; ok {
		return name
	}
	return "Unknown"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MessageDumper provides structured logging of PipeWire messages
type MessageDumper struct {
	logger *Logger
}

// NewMessageDumper creates a new message dumper
func NewMessageDumper(logger *Logger) *MessageDumper {
	return &MessageDumper{logger: logger}
}

// DumpMethodCall logs a method call in detail
func (d *MessageDumper) DumpMethodCall(objectID uint32, methodName string, methodID uint32, args []byte) {
	if !d.logger.IsVerbose(LogLevelDebug) {
		return
	}

	d.logger.mutex.Lock()
	defer d.logger.mutex.Unlock()

	fmt.Printf("[DEBUG] ▶ Method Call\n")
	fmt.Printf("[DEBUG]   Object: %d\n", objectID)
	fmt.Printf("[DEBUG]   Method: %s (id=%d)\n", methodName, methodID)
	fmt.Printf("[DEBUG]   Args Size: %d bytes\n", len(args))

	if len(args) > 0 {
		fmt.Printf("[DEBUG]   Args Data:\n")
		d.logger.dumpBinaryInternal("args", args, 0, len(args))
	}
}

// DumpEvent logs a server event in detail
func (d *MessageDumper) DumpEvent(objectID uint32, eventName string, eventID uint32, args []byte) {
	if !d.logger.IsVerbose(LogLevelDebug) {
		return
	}

	d.logger.mutex.Lock()
	defer d.logger.mutex.Unlock()

	fmt.Printf("[DEBUG] ◀ Event\n")
	fmt.Printf("[DEBUG]   Object: %d\n", objectID)
	fmt.Printf("[DEBUG]   Event: %s (id=%d)\n", eventName, eventID)
	fmt.Printf("[DEBUG]   Args Size: %d bytes\n", len(args))

	if len(args) > 0 {
		fmt.Printf("[DEBUG]   Args Data:\n")
		d.logger.dumpBinaryInternal("args", args, 0, len(args))
	}
}

// DumpObject logs an object's properties
func (d *MessageDumper) DumpObject(objectID uint32, objectType string, properties map[string]string) {
	if !d.logger.IsVerbose(LogLevelInfo) {
		return
	}

	d.logger.mutex.Lock()
	defer d.logger.mutex.Unlock()

	fmt.Printf("[DEBUG] Object: %s (id=%d)\n", objectType, objectID)
	for key, value := range properties {
		fmt.Printf("[DEBUG]   %s: %s\n", key, value)
	}
}
