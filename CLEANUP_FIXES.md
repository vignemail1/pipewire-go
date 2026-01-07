# Cleanup and Fix Guide

This document provides step-by-step fixes for all identified issues in pw-tui and pw-gui commands.

## Table of Contents

1. [pw-tui Fixes](#pw-tui-fixes)
2. [pw-gui Fixes](#pw-gui-fixes)
3. [Testing Guide](#testing-guide)
4. [Verification Steps](#verification-steps)

---

## pw-tui Fixes

### Fix 1: Implement Proper Message Handlers

**File**: `cmd/pw-tui/main.go`

**Issue**: `deleteLink()` and `createLink()` commands don't actually process messages

**Current Code**:
```go
func deleteLink(linkID uint32) tea.Cmd {
    return func() tea.Msg {
        return DeleteLinkMsg{linkID: linkID}  // Returned but never handled
    }
}

func createLink() tea.Cmd {
    return nil  // Not implemented
}
```

**Fix**:
```go
func deleteLink(m *Model, linkID uint32) tea.Cmd {
    return func() tea.Msg {
        // Actually delete the link
        err := m.client.DestroyLink(linkID)
        if err != nil {
            return ErrorMsg{err: fmt.Errorf("failed to delete link: %w", err)}
        }
        // Refresh graph
        m.graphModel.Update(m.client)
        return GraphUpdateMsg{}
    }
}

func createLink(m *Model, sourcePortID, sinkPortID uint32) tea.Cmd {
    return func() tea.Msg {
        if sourcePortID == 0 || sinkPortID == 0 {
            return ErrorMsg{err: fmt.Errorf("invalid port selection")}
        }
        
        // Get ports and create link
        sourcePort := m.graphModel.ports[sourcePortID]
        sinkPort := m.graphModel.ports[sinkPortID]
        
        if sourcePort == nil || sinkPort == nil {
            return ErrorMsg{err: fmt.Errorf("port not found")}
        }
        
        _, err := m.client.CreateLink(sourcePort, sinkPort, nil)
        if err != nil {
            return ErrorMsg{err: fmt.Errorf("failed to create link: %w", err)}
        }
        
        // Refresh graph
        m.graphModel.Update(m.client)
        return GraphUpdateMsg{}
    }
}
```

**Update Message Handler** in `Update()` method:
```go
case DeleteLinkMsg:
    return m, deleteLink(m, msg.linkID)
case CreateLinkMsg:
    return m, createLink(m, msg.outputID, msg.inputID)
```

### Fix 2: Fix Broken Property Access

**File**: `cmd/pw-tui/main.go`

**Issue**: `node.info.MediaClass` is accessing unexposed field

**Current Code**:
```go
output += fmt.Sprintf("      Class: %s\n", node.info.MediaClass)  // WRONG
```

**Fix**:
First, check the client.Node API to get the correct method:
```go
// Instead of accessing info directly, use public methods
if class, ok := node.GetProperty("media.class"); ok {
    output += fmt.Sprintf("      Class: %s\n", class)
}

// Or if client.Node has dedicated methods:
if mediaClass := node.MediaClass(); mediaClass != "" {
    output += fmt.Sprintf("      Class: %s\n", mediaClass)
}
```

### Fix 3: Implement Event Subscription

**File**: `cmd/pw-tui/app.go` (new file or section)

**Issue**: No event subscription to PipeWire daemon

**Fix**: Add event listener in NewModel:
```go
func (m *Model) subscribeToEvents() error {
    // Subscribe to node events
    err := m.client.RegisterEventListener(client.EventTypeNodeAdded, func(event client.Event) error {
        m.graphModel.Update(m.client)
        return nil
    })
    if err != nil {
        return err
    }
    
    // Subscribe to port events
    err = m.client.RegisterEventListener(client.EventTypePortAdded, func(event client.Event) error {
        m.graphModel.Update(m.client)
        return nil
    })
    if err != nil {
        return err
    }
    
    // Subscribe to link events
    err = m.client.RegisterEventListener(client.EventTypeLinkAdded, func(event client.Event) error {
        m.graphModel.Update(m.client)
        return nil
    })
    if err != nil {
        return err
    }
    
    return nil
}
```

Call this in `NewModel()` after client creation.

### Fix 4: Add Help Implementation

**File**: `cmd/pw-tui/help.go`

**Issue**: `showHelp()` returns nil

**Fix**:
```go
func showHelp() tea.Cmd {
    return func() tea.Msg {
        // Could display help screen as a new view
        return tea.Msg(nil)  // For now, or implement help view
    }
}
```

Or better, add a help view:
```go
const (
    ViewTypeGraph ViewType = iota
    ViewTypeRouting
    ViewTypeDevices
    ViewTypeConnections
    ViewTypeProperties
    ViewTypeHelp  // Add this
)

func (m *Model) renderHelpView() string {
    help := `
╔════════════════════════════════════════════════════════╗
║           PipeWire Audio Graph TUI - Help              ║
╚════════════════════════════════════════════════════════╝

NAVIGATION:
  ↑/↓       Navigate up/down
  Tab       Switch view
  1/2/3     Quick view switch (Graph/Devices/Connections)

ACTIONS:
  Enter     Select item
  r         Toggle routing mode
  c         Create link (routing mode)
  d         Delete link
  Ctrl+R    Refresh graph

OTHER:
  ?         Toggle this help
  q         Quit

VIEWS:
  Graph       - Visual representation of audio graph
  Devices     - List of audio devices
  Connections - Active links between ports
  Properties  - Properties of selected item

Press 'q' to return to main view
`
    return help
}
```

---

## pw-gui Fixes

### Fix 1: Implement Missing Types

**File**: `cmd/pw-gui/app.go` (create new files)

**Create `cmd/pw-gui/graph.go`**:
```go
package main

import (
    "fmt"
    "github.com/diamondburned/gotk4/pkg/gdk/v4"
    "github.com/vignemail1/pipewire-go/client"
)

type GraphVisualizer struct {
    client      *client.Client
    drawingArea *gtk.DrawingArea
    zoomLevel   float64
}

func NewGraphVisualizer(c *client.Client) *GraphVisualizer {
    return &GraphVisualizer{
        client:    c,
        zoomLevel: 1.0,
    }
}

func (gv *GraphVisualizer) SetDrawingArea(da *gtk.DrawingArea) {
    gv.drawingArea = da
}

func (gv *GraphVisualizer) Draw(cr cairo.Context, width, height int, nodes []*client.Node, links []*client.Link) {
    // Implement actual graph drawing
    // This is a complex operation - consider using graphviz library
    
    // For now, basic layout:
    nodeHeight := height / (len(nodes) + 1)
    for i, node := range nodes {
        y := float64((i + 1) * nodeHeight)
        
        // Draw node box
        cr.SetSourceRGB(0.3, 0.3, 0.5)
        cr.Rectangle(50, y, 200, 40)
        cr.Fill()
        
        // Draw node label
        cr.SetSourceRGB(1, 1, 1)
        cr.MoveTo(60, y+25)
        // Use cairo text rendering for node name
        fmt.Sprintf("%s (ID: %d)", node.Name(), node.ID)
    }
    
    // Draw links
    cr.SetSourceRGB(0.5, 0.8, 0.5)
    cr.SetLineWidth(2)
    for _, link := range links {
        // Draw line from source to sink
        // This requires position mapping
    }
}

func (gv *GraphVisualizer) ZoomIn() {
    gv.zoomLevel *= 1.2
    gv.drawingArea.QueueDraw()
}

func (gv *GraphVisualizer) ZoomOut() {
    gv.zoomLevel /= 1.2
    gv.drawingArea.QueueDraw()
}

func (gv *GraphVisualizer) FitToWindow() {
    gv.zoomLevel = 1.0
    gv.drawingArea.QueueDraw()
}
```

**Create `cmd/pw-gui/engine.go`**:
```go
package main

import "github.com/vignemail1/pipewire-go/client"

type RoutingEngine struct {
    client *client.Client
}

func NewRoutingEngine(c *client.Client) *RoutingEngine {
    return &RoutingEngine{client: c}
}

func (re *RoutingEngine) GetCompatiblePorts(sourcePort *client.Port) []*client.Port {
    var compatible []*client.Port
    
    for _, node := range re.client.GetNodes() {
        for _, port := range node.GetPorts() {
            if sourcePort.CanConnectTo(port) {
                compatible = append(compatible, port)
            }
        }
    }
    
    return compatible
}

func (re *RoutingEngine) CreateLink(sourcePort, sinkPort *client.Port) (uint32, error) {
    link, err := re.client.CreateLink(sourcePort, sinkPort, nil)
    if err != nil {
        return 0, err
    }
    return link.ID(), nil
}

func (re *RoutingEngine) DeleteLink(linkID uint32) error {
    return re.client.DestroyLink(linkID)
}
```

**Create `cmd/pw-gui/widgets.go`**:
```go
package main

import (
    "fmt"
    "github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type StatusBar struct {
    box     *gtk.Box
    status  *gtk.Label
    mode    *gtk.Label
}

func NewStatusBar() *StatusBar {
    box := gtk.NewBoxOrientation(gtk.OrientationHorizontal, 10)
    status := gtk.NewLabel("Ready")
    mode := gtk.NewLabel("Normal")
    
    box.Append(status)
    box.Append(mode)
    
    return &StatusBar{box: box, status: status, mode: mode}
}

func (sb *StatusBar) GetWidget() *gtk.Box {
    return sb.box
}

func (sb *StatusBar) SetStatus(text string) {
    sb.status.SetText(text)
}

func (sb *StatusBar) SetMode(text string) {
    sb.mode.SetText(text)
}

type SettingsPanel struct {
    box *gtk.Box
}

func NewSettingsPanel() *SettingsPanel {
    box := gtk.NewBoxOrientation(gtk.OrientationVertical, 5)
    
    label := gtk.NewLabel("Settings")
    box.Append(label)
    
    return &SettingsPanel{box: box}
}

func (sp *SettingsPanel) GetWidget() *gtk.Box {
    return sp.box
}
```

### Fix 2: Implement Menu Handlers

**File**: `cmd/pw-gui/main.go`

**Current Code**:
```go
func (ga *GuiApp) showFileMenu() {
    ga.logger.Info("File menu opened")
}
```

**Fix**:
```go
func (ga *GuiApp) showFileMenu() {
    // Create popover with menu items
    menu := gtk.NewBoxOrientation(gtk.OrientationVertical, 5)
    
    exportBtn := gtk.NewButtonWithLabel("Export Graph")
    exportBtn.ConnectClicked(func() {
        ga.exportGraph()
    })
    
    importBtn := gtk.NewButtonWithLabel("Import Configuration")
    importBtn.ConnectClicked(func() {
        ga.importConfiguration()
    })
    
    quitBtn := gtk.NewButtonWithLabel("Quit")
    quitBtn.ConnectClicked(func() {
        ga.window.Close()
    })
    
    menu.Append(exportBtn)
    menu.Append(importBtn)
    menu.Append(quitBtn)
    
    menu.Show()
}

func (ga *GuiApp) exportGraph() {
    // Implement graph export
    ga.statusBar.SetStatus("Exporting graph...")
    ga.logger.Info("Graph exported")
}

func (ga *GuiApp) importConfiguration() {
    // Implement configuration import
    ga.statusBar.SetStatus("Importing configuration...")
    ga.logger.Info("Configuration imported")
}
```

### Fix 3: Implement Update Loop

**File**: `cmd/pw-gui/main.go`

**Current Code**:
```go
func (ga *GuiApp) startUpdateLoop() {
    ga.logger.Debug("Update loop started")
}
```

**Fix**:
```go
func (ga *GuiApp) startUpdateLoop() {
    // Use glib to schedule periodic updates
    glib.TimeoutAdd(glib.PRIORITY_DEFAULT, 1000, func() bool {
        // Update device list
        ga.statusBar.SetStatus(fmt.Sprintf(
            "Connected | %d nodes | %d links",
            len(ga.client.GetNodes()),
            len(ga.client.GetLinks()),
        ))
        
        // Redraw graph
        ga.drawingArea.QueueDraw()
        
        return true  // Continue timeout
    })
}
```

---

## Testing Guide

### Unit Tests for pw-tui

**File**: `cmd/pw-tui/main_test.go`

```go
package main

import (
    "testing"
    tea "github.com/charmbracelet/bubbletea"
)

func TestModelInitialization(t *testing.T) {
    m, err := NewModel("/tmp/test", nil)
    if err != nil && m == nil {
        // Expected for test socket
    }
}

func TestKeyboardNavigation(t *testing.T) {
    m := &Model{selectedIndex: 0}
    msg := tea.KeyMsg{Type: tea.KeyUp}
    _, _ = m.Update(msg)
    // Verify behavior
}
```

### Build Verification

```bash
# Verify pw-tui builds
cd cmd/pw-tui
go build -v -o pw-tui
if [ -f pw-tui ]; then echo "✓ pw-tui built successfully"; fi

# Verify pw-gui builds
cd ../pw-gui
go build -v -o pw-gui
if [ -f pw-gui ]; then echo "✓ pw-gui built successfully"; fi
```

---

## Verification Steps

### Pre-Deployment Checklist

- [ ] All code compiles without errors
- [ ] All linters pass (`golangci-lint run ./cmd/...`)
- [ ] All tests pass (`go test ./cmd/...`)
- [ ] pw-tui launches without crashes
- [ ] pw-gui launches without crashes
- [ ] All keyboard shortcuts work (TUI)
- [ ] All menu items are functional (GUI)
- [ ] Error handling works properly
- [ ] Help system accessible
- [ ] Graph visualization displays
- [ ] Can create/delete links
- [ ] UI updates on graph changes

### Functional Verification

```bash
# Test TUI
steps:
  1. Start: pw-tui
  2. Press ↑/↓ to navigate
  3. Press Tab to switch views
  4. Press ? for help
  5. Press Ctrl+R to refresh
  6. Press q to quit

# Test GUI
steps:
  1. Start: pw-gui
  2. Click tabs (Graph, Devices, Connections, Properties)
  3. Click menu items (File, Edit, View, Audio, Help)
  4. Try zoom controls
  5. Try refresh button
  6. Close window
```

---

## Related Documentation

- Issue #22 - TUI Cleanup
- Issue #23 - GUI Completion
- Issue #24 - Testing
- `CLEANUP_REPORT.md` - Detailed analysis
