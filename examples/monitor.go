// examples/monitor.go
// Real-time monitoring of PipeWire audio graph

package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/vignemail1/pipewire-go/client"
	"github.com/vignemail1/pipewire-go/verbose"
)

func main() {
	var (
		interval = flag.Duration("interval", 5*time.Second, "Update interval")
		verbose_ = flag.Bool("v", false, "Verbose output")
		watch    = flag.String("watch", "all", "Watch: all, nodes, ports, links")
	)
	flag.Parse()

	// Create logger
	level := verbose.LogLevelInfo
	if *verbose_ {
		level = verbose.LogLevelDebug
	}
	logger := verbose.NewLogger(level, false)

	// Create client
	c, err := client.NewClient("/run/pipewire-0", logger)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	// Register event listeners
	if *watch == "all" || *watch == "nodes" {
		c.On(client.EventNodeAdded, func(e *client.Event) {
			if node, ok := e.Object.(*client.Node); ok {
				fmt.Printf("[EVENT] Node added: %s\n", node.Name())
			}
		})
		c.On(client.EventNodeRemoved, func(e *client.Event) {
			fmt.Printf("[EVENT] Node removed: ID=%d\n", e.Object)
		})
	}

	if *watch == "all" || *watch == "ports" {
		c.On(client.EventPortAdded, func(e *client.Event) {
			if port, ok := e.Object.(*client.Port); ok {
				fmt.Printf("[EVENT] Port added: %s\n", port.Name)
			}
		})
		c.On(client.EventPortRemoved, func(e *client.Event) {
			fmt.Printf("[EVENT] Port removed: ID=%d\n", e.Object)
		})
	}

	if *watch == "all" || *watch == "links" {
		c.On(client.EventLinkAdded, func(e *client.Event) {
			if link, ok := e.Object.(*client.Link); ok {
				fmt.Printf("[EVENT] Link added: %s\n", link.String())
			}
		})
		c.On(client.EventLinkRemoved, func(e *client.Event) {
			fmt.Printf("[EVENT] Link removed: ID=%d\n", e.Object)
		})
	}

	// Initial state
	printGraphState(c)

	// Monitor loop
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Print("\033[2J\033[H") // Clear screen
			printGraphState(c)
		}
	}
}

func printGraphState(c *client.Client) {
	stats := c.GetStatistics()
	
	fmt.Printf("PipeWire Audio Graph - %s\n", time.Now().Format("15:04:05"))
	fmt.Println("=" * 80)
	fmt.Printf("Nodes: %d | Ports: %d | Links: %d\n\n", stats["nodes"], stats["ports"], stats["links"])

	// Group nodes by direction
	nodes := c.GetNodes()
	
	var inputs, outputs []*client.Node
	for _, node := range nodes {
		dir := node.GetDirection()
		if dir == client.NodeDirectionCapture {
			inputs = append(inputs, node)
		} else if dir == client.NodeDirectionPlayback {
			outputs = append(outputs, node)
		}
	}

	// Display input devices (sources)
	if len(inputs) > 0 {
		fmt.Printf("INPUT DEVICES (%d)\n", len(inputs))
		fmt.Println("-" * 80)
		for _, node := range inputs {
			displayNode(node)
		}
		fmt.Println()
	}

	// Display output devices (sinks)
	if len(outputs) > 0 {
		fmt.Printf("OUTPUT DEVICES (%d)\n", len(outputs))
		fmt.Println("-" * 80)
		for _, node := range outputs {
			displayNode(node)
		}
		fmt.Println()
	}

	// Display connections
	links := c.GetLinks()
	if len(links) > 0 {
		fmt.Printf("CONNECTIONS (%d)\n", len(links))
		fmt.Println("-" * 80)
		for _, link := range links {
			outputName := "?"
			inputName := "?"
			if link.Output != nil && link.Output.ParentNode != nil {
				outputName = fmt.Sprintf("%s:%s", link.Output.ParentNode.Name(), link.Output.Name)
			}
			if link.Input != nil && link.Input.ParentNode != nil {
				inputName = fmt.Sprintf("%s:%s", link.Input.ParentNode.Name(), link.Input.Name)
			}

			status := "●"
			if !link.IsActive() {
				status = "○"
			}
			fmt.Printf("%s %s → %s\n", status, outputName, inputName)
		}
	}
}

func displayNode(node *client.Node) {
	state := ""
	switch node.GetState() {
	case client.NodeStateRunning:
		state = "▶"
	case client.NodeStateIdle:
		state = "⏸"
	case client.NodeStateSuspended:
		state = "⏹"
	case client.NodeStateError:
		state = "✗"
	default:
		state = "?"
	}

	fmt.Printf("%s %s\n", state, node.Name())
	fmt.Printf("   Class: %s\n", node.info.MediaClass)
	if node.GetSampleRate() > 0 {
		fmt.Printf("   Rate: %d Hz / %d ch\n", node.GetSampleRate(), node.GetChannels())
	}

	ports := node.GetPorts()
	if len(ports) > 0 {
		fmt.Printf("   Ports: ")
		for i, port := range ports {
			dir := "→"
			if port.Direction == client.PortDirectionInput {
				dir = "←"
			}
			fmt.Printf("%s%s", dir, port.Name)
			if i < len(ports)-1 {
				fmt.Printf(", ")
			}
		}
		fmt.Println()
	}

	// Show connections for each port
	for _, port := range ports {
		if port.IsConnected() {
			fmt.Printf("   Connected ports: ")
			for _, connPort := range port.GetConnectedPorts() {
				fmt.Printf("%s ", connPort.Name)
			}
			fmt.Println()
		}
	}

	fmt.Println()
}
