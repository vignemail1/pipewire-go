# Technical Implementation Details

**For:** vignemail1/pipewire-go  
**Date:** January 7, 2026, 03:44 CET

---

## 🎯 RELEASE v0.2.0 - PRODUCTION READY

**Status:** ✅ RELEASED  
**Tag:** v0.2.0  
**Commit:** f56b8aa  
**Release:** https://github.com/vignemail1/pipewire-go/releases/tag/v0.2.0

---

## 🎨 Virtual Node Creation (Issue #46)

### Overview

**Feature:** Allow users to programmatically create and manage virtual audio nodes via library API, pw-gui, and pw-tui.

**Scope:**
- Core library API for virtual node creation
- pw-gui creation wizard with UI
- pw-tui multi-step terminal wizard
- CLI tool `pw-virtual`
- Pre-configured node presets

### Core API Implementation

**File:** `core/virtual_node.go`

#### VirtualNode Structures

```go
// VirtualNodeType defines the type of virtual node
type VirtualNodeType string

const (
    VirtualNode_Sink       VirtualNodeType = "sink"
    VirtualNode_Source     VirtualNodeType = "source"
    VirtualNode_Filter     VirtualNodeType = "filter"
    VirtualNode_Loopback   VirtualNodeType = "loopback"
)

// VirtualNodeFactory type
type VirtualNodeFactory string

const (
    Factory_NullAudioSink     VirtualNodeFactory = "support.null-audio-sink"
    Factory_NullAudioSource   VirtualNodeFactory = "support.null-audio-source"
    Factory_Adapter           VirtualNodeFactory = "adapter"
    Factory_Loopback          VirtualNodeFactory = "libpipewire-module-loopback"
    Factory_FilterChain       VirtualNodeFactory = "filter-chain"
)

// VirtualNodeConfig holds configuration for virtual node
type VirtualNodeConfig struct {
    // Basic info
    Name        string
    Description string
    Type        VirtualNodeType
    Factory     VirtualNodeFactory
    
    // Audio properties
    Channels    uint32
    SampleRate  uint32
    BitDepth    uint32
    ChannelLayout string // "FL FR" for stereo
    
    // Behavior
    Passive     bool
    Virtual     bool
    Exclusive   bool
    DontReconnect bool
    
    // Advanced
    Latency     string // "1024/48000" format
    Priority    int
    
    // Extra properties
    CustomProps map[string]interface{}
}

// VirtualNode represents a virtual node in the graph
type VirtualNode struct {
    ID         uint32
    Config     VirtualNodeConfig
    Ports      []*Port
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

#### Client Methods

```go
// CreateVirtualNode creates a new virtual node in the graph
func (c *Client) CreateVirtualNode(config VirtualNodeConfig) (*VirtualNode, error)

// GetVirtualNode retrieves a virtual node by ID
func (c *Client) GetVirtualNode(id uint32) (*VirtualNode, error)

// DeleteVirtualNode removes a virtual node from the graph
func (v *VirtualNode) Delete() error

// UpdateProperty updates a node property
func (v *VirtualNode) UpdateProperty(key string, value interface{}) error

// GetProperty retrieves a node property
func (v *VirtualNode) GetProperty(key string) (interface{}, error)

// GetPorts returns all ports belonging to this node
func (v *VirtualNode) GetPorts() ([]*Port, error)

// Refresh syncs node state with daemon
func (v *VirtualNode) Refresh() error
```

#### Preset Configurations

```go
// GetVirtualNodePreset returns a preconfigured node
func GetVirtualNodePreset(preset string) VirtualNodeConfig

// Available presets:
// "null-sink"     - Null audio sink (discard)
// "null-source"   - Null audio source (silence)
// "loopback"      - Virtual loopback pair
// "recording"     - Recording virtual sink (passive)
// "monitoring"    - Monitoring virtual source
// "default"       - Default stereo sink

var presets = map[string]VirtualNodeConfig{
    "null-sink": {
        Name:        "Null Sink",
        Description: "Discards all audio",
        Type:        VirtualNode_Sink,
        Factory:     Factory_NullAudioSink,
        Channels:    2,
        SampleRate:  48000,
        BitDepth:    32,
        Passive:     true,
    },
    "loopback": {
        Name:        "Virtual Loopback",
        Description: "Virtual audio loopback pair",
        Type:        VirtualNode_Loopback,
        Factory:     Factory_Loopback,
        Channels:    2,
        SampleRate:  48000,
        BitDepth:    32,
    },
    "recording": {
        Name:        "Recording",
        Description: "Virtual recording sink",
        Type:        VirtualNode_Sink,
        Factory:     Factory_NullAudioSink,
        Channels:    2,
        SampleRate:  48000,
        BitDepth:    32,
        Passive:     true,
    },
    // ... more presets
}
```

#### Usage Example

```go
package main

