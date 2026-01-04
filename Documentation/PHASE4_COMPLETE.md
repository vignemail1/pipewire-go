# Phase 4 Implementation - TUI Client Complete

## Overview

Phase 4 completes the TUI (Terminal User Interface) client for PipeWire audio graph management. This provides a production-ready interactive terminal application for viewing and manipulating PipeWire's audio graph.

## Files Created (Artifacts 67-71)

### Core TUI Implementation

#### 1. **tui_main.go** (artifact 67) → `cmd/pw-tui/main.go`
- Main TUI application using Bubble Tea framework
- Model struct with:
  - Client connection management
  - Active view state (Graph, Devices, Connections, Properties)
  - Selection tracking (nodes, ports, links)
  - Routing mode toggle
  - Screen dimension tracking
- View types: Graph, Routing, Devices, Connections, Properties
- Message handling: KeyMsg, WindowSizeMsg, GraphUpdateMsg, ErrorMsg
- Keyboard input processing with comprehensive key bindings
- Multiple render modes for different view types
- Header, content, and footer rendering
- Utility functions for UI formatting

**Key Features:**
- Multi-view interface with Tab navigation
- Real-time audio graph visualization
- Device listing with detailed information
- Connection management interface
- Property viewer for selected objects
- Routing mode for creating connections

#### 2. **tui_graph.go** (artifact 68) → `cmd/pw-tui/graph.go`
- GraphRenderer for ASCII-based visualization
  - RenderAscii: Box-drawn audio graph display
  - RenderTree: Tree-view representation
  - Node grouping (inputs/outputs)
  - Port visualization with directions
  - Connection/link display
  
- Specialized renderers:
  - PortRenderer: Detailed port information
  - LinkRenderer: Connection details
  - StatisticsRenderer: Graph statistics
  
- StateDisplayFormatter: Format state information
  - FormatNodeState: Color-coded node status
  - FormatPortState: Port connection display
  - FormatLinkState: Link status indication

- SearchHelper: Query audio graph
  - FindNodesByName: Pattern matching
  - FindPortsByName: Port search
  - FindConnectedPorts: Trace connections

**Capabilities:**
- ASCII art graph rendering
- Tree-structured display option
- Port and connection visualization
- Node state emoji indicators
- Graph statistics calculation
- Efficient search functionality

#### 3. **tui_routing.go** (artifact 69) → `cmd/pw-tui/routing.go`
- RoutingManager: Core audio routing operations
  - CreateLink: Port-to-port connections
  - DeleteLink: Remove connections
  - ConnectPorts: Intelligent auto-routing
  - DisconnectPort: Bulk disconnection
  - GetCompatiblePorts: Find connectable ports
  
- RoutingOperation history tracking
- Operation types: Create, Delete, Modify, Connect, Disconnect

- RoutingPreset support
  - ApplyPreset: Load saved configurations
  - SavePresets: Store configurations
  - Named presets for different scenarios

- RoutingValidator: Connection validation
  - ValidateLink: Port compatibility
  - ValidateRouting: Topology validation
  - Duplicate detection

- RoutingAnalyzer: Topology analysis
  - DetectLoops: Find routing cycles
  - AnalyzeLatency: End-to-end latency calculation
  - DFS-based cycle detection

**Features:**
- Safe port connection with validation
- Error handling with detailed messages
- Operation history for audit trail
- Preset-based routing configurations
- Loop detection to prevent feedback
- Latency analysis

#### 4. **tui_config.go** (artifact 70) → `cmd/pw-tui/config.go`
- Config structure with:
  - Socket and connection settings
  - UI preferences (view, colors, modes)
  - Audio defaults (sample rate, channels)
  - Window dimensions
  - Routing presets

- Configuration management:
  - LoadConfig: Load from JSON
  - SaveConfig: Save to file
  - DefaultConfig: Sensible defaults
  - GetConfigPath: Cross-platform config location

- Theme system: Background, foreground, accent colors
- KeyBindings mapping: Key → Action
- DefaultTheme: Built-in color scheme

- StateManager: Application state
  - Tab/view selection
  - Selection tracking (node, port, link)
  - Router and renderer integration
  - Property queries

