package tests

import (
	"github.com/vignemail1/pipewire-go/client"
)

// CreateTestNode creates a test node fixture
func CreateTestNode(id uint32, name string) *client.GlobalObject {
	return &client.GlobalObject{
		ID:      id,
		Type:    "Node",
		Version: 3,
		Props: map[string]string{
			"node.name":        name,
			"node.description": "Test node",
			"media.class":      "Audio/Sink",
		},
	}
}

// CreateTestPort creates a test port fixture
func CreateTestPort(id uint32, nodeID uint32, direction string) *client.GlobalObject {
	return &client.GlobalObject{
		ID:      id,
		Type:    "Port",
		Version: 3,
		Props: map[string]string{
			"port.name":      "test-port",
			"port.direction": direction,
			"port.node":      string(rune(nodeID)),
			"format.dsp":     "32 bit float mono audio",
		},
	}
}

// CreateTestLink creates a test link fixture
func CreateTestLink(id uint32, outputPort uint32, inputPort uint32) *client.GlobalObject {
	return &client.GlobalObject{
		ID:      id,
		Type:    "Link",
		Version: 3,
		Props: map[string]string{
			"link.output.port": string(rune(outputPort)),
			"link.input.port":  string(rune(inputPort)),
			"link.state":       "active",
		},
	}
}
