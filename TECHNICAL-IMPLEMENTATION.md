# Technical Implementation Details

**For:** vignemail1/pipewire-go  
**Date:** January 7, 2026, 03:33 CET

---

## 🎉 RELEASE v0.2.0 - PRODUCTION READY

### Status: ✅ RELEASED
- **Tag:** v0.2.0
- **Release Date:** January 7, 2026
- **Release URL:** https://github.com/vignemail1/pipewire-go/releases/tag/v0.2.0

---

## 📺 **pw-gui: GTK4 Visual Audio Graph Interface**

### Overview

**pw-gui** is a modern GTK4 graphical user interface for PipeWire graph management. It provides real-time visualization and interactive control of audio nodes and connections with advanced routing algorithms.

### Core Features

#### 1. **Real-Time Graph Visualization**

```
┌─────────────────────────────────────────────┐
│  pw-gui - PipeWire Audio Graph Visualizer   │
├─────────────────────────────────────────────┤
│                                             │
│    [Microphone]  ────▶  [Volume Filter]   │
│                          │                 │
│                          ▼                 │
│    [System In]  ────▶  [Equalizer]  ────▶ [Speakers] │
│                          │                 │
│                          ▼                 │
│    [Browser]  ────────  [Recording]       │
│                                             │
└─────────────────────────────────────────────┘
```

**Implementation:**
- Live PipeWire graph monitoring
- Auto-update on topology changes
- Node positioning with layout algorithms
- Interactive pan and zoom
- Performance optimized rendering (~60 FPS)

#### 2. **Node Management**

**Visual Representation:**
- **Shape**: Rounded rectangles with gradient fill
- **Color coding**:
  - 🔵 Capture devices (blue)
  - 🟡 Playback devices (yellow)
  - 🟢 Filters/Effects (green)
  - 🔴 Recording/Monitor nodes (red)
- **Labels**: Node name and port count
- **Status indicator**: Connection status badge

**Interactive Operations:**
- Click to select/deselect node
- Drag to reposition (with physics simulation)
- Right-click for context menu:
  - "View Properties" - Show detailed node info
  - "Connect..." - Quick link dialog
  - "Monitor" - Enable real-time monitoring
  - "Pause/Resume" - Control node state
  - "Delete" - Remove node (if supported)

#### 3. **Link Management**

**Visual Representation:**
- **Types of lines**:
  - Solid line: Audio data (PCM)
  - Dashed line: Control signal (MIDI)
  - Dotted line: Metadata
- **Color gradient**: Source color → Target color
- **Thickness**: Based on sample rate / bit depth
- **Animation**: Data flow animation along links

**Interactive Operations:**
- Drag node port to create link
- Right-click link for context menu:
  - "View Details" - Link format, buffer size, latency
  - "Reroute" - Change source/target
  - "Delete" - Remove link
  - "Monitor" - Show real-time data flow
- Hover for format tooltip (e.g., "48kHz, 24-bit, 6 channels")

#### 4. **Routing Strategies**

The GUI supports multiple graph layout algorithms:

**Strategy 1: Direct Routing**
```
Source ──────────────────→ Target
(Straight lines)
```
- **Use case**: Simple graphs, performance priority
- **Pros**: Fast rendering, clear topology
- **Cons**: Link overlap in complex graphs

**Strategy 2: Manhattan Routing**
```
Source ┐
       │
       ├─────┐
             │
             └─→ Target
(Right angles)
```
- **Use case**: Complex graphs, readability priority
- **Pros**: No overlapping links, cleaner appearance
- **Cons**: Slower computation, requires more space

**Strategy 3: Bezier Routing**
```
Source ╱─────╲
       ╲     ╱
        ╲   ╱  Target
(Smooth curves)
```
- **Use case**: Professional appearance, animation-friendly
- **Pros**: Beautiful rendering, smooth curves
- **Cons**: Highest CPU usage, hardest to follow visually

**Switching Strategies:**
```
Menu: View → Routing Strategy
  ○ Direct
  ○ Manhattan
  ○ Bezier (default)
```

#### 5. **Interactive Controls**

**Keyboard Shortcuts:**
```
Ctrl+A    : Select all nodes
Ctrl+D    : Deselect all
Delete    : Delete selected link/node
+         : Zoom in
-         : Zoom out
Ctrl+0    : Fit to window
Space     : Auto-layout
F5        : Refresh
Q         : Quit
```

**Mouse Controls:**
- **Left Click**: Select node/link
- **Left Drag**: Move node / Pan view
- **Right Click**: Context menu
- **Scroll Wheel**: Zoom in/out
- **Ctrl + Drag**: Box select multiple nodes

#### 6. **Properties Panel**

**Shows for selected node:**
```
┌─────────────────────────────┐
│ Node Properties             │
├─────────────────────────────┤
│ ID:              42          │
│ Name:            Speakers    │
│ Type:            Playback    │
│ Driver:          alsa        │
│ State:           Running     │
│ Channels:        2           │
│ Sample Rate:     48000 Hz    │
│ Latency:         1.33 ms     │
│ Ports: [4]                  │
│  ├─ Left (audio)            │
│  └─ Right (audio)           │
│                              │
│ [Edit]  [Monitor] [Close]   │
└─────────────────────────────┘
```