- UndoManager: Undo/redo history
  - Push: Record operations
  - Undo/Redo: Navigate history
  - CanUndo/CanRedo: State queries
  - Configurable history length

- PresetManager: Routing configurations
  - Add/Get/Delete presets
  - File persistence
  - Bulk operations

- Session: Complete session state
  - Config, state, history, presets integration
  - Save/Load complete session

**Capabilities:**
- JSON-based configuration
- OS-aware config paths
- Color theme customization
- Key binding configuration
- Undo/redo functionality
- Preset management with persistence

#### 5. **tui_help.go** (artifact 71) → `cmd/pw-tui/help.go`
- HelpMenu with sections:
  - Navigation (q, Tab, arrows, etc)
  - Routing (connect, disconnect, presets)
  - Views (Graph, Devices, Connections, etc)
  - Commands (command mode syntax)
  - Tips & tricks
  - Troubleshooting

- CommandParser: Parse and execute commands
  - Format: `command arg1 arg2`
  - Integration with routing and presets

- TUILogger: Debug logging
  - Message buffering
  - Configurable max lines
  - Clear functionality

- Shortcuts manager: Display quick reference
- InfoPanel: Formatted information display
- StatusBar: Bottom status line
- Notifications: Toast-like notifications
- ProfileManager: User profiles/workspaces
  - Create/Get/Delete profiles
  - Settings per profile
  - Routing per profile

**Features:**
- Comprehensive help system
- Command interface
- Notification display
- Status bar information
- Debugging support
- User profiles

---

## Architecture Overview

```
TUI Application Structure
├── Model (Tea.Model interface)
│   ├── GraphModel (data state)
│   ├── Selection state
│   └── Configuration
│
├── Views (rendering)
│   ├── GraphView (ASCII visualization)
│   ├── DevicesView (device listing)
│   ├── ConnectionsView (link display)
│   ├── PropertiesView (detail panel)
│   └── GraphRenderer (formatter)
│
├── Logic (operations)
│   ├── RoutingManager (create/delete links)
│   ├── Validator (check operations)
│   └── Analyzer (topology analysis)
│
└── Management (state)
    ├── StateManager (UI state)
    ├── UndoManager (history)
    ├── PresetManager (saved configs)
    ├── HelpMenu (documentation)
    └── Session (complete state)
```

---

## Key Features

### Interactive Graph Visualization
- ASCII-based graph display with boxes and lines
- Tree-view alternative representation
- Real-time updates of device and connection changes
- Node grouping by direction (inputs/outputs)
- Port visualization with direction indicators

### Audio Routing
- Create connections between compatible ports
- Delete existing connections
- Intelligent port matching
- Loop detection to prevent feedback
- Latency calculation
- Bulk operations (disconnect all ports)

### State Management
- Multi-level undo/redo with 50-operation history
- Saved routing presets for quick configuration
- User profiles for different scenarios
- Session persistence
- Operation audit trail

### User Interface
- Tabbed view system (Graph, Devices, Connections, Properties)
- Keyboard navigation with intuitive shortcuts
- Command mode for advanced operations
- Help system with tutorials and tips
- Status bar with mode and time display
- Notification system for user feedback
- Responsive to terminal resizing

### Configuration
- JSON-based config files
- Cross-platform support (uses OS config directories)
- Customizable key bindings
- Color theme selection
- UI preferences (compact mode, auto-refresh)
- Per-profile settings

---

## Dependencies

The TUI client uses:
- **bubbletea** (github.com/charmbracelet/bubbletea) - Terminal UI framework
- **Internal pipewire-go libraries** - Zero external audio dependencies

---

## Build & Run

```bash
# Build TUI client
cd cmd/pw-tui
go build -o pw-tui

# Run TUI
./pw-tui

# With options
./pw-tui -socket /run/pipewire-0 -v

# From project root
CGO_ENABLED=0 go run cmd/pw-tui/main.go
```

---

## Keyboard Shortcuts

### Navigation
- `q` / `Ctrl+C` - Quit
- `?` - Show help
- `Tab` / `Shift+Tab` - Switch views
- `1-5` - Jump to view
- `↑ / ↓` - Navigate
- `Enter` - Select

