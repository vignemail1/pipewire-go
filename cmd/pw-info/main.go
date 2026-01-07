// pw-info - Display detailed object information
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/vignemail1/pipewire-go/client"
)

func main() {
	var (
		asJSON = flag.Bool("json", false, "Output as JSON")
	)
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: pw-info [OPTIONS] <object-id>\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Parse object ID
	id, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		log.Fatalf("Invalid object ID: %v", err)
	}

	// Connect to PipeWire
	c, err := client.NewClient("pw-info")
	if err != nil {
		log.Fatalf("Failed to connect to PipeWire: %v", err)
	}
	defer c.Disconnect()

	registry := c.GetRegistry()
	obj := registry.GetObject(uint32(id))

	if obj == nil {
		fmt.Fprintf(os.Stderr, "Object %d not found\n", id)
		os.Exit(1)
	}

	if *asJSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		encoder.Encode(obj)
	} else {
		displayObject(obj)
	}
}

func displayObject(obj *client.GlobalObject) {
	fmt.Printf("Object ID: %d\n", obj.ID)
	fmt.Printf("Type: %s\n", obj.Type)
	fmt.Printf("Version: %d\n", obj.Version)
	fmt.Println()

	fmt.Println("Properties:")
	for key, value := range obj.Props {
		fmt.Printf("  %s = %s\n", key, value)
	}
}
