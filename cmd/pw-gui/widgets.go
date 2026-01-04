// Package widgets - GUI Widgets and Dialogs
// cmd/pw-gui/widgets.go
// Reusable GUI components

package main

import (
	"fmt"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/vignemail1/pipewire-go/client"
)

// DeviceListWidget represents a list of audio devices
type DeviceListWidget struct {
	box           *gtk.Box
	listView      *gtk.ListView
	listModel     *gtk.StringList
	selectedIndex int
	onSelect      func(nodeID uint32)
}

// NewDeviceListWidget creates a new device list widget
func NewDeviceListWidget(nodes map[uint32]*client.Node) *DeviceListWidget {
	items := []string{}
	for _, node := range nodes {
		items = append(items, fmt.Sprintf("[%d] %s", node.ID, node.Name()))
	}

	listModel := gtk.NewStringList(items)

	factory := gtk.NewSignalListItemFactory()
	factory.ConnectSetup(func(_ *gtk.SignalListItemFactory, obj *gtk.ListItem) {
		label := gtk.NewLabel("")
		obj.SetChild(label)
	})

	factory.ConnectBind(func(_ *gtk.SignalListItemFactory, obj *gtk.ListItem) {
		label := obj.Child().(*gtk.Label)
		item := obj.Item().(*gtk.StringObject)
		label.SetText(item.String())
	})

	listView := gtk.NewListView(gtk.NewSingleSelection(listModel), factory)

	box := gtk.NewBoxOrientation(gtk.OrientationVertical, 0)
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(listView)
	scrolled.SetPropagateNaturalWidth(true)
	scrolled.SetPropagateNaturalHeight(true)
	box.Append(scrolled)

	return &DeviceListWidget{
		box:       box,
		listView:  listView,
		listModel: listModel,
	}
}

// GetWidget returns the underlying widget
func (dlw *DeviceListWidget) GetWidget() gtk.Widgetter {
	return dlw.box
}

// GetSelected returns the selected item
func (dlw *DeviceListWidget) GetSelected() int {
	return dlw.selectedIndex
}

// PortListWidget represents a list of ports
type PortListWidget struct {
	box       *gtk.Box
	listView  *gtk.ListView
	listModel *gtk.StringList
	ports     []*client.Port
	onSelect  func(portID uint32)
}

// NewPortListWidget creates a new port list widget
func NewPortListWidget(ports []*client.Port) *PortListWidget {
	items := []string{}
	for _, port := range ports {
		direction := "→"
		if port.Direction == client.PortDirectionInput {
			direction = "←"
		}
		items = append(items, fmt.Sprintf("%s [%d] %s", direction, port.ID, port.Name))
	}

	listModel := gtk.NewStringList(items)

	factory := gtk.NewSignalListItemFactory()
	factory.ConnectSetup(func(_ *gtk.SignalListItemFactory, obj *gtk.ListItem) {
		label := gtk.NewLabel("")
		obj.SetChild(label)
	})

	factory.ConnectBind(func(_ *gtk.SignalListItemFactory, obj *gtk.ListItem) {
		label := obj.Child().(*gtk.Label)
		item := obj.Item().(*gtk.StringObject)
		label.SetText(item.String())
	})

	listView := gtk.NewListView(gtk.NewSingleSelection(listModel), factory)

	box := gtk.NewBoxOrientation(gtk.OrientationVertical, 0)
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(listView)
	scrolled.SetPropagateNaturalWidth(true)
	scrolled.SetPropagateNaturalHeight(true)
	box.Append(scrolled)

	return &PortListWidget{
		box:       box,
		listView:  listView,
		listModel: listModel,
		ports:     ports,
	}
}

// GetWidget returns the underlying widget
func (plw *PortListWidget) GetWidget() gtk.Widgetter {
	return plw.box
}

// PropertiesPanel displays object properties
type PropertiesPanel struct {
	box    *gtk.Box
	scroll *gtk.ScrolledWindow
	label  *gtk.Label
}

// NewPropertiesPanel creates a new properties panel
func NewPropertiesPanel() *PropertiesPanel {
	label := gtk.NewLabel("")
	label.SetWrapMode(gtk.WrapWord)

	scroll := gtk.NewScrolledWindow()
	scroll.SetChild(label)

	box := gtk.NewBoxOrientation(gtk.OrientationVertical, 0)
	box.Append(scroll)

	return &PropertiesPanel{
		box:    box,
		scroll: scroll,
		label:  label,
	}
}

// SetProperties sets the properties to display
func (pp *PropertiesPanel) SetProperties(props map[string]string) {
	text := ""
	for key, value := range props {
		text += fmt.Sprintf("%s: %s\n", key, value)
	}
	pp.label.SetText(text)
}

// GetWidget returns the underlying widget
func (pp *PropertiesPanel) GetWidget() gtk.Widgetter {
	return pp.box
}

// AudioMeterWidget displays audio level meter
type AudioMeterWidget struct {
	box      *gtk.Box
	progress *gtk.ProgressBar
	label    *gtk.Label
	level    float64
}

