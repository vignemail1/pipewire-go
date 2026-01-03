// examples/list_devices.go
// List all audio devices and their ports

package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/vignemail1/pipewire-go/client"
	"github.com/vignemail1/pipewire-go/verbose"
)

func main() {
	var (
		verbose_ = flag.Bool("v", false, "Verbose output")
		json     = flag.Bool("json", false, "JSON output")
		detailed = flag.Bool("d", false, "Detailed information")
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

	// Ping to verify connection
	if err := c.Ping(); err != nil {
		log.Fatalf("Failed to ping daemon: %v", err)
	}

	// List devices
	if *json {
		listDevicesJSON(c)
	} else if *detailed {
		listDevicesDetailed(c)
	} else {
		listDevices(c)
	}
}

func listDevices(c *client.Client) {
	nodes := c.GetNodes()

	if len(nodes) == 0 {
		fmt.Println("No devices found")
		return
	}

	fmt.Printf("Audio Devices (%d total)\n", len(nodes))
	fmt.Println("=" * 80)

	for _, node := range nodes {
		fmt.Printf("[%2d] %s\n", node.ID, node.Name())
		fmt.Printf("     Type: %s\n", node.info.MediaClass)
		fmt.Printf("     Dir:  %s\n", node.GetDirection())
		fmt.Printf("     State: %s\n", node.GetState())
		if sr := node.GetSampleRate(); sr > 0 {
			fmt.Printf("     Rate: %d Hz\n", sr)
		}
		if ch := node.GetChannels(); ch > 0 {
			fmt.Printf("     Chan: %d\n", ch)
		}

		// List ports
		ports := node.GetPorts()
		if len(ports) > 0 {
			fmt.Printf("     Ports: %d\n", len(ports))
			for _, port := range ports {
				dir := "→"
				if port.Direction == client.PortDirectionInput {
					dir = "←"
				}
				fmt.Printf("       %s [%2d] %s\n", dir, port.ID, port.Name)
			}
		}

		fmt.Println()
	}
}

func listDevicesDetailed(c *client.Client) {
	nodes := c.GetNodes()

	fmt.Printf("Audio Devices Detailed (%d total)\n", len(nodes))
	fmt.Println("=" * 80)

	for _, node := range nodes {
		fmt.Printf("\nDevice [%d]: %s\n", node.ID, node.Name())
		fmt.Println("Properties:")
		for k, v := range node.GetProperties() {
			fmt.Printf("  %s = %s\n", k, v)
		}

		ports := node.GetPorts()
		if len(ports) > 0 {
			fmt.Printf("\nPorts (%d):\n", len(ports))
			for _, port := range ports {
				fmt.Printf("  [%d] %s\n", port.ID, port.Name)
				fmt.Printf("    Direction: %s\n", port.Direction.String())
				fmt.Printf("    Type: %s\n", port.Type.String())
				fmt.Printf("    Connected: %v\n", port.IsConnected())

				if len(port.GetProperties()) > 0 {
					fmt.Printf("    Properties:\n")
					for k, v := range port.GetProperties() {
						fmt.Printf("      %s = %s\n", k, v)
					}
				}
			}
		}
	}
}

func listDevicesJSON(c *client.Client) {
	nodes := c.GetNodes()

	fmt.Println("[")
	for i, node := range nodes {
		fmt.Printf("  {\n")
		fmt.Printf("    \"id\": %d,\n", node.ID)
		fmt.Printf("    \"name\": \"%s\",\n", node.Name())
		fmt.Printf("    \"type\": \"%s\",\n", node.info.MediaClass)
		fmt.Printf("    \"direction\": \"%s\",\n", node.GetDirection())
		fmt.Printf("    \"state\": \"%s\",\n", node.GetState())
		fmt.Printf("    \"rate\": %d,\n", node.GetSampleRate())
		fmt.Printf("    \"channels\": %d,\n", node.GetChannels())

		ports := node.GetPorts()
		fmt.Printf("    \"ports\": [\n")
		for j, port := range ports {
			fmt.Printf("      {\"id\": %d, \"name\": \"%s\", \"direction\": \"%s\"", port.ID, port.Name, port.Direction.String())
			if j < len(ports)-1 {
				fmt.Printf(",")
			}
			fmt.Printf("\n")
		}
		fmt.Printf("    ]\n")
		fmt.Printf("  }")
		if i < len(nodes)-1 {
			fmt.Printf(",")
		}
		fmt.Printf("\n")
	}
	fmt.Println("]")
}