import (
    pw "github.com/vignemail1/pipewire-go"
)

func main() {
    client, _ := pw.NewClient("my-app")
    defer client.Disconnect()
    
    // Method 1: Use preset
    config := pw.GetVirtualNodePreset("recording")
    config.Name = "My Recording Sink"
    
    // Method 2: Custom configuration
    config := pw.VirtualNodeConfig{
        Name:        "VST Processing",
        Description: "Virtual sink for VST processing",
        Type:        pw.VirtualNode_Sink,
        Factory:     pw.Factory_NullAudioSink,
        Channels:    2,
        SampleRate:  48000,
        BitDepth:    32,
        Passive:     true,
    }
    
    // Create the node
    virtualNode, err := client.CreateVirtualNode(config)
    if err != nil {
        panic(err)
    }
    
    // The node is now in the graph
    fmt.Printf("Created virtual node ID: %d\n", virtualNode.ID)
    
    // Update property
    virtualNode.UpdateProperty("node.description", "Updated VST Sink")
    
    // Get ports
    ports, _ := virtualNode.GetPorts()
    fmt.Printf("Ports: %+v\n", ports)
    
    // Clean up
    defer virtualNode.Delete()
}
```

---

## 📺 pw-gui Virtual Node Creation

### Menu Integration

**Main Menu:** `File → Create Virtual Node`

### Creation Dialog

```
┌────────────────────────────────────────────────────────────────────────────┐
│ Create Virtual Node                                                    [×] │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                            │
│ Type:  ○ Sink   ○ Source   ○ Filter   ○ Loopback                          │
│                                                                            │
│ Name:  [                                                            ]      │
│ Desc:  [                                                            ]      │
│                                                                            │
│ ┌────────────────────────────────────────────────────────────────────────┐ │
│ │ Audio Configuration                                                  │ │
│ ├────────────────────────────────────────────────────────────────────────┤ │
│ │ Channels:     [2 ▼]                                                  │ │
│ │ Sample Rate:  [48000 ▼]                                              │ │
│ │ Bit Depth:    [32 ▼]                                                 │ │
│ │ Layout:       [FL FR ▼]                                              │ │
│ └────────────────────────────────────────────────────────────────────────┘ │
│                                                                            │
│ ┌────────────────────────────────────────────────────────────────────────┐ │
│ │ Options                                                              │ │
│ ├────────────────────────────────────────────────────────────────────────┤ │
│ │ ☐ Passive (don't hold graph)                                         │ │
│ │ ☐ Virtual                                                            │ │
│ │ ☐ Exclusive                                                          │ │
│ │ ☐ Don't reconnect                                                    │ │
│ └────────────────────────────────────────────────────────────────────────┘ │
│                                                                            │
│ Preset: [Default ▼] or [Loopback ▼] [Recording ▼]                        │
│                                                                            │
│ [Create] [Cancel] [Load Preset...]                                       │
└────────────────────────────────────────────────────────────────────────────┘
```

### Features

✅ **Preset Dropdown**
- Default, Loopback, Recording, Monitoring, Null, Custom
- Auto-populate form from preset

✅ **Type Selection**
- Sink (output), Source (input), Filter, Loopback
- Visual icons for each type

✅ **Audio Configuration Panel**
- Channels: 1-8 spinbox
- Sample Rate: 44.1kHz, 48kHz, 96kHz, 192kHz
- Bit Depth: 16, 24, 32 options
- Channel Layout: FL FR, FL FR LFE, etc.

✅ **Advanced Options**
- Passive: Don't hold graph playing
- Virtual: Mark as virtual
- Exclusive: Exclusive access
- Don't Reconnect

✅ **Validation**
- Name cannot be empty
- Name must be unique
- Valid channel range
- Real-time error display

✅ **After Creation**
- Show in graph automatically
- Display in Nodes list
- Enable immediate connections
- Show success notification

### Implementation Steps

1. Add "Create Virtual Node" menu item
2. Create GTK4 dialog widget
3. Implement form validation
4. Connect to library API
5. Handle creation result
6. Auto-refresh graph
7. UI tests (5+)

---

## 🖥️ pw-tui Virtual Node Creation

### Menu Integration

**Main Menu:** `Create → Virtual Node` or press `V`

### Multi-Step Wizard

#### Step 1: Type & Naming

```
┌────────────────────────────────────────────────────────────────────────────┐
│ CREATE VIRTUAL NODE                                                        │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                            │
│ Type:                                                                      │
│ ○ Sink       - Audio output device                                        │
│ ○ Source     - Audio input device                                         │
│ ○ Filter     - Audio filter/processor                                     │
│ ○ Loopback   - Virtual audio pair                                         │
│                                                                            │
│ Name: [_____________________________________]                            │
│       (e.g., "Recording", "VST Processing")                              │
│                                                                            │
│ Description: [_______________________________]                             │
│                                                                            │
│ Use preset: [Default ▼]                                                   │
│   • Default      • Loopback  • Recording                                  │
│   • Monitoring   • Null      • Custom                                     │
│                                                                            │
│ [Next] [Cancel]                                                           │
└────────────────────────────────────────────────────────────────────────────┘
```

**Keyboard Navigation:**
- `Tab`/`Shift+Tab` : Navigate fields
- `↑/↓` : Change preset/type options
- `Enter` : Confirm and next
- `Q` : Cancel

#### Step 2: Audio Configuration

```
┌────────────────────────────────────────────────────────────────────────────┐
│ VIRTUAL NODE - AUDIO CONFIG                                                │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                            │
│ Channels:        [2 ▲ ▼] (1-8)                                            │
│ Sample Rate:     [48000 ▼]  44.1k 48k 96k 192k                            │
│ Bit Depth:       [32 ▼]     16 24 32                                      │
│ Channel Layout:  [FL FR ▼]                                                │
│                                                                            │
│ ☐ Passive    (Don't hold graph playing)                                   │
│ ☐ Virtual    (Mark as virtual node)                                       │
│ ☐ Exclusive  (Exclusive access)                                           │
│ ☐ DontReconnect                                                            │
│                                                                            │
│ [< Back] [Create] [Cancel]                                               │
└────────────────────────────────────────────────────────────────────────────┘
```

**Interactive Elements:**
- Spinbox for channels (↑/↓)
- Dropdown for sample rate
- Dropdown for bit depth
- Dropdown for layout
- Checkboxes for options (Space to toggle)

#### Step 3: Review & Confirm

```
┌────────────────────────────────────────────────────────────────────────────┐
│ REVIEW - VIRTUAL NODE CONFIG                                               │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                            │
│ Name:            Recording                                                 │
│ Type:            Sink (Audio Output)                                       │
│ Description:     Recording virtual sink                                    │
│ Channels:        2 (Stereo)                                                │
│ Sample Rate:     48000 Hz                                                  │
│ Bit Depth:       32-bit                                                    │
│ Layout:          Front Left + Right                                        │
│ Passive:         Yes                                                       │
│                                                                            │
│ This node will be available in the graph immediately.                     │
│ You can connect other nodes to it.                                        │
│                                                                            │
│ [< Back] [Create] [Cancel]                                               │
└────────────────────────────────────────────────────────────────────────────┘
```

#### Success Confirmation

```
┌────────────────────────────────────────────────────────────────────────────┐
│ ✓ Virtual Node Created Successfully                                        │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                            │
│ Recording (ID: 100)                                                        │
│ ├─ Left (FL)      [Input Port]                                             │
│ └─ Right (FR)     [Input Port]                                             │
│                                                                            │
│ You can now route audio to this node.                                     │
│                                                                            │
│ [OK] [Show in Graph]                                                      │
└────────────────────────────────────────────────────────────────────────────┘
```

### Keyboard Shortcuts

```
Global:
  V           : Create virtual node
  Ctrl+C      : Cancel wizard
  Ctrl+Z      : Undo (undo last step)
  Q           : Quit wizard

In Wizard:
  Tab         : Next field
  Shift+Tab   : Previous field
  ↑/↓         : Change dropdown/option
  Enter       : Confirm/Next step
  Space       : Toggle checkbox
  Backspace   : Delete character in text field
```

### Implementation Steps

1. Add "Create" menu option
2. Implement wizard framework (3 steps)
3. Step 1: Type selection & naming
4. Step 2: Audio configuration
5. Step 3: Review & confirm
6. Connect to library API
7. Handle creation result
8. Display success/error
9. UI tests (8+)
10. Integration tests (3+)

---

## 🖨️ CLI Tool: pw-virtual

### Commands

**Create Virtual Node:**
```bash
pw-virtual create sink \
  --name "Recording" \
  --channels 2 \
  --rate 48000 \
  --passive

pw-virtual create loopback \
  --name "Virtual Pair" \
  --preset loopback
```

**List Virtual Nodes:**
```bash
pw-virtual list

# Output:
# ID   Name           Type      Channels  Rate     Passive
# 100  Recording      Sink      2         48000    yes
# 101  Virtual Pair   Loopback  2         48000    no
```

**Get Node Info:**
```bash
pw-virtual info 100

# Output:
# Node ID: 100
# Name: Recording
# Type: Sink (Audio Output)
# Description: Recording virtual sink
# Channels: 2 (FL FR)
# Sample Rate: 48000 Hz
# Bit Depth: 32-bit
# Passive: yes
# Ports:
#   - 100:0 [Left/Front Left]
#   - 100:1 [Right/Front Right]
```

**Delete Virtual Node:**
```bash
pw-virtual delete 100
# Virtual node 100 deleted
```

**Update Property:**
```bash
pw-virtual update 100 node.description "New Description"
# Property updated
```

---

## 🔧 Virtual Node Properties

**Common Properties:**
```
node.name              - Name of the node
node.description       - Description
node.nick              - Short name
node.virtual           - Is virtual (bool)
node.passive           - Don't hold graph (bool)
node.exclusive         - Exclusive access (bool)
node.dont-reconnect    - Don't auto-reconnect (bool)
node.latency           - Latency (format: "samples/rate")
node.lock-quantum      - Lock quantum size
priority.driver        - Priority level

audio.position         - Channel layout ("FL FR")
audio.format           - Sample format ("F32", "S32", etc.)
audio.rate             - Sample rate (44100, 48000, etc.)
```

---

## 📊 Virtual Node Lifecycle

```
1. Create
   ↓
   Client calls CreateVirtualNode(config)
   ↓
   Library sends to daemon
   ↓
   Daemon creates spa-node via factory
   ↓
   Daemon returns node ID to client
   ↓
   VirtualNode object created with ID

2. Active (In Graph)
   ↓
   Node is available in registry
   ↓
   Ports are created
   ↓
   Can be connected to other nodes
   ↓
   Audio can flow through

3. Modify
   ↓
   UpdateProperty(key, value)
   ↓
   Library sends to daemon
   ↓
   Daemon updates property
   ↓
   Client receives update

4. Delete
   ↓
   Delete() method called
   ↓
   Library sends delete request
   ↓
   Daemon removes node
   ↓
   All links to node are destroyed
   ↓
   Node removed from registry
```

---

## 🎯 Error Handling

**Potential Errors:**

```go
// VirtualNodeError types
type VirtualNodeCreateError struct {
    Reason string // "invalid_config", "factory_not_available", etc.
}

type VirtualNodePropertyError struct {
    Property string
    Message  string
}

type VirtualNodeNotFoundError struct {
    NodeID uint32
}
```

**Example Handling:**

```go
config := pw.VirtualNodeConfig{
    Name: "Recording",
    Channels: 0, // Invalid!
}

_, err := client.CreateVirtualNode(config)
if err != nil {
    switch err := err.(type) {
    case *pw.VirtualNodeCreateError:
        fmt.Printf("Creation failed: %s\n", err.Reason)
    default:
        fmt.Printf("Error: %v\n", err)
    }
}
```

---

## ✅ Testing Strategy

**Unit Tests (Core API):**
- Create virtual node with valid config
- Create with invalid config (validation)
- Update node properties
- Delete node
- Get node properties
- Preset loading

**Integration Tests:**
- Create node in live daemon
- Verify node appears in registry
- Connect nodes to virtual node
- Monitor data flow
- Delete and verify removal

**UI Tests (pw-gui):**
- Dialog opens/closes
- Form validation
- Preset loading
- Node creation
- Success/error display

**UI Tests (pw-tui):**
- Wizard navigation (all 3 steps)
- Keyboard input handling
- Form validation
- Node creation
- Message display

---

## 📝 Documentation Needed

- ✅ GoDoc API documentation
- ✅ Usage guide with examples
- ✅ Preset reference
- ✅ CLI reference
- ✅ Architecture diagrams
- ✅ Troubleshooting guide

---

## 🔴 CRITICAL FIX: Issue #27 - Race Condition

### The Problem

**Files affected:**
- `client/registry.go`
- `client/event_dispatcher.go`

**Why it's critical:**
Multiple goroutines access shared maps without synchronization, causing:
- Data corruption on concurrent writes
- Nil pointer dereference on reads
- Memory leaks from incomplete cleanup
- Unpredictable behavior in production

### The Solution

#### Step 1: Update Registry struct

**File:** `client/registry.go`

**Before:**
```go
type Registry struct {
    nodes map[uint32]*Node
    ports map[uint32]*Port
    links map[uint32]*Link
}
```

**After:**
```go
type Registry struct {
    mu    sync.RWMutex      // ADD THIS LINE
    nodes map[uint32]*Node
    ports map[uint32]*Port
    links map[uint32]*Link
}
```

#### Step 2: Protect read operations in Registry

**Example - GetNode:**
```go
func (r *Registry) GetNode(id uint32) *Node {
    r.mu.RLock()           // ADD THIS
    defer r.mu.RUnlock()   // ADD THIS
    return r.nodes[id]
}
```

**Pattern for all read methods:**
```go
// Before
func (r *Registry) GetNodes() map[uint32]*Node {
    return r.nodes
}

// After
func (r *Registry) GetNodes() map[uint32]*Node {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.nodes  // SAFE: Lock held during read
}
```

#### Step 3: Protect write operations in Registry

**Example - AddNode:**
```go
func (r *Registry) AddNode(node *Node) {
    r.mu.Lock()           // ADD THIS (not RLock!)
    defer r.mu.Unlock()   // ADD THIS
    r.nodes[node.ID] = node
}
```

**Pattern for all write methods:**
```go
// Before
func (r *Registry) RemoveNode(id uint32) {
    delete(r.nodes, id)
}

// After
func (r *Registry) RemoveNode(id uint32) {
    r.mu.Lock()
    defer r.mu.Unlock()
    delete(r.nodes, id)
}
```

#### Step 4: Update EventDispatcher struct

**File:** `client/event_dispatcher.go`

**Before:**
```go
type EventDispatcher struct {
    listeners map[string][]Handler
}
```

**After:**
```go
type EventDispatcher struct {
    mu        sync.RWMutex
    listeners map[string][]Handler
}
```

#### Step 5: Protect EventDispatcher methods

**Subscribe example:**
```go
// Before
func (ed *EventDispatcher) Subscribe(eventType string, handler Handler) {
    ed.listeners[eventType] = append(ed.listeners[eventType], handler)
}

// After
func (ed *EventDispatcher) Subscribe(eventType string, handler Handler) {
    ed.mu.Lock()
    defer ed.mu.Unlock()
    ed.listeners[eventType] = append(ed.listeners[eventType], handler)
}
```

**Dispatch example:**
```go
// Before
func (ed *EventDispatcher) Dispatch(event Event) {
    handlers := ed.listeners[event.Type]
    for _, handler := range handlers {
        handler(event)
    }
}

// After
func (ed *EventDispatcher) Dispatch(event Event) {
    ed.mu.RLock()
    handlers := ed.listeners[event.Type]
    ed.mu.RUnlock()
    
    for _, handler := range handlers {
        handler(event)  // SAFE: Call outside lock to avoid deadlock
    }
}
```

### Verification

#### Test with race detector
```bash
# Test just the client package
go test -v -race ./client/...

# Test everything
go test -v -race ./...
```

#### Expected output
```
✅ OK - no race conditions detected
PASS
```

#### If it fails
```
==================
WARNING: DATA RACE
==================
Write at 0x00c0001234567890 by goroutine 42:
    github.com/vignemail1/pipewire-go/client.(*Registry).AddNode()
        client/registry.go:45 +0x123
```

**Solution:** Make sure the lock is held for that operation.

---

## 🟠 HIGH: Issue #28 - Integration Tests

### Docker Setup

**File:** `tests/docker-compose.yml` (create new)

```yaml
version: '3.8'

services:
  pipewire:
    image: fedora:latest
    volumes:
      - /run/dbus:/run/dbus
    environment:
      - DBUS_SYSTEM_BUS_ADDRESS=unix:path=/run/dbus/system_bus_socket
    entrypoint: |
      bash -c '
        dnf install -y pipewire pipewire-utils
        pipewired -d
        sleep infinity
      '
    networks:
      - test

networks:
  test:
    driver: bridge
```

### Integration Test Structure

**File:** `tests/integration_test.go` (create new)

```go
// +build integration

package tests

import (
    "testing"
    "github.com/vignemail1/pipewire-go/client"
    "github.com/vignemail1/pipewire-go/core"
)

func TestConnectionWithRealPipeWire(t *testing.T) {
    // Start PipeWire connection
    c, err := client.NewClient(":0")  // Connect to PipeWire daemon
    if err != nil {
        t.Fatalf("Failed to connect: %v", err)
    }
    defer c.Close()
    
    // Test basic operations
    nodes, err := c.GetNodes()
    if err != nil {
        t.Fatalf("Failed to get nodes: %v", err)
    }
    
    if len(nodes) == 0 {
        t.Error("Expected nodes but got none")
    }
}

func TestGraphEnumeration(t *testing.T) {
    // Test complete graph enumeration
    // ...
}

func TestLinkManagement(t *testing.T) {
    // Test create/destroy links
    // ...
}
```

### Run Integration Tests

```bash
# Start services
docker-compose -f tests/docker-compose.yml up -d

# Run tests
go test -v -tags=integration ./tests/...

# Stop services
docker-compose -f tests/docker-compose.yml down
```

---

## 🟠 HIGH: Issue #29 - Error Handling

### Define Error Types

**File:** `errors.go` (create new)

```go
package pipewireio

import "fmt"

// ConnectionError represents a connection failure
type ConnectionError struct {
    Reason string
    Err    error
}

func (e *ConnectionError) Error() string {
    return fmt.Sprintf("connection failed: %s", e.Reason)
}

func (e *ConnectionError) Unwrap() error {
    return e.Err
}

// ProtocolError represents a protocol violation
type ProtocolError struct {
    Message string
    Data    []byte
}

func (e *ProtocolError) Error() string {
    return fmt.Sprintf("protocol error: %s", e.Message)
}

// ValidationError represents invalid input
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error in field %q: %s", e.Field, e.Message)
}

func (e *ValidationError) Unwrap() error {
    return nil
}
```

### Use in Functions

**Before:**
```go
func (c *Client) CreateLink(link *Link) error {
    if link == nil {
        return errors.New("link is nil")
    }
    // ...
}
```

**After:**
```go
func (c *Client) CreateLink(link *Link) error {
    if link == nil {
        return &ValidationError{
            Field:   "link",
            Message: "link cannot be nil",
        }
    }
    // ...
}
```

### Use error.Is pattern

```go
err := client.CreateLink(link)
if errors.Is(err, &ValidationError{}) {
    // Handle validation error
}

var valErr *ValidationError
if errors.As(err, &valErr) {
    log.Printf("Validation failed in field %s: %s", valErr.Field, valErr.Message)
}
```

---

## ✅ Testing Commands

### Race Detection
```bash
go test -race -v ./...
```

### Integration Tests
```bash
docker-compose -f tests/docker-compose.yml up -d
go test -v -tags=integration ./tests/...
docker-compose -f tests/docker-compose.yml down
```

### Coverage Report
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 📊 Implementation Checklist

- [x] **v0.2.0** Production Release Complete
- [x] **#27** Race condition fixed and tested
- [x] **#28** Docker setup and integration tests
- [x] **#29** Custom error types implemented
- [x] **#30** pw-gui types and functions defined
- [x] **#31** UI tests added
- [x] **#32** All GoDoc comments added
- [x] **#33** CLI tools implemented
- [x] **#34** TUI event handlers completed
- [ ] **#46** Virtual node creation API & UI (IN PROGRESS)
- [ ] **#42** Complete pw-gui rendering
- [ ] **#43** pw-connect link creation
- [ ] **#44** Complete CI/CD pipeline

---

**Next: Implement Issue #46 - Virtual Node Creation** 🚀

Implementation priority:
1. Core API (4-5h)
2. pw-gui integration (2-3h)
3. pw-tui integration (2-3h)
4. CLI tool (1h)

Estimated total effort: **6-8 hours**

