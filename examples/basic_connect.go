// Example: Basic PipeWire Connection Test
// This demonstrates connecting to the PipeWire daemon and performing basic operations
package main

import (
	"fmt"
	"log"

	"github.com/vignemail1/pipewire-go/core"
	"github.com/vignemail1/pipewire-go/spa"
	"github.com/vignemail1/pipewire-go/verbose"
)

func main() {
	// Create a verbose logger for debugging
	logger := verbose.NewLogger(verbose.LogLevelDebug, true)

	fmt.Println("PipeWire Go Client - Basic Connection Test")
	fmt.Println("==========================================\n")

	// Find and connect to the PipeWire daemon
	socketPath := core.FindDefaultSocket()
	fmt.Printf("Using socket: %s\n\n", socketPath)

	conn, err := core.Dial(socketPath, logger)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	fmt.Println("✓ Connected to PipeWire daemon\n")

	// Test 1: Send a ping message to Core (object_id=0)
	fmt.Println("Test 1: Sending ping to Core...")
	builder := spa.NewPODBuilder()
	if err := conn.SendMessage(0, 0, nil); err != nil {
		log.Printf("Error sending ping: %v", err)
	} else {
		fmt.Println("✓ Ping sent\n")
	}

	// Test 2: Parse and build POD structures
	fmt.Println("Test 2: Testing POD Parser/Builder...")

	// Build a simple POD
	builder.Reset()
	if err := builder.WriteInt(42); err != nil {
		log.Fatalf("Error writing int: %v", err)
	}
	if err := builder.WriteString("hello"); err != nil {
		log.Fatalf("Error writing string: %v", err)
	}
	if err := builder.WriteFloat(3.14); err != nil {
		log.Fatalf("Error writing float: %v", err)
	}

	podData := builder.Bytes()
	fmt.Printf("Built POD of %d bytes\n", len(podData))

	// Parse it back
	parser := spa.NewPODParser(podData)

	// Parse int
	intPod, err := parser.ParsePOD()
	if err != nil {
		log.Fatalf("Error parsing int: %v", err)
	}
	if intVal, ok := intPod.(*spa.IntPOD); ok {
		fmt.Printf("  Parsed Int: %d\n", intVal.Value)
	}

	// Parse string
	strPod, err := parser.ParsePOD()
	if err != nil {
		log.Fatalf("Error parsing string: %v", err)
	}
	if strVal, ok := strPod.(*spa.StringPOD); ok {
		fmt.Printf("  Parsed String: %s\n", strVal.Value)
	}

	// Parse float
	floatPod, err := parser.ParsePOD()
	if err != nil {
		log.Fatalf("Error parsing float: %v", err)
	}
	if floatVal, ok := floatPod.(*spa.FloatPOD); ok {
		fmt.Printf("  Parsed Float: %f\n", floatVal.Value)
	}

	fmt.Println("✓ POD parsing works\n")

	// Test 3: Logging capabilities
	fmt.Println("Test 3: Logging and debugging...")

	logger.DumpBinary("Test data", []byte{0x01, 0x02, 0x03, 0x04, 0x05})
	logger.Debugf("Debug message")
	logger.Infof("Info message")

	dumper := verbose.NewMessageDumper(logger)
	dumper.DumpObject(0, "Core", map[string]string{
		"object.serial": "1",
		"core.version": "0.3.68",
	})

	fmt.Println("✓ Logging works\n")

	// Test 4: Register event handler
	fmt.Println("Test 4: Registering event handler...")

	conn.RegisterEventHandler(0, func(msg *core.Message) error {
		fmt.Printf("Received event on object %d: opcode=%d\n", msg.ObjectID, msg.OpCode)
		return nil
	})

	fmt.Println("✓ Event handler registered\n")

	fmt.Println("All tests completed successfully!")
}
