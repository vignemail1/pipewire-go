// examples/audio_routing.go
// Create and manage audio links between ports

package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/vignemail1/pipewire-go/client"
	"github.com/vignemail1/pipewire-go/verbose"
)

func main() {
	var (
		action  = flag.String("action", "list", "Action: list, create, remove, info")
		from    = flag.String("from", "", "Source port name (for create)")
		to      = flag.String("to", "", "Destination port name (for create)")
		linkID  = flag.Int("id", 0, "Link ID (for info/remove)")
		verbose_ = flag.Bool("v", false, "Verbose output")
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

	switch *action {
	case "list":
		listLinks(c)
	case "create":
		if *from == "" || *to == "" {
			log.Fatal("create action requires -from and -to options")
		}
		createLink(c, *from, *to)
	case "remove":
		if *linkID == 0 {
			log.Fatal("remove action requires -id option")
		}
		removeLink(c, uint32(*linkID))
	case "info":
		if *linkID == 0 {
			log.Fatal("info action requires -id option")
		}
		linkInfo(c, uint32(*linkID))
	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}

func listLinks(c *client.Client) {
	links := c.GetLinks()

	if len(links) == 0 {
		fmt.Println("No audio links found")
		return
	}

	fmt.Printf("Audio Links (%d total)\n", len(links))
	fmt.Println("=" * 80)

	for _, link := range links {
		outputName := "unknown"
		inputName := "unknown"

		if link.Output != nil {
			outputName = link.Output.Name
		}
		if link.Input != nil {
			inputName = link.Input.Name
		}

		status := "inactive"
		if link.IsActive() {
			status = "active"
		}

		fmt.Printf("[%2d] %s → %s [%s]\n", link.ID, outputName, inputName, status)
	}
}

func createLink(c *client.Client, from string, to string) {
	// Find output port
	var outputPort *client.Port
	for _, port := range c.GetPorts() {
		if strings.Contains(port.Name, from) && port.IsOutput() {
			outputPort = port
			break
		}
	}

	if outputPort == nil {
		log.Fatalf("Output port matching '%s' not found", from)
	}

	// Find input port
	var inputPort *client.Port
	for _, port := range c.GetPorts() {
		if strings.Contains(port.Name, to) && port.IsInput() {
			inputPort = port
			break
		}
	}

	if inputPort == nil {
		log.Fatalf("Input port matching '%s' not found", to)
	}

	// Create link
	fmt.Printf("Creating link: %s → %s\n", outputPort.Name, inputPort.Name)
	link, err := c.CreateLink(outputPort, inputPort)
	if err != nil {
		log.Fatalf("Failed to create link: %v", err)
	}

	fmt.Printf("✓ Link created: [%d] %s → %s\n", link.ID, outputPort.Name, inputPort.Name)
}

func removeLink(c *client.Client, linkID uint32) {
	link := c.GetLink(linkID)
	if link == nil {
		log.Fatalf("Link not found: %d", linkID)
	}

	outputName := "unknown"
	inputName := "unknown"
	if link.Output != nil {
		outputName = link.Output.Name
	}
	if link.Input != nil {
		inputName = link.Input.Name
	}

	fmt.Printf("Removing link: %s → %s\n", outputName, inputName)
	if err := c.RemoveLink(link); err != nil {
		log.Fatalf("Failed to remove link: %v", err)
	}

	fmt.Println("✓ Link removed")
}

func linkInfo(c *client.Client, linkID uint32) {
	link := c.GetLink(linkID)
	if link == nil {
		log.Fatalf("Link not found: %d", linkID)
	}

	fmt.Printf("Link [%d]\n", link.ID)
	fmt.Printf("Output Port: %d\n", link.OutputPort)
	if link.Output != nil {
		fmt.Printf("  %s (Node %d)\n", link.Output.Name, link.Output.NodeID)
	}

	fmt.Printf("Input Port: %d\n", link.InputPort)
	if link.Input != nil {
		fmt.Printf("  %s (Node %d)\n", link.Input.Name, link.Input.NodeID)
	}

	fmt.Printf("Status: %v\n", link.IsActive())
	fmt.Printf("Properties:\n")
	for k, v := range link.GetProperties() {
		fmt.Printf("  %s = %s\n", k, v)
	}
}
