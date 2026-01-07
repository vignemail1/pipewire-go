// pw-connect - Create and manage links
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/vignemail1/pipewire-go/client"
)

func main() {
	var (
		disconnect = flag.Bool("disconnect", false, "Disconnect (remove link)")
	)
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: pw-connect [OPTIONS] <output-port> <input-port>\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Parse port IDs
	output, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		log.Fatalf("Invalid output port ID: %v", err)
	}

	input, err := strconv.ParseUint(args[1], 10, 32)
	if err != nil {
		log.Fatalf("Invalid input port ID: %v", err)
	}

	// Connect to PipeWire
	c, err := client.NewClient("pw-connect")
	if err != nil {
		log.Fatalf("Failed to connect to PipeWire: %v", err)
	}
	defer c.Disconnect()

	if *disconnect {
		fmt.Printf("Disconnecting port %d -> %d...\n", output, input)
		fmt.Println("[Not yet implemented]")
	} else {
		fmt.Printf("Connecting port %d -> %d...\n", output, input)
		fmt.Println("[Not yet implemented]")
	}
}