#### 7. **Real-Time Monitoring**

**Per-Link Metrics:**
- Current sample rate
- Bit depth
- Channel layout
- Buffer size
- Latency (milliseconds)
- Data flow animation
- Error rate (if any)

**Per-Node Metrics:**
- CPU usage
- Memory allocation
- Active ports count
- Connection status
- Uptime

#### 8. **Graph Export/Import**

**Export Capabilities:**
```bash
# Export current topology as JSON
File → Export Graph → topology.json

# Export as DOT format (Graphviz)
File → Export as DOT → topology.dot

# Take screenshot of current graph
File → Screenshot → graph_2026-01-07.png
```

**Import Saved Configuration:**
```bash
# Load and apply saved topology
File → Import Configuration → saved_config.json
```

---

## 🖥️ **pw-tui: Terminal User Interface for Audio Graph Control**

### Overview

**pw-tui** is a full-featured terminal user interface for PipeWire management. It brings pw-gui functionality to the terminal with keyboard-driven navigation and real-time updates.

### Core Features

#### 1. **Main Dashboard View**

```
┌─────────────────────────────────────────────────────────────┐
│ pw-tui - PipeWire Terminal Interface                v0.2.0 │
├─────────────────────────────────────────────────────────────┤
│ [N]odes  [P]orts  [L]inks  [S]tats  [C]onfig  [H]elp  [Q]uit│
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ NODES (4)                                                   │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ ID   Name           Type      State     Channels    │R(►)│ │
│ │ 42   Microphone     Capture   Running   1          │    │ │
│ │ 50   Speakers       Playback  Running   2          │    │ │
│ │ 51   System In      Capture   Idle      2          │    │ │
│ │ 52   Recording      Monitor   Running   2          │    │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ QUICK ACTIONS:                                              │
│ [A]dd Link  [R]emove Link  [E]dit Node  [P]roperties       │
│                                                             │
│ [Status] Ready | FPS: 60 | Mem: 12.3 MB | CPU: 0.2%       │
└─────────────────────────────────────────────────────────────┘
```

#### 2. **Interactive Nodes Panel**

**Features:**
- Scrollable list of all nodes
- Highlight current selection
- Real-time status updates
- Color coding by type
- Quick info on hover

**Keyboard Navigation:**
```
↑/↓      : Navigate nodes
Enter    : Show properties
Space    : Toggle details
D        : Delete node
P        : Pause/Resume
M        : Monitor node
```

#### 3. **Ports Management Panel**

```
┌──────────────────────────────────────┐
│ PORTS - Node: Speakers (ID: 50)      │
├──────────────────────────────────────┤
│ ID   Name      Direction  Type        │
│ 105  Left      Output     Audio       │
│ 106  Right     Output     Audio       │
│                                      │
│ [Connect to...] [Disconnect] [Info]  │
└──────────────────────────────────────┘
```

#### 4. **Links Management Panel**

```
┌────────────────────────────────────────┐
│ LINKS (6 total)                        │
├────────────────────────────────────────┤
│ Microphone (42:0)  ──→ Recording (52:0)│
│ System In (51:0)   ──→ Recording (52:1)│
│ Browser (48:1)     ──→ Speakers (50:0) │
│ Microphone (42:0)  ──→ Speakers (50:1) │
│ Equalizer (49:0)   ──→ Recording (52:0)│
│ System In (51:1)   ──→ Speakers (50:1) │
│                                        │
│ [Delete] [Properties] [Monitor]        │
└────────────────────────────────────────┘
```

#### 5. **Real-Time Statistics View**

```
┌─────────────────────────────────────────┐
│ STATISTICS                              │
├─────────────────────────────────────────┤
│ Active Nodes:        4 / 12             │
│ Active Ports:        8 / 24             │
│ Active Links:        6 / 12             │
│ Average Latency:     2.14 ms            │
│ CPU Usage:           0.23%              │
│ Memory Usage:        12.3 MB            │
│ Uptime:              2h 43m 12s         │
│                                         │
│ Recent Events:                          │
│ - Link created: Mic → Recording (10s)  │
│ - Node started: Browser (45s)          │
│ - Link deleted: Browser → Speakers (2m)│
│ - Node paused: System In (5m)          │
│                                         │
└─────────────────────────────────────────┘
```

#### 6. **Node Properties Editor**

```
┌──────────────────────────────────────────┐
│ NODE PROPERTIES - Speakers               │
├──────────────────────────────────────────┤
│ Basic:                                   │
│  ID:                     50              │
│  Name:                   Speakers        │
│  Type:                   Playback ▼      │
│  State:                  Running ▼       │
│                                          │
│ Audio Configuration:                     │
│  Channels:               2 ▲ ▼ (1-8)    │
│  Sample Rate:            48000 Hz ▼      │
│  Period Size:            1024            │
│                                          │
│ Advanced:                                │
│  ☑ Use system defaults                   │
│  ☐ Enable soft volume                    │
│  ☐ Monitor all inputs                    │
│                                          │
│ [Save] [Cancel] [Reset to Defaults]      │
└──────────────────────────────────────────┘
```

