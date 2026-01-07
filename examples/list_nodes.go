// Example: List Nodes and Ports
// This example demonstrates how to:
// - Connect to the PipeWire daemon
// - Enumerate all nodes
// - Display node properties and capabilities
// - List ports for each node
//
// Usage: go run examples/list_nodes.go

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/vignemail1/pipewire-go/client"
	"github.com/vignemail1/pipewire-go/verbose"
)

func main() {
	socketPath := flag.String("socket", "/run/user/1000/pipewire-0", "PipeWire socket path")
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
		fmt.Fprintf(os.Stderr, "Make sure PipeWire is running and socket is at: %s\n", *socketPath)
		os.Exit(1)
	}
	defer conn.Close()

	// Wait for connection to be ready
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.WaitUntilReady(ctx); err != nil {
		logger.Errorf("Connection not ready: %v", err)
		fmt.Fprintf(os.Stderr, "Error: PipeWire connection failed to initialize\n")
		os.Exit(1)
	}

	logger.Infof("Connected to PipeWire daemon")

	// Enumerate all nodes
	nodes := conn.GetNodes()
	if len(nodes) == 0 {
		fmt.Println("No nodes found in PipeWire")
		return
	}

	fmt.Printf("\n=== PipeWire Audio Nodes ===\n\n")
	fmt.Printf("Total Nodes: %d\n\n", len(nodes))

	// Display each node
	for i, node := range nodes {
		fmt.Printf("[%d] %s\n", i+1, node.Name())
		fmt.Printf("    ID: %d\n", node.ID)
		fmt.Printf("    Description: %s\n", node.Description())
		fmt.Printf("    Direction: %s\n", node.GetDirection())
		fmt.Printf("    State: %s\n", node.GetState())

		// Display audio properties if available
		if rate := node.GetSampleRate(); rate > 0 {
			fmt.Printf("    Sample Rate: %d Hz\n", rate)
		}
		if channels := node.GetChannels(); channels > 0 {
			fmt.Printf("    Channels: %d\n", channels)
		}

		// Display media class if available
		if mediaClass, ok := node.GetProperty("media.class"); ok {
			fmt.Printf("    Media Class: %s\n", mediaClass)
		}

		// Try to query node parameters
		if params, err := node.GetParams(client.ParamIDFormat); err == nil {
			if formatInfo, ok := params.(map[string]interface{}); ok {
				if rate, hasRate := formatInfo["rate"]; hasRate {
					fmt.Printf("    Format Rate: %v Hz\n", rate)
				}
			}
		}

		// List ports for this node
		ports := node.GetPorts()
		if len(ports) > 0 {
			fmt.Printf("\n    Ports: (%d)\n", len(ports))
			for j, port := range ports {
				fmt.Printf("      [%d.%d] %s\n", i+1, j+1, port.Name())
				fmt.Printf("            Direction: %s\n", port.Direction())
				fmt.Printf("            Type: %s\n", port.Type())

				// Display port format information
				if format, err := port.GetFormat(); err == nil && format != nil {
					fmt.Printf("            Current Format: %s\n", format.String())
				}

				// Display supported formats
				supportedFormats := port.GetSupportedFormats()
				if len(supportedFormats) > 0 {
					fmt.Printf("            Supported Formats:\n")
					for _, fmt := range supportedFormats {
						fmt.Printf("              - %s\n", fmt.String())
					}
				}

				// Display connection status
				if port.IsConnected() {
					fmt.Printf("            Status: Connected\n")
				} else {
					fmt.Printf("            Status: Disconnected\n")
				}
			}
		}
		fmt.Println()
	}

	// Display summary
	fmt.Println("\n=== Summary ===")
	fmt.Printf("Nodes: %d\n", len(nodes))
	fmt.Printf("Total Ports: %d\n", len(conn.GetPorts()))
	fmt.Printf("Active Links: %d\n", len(conn.GetLinks()))
}
