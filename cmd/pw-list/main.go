// pw-list - List PipeWire objects
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/vignemail1/pipewire-go/client"
)

func main() {
	var (
		listNodes   = flag.Bool("nodes", false, "List all nodes")
		listPorts   = flag.Bool("ports", false, "List all ports")
		listLinks   = flag.Bool("links", false, "List all links")
		listAll     = flag.Bool("all", false, "List all objects")
		asJSON      = flag.Bool("json", false, "Output as JSON")
		filter      = flag.String("filter", "", "Filter by name (contains)")
		showProps   = flag.Bool("props", false, "Show properties")
	)
	flag.Parse()

	// Connect to PipeWire
	c, err := client.NewClient("pw-list")
	if err != nil {
		log.Fatalf("Failed to connect to PipeWire: %v", err)
	}
	defer c.Disconnect()

	registry := c.GetRegistry()

	// Determine what to list
	if !*listNodes && !*listPorts && !*listLinks && !*listAll {
		*listAll = true
	}

	var objects []*client.GlobalObject

	if *listAll {
		objects = registry.ListAll()
	} else if *listNodes {
		objects = registry.ListNodes()
	} else if *listPorts {
		objects = registry.ListPorts()
	} else if *listLinks {
		objects = registry.ListLinks()
	}

	// Apply filter
	if *filter != "" {
		objects = filterObjects(objects, *filter)
	}

	// Output
	if *asJSON {
		outputJSON(objects)
	} else {
		outputText(objects, *showProps)
	}
}

func filterObjects(objects []*client.GlobalObject, pattern string) []*client.GlobalObject {
	pattern = strings.ToLower(pattern)
	var filtered []*client.GlobalObject

	for _, obj := range objects {
		name := strings.ToLower(obj.Props["node.name"])
		if strings.Contains(name, pattern) {
			filtered = append(filtered, obj)
			continue
		}

		name = strings.ToLower(obj.Props["port.name"])
		if strings.Contains(name, pattern) {
			filtered = append(filtered, obj)
		}
	}

	return filtered
}

func outputText(objects []*client.GlobalObject, showProps bool) {
	if len(objects) == 0 {
		fmt.Println("No objects found")
		return
	}

	for _, obj := range objects {
		fmt.Printf("%d\t%s\tv%d\n", obj.ID, obj.Type, obj.Version)

		if name, ok := obj.Props["node.name"]; ok {
			fmt.Printf("  Name: %s\n", name)
		}
		if direction, ok := obj.Props["port.direction"]; ok {
			fmt.Printf("  Direction: %s\n", direction)
		}

		if showProps {
			fmt.Println("  Properties:")
			for k, v := range obj.Props {
				fmt.Printf("    %s: %s\n", k, v)
			}
		}
		fmt.Println()
	}
}

func outputJSON(objects []*client.GlobalObject) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(objects); err != nil {
		log.Fatalf("Failed to encode JSON: %v", err)
	}
}