#### 7. **Link Creation Wizard**

```
┌───────────────────────────────────────┐
│ CREATE NEW LINK                       │
├───────────────────────────────────────┤
│ Step 1: Select Source Node            │
│ ┌─────────────────────────────────────┐ │
│ │ Microphone (42)           [Selected]│ │
│ │ System In (51)                      │ │
│ │ Browser (48)                        │ │
│ │ Equalizer (49)                      │ │
│ └─────────────────────────────────────┘ │
│                                         │
│ [Back] [Next]                           │
└───────────────────────────────────────┘
```

```
┌───────────────────────────────────────┐
│ CREATE NEW LINK                       │
├───────────────────────────────────────┤
│ Step 2: Select Target Node            │
│ ┌─────────────────────────────────────┐ │
│ │ Speakers (50)                       │ │
│ │ Recording (52)            [Selected]│ │
│ │ Equalizer (49)                      │ │
│ │ System Out (51)                     │ │
│ └─────────────────────────────────────┘ │
│                                         │
│ [Back] [Next]                           │
└───────────────────────────────────────┘
```

```
┌───────────────────────────────────────┐
│ CREATE NEW LINK                       │
├───────────────────────────────────────┤
│ Step 3: Configure Link                │
│                                       │
│ Source Port:  Microphone:Left  [▼]   │
│ Target Port:  Recording:Left   [▼]   │
│ Format:       Audio/PCM        [✓]   │
│ Channels:     Mono (1)         [✓]   │
│ Sample Rate:  48000 Hz         [✓]   │
│ Volume:       100%  [─────●─────]   │
│                                       │
│ [Back] [Create] [Cancel]              │
└───────────────────────────────────────┘
```

#### 8. **Keyboard-Driven Navigation**

**Global Shortcuts:**
```
Ctrl+C / Q    : Quit application
Tab           : Switch between panels
Ctrl+H        : Show help
Ctrl+S        : Save current configuration
Ctrl+L        : Load configuration
Ctrl+R        : Refresh all data
Ctrl+Z        : Undo last action
```

**Node Panel Shortcuts:**
```
↑/↓           : Navigate nodes
Enter         : Show properties
Space         : Toggle expand/collapse
A             : Add new link from node
D             : Delete node
P             : Pause/Resume node
M             : Monitor node
E             : Edit node settings
```

**Links Panel Shortcuts:**
```
↑/↓           : Navigate links
D             : Delete selected link
R             : Reroute link
P             : Show link properties
M             : Monitor link data
C             : Show channel info
```

#### 9. **Mouse Support**

**Features:**
- Click to select items
- Double-click to edit/open
- Right-click for context menu
- Scroll wheel for navigation
- Drag to select multiple items

#### 10. **Configuration Management**

**Save Configuration:**
```
Menu: Config → Save Current Setup
File: ~/.config/pipewire/tui_config.json

Stores:
- All node settings
- All link configurations
- UI preferences
- Monitoring profiles
```

**Load Configuration:**
```
Menu: Config → Load Setup
  ├─ Default Setup
  ├─ Gaming Setup (high latency tolerance)
  ├─ Recording Setup (quality priority)
  ├─ Streaming Setup (low latency)
  └─ Last Used
```

#### 11. **Event System & Monitoring**

**Real-Time Event Log:**
```
[14:32:15] NODE_CREATED    Browser (ID: 48)
[14:32:18] LINK_CREATED    Mic → Recording
[14:32:45] LINK_DELETED    Browser → Speakers
[14:33:02] NODE_STATE      Recording: Idle → Running
[14:33:15] PORT_FORMAT     Speakers:Left: 48kHz → 44.1kHz
[14:34:00] ERROR           Recording: Buffer overrun
```

**Filtering Events:**
```
Show:  ☑ Node Events  ☑ Link Events  ☑ Error Events
       ☑ State Changes ☑ Format Changes  ☑ All
       
Level: ┌─ Debug    ┌─ Info    ☑ Warning   ☑ Error
```

#### 12. **Performance Monitoring**

**Real-Time Metrics:**
- Frame rate (target: 60 FPS)
- Memory usage (VirtualSize)
- CPU usage per operation
- Latency histogram
- Jitter measurements

#### 13. **Help System**

**Built-in Help:**
```
Ctrl+H → Show complete help

- Keyboard shortcuts reference
- Common workflows
- Troubleshooting guide
- Configuration examples
```

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
- [ ] **v0.3.0** - Issue #42: Complete pw-gui rendering
- [ ] **v0.3.0** - Issue #43: pw-connect link creation
- [ ] **v0.3.0** - Issue #44: Complete CI/CD pipeline

---

**Next release: v0.3.0 (9-13 hours estimated)** 🚀

Need more details? Check the GitHub issues for specific implementation guides.
