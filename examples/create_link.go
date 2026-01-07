// Example: Create Audio Link
// This example demonstrates how to:
// - Connect to the PipeWire daemon
// - Find specific nodes by name
// - Select compatible ports
// - Create an audio link between ports
// - Handle link creation errors
//
// Usage: go run examples/create_link.go -source "alsa_output.pci-0000_00_1f.3.analog-stereo" -sink "alsa_input.usb-0123_USB_Audio-00.mono-fallback"

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/vignemail1/pipewire-go/client"
	"github.com/vignemail1/pipewire-go/verbose"
)

func main() {
	socketPath := flag.String("socket", "/run/user/1000/pipewire-0", "PipeWire socket path")
	sourceName := flag.String("source", "alsa_output.pci-0000_00_1f.3.analog-stereo", "Source node name")
	sinkName := flag.String("sink", "", "Sink node name (optional)")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	// Create logger
	logLevel := verbose.LogLevelInfo
	if *verbose {
		logLevel = verbose.LogLevelDebug
	}
	logger := verbose.NewLogger(logLevel, true)

	// Connect to PipeWire daemon
	logger.Infof("Connecting to PipeWire daemon at %s", *socketPath)
	conn, err := client.NewClient(*socketPath, logger)
	if err != nil {
		logger.Errorf("Failed to connect: %v", err)
		fmt.Fprintf(os.Stderr, "Error: Could not connect to PipeWire daemon\n")
		os.Exit(1)
	}
	defer conn.Close()

	// Wait for connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.WaitUntilReady(ctx); err != nil {
		logger.Errorf("Connection not ready: %v", err)
		fmt.Fprintf(os.Stderr, "Error: PipeWire connection failed\n")
		os.Exit(1)
	}

	logger.Infof("Connected to PipeWire daemon")

	// Find source node by name
	fmt.Printf("\n=== Creating Audio Link ===\n\n")
	fmt.Printf("Searching for source node: %s\n", *sourceName)

	nodes := conn.GetNodes()
	var sourceNode *client.Node
	for _, node := range nodes {
		if node.Name() == *sourceName {
			sourceNode = node
			break
		}
	}

	if sourceNode == nil {
		fmt.Fprintf(os.Stderr, "Error: Source node '%s' not found\n", *sourceName)
		fmt.Println("\nAvailable nodes:")
		for _, node := range nodes {
			fmt.Printf("  - %s (ID: %d)\n", node.Name(), node.ID)
		}
		os.Exit(1)
	}

	fmt.Printf("✓ Found source node: %s\n", sourceNode.Name())

	// Get output ports from source
	outputPorts := sourceNode.GetPortsByDirection(client.PortDirectionOutput)
	if len(outputPorts) == 0 {
		fmt.Fprintf(os.Stderr, "Error: Source node has no output ports\n")
		os.Exit(1)
	}

	fmt.Printf("✓ Found %d output ports\n", len(outputPorts))

	// Find sink node
	var sinkNode *client.Node

	if *sinkName != "" {
		// Search by specific name
		fmt.Printf("\nSearching for sink node: %s\n", *sinkName)
		for _, node := range nodes {
			if node.Name() == *sinkName {
				sinkNode = node
				break
			}
		}

		if sinkNode == nil {
			fmt.Fprintf(os.Stderr, "Error: Sink node '%s' not found\n", *sinkName)
			os.Exit(1)
		}
	} else {
		// Find first compatible sink
		fmt.Println("\nSearching for compatible sink node...")
		for _, node := range nodes {
			if node.ID == sourceNode.ID {
				continue // Skip source node
			}

			inputPorts := node.GetPortsByDirection(client.PortDirectionInput)
			if len(inputPorts) > 0 {
				sinkNode = node
				break
			}
		}
	}

	if sinkNode == nil {
		fmt.Fprintf(os.Stderr, "Error: No compatible sink node found\n")
		os.Exit(1)
	}

	fmt.Printf("✓ Found sink node: %s\n", sinkNode.Name())

	// Get input ports from sink
	inputPorts := sinkNode.GetPortsByDirection(client.PortDirectionInput)
	if len(inputPorts) == 0 {
		fmt.Fprintf(os.Stderr, "Error: Sink node has no input ports\n")
		os.Exit(1)
	}

	fmt.Printf("✓ Found %d input ports\n", len(inputPorts))

	// Find compatible ports
	fmt.Println("\nSearching for compatible port pair...")
	var selectedOutput, selectedInput *client.Port

	for _, outPort := range outputPorts {
		for _, inPort := range inputPorts {
			// Check if ports can connect
			if outPort.CanConnectTo(inPort) {
				// Check format compatibility
				supported := inPort.GetSupportedFormats()
				if len(supported) > 0 {
					selectedOutput = outPort
					selectedInput = inPort
					break
				}
			}
		}
		if selectedOutput != nil {
			break
		}
	}

	if selectedOutput == nil || selectedInput == nil {
		fmt.Fprintf(os.Stderr, "Error: No compatible port pair found\n")
		os.Exit(1)
	}

	fmt.Printf("✓ Compatible ports found\n")
	fmt.Printf("  Output: %s\n", selectedOutput.Name())
	fmt.Printf("  Input: %s\n", selectedInput.Name())

	// Create link
	fmt.Println("\nCreating audio link...")

	linkParams := &client.LinkParams{
		Properties: make(map[string]string),
	}

	link, err := conn.CreateLink(selectedOutput, selectedInput, linkParams)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create link: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Link created successfully!\n")
	fmt.Printf("  Link ID: %d\n", link.ID())
	fmt.Printf("  Source: %s → %s\n", sourceNode.Name(), selectedOutput.Name())
	fmt.Printf("  Destination: %s → %s\n", sinkNode.Name(), selectedInput.Name())
	fmt.Printf("\nAudio is now routed from source to sink.\n")
}
