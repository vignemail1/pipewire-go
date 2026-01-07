// Example: Monitor Audio Graph
// This example demonstrates how to:
// - Connect to the PipeWire daemon
// - Register event listeners
// - Display events as they occur
// - Monitor node and link changes
//
// Usage: go run examples/monitor_graph.go [-duration=30s]

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vignemail1/pipewire-go/client"
	"github.com/vignemail1/pipewire-go/verbose"
)

func main() {
	socketPath := flag.String("socket", "/run/user/1000/pipewire-0", "PipeWire socket path")
	duration := flag.Duration("duration", 0, "Monitor duration (0 = infinite)")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	// Create logger
	logLevel := verbose.LogLevelInfo
	if *verbose {
		logLevel = verbose.LogLevelDebug
	}
	logger := verbose.NewLogger(logLevel, true)

	// Connect to PipeWire daemon
	fmt.Printf("\n=== PipeWire Audio Graph Monitor ===\n\n")
	fmt.Printf("Connecting to %s...\n", *socketPath)

	conn, err := client.NewClient(*socketPath, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not connect to PipeWire daemon\n")
		fmt.Fprintf(os.Stderr, "Make sure PipeWire is running and socket is at: %s\n", *socketPath)
		os.Exit(1)
	}
	defer conn.Close()

	// Wait for connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.WaitUntilReady(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: PipeWire connection failed\n")
		os.Exit(1)
	}

	fmt.Println("✓ Connected to PipeWire daemon")

	// Print initial state
	nodes := conn.GetNodes()
	links := conn.GetLinks()
	fmt.Printf("\nInitial state: %d nodes, %d links\n\n", len(nodes), len(links))

	// Register event listeners
	fmt.Println("Listening for events... (press Ctrl+C to stop)\n")

	// Create a channel for events
	eventChan := make(chan string, 100)

	// Register node event listener
	if err := conn.RegisterEventListener(client.EventTypeNodeAdded, func(event client.Event) error {
		eventChan <- fmt.Sprintf("[%s] Node added: %v", time.Now().Format("15:04:05"), event.Data)
		return nil
	}); err != nil {
		logger.Warnf("Could not register NodeAdded listener: %v", err)
	}

	if err := conn.RegisterEventListener(client.EventTypeNodeRemoved, func(event client.Event) error {
		eventChan <- fmt.Sprintf("[%s] Node removed: %v", time.Now().Format("15:04:05"), event.Data)
		return nil
	}); err != nil {
		logger.Warnf("Could not register NodeRemoved listener: %v", err)
	}

	// Register port event listeners
	if err := conn.RegisterEventListener(client.EventTypePortAdded, func(event client.Event) error {
		eventChan <- fmt.Sprintf("[%s] Port added: %v", time.Now().Format("15:04:05"), event.Data)
		return nil
	}); err != nil {
		logger.Warnf("Could not register PortAdded listener: %v", err)
	}

	if err := conn.RegisterEventListener(client.EventTypePortRemoved, func(event client.Event) error {
		eventChan <- fmt.Sprintf("[%s] Port removed: %v", time.Now().Format("15:04:05"), event.Data)
		return nil
	}); err != nil {
		logger.Warnf("Could not register PortRemoved listener: %v", err)
	}

	// Register link event listeners
	if err := conn.RegisterEventListener(client.EventTypeLinkAdded, func(event client.Event) error {
		eventChan <- fmt.Sprintf("[%s] ✓ Link created: %v", time.Now().Format("15:04:05"), event.Data)
		return nil
	}); err != nil {
		logger.Warnf("Could not register LinkAdded listener: %v", err)
	}

	if err := conn.RegisterEventListener(client.EventTypeLinkRemoved, func(event client.Event) error {
		eventChan <- fmt.Sprintf("[%s] ✗ Link destroyed: %v", time.Now().Format("15:04:05"), event.Data)
		return nil
	}); err != nil {
		logger.Warnf("Could not register LinkRemoved listener: %v", err)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Setup timeout if duration specified
	var timeoutChan <-chan time.Time
	if *duration > 0 {
		time.AfterFunc(*duration, func() {
			sigChan <- syscall.SIGTERM
		})
	}

	// Event loop
	monitorStart := time.Now()
	eventCount := 0

	for {
		select {
		case event := <-eventChan:
			fmt.Println(event)
			eventCount++

		case <-sigChan:
			fmt.Printf("\n\nMonitoring stopped.\n")
			elapsed := time.Since(monitorStart)
			fmt.Printf("Duration: %v\n", elapsed)
			fmt.Printf("Events captured: %d\n", eventCount)

			// Print final state
			nodes := conn.GetNodes()
			links := conn.GetLinks()
			fmt.Printf("Final state: %d nodes, %d links\n", len(nodes), len(links))

			return

		case <-timeoutChan:
			fmt.Printf("\n\nMonitoring timeout reached.\n")
			return
	}
	}
}