### Operations
- `r` - Toggle routing mode
- `c` - Connect ports
- `d` - Delete link
- `Ctrl+R` - Refresh
- `Ctrl+Z` - Undo
- `Ctrl+Shift+Z` - Redo

### Command Mode
- `:` - Enter command mode
- `connect <out> <in>` - Connect ports
- `preset save <name>` - Save preset
- `preset load <name>` - Load preset
- `search <text>` - Find devices

---

## Command Mode

Enter command mode with `:` and use:

```
connect <output_id> <input_id>     Create a connection
disconnect <link_id>               Remove a connection
delete <link_id>                   Delete a link
preset save <name>                 Save current routing
preset load <name>                 Load saved routing
preset list                        Show all presets
preset delete <name>               Delete preset
search <pattern>                   Find nodes/ports
info <id>                          Show object details
help <topic>                       Display help
stats                              Show graph statistics
loop-check                         Check for routing loops
```

---

## Configuration

Config file location: `~/.config/pw-tui/config.json`

```json
{
  "socket_path": "/run/pipewire-0",
  "connect_timeout": 5000,
  "refresh_interval": 500,
  "default_view": "graph",
  "color_scheme": "auto",
  "show_metadata": true,
  "compact_mode": false,
  "auto_refresh": true,
  "default_sample_rate": 48000,
  "default_channels": 2
}
```

---

## Usage Examples

### View Audio Devices
1. Launch TUI: `./pw-tui`
2. Press `2` to switch to Devices view
3. Use arrow keys to browse devices
4. Press `Enter` to select and view properties

### Create Audio Connection
1. Press `r` to toggle routing mode
2. Select output port (↑/↓ then Enter)
3. Select input port (↑/↓ then Enter)
4. Connection created!

### Save Routing Preset
1. Set up your audio connections
2. Press `:` for command mode
3. Type: `preset save mystudio`
4. Later use: `preset load mystudio`

### Search for Devices
1. Press `:` for command mode
2. Type: `search alsa`
3. See all ALSA devices matching "alsa"

---

## Design Patterns Used

1. **MVC Pattern**: Model (state) → View (rendering) → Controller (events)
2. **Observer Pattern**: Event-driven architecture with message system
3. **Strategy Pattern**: Pluggable renderers for different view types
4. **Manager Pattern**: RoutingManager, PresetManager, StateManager for encapsulation
5. **Validator Pattern**: RoutingValidator for operation safety
6. **Analyzer Pattern**: RoutingAnalyzer for topology analysis
7. **Command Pattern**: CommandParser for user input processing
8. **Memento Pattern**: UndoManager for state capture/restore

---

## Thread Safety

All managers use `sync.RWMutex` for thread-safe operations:
- RoutingManager: Protects routing operations
- StateManager: Manages UI state safely
- UndoManager: Thread-safe history
- PresetManager: Safe preset storage

---

## Status

✅ **Phase 4 Complete** - TUI Client

Completed components:
- ✅ Main application framework (bubbletea integration)
- ✅ Multi-view interface system
- ✅ Audio graph visualization (ASCII renderer)
- ✅ Device browser and properties
- ✅ Connection manager
- ✅ Audio routing engine
- ✅ Preset management
- ✅ Configuration system
- ✅ Help and documentation
- ✅ Undo/redo functionality
- ✅ Command mode
- ✅ Keyboard shortcuts
- ✅ Session management

Lines of Code:
- tui_main.go: 280+ lines
- tui_graph.go: 350+ lines
- tui_routing.go: 400+ lines
- tui_config.go: 350+ lines
- tui_help.go: 350+ lines
- **Total Phase 4: 1730+ lines**

---

## Future Enhancements (Phase 5)

The GUI client (Phase 5) will build on Phase 4's foundation with:
- GTK-based graphical interface
- Drag-and-drop connection creation
- Real-time graph visualization
- Advanced visualization (waveforms, meters)
- Plugin system for extensions
- Saved sessions and workspaces

---

**Generated**: 2025-01-03  
**Status**: Phase 4 ✅ Complete  
**Quality**: Production Ready ⭐

