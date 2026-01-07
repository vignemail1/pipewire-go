// Example: Error handling patterns
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/vignemail1/pipewire-go"
	"github.com/vignemail1/pipewire-go/client"
)

func main() {
	fmt.Println("=== Error Handling Examples ===")

	// Example 1: Connection errors
	demoConnectionErrors()

	// Example 2: Validation errors
	demoValidationErrors()

	// Example 3: Using error.Is pattern
	demoErrorIsPattern()

	// Example 4: Using error.As pattern
	demoErrorAsPattern()
}

func demoConnectionErrors() {
	fmt.Println("\n--- Connection Error Example ---")

	// Simulate connection error
	err := &pipewireio.ConnectionError{
		Reason:  "socket not available",
		Address: "pipewire-0",
	}

	fmt.Printf("Error: %v\n", err)

	// Check if it's a connection error
	if errors.Is(err, &pipewireio.ConnectionError{}) {
		fmt.Println("This is a connection error")
	}
}

func demoValidationErrors() {
	fmt.Println("\n--- Validation Error Example ---")

	// Simulate validation error
	err := &pipewireio.ValidationError{
		Field:   "port_id",
		Value:   -1,
		Message: "port ID must be positive",
	}

	fmt.Printf("Error: %v\n", err)
}

func demoErrorIsPattern() {
	fmt.Println("\n--- error.Is Pattern Example ---")

	// Try to connect (will fail)
	c, err := client.NewClient("invalid-address")
	if err != nil {
		// Check for specific error types
		if errors.Is(err, pipewireio.ErrNotConnected) {
			fmt.Println("Connection error detected with error.Is")
		}
		return
	}
	defer c.Disconnect()
}

func demoErrorAsPattern() {
	fmt.Println("\n--- error.As Pattern Example ---")

	// Simulate an error
	err := &pipewireio.ValidationError{
		Field:   "node_name",
		Value:   "",
		Message: "name cannot be empty",
	}

	// Extract detailed error information
	var valErr *pipewireio.ValidationError
	if errors.As(err, &valErr) {
		log.Printf("Validation failed:")
		log.Printf("  Field: %s", valErr.Field)
		log.Printf("  Value: %v", valErr.Value)
		log.Printf("  Message: %s", valErr.Message)
	}
}
