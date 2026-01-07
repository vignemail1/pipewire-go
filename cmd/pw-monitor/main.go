// pw-monitor - Monitor PipeWire graph changes
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/vignemail1/pipewire-go/client"
)

func main() {
	var (
		followEvents = flag.Bool("follow", false, "Keep monitoring (Ctrl+C to stop)")
	)
	flag.Parse()

	// Connect to PipeWire
	c, err := client.NewClient("pw-monitor")
	if err != nil {
		log.Fatalf("Failed to connect to PipeWire: %v", err)
	}
	defer c.Disconnect()

	registry := c.GetRegistry()

	// Display initial state
	fmt.Println("=== Initial PipeWire State ===")
	fmt.Printf("Total objects: %d\n", registry.CountObjects())
	fmt.Printf("Nodes: %d\n", registry.Count("Node"))
	fmt.Printf("Ports: %d\n", registry.Count("Port"))
	fmt.Printf("Links: %d\n", registry.Count("Link"))
	fmt.Println()

	if !*followEvents {
		return
	}

	// Monitor events
	fmt.Println("=== Monitoring Events ===")
	fmt.Println("Press Ctrl+C to stop...\n")

	registry.OnGlobalAdded(func(obj *client.GlobalObject) {
		ts := time.Now().Format("15:04:05")
		name := obj.Props["node.name"]
		if name == "" {
			name = obj.Props["port.name"]
		}

		if name == "" {
			name = "(unnamed)"
		}

		fmt.Printf("[%s] NEW %s: %s (ID: %d)\n", ts, obj.Type, name, obj.ID)
	})

	// Keep running
	select {}
}
