// Package help - Help and Documentation
// cmd/pw-tui/help.go
// Help system and documentation

package main

import "fmt"

// HelpMenu represents the help menu
type HelpMenu struct {
	sections map[string]string
}

// NewHelpMenu creates a new help menu
func NewHelpMenu() *HelpMenu {
	return &HelpMenu{
		sections: map[string]string{
			"navigation": `
NAVIGATION KEYS
───────────────────────────────────────────────────────────────
q, Ctrl+C       Quit the application
?               Show this help menu
Tab             Switch to next view
Shift+Tab       Switch to previous view
1, 2, 3, 4, 5   Jump to specific view (Graph, Devices, Connections, etc)
↑, ↓            Navigate up/down in lists
←, →            Navigate left/right (in some contexts)
Enter           Select item
Esc             Cancel/Go back
`,
			"routing": `
ROUTING & CONNECTIONS
───────────────────────────────────────────────────────────────
r               Toggle routing mode (connection creation)
c               Connect selected ports
d               Delete selected connection
Ctrl+R          Refresh graph
:               Enter command mode
`,
			"views": `
VIEWS
───────────────────────────────────────────────────────────────
Graph (1)       Visual representation of the audio graph
Devices (2)     List of audio devices and their properties
Connections (3) List of active audio connections/links
Properties (4)  Detailed properties of selected items
Stats (5)       Graph statistics and analysis
`,
			"commands": `
COMMAND MODE (press :)
───────────────────────────────────────────────────────────────
connect <out> <in>    Connect output port to input port
disconnect <id>       Disconnect a link
delete <id>           Delete a link
preset save <name>    Save current routing as preset
preset load <name>    Load a routing preset
preset list           List available presets
preset delete <name>  Delete a routing preset
search <text>         Search for nodes/ports
help <topic>          Show help on topic
info <id>             Show detailed info about a node/port
`,
			"tips": `
TIPS & TRICKS
───────────────────────────────────────────────────────────────
• Use arrow keys to navigate, Enter to select
• In routing mode, select output port then input port to connect
• Presets let you save and restore complex routing configurations
• Use Ctrl+R to refresh if the graph gets out of sync
• Search (/) can help find specific devices quickly
• Hover over items to see additional details in status bar
• Routing history can be undone with Ctrl+Z
`,
			"troubleshooting": `
TROUBLESHOOTING
───────────────────────────────────────────────────────────────
Q: No devices showing up?
A: Make sure PipeWire is running: systemctl --user status pipewire

Q: Can't connect ports?
A: Check that ports have compatible types (audio/MIDI)
   Verify directions are correct (output → input)

Q: Connection breaks after creating link?
A: This might be normal if nodes are busy
   Use Ctrl+R to refresh and verify the connection exists

Q: TUI is slow?
A: Try enabling compact mode in settings
   Reduce refresh interval if running on slow hardware
`,
		},
	}
}

// GetSection returns help text for a section
func (hm *HelpMenu) GetSection(section string) string {
	if text, exists := hm.sections[section]; exists {
		return text
	}
	return "Unknown help section. Available: navigation, routing, views, commands, tips, troubleshooting"
}

// GetAllHelp returns all help text
func (hm *HelpMenu) GetAllHelp() string {
	help := "PipeWire TUI - Complete Help\n"
	help += "=" + repeatString("=", 59) + "\n\n"

	sections := []string{"navigation", "routing", "views", "commands", "tips", "troubleshooting"}
	for _, section := range sections {
		help += hm.sections[section]
		help += "\n"
	}

	return help
}

// CommandParser parses and executes commands
type CommandParser struct {
	routingManager *RoutingManager
	presetManager  *PresetManager
	stateManager   *StateManager
}

// NewCommandParser creates a new command parser
func NewCommandParser(rm *RoutingManager, pm *PresetManager, sm *StateManager) *CommandParser {
	return &CommandParser{
		routingManager: rm,
		presetManager:  pm,
		stateManager:   sm,
	}
}

// ParseAndExecute parses a command and executes it
func (cp *CommandParser) ParseAndExecute(command string) (string, error) {
	// Parse command
	// Format: command arg1 arg2 ...
	
	// For now, return a placeholder
	return fmt.Sprintf("Command executed: %s", command), nil
}

// Logger for TUI debug output
type TUILogger struct {
	messages []string
	maxLines int
}

// NewTUILogger creates a new TUI logger
func NewTUILogger(maxLines int) *TUILogger {
	return &TUILogger{
		messages: make([]string, 0),
		maxLines: maxLines,
	}
}

// Log adds a log message
func (tl *TUILogger) Log(message string) {
	tl.messages = append(tl.messages, message)
	
	// Keep messages under control
	if len(tl.messages) > tl.maxLines {
		tl.messages = tl.messages[1:]
	}
}

// GetMessages returns all log messages
func (tl *TUILogger) GetMessages() []string {
	return tl.messages
}

// Clear clears all messages
func (tl *TUILogger) Clear() {
	tl.messages = make([]string, 0)
}

// Shortcuts represents keyboard shortcuts
type Shortcuts struct {
	shortcuts map[string]string
}

