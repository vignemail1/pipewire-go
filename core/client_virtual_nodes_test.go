package core

import (
	"testing"
)

func TestClientCreateVirtualNode(t *testing.T) {
	client := &Client{
		connected:        true,
		nextVirtualNodeID: 0,
	}

	config := GetVirtualNodePreset("recording")
	config.Name = "Test Recording Sink"

	virtualNode, err := client.CreateVirtualNode(config)
	if err != nil {
		t.Errorf("CreateVirtualNode() error = %v, want nil", err)
	}

	if virtualNode == nil {
		t.Fatalf("CreateVirtualNode() returned nil, want VirtualNode")
	}

	if virtualNode.ID == 0 {
		t.Errorf("CreateVirtualNode() ID = 0, want non-zero")
	}

	if virtualNode.Config.Name != config.Name {
		t.Errorf("CreateVirtualNode() Name = %q, want %q", virtualNode.Config.Name, config.Name)
	}
}

func TestClientCreateVirtualNodeValidation(t *testing.T) {
	client := &Client{
		connected:        true,
		nextVirtualNodeID: 0,
	}

	// Invalid config (empty name)
	config := VirtualNodeConfig{
		Name:       "",
		Type:       VirtualNode_Sink,
		Factory:    Factory_NullAudioSink,
		Channels:   2,
		SampleRate: 48000,
	}

	_, err := client.CreateVirtualNode(config)
	if err == nil {
		t.Errorf("CreateVirtualNode() with invalid config error = nil, want error")
	}
}

func TestClientCreateVirtualNodeNotConnected(t *testing.T) {
	client := &Client{
		connected: false,
	}

	config := GetVirtualNodePreset("recording")

	_, err := client.CreateVirtualNode(config)
	if err == nil {
		t.Errorf("CreateVirtualNode() on disconnected client error = nil, want error")
	}
}

func TestClientGetVirtualNode(t *testing.T) {
	client := &Client{
		connected:        true,
		nextVirtualNodeID: 0,
	}

	config := GetVirtualNodePreset("recording")
	virtualNode, err := client.CreateVirtualNode(config)
	if err != nil {
		t.Fatalf("CreateVirtualNode() error = %v", err)
	}

	nodeID := virtualNode.ID

	// Retrieve the node
	retrieved, err := client.GetVirtualNode(nodeID)
	if err != nil {
		t.Errorf("GetVirtualNode() error = %v, want nil", err)
	}

	if retrieved.ID != nodeID {
		t.Errorf("GetVirtualNode() ID = %d, want %d", retrieved.ID, nodeID)
	}
}

func TestClientGetVirtualNodeNotFound(t *testing.T) {
	client := &Client{
		connected:        true,
		nextVirtualNodeID: 0,
	}

	_, err := client.GetVirtualNode(999)
	if err == nil {
		t.Errorf("GetVirtualNode() with non-existent ID error = nil, want error")
	}
}

func TestClientListVirtualNodes(t *testing.T) {
	client := &Client{
		connected:        true,
		nextVirtualNodeID: 0,
	}

	// Create multiple nodes
	for i := 0; i < 3; i++ {
		config := GetVirtualNodePreset("recording")
		_, err := client.CreateVirtualNode(config)
		if err != nil {
			t.Fatalf("CreateVirtualNode() error = %v", err)
		}
	}

	nodes := client.ListVirtualNodes()
	if len(nodes) != 3 {
		t.Errorf("ListVirtualNodes() returned %d nodes, want 3", len(nodes))
	}
}

func TestClientListVirtualNodesEmpty(t *testing.T) {
	client := &Client{
		connected:        true,
		nextVirtualNodeID: 0,
	}

	nodes := client.ListVirtualNodes()
	if len(nodes) != 0 {
		t.Errorf("ListVirtualNodes() on empty client returned %d nodes, want 0", len(nodes))
	}
}

