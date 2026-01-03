// Package config - TUI Configuration
// cmd/pw-tui/config.go
// Configuration and state management for the TUI

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Config represents the TUI configuration
type Config struct {
	// Socket and connection settings
	SocketPath       string `json:"socket_path"`
	ConnectTimeout   int    `json:"connect_timeout"`
	RefreshInterval  int    `json:"refresh_interval"`

	// UI Settings
	DefaultView      string `json:"default_view"`
	ColorScheme      string `json:"color_scheme"`
	ShowMetadata     bool   `json:"show_metadata"`
	CompactMode      bool   `json:"compact_mode"`
	AutoRefresh      bool   `json:"auto_refresh"`

	// Audio settings
	DefaultSampleRate uint32 `json:"default_sample_rate"`
	DefaultChannels   uint32 `json:"default_channels"`

	// Routing presets
	Presets map[string]*RoutingPreset `json:"presets"`

	// Window size
	MinWidth  int `json:"min_width"`
	MinHeight int `json:"min_height"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		SocketPath:       "/run/pipewire-0",
		ConnectTimeout:   5000, // milliseconds
		RefreshInterval:  500,  // milliseconds
		DefaultView:      "graph",
		ColorScheme:      "auto",
		ShowMetadata:     true,
		CompactMode:      false,
		AutoRefresh:      true,
		DefaultSampleRate: 48000,
		DefaultChannels:   2,
		Presets:          make(map[string]*RoutingPreset),
		MinWidth:         80,
		MinHeight:        24,
	}
}

// LoadConfig loads configuration from file
func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := DefaultConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// SaveConfig saves configuration to file
func (c *Config) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}

// GetConfigPath returns the config file path
func GetConfigPath() string {
	if configDir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(configDir, "pw-tui", "config.json")
	}
	return filepath.Join(os.Getenv("HOME"), ".config", "pw-tui", "config.json")
}

// KeyBindings represents keyboard bindings
type KeyBindings struct {
	Bindings map[string]string `json:"bindings"`
}

// DefaultKeyBindings returns default key bindings
func DefaultKeyBindings() *KeyBindings {
	return &KeyBindings{
		Bindings: map[string]string{
			"q":        "quit",
			"ctrl+c":   "quit",
			"?":        "help",
			"tab":      "next_view",
			"shift+tab": "prev_view",
			"up":       "select_up",
			"down":     "select_down",
			"enter":    "select",
			"r":        "toggle_routing",
			"d":        "delete",
			"c":        "connect",
			"ctrl+r":   "refresh",
			":":        "command",
		},
	}
}

// Theme represents the color theme
type Theme struct {
	Name           string `json:"name"`
	Background     string `json:"background"`
	Foreground     string `json:"foreground"`
	Primary        string `json:"primary"`
	Secondary      string `json:"secondary"`
	Success        string `json:"success"`
	Error          string `json:"error"`
	Warning        string `json:"warning"`
	Border         string `json:"border"`
	Selection      string `json:"selection"`
}

// DefaultTheme returns the default theme
func DefaultTheme() *Theme {
	return &Theme{
		Name:       "default",
		Background: "#000000",
		Foreground: "#ffffff",
		Primary:    "#00ff00",
		Secondary:  "#00aaff",
		Success:    "#00ff00",
		Error:      "#ff0000",
		Warning:    "#ffaa00",
		Border:     "#555555",
		Selection:  "#0055ff",
	}
}

// StateManager manages the application state
type StateManager struct {
	config         *Config
	theme          *Theme
	keyBindings    *KeyBindings
	routingManager *RoutingManager
	graphRenderer  *GraphRenderer
	selectedTab    int
	selectedNode   uint32
	selectedPort   uint32
	selectedLink   uint32
}

// NewStateManager creates a new state manager
func NewStateManager(config *Config) *StateManager {
	return &StateManager{
		config:       config,
		theme:        DefaultTheme(),
		keyBindings:  DefaultKeyBindings(),
		selectedTab:  0,
	}
}

// SetRouting sets the routing manager
func (sm *StateManager) SetRouting(rm *RoutingManager) {
	sm.routingManager = rm
}

// SetGraphRenderer sets the graph renderer
func (sm *StateManager) SetGraphRenderer(gr *GraphRenderer) {
	sm.graphRenderer = gr
}

// SelectNode selects a node
func (sm *StateManager) SelectNode(nodeID uint32) {
	sm.selectedNode = nodeID
}

// SelectPort selects a port
func (sm *StateManager) SelectPort(portID uint32) {
	sm.selectedPort = portID
}

// SelectLink selects a link
func (sm *StateManager) SelectLink(linkID uint32) {
	sm.selectedLink = linkID
}

// GetSelectedNode returns the currently selected node ID
func (sm *StateManager) GetSelectedNode() uint32 {
	return sm.selectedNode
}

// GetSelectedPort returns the currently selected port ID
func (sm *StateManager) GetSelectedPort() uint32 {
	return sm.selectedPort
}

// GetSelectedLink returns the currently selected link ID
func (sm *StateManager) GetSelectedLink() uint32 {
	return sm.selectedLink
}

// UndoManager manages undo/redo for routing operations
type UndoManager struct {
	history       []*RoutingOperation
	currentIndex  int
	maxHistoryLen int
}

// NewUndoManager creates a new undo manager
func NewUndoManager(maxHistoryLen int) *UndoManager {
	return &UndoManager{
		history:       make([]*RoutingOperation, 0),
		currentIndex:  -1,
		maxHistoryLen: maxHistoryLen,
	}
}

// Push adds an operation to the undo history
func (um *UndoManager) Push(op *RoutingOperation) {
	// Remove any redo history
	if um.currentIndex < len(um.history)-1 {
		um.history = um.history[:um.currentIndex+1]
	}

	// Add new operation
	um.history = append(um.history, op)
	um.currentIndex++

	// Limit history size
	if len(um.history) > um.maxHistoryLen {
		um.history = um.history[1:]
		um.currentIndex--
	}
}

// Undo returns the last operation (for undoing)
func (um *UndoManager) Undo() *RoutingOperation {
	if um.currentIndex > 0 {
		um.currentIndex--
		return um.history[um.currentIndex]
	}
	return nil
}

// Redo returns the next operation (for redoing)
func (um *UndoManager) Redo() *RoutingOperation {
	if um.currentIndex < len(um.history)-1 {
		um.currentIndex++
		return um.history[um.currentIndex]
	}
	return nil
}

// CanUndo checks if undo is available
func (um *UndoManager) CanUndo() bool {
	return um.currentIndex > 0
}

// CanRedo checks if redo is available
func (um *UndoManager) CanRedo() bool {
	return um.currentIndex < len(um.history)-1
}

// GetHistory returns the operation history
func (um *UndoManager) GetHistory() []*RoutingOperation {
	return um.history
}

// PresetManager manages routing presets
type PresetManager struct {
	presets map[string]*RoutingPreset
}

// NewPresetManager creates a new preset manager
func NewPresetManager() *PresetManager {
	return &PresetManager{
		presets: make(map[string]*RoutingPreset),
	}
}

// AddPreset adds a routing preset
func (pm *PresetManager) AddPreset(name string, preset *RoutingPreset) error {
	if name == "" {
		return fmt.Errorf("preset name cannot be empty")
	}
	pm.presets[name] = preset
	return nil
}

// GetPreset retrieves a routing preset
func (pm *PresetManager) GetPreset(name string) (*RoutingPreset, bool) {
	preset, exists := pm.presets[name]
	return preset, exists
}

// DeletePreset removes a routing preset
func (pm *PresetManager) DeletePreset(name string) {
	delete(pm.presets, name)
}

// ListPresets returns all preset names
func (pm *PresetManager) ListPresets() []string {
	names := make([]string, 0, len(pm.presets))
	for name := range pm.presets {
		names = append(names, name)
	}
	return names
}

// SavePresetsToFile saves presets to a JSON file
func (pm *PresetManager) SavePresetsToFile(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(pm.presets, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}

// LoadPresetsFromFile loads presets from a JSON file
func (pm *PresetManager) LoadPresetsFromFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &pm.presets); err != nil {
		return err
	}

	return nil
}

// Session represents a TUI session
type Session struct {
	config         *Config
	stateManager   *StateManager
	undoManager    *UndoManager
	presetManager  *PresetManager
	routingManager *RoutingManager
}

// NewSession creates a new TUI session
func NewSession(config *Config, routingManager *RoutingManager) *Session {
	return &Session{
		config:         config,
		stateManager:   NewStateManager(config),
		undoManager:    NewUndoManager(50),
		presetManager:  NewPresetManager(),
		routingManager: routingManager,
	}
}

// Save saves the session state
func (s *Session) Save() error {
	presetPath := filepath.Join(filepath.Dir(GetConfigPath()), "presets.json")
	return s.presetManager.SavePresetsToFile(presetPath)
}

// Load loads the session state
func (s *Session) Load() error {
	presetPath := filepath.Join(filepath.Dir(GetConfigPath()), "presets.json")
	if _, err := os.Stat(presetPath); os.IsNotExist(err) {
		return nil // File doesn't exist yet
	}
	return s.presetManager.LoadPresetsFromFile(presetPath)
}