// NewAudioMeterWidget creates a new audio meter widget
func NewAudioMeterWidget(label string) *AudioMeterWidget {
	lbl := gtk.NewLabel(label)
	progress := gtk.NewProgressBar()
	progress.SetFraction(0.0)

	box := gtk.NewBoxOrientation(gtk.OrientationHorizontal, 5)
	box.Append(lbl)
	box.Append(progress)

	return &AudioMeterWidget{
		box:      box,
		progress: progress,
		label:    lbl,
		level:    0.0,
	}
}

// SetLevel sets the meter level (0.0 to 1.0)
func (amw *AudioMeterWidget) SetLevel(level float64) {
	if level < 0.0 {
		level = 0.0
	}
	if level > 1.0 {
		level = 1.0
	}
	amw.level = level
	amw.progress.SetFraction(level)
}

// GetWidget returns the underlying widget
func (amw *AudioMeterWidget) GetWidget() gtk.Widgetter {
	return amw.box
}

// PresetComboWidget represents a preset selector
type PresetComboWidget struct {
	box       *gtk.Box
	combo     *gtk.ComboBoxText
	onSelect  func(preset string)
}

// NewPresetComboWidget creates a new preset combo widget
func NewPresetComboWidget(presets []string) *PresetComboWidget {
	combo := gtk.NewComboBoxText()
	for _, preset := range presets {
		combo.AppendText(preset)
	}

	label := gtk.NewLabel("Presets:")
	box := gtk.NewBoxOrientation(gtk.OrientationHorizontal, 5)
	box.Append(label)
	box.Append(combo)

	return &PresetComboWidget{
		box:   box,
		combo: combo,
	}
}

// GetWidget returns the underlying widget
func (pcw *PresetComboWidget) GetWidget() gtk.Widgetter {
	return pcw.box
}

// GetSelected returns the selected preset
func (pcw *PresetComboWidget) GetSelected() string {
	return pcw.combo.ActiveID()
}

// NotificationWidget displays notifications
type NotificationWidget struct {
	box       *gtk.Box
	message   *gtk.Label
	closeBtn  *gtk.Button
	visible   bool
	notifType string
}

// NewNotificationWidget creates a new notification widget
func NewNotificationWidget() *NotificationWidget {
	message := gtk.NewLabel("")
	message.SetWrapMode(gtk.WrapWord)

	closeBtn := gtk.NewButtonWithLabel("×")

	box := gtk.NewBoxOrientation(gtk.OrientationHorizontal, 10)
	box.SetMarginTop(10)
	box.SetMarginBottom(10)
	box.SetMarginStart(10)
	box.SetMarginEnd(10)
	box.Append(message)
	box.Append(closeBtn)

	return &NotificationWidget{
		box:      box,
		message:  message,
		closeBtn: closeBtn,
		visible:  false,
	}
}

// Show displays a notification
func (nw *NotificationWidget) Show(message string, notifType string) {
	nw.message.SetText(message)
	nw.notifType = notifType
	nw.visible = true

	// Apply color based on type
	switch notifType {
	case "info":
		nw.box.AddCSSClass("info-notification")
	case "warning":
		nw.box.AddCSSClass("warning-notification")
	case "error":
		nw.box.AddCSSClass("error-notification")
	}
}

// Hide hides the notification
func (nw *NotificationWidget) Hide() {
	nw.visible = false
	nw.message.SetText("")
}

// GetWidget returns the underlying widget
func (nw *NotificationWidget) GetWidget() gtk.Widgetter {
	return nw.box
}

// SearchWidget represents a search/filter widget
type SearchWidget struct {
	box       *gtk.Box
	entry     *gtk.SearchEntry
	onSearch  func(query string)
}

// NewSearchWidget creates a new search widget
func NewSearchWidget() *SearchWidget {
	entry := gtk.NewSearchEntry()
	entry.SetPlaceholderText("Search devices, ports...")

	box := gtk.NewBoxOrientation(gtk.OrientationHorizontal, 0)
	box.Append(entry)

	return &SearchWidget{
		box:   box,
		entry: entry,
	}
}

// GetWidget returns the underlying widget
func (sw *SearchWidget) GetWidget() gtk.Widgetter {
	return sw.box
}

// GetSearchQuery returns the current search query
func (sw *SearchWidget) GetSearchQuery() string {
	return sw.entry.Text()
}

// InfoPanel displays detailed information
type InfoPanel struct {
	box    *gtk.Box
	title  *gtk.Label
	scroll *gtk.ScrolledWindow
	text   *gtk.Label
}

// NewInfoPanel creates a new info panel
func NewInfoPanel(title string) *InfoPanel {
	titleLabel := gtk.NewLabel(title)
	titleLabel.AddCSSClass("title")

	textLabel := gtk.NewLabel("")
	textLabel.SetWrapMode(gtk.WrapWord)

	scroll := gtk.NewScrolledWindow()
	scroll.SetChild(textLabel)

	box := gtk.NewBoxOrientation(gtk.OrientationVertical, 5)
	box.Append(titleLabel)
	box.Append(scroll)

	return &InfoPanel{
		box:    box,
		title:  titleLabel,
		scroll: scroll,
		text:   textLabel,
	}
}

// SetContent sets the panel content
func (ip *InfoPanel) SetContent(content string) {
	ip.text.SetText(content)
}

// GetWidget returns the underlying widget
func (ip *InfoPanel) GetWidget() gtk.Widgetter {
	return ip.box
}