func TestClientDeleteVirtualNode(t *testing.T) {
	client := &Client{
		connected:        true,
		nextVirtualNodeID: 0,
	}

	config := GetVirtualNodePreset("recording")
	virtualNode, err := client.CreateVirtualNode(config)
	if err != nil {
		t.Fatalf("CreateVirtualNode() error = %v", err)
	}

	nodeID := virtualNode.ID

	// Delete the node
	err = client.DeleteVirtualNode(nodeID)
	if err != nil {
		t.Errorf("DeleteVirtualNode() error = %v, want nil", err)
	}

	// Verify it's deleted
	_, err = client.GetVirtualNode(nodeID)
	if err == nil {
		t.Errorf("GetVirtualNode() after delete error = nil, want error")
	}
}

func TestClientDeleteVirtualNodeNotFound(t *testing.T) {
	client := &Client{
		connected:        true,
		nextVirtualNodeID: 0,
	}

	err := client.DeleteVirtualNode(999)
	if err == nil {
		t.Errorf("DeleteVirtualNode() with non-existent ID error = nil, want error")
	}
}

func TestClientUpdateVirtualNodeProperty(t *testing.T) {
	client := &Client{
		connected:        true,
		nextVirtualNodeID: 0,
	}

	config := GetVirtualNodePreset("recording")
	virtualNode, err := client.CreateVirtualNode(config)
	if err != nil {
		t.Fatalf("CreateVirtualNode() error = %v", err)
	}

	nodeID := virtualNode.ID

	// Update property
	err = client.UpdateVirtualNodeProperty(nodeID, "test.key", "test.value")
	if err != nil {
		t.Errorf("UpdateVirtualNodeProperty() error = %v, want nil", err)
	}

	// Verify it was updated
	value, err := client.GetVirtualNodeProperty(nodeID, "test.key")
	if err != nil {
		t.Errorf("GetVirtualNodeProperty() error = %v, want nil", err)
	}

	if value != "test.value" {
		t.Errorf("GetVirtualNodeProperty() returned %v, want 'test.value'", value)
	}
}

func TestClientGetVirtualNodeProperty(t *testing.T) {
	client := &Client{
		connected:        true,
		nextVirtualNodeID: 0,
	}

	config := GetVirtualNodePreset("recording")
	virtualNode, err := client.CreateVirtualNode(config)
	if err != nil {
		t.Fatalf("CreateVirtualNode() error = %v", err)
	}

	nodeID := virtualNode.ID

	// Set property first
	client.UpdateVirtualNodeProperty(nodeID, "test.key", "test.value")

	// Get property
	value, err := client.GetVirtualNodeProperty(nodeID, "test.key")
	if err != nil {
		t.Errorf("GetVirtualNodeProperty() error = %v, want nil", err)
	}

	if value != "test.value" {
		t.Errorf("GetVirtualNodeProperty() returned %v, want 'test.value'", value)
	}
}

func TestClientGenerateVirtualNodeID(t *testing.T) {
	client := &Client{
		connected:        true,
		nextVirtualNodeID: 0,
	}

	id1 := client.generateVirtualNodeID()
	id2 := client.generateVirtualNodeID()
	id3 := client.generateVirtualNodeID()

	if id1 == id2 || id2 == id3 || id1 == id3 {
		t.Errorf("generateVirtualNodeID() generated duplicate IDs: %d, %d, %d", id1, id2, id3)
	}

	if id1 < 1000 {
		t.Errorf("generateVirtualNodeID() returned ID < 1000: %d", id1)
	}
}

func TestClientGetMediaClass(t *testing.T) {
	client := &Client{}

	tests := []struct {
		nodeType VirtualNodeType
		want     string
	}{
		{VirtualNode_Sink, "Audio/Sink"},
		{VirtualNode_Source, "Audio/Source"},
		{VirtualNode_Filter, "Audio/Effect"},
		{VirtualNode_Loopback, "Audio/Sink"},
	}

	for _, tt := range tests {
		got := client.getMediaClass(tt.nodeType)
		if got != tt.want {
			t.Errorf("getMediaClass(%q) = %q, want %q", tt.nodeType, got, tt.want)
		}
	}
}