// NewShortcuts creates a new shortcuts manager
func NewShortcuts() *Shortcuts {
	return &Shortcuts{
		shortcuts: map[string]string{
			"q":          "Quit",
			"?":          "Help",
			"Tab":        "Next View",
			"Shift+Tab":  "Prev View",
			"↑/↓":        "Navigate",
			"Enter":      "Select",
			"r":          "Routing Mode",
			"c":          "Connect",
			"d":          "Delete",
			"Ctrl+R":     "Refresh",
			"Ctrl+Z":     "Undo",
			"Ctrl+Shift+Z": "Redo",
		},
	}
}

// GetShortcuts returns all shortcuts
func (s *Shortcuts) GetShortcuts() map[string]string {
	return s.shortcuts
}

// GetShortcutsString returns shortcuts as formatted string
func (s *Shortcuts) GetShortcutsString() string {
	output := "Common Shortcuts:\n"
	for key, action := range s.shortcuts {
		output += fmt.Sprintf("  %-12s - %s\n", key, action)
	}
	return output
}

// InfoPanel displays detailed information
type InfoPanel struct {
	title   string
	content string
	width   int
	height  int
}

// NewInfoPanel creates a new info panel
func NewInfoPanel(title string, width, height int) *InfoPanel {
	return &InfoPanel{
		title:   title,
		content: "",
		width:   width,
		height:  height,
	}
}

// SetContent sets the panel content
func (ip *InfoPanel) SetContent(content string) {
	ip.content = content
}

// Render renders the info panel
func (ip *InfoPanel) Render() string {
	output := "╔" + repeatString("═", ip.width-2) + "╗\n"
	output += "║ " + ip.title + repeatString(" ", ip.width-len(ip.title)-3) + "║\n"
	output += "╠" + repeatString("═", ip.width-2) + "╣\n"
	output += ip.content
	output += "╚" + repeatString("═", ip.width-2) + "╝"
	return output
}

// StatusBar displays status information
type StatusBar struct {
	width   int
	message string
	mode    string
	time    string
}

// NewStatusBar creates a new status bar
func NewStatusBar(width int) *StatusBar {
	return &StatusBar{
		width:   width,
		message: "Ready",
		mode:    "Normal",
		time:    "",
	}
}

// SetMessage sets the status message
func (sb *StatusBar) SetMessage(message string) {
	sb.message = message
}

// SetMode sets the current mode
func (sb *StatusBar) SetMode(mode string) {
	sb.mode = mode
}

// SetTime sets the current time display
func (sb *StatusBar) SetTime(time string) {
	sb.time = time
}

// Render renders the status bar
func (sb *StatusBar) Render() string {
	bar := fmt.Sprintf(" %s | %s | %s", sb.message, sb.mode, sb.time)
	bar += repeatString(" ", sb.width-len(bar)-1)
	return bar
}

// Notifications manages notification display
type Notifications struct {
	notifications []Notification
	maxNotifs     int
}

// Notification represents a single notification
type Notification struct {
	Type    string // "info", "warning", "error"
	Message string
	Created int64
}

// NewNotifications creates a new notifications manager
func NewNotifications(maxNotifs int) *Notifications {
	return &Notifications{
		notifications: make([]Notification, 0),
		maxNotifs:     maxNotifs,
	}
}

// Add adds a notification
func (n *Notifications) Add(notifType, message string) {
	notif := Notification{
		Type:    notifType,
		Message: message,
		Created: getTimestamp(),
	}
	
	n.notifications = append(n.notifications, notif)
	
	// Keep notifications under control
	if len(n.notifications) > n.maxNotifs {
		n.notifications = n.notifications[1:]
	}
}

// GetNotifications returns all notifications
func (n *Notifications) GetNotifications() []Notification {
	return n.notifications
}

// Clear clears all notifications
func (n *Notifications) Clear() {
	n.notifications = make([]Notification, 0)
}

// Helper function to get current timestamp
func getTimestamp() int64 {
	// Placeholder - would use time.Now().UnixNano() in real implementation
	return 0
}

// ProfileManager manages user profiles/workspaces
type ProfileManager struct {
	profiles map[string]*Profile
}

// Profile represents a user profile/workspace
type Profile struct {
	Name     string
	Routing  *RoutingPreset
	Settings map[string]string
}

// NewProfileManager creates a new profile manager
func NewProfileManager() *ProfileManager {
	return &ProfileManager{
		profiles: make(map[string]*Profile),
	}
}

// CreateProfile creates a new profile
func (pm *ProfileManager) CreateProfile(name string) *Profile {
	profile := &Profile{
		Name:     name,
		Routing:  &RoutingPreset{Name: name},
		Settings: make(map[string]string),
	}
	pm.profiles[name] = profile
	return profile
}

// GetProfile retrieves a profile
func (pm *ProfileManager) GetProfile(name string) *Profile {
	return pm.profiles[name]
}

// ListProfiles returns all profile names
func (pm *ProfileManager) ListProfiles() []string {
	names := make([]string, 0, len(pm.profiles))
	for name := range pm.profiles {
		names = append(names, name)
	}
	return names
}

// DeleteProfile removes a profile
func (pm *ProfileManager) DeleteProfile(name string) {
	delete(pm.profiles, name)
}
