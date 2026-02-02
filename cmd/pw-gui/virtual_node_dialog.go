package main

import (
	"fmt"
	"log"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	core "github.com/vignemail1/pipewire-go/core"
)

// VirtualNodeCreationDialog is the main dialog for creating virtual nodes
type VirtualNodeCreationDialog struct {
	dialog            *gtk.Dialog
	client            *core.Client
	parent            *gtk.ApplicationWindow
	typeCombo         *gtk.ComboBoxText
	presetCombo       *gtk.ComboBoxText
	nameEntry         *gtk.Entry
	descEntry         *gtk.Entry
	channelsSpinner   *gtk.SpinButton
	rateCombo         *gtk.ComboBoxText
	bitDepthCombo     *gtk.ComboBoxText
	layoutCombo       *gtk.ComboBoxText
	passiveCheck      *gtk.CheckButton
	virtualCheck      *gtk.CheckButton
	exclusiveCheck    *gtk.CheckButton
	dontReconnectCheck *gtk.CheckButton
	errorLabel        *gtk.Label
}

// NewVirtualNodeCreationDialog creates a new GTK4 dialog for virtual node creation
func NewVirtualNodeCreationDialog(parent *gtk.ApplicationWindow, client *core.Client) *VirtualNodeCreationDialog {
	d := &VirtualNodeCreationDialog{
		client: client,
		parent: parent,
	}

	d.build()
	return d
}

func (d *VirtualNodeCreationDialog) build() {
	d.dialog = gtk.NewDialog()
	d.dialog.SetTransientFor(d.parent)
	d.dialog.SetModal(true)
	d.dialog.SetTitle("Create Virtual Node")
	d.dialog.SetDefaultSize(700, 600)

	// Content area
	content := d.dialog.ContentArea()
	content.SetMarginTop(12)
	content.SetMarginBottom(12)
	content.SetMarginStart(12)
	content.SetMarginEnd(12)

	// Main vertical box
	mainBox := gtk.NewBoxWithOrientation(gtk.OrientationVertical, 6)

	// === SECTION 1: Type & Preset ===

typeFrame := gtk.NewFrame("Node Type & Template")
	typeBox := gtk.NewBoxWithOrientation(gtk.OrientationVertical, 6)
	typeBox.SetMarginTop(6)
	typeBox.SetMarginBottom(6)
	typeBox.SetMarginStart(6)
	typeBox.SetMarginEnd(6)

	// Type selector
	typeHBox := gtk.NewBoxWithOrientation(gtk.OrientationHorizontal, 6)
	typeLabel := gtk.NewLabel("Node Type:")
	typeLabel.SetWidthRequest(120)
	d.typeCombo = gtk.NewComboBoxText()
	d.typeCombo.Append("sink", "Audio Sink (Output)")
	d.typeCombo.Append("source", "Audio Source (Input)")
	d.typeCombo.Append("filter", "Audio Filter (Processing)")
	d.typeCombo.Append("loopback", "Virtual Loopback")
	d.typeCombo.SetActive(0)
	typeHBox.Append(typeLabel)
	typeHBox.SetHExpand(true)
	typeHBox.Append(d.typeCombo)
	typeBox.Append(typeHBox)

	// Preset selector
	presetHBox := gtk.NewBoxWithOrientation(gtk.OrientationHorizontal, 6)
	presetLabel := gtk.NewLabel("Preset Template:")
	presetLabel.SetWidthRequest(120)
	d.presetCombo = gtk.NewComboBoxText()
	d.populatePresets()
	presetHBox.Append(presetLabel)
	presetHBox.SetHExpand(true)
	presetHBox.Append(d.presetCombo)
	typeBox.Append(presetHBox)
	typeFrame.SetChild(typeBox)
	mainBox.Append(typeFrame)

	// === SECTION 2: Basic Properties ===
	basicFrame := gtk.NewFrame("Basic Properties")
	basicBox := gtk.NewBoxWithOrientation(gtk.OrientationVertical, 6)
	basicBox.SetMarginTop(6)
	basicBox.SetMarginBottom(6)
	basicBox.SetMarginStart(6)
	basicBox.SetMarginEnd(6)

	// Name
	nameHBox := gtk.NewBoxWithOrientation(gtk.OrientationHorizontal, 6)
	nameLabel := gtk.NewLabel("Name:")
	nameLabel.SetWidthRequest(120)
	d.nameEntry = gtk.NewEntry()
	d.nameEntry.SetPlaceholderText("e.g., Recording Sink")
	nameHBox.Append(nameLabel)
	nameHBox.SetHExpand(true)
	nameHBox.Append(d.nameEntry)
	basicBox.Append(nameHBox)

	// Description
	descHBox := gtk.NewBoxWithOrientation(gtk.OrientationHorizontal, 6)
	descLabel := gtk.NewLabel("Description:")
	descLabel.SetWidthRequest(120)
	d.descEntry = gtk.NewEntry()
	d.descEntry.SetPlaceholderText("Optional description")
	descHBox.Append(descLabel)
	descHBox.SetHExpand(true)
	descHBox.Append(d.descEntry)
	basicBox.Append(descHBox)
	basicFrame.SetChild(basicBox)
	mainBox.Append(basicFrame)

	// === SECTION 3: Audio Configuration ===
	audioFrame := gtk.NewFrame("Audio Configuration")
	audioBox := gtk.NewBoxWithOrientation(gtk.OrientationVertical, 6)
	audioBox.SetMarginTop(6)
	audioBox.SetMarginBottom(6)
	audioBox.SetMarginStart(6)
	audioBox.SetMarginEnd(6)

	// Channels
	channelsHBox := gtk.NewBoxWithOrientation(gtk.OrientationHorizontal, 6)
	channelsLabel := gtk.NewLabel("Channels:")
	channelsLabel.SetWidthRequest(120)
	d.channelsSpinner = gtk.NewSpinButtonWithRange(1, 8, 1)
	d.channelsSpinner.SetValue(2)
	channelsHBox.Append(channelsLabel)
	channelsHBox.SetHExpand(true)
	channelsHBox.Append(d.channelsSpinner)
	audioBox.Append(channelsHBox)

	// Sample Rate
	rateHBox := gtk.NewBoxWithOrientation(gtk.OrientationHorizontal, 6)
	rateLabel := gtk.NewLabel("Sample Rate:")
	rateLabel.SetWidthRequest(120)
	d.rateCombo = gtk.NewComboBoxText()
	d.rateCombo.Append("44100", "44.1 kHz")
	d.rateCombo.Append("48000", "48 kHz")
	d.rateCombo.Append("96000", "96 kHz")
	d.rateCombo.Append("192000", "192 kHz")
	d.rateCombo.SetActive(1) // Default 48kHz
	rateHBox.Append(rateLabel)
	rateHBox.SetHExpand(true)
	rateHBox.Append(d.rateCombo)
	audioBox.Append(rateHBox)

	// Bit Depth
	bitDepthHBox := gtk.NewBoxWithOrientation(gtk.OrientationHorizontal, 6)
	bitDepthLabel := gtk.NewLabel("Bit Depth:")
	bitDepthLabel.SetWidthRequest(120)
	d.bitDepthCombo = gtk.NewComboBoxText()
	d.bitDepthCombo.Append("16", "16-bit")
	d.bitDepthCombo.Append("24", "24-bit")
	d.bitDepthCombo.Append("32", "32-bit")
	d.bitDepthCombo.SetActive(2) // Default 32-bit
	bitDepthHBox.Append(bitDepthLabel)
	bitDepthHBox.SetHExpand(true)
	bitDepthHBox.Append(d.bitDepthCombo)
	audioBox.Append(bitDepthHBox)

	// Channel Layout
	layoutHBox := gtk.NewBoxWithOrientation(gtk.OrientationHorizontal, 6)
	layoutLabel := gtk.NewLabel("Channel Layout:")
	layoutLabel.SetWidthRequest(120)
	d.layoutCombo = gtk.NewComboBoxText()
	d.layoutCombo.Append("FL,FR", "Stereo (L/R)")
	d.layoutCombo.Append("FL,FR,FC", "2.1 (L/R/Center)")
	d.layoutCombo.Append("FL,FR,RL,RR", "Quad (L/R/Rear)")
	d.layoutCombo.Append("FL,FR,FC,RL,RR,LFE", "5.1 Surround")
	d.layoutCombo.SetActive(0) // Default stereo
	layoutHBox.Append(layoutLabel)
	layoutHBox.SetHExpand(true)
	layoutHBox.Append(d.layoutCombo)
	audioBox.Append(layoutHBox)
	audioFrame.SetChild(audioBox)
	mainBox.Append(audioFrame)

	// === SECTION 4: Advanced Options ===
	advFrame := gtk.NewFrame("Advanced Options")
	advBox := gtk.NewBoxWithOrientation(gtk.OrientationVertical, 6)
	advBox.SetMarginTop(6)
	advBox.SetMarginBottom(6)
	advBox.SetMarginStart(6)
	advBox.SetMarginEnd(6)

	d.passiveCheck = gtk.NewCheckButtonWithLabel("Passive (don't route audio automatically)")
	advBox.Append(d.passiveCheck)

	d.virtualCheck = gtk.NewCheckButtonWithLabel("Virtual (internal node, hidden from external apps)")
	advBox.Append(d.virtualCheck)

	d.exclusiveCheck = gtk.NewCheckButtonWithLabel("Exclusive (exclusive access to device)")
	advBox.Append(d.exclusiveCheck)

	d.dontReconnectCheck = gtk.NewCheckButtonWithLabel("Don't Reconnect (prevent auto-reconnection)")
	advBox.Append(d.dontReconnectCheck)

	advFrame.SetChild(advBox)
	mainBox.Append(advFrame)

	// === ERROR DISPLAY ===
	d.errorLabel = gtk.NewLabel("")
	d.errorLabel.SetMarkup("")
	mainBox.Append(d.errorLabel)

	// Wrap in scrollable
	scroll := gtk.NewScrolledWindow()
	scroll.SetChild(mainBox)
	scroll.SetVExpand(true)
	content.Append(scroll)

	// Action buttons
	d.dialog.AddButton("Cancel", int(gtk.ResponseCancel))
	d.dialog.AddButton("Create", int(gtk.ResponseAccept))

	// Connect signal
	d.dialog.ConnectResponse(func(responseID int) {
		if gtk.ResponseType(responseID) == gtk.ResponseAccept {
			d.onCreateClicked()
		}
	})
}

func (d *VirtualNodeCreationDialog) populatePresets() {
	presets := getVirtualNodePresets()
	for _, p := range presets {
		d.presetCombo.Append(p.ID, formatPresetLabel(p))
	}
	if len(presets) > 0 {
		d.presetCombo.SetActive(0)
	}
}

func (d *VirtualNodeCreationDialog) onCreateClicked() {
	// Validate inputs
	name, _ := d.nameEntry.Text()
	if name == "" {
		d.showError("Node name cannot be empty")
		return
	}

	// Build config
	config := core.VirtualNodeConfig{
		Name:        name,
		Description: d.getDescription(),
		Type:        d.getNodeType(),
		Factory:     d.getFactory(),
		Channels:    int(d.channelsSpinner.Value()),
		SampleRate:  d.getSampleRate(),
		BitDepth:    d.getBitDepth(),
		Layout:      d.getLayout(),
		Passive:     d.passiveCheck.Active(),
		Virtual:     d.virtualCheck.Active(),
		Exclusive:   d.exclusiveCheck.Active(),
		DontReconnect: d.dontReconnectCheck.Active(),
	}

	// Validate config
	if err := config.Validate(); err != nil {
		d.showError(fmt.Sprintf("Invalid configuration: %v", err))
		return
	}

	// Create virtual node
	virtualNode, err := d.client.CreateVirtualNode(config)
	if err != nil {
		d.showError(fmt.Sprintf("Failed to create node: %v", err))
		return
	}

	log.Printf("Virtual node created successfully (ID: %d)\n", virtualNode.ID)
	d.dialog.Close()
}

func (d *VirtualNodeCreationDialog) showError(msg string) {
	d.errorLabel.SetMarkup(fmt.Sprintf("<span foreground='red'><b>Error:</b> %s</span>", msg))
}

func (d *VirtualNodeCreationDialog) getNodeType() core.VirtualNodeType {
	active := d.typeCombo.ActiveID()
	switch active {
	case "source":
		return core.VirtualNode_Source
	case "filter":
		return core.VirtualNode_Filter
	case "loopback":
		return core.VirtualNode_Loopback
	default:
		return core.VirtualNode_Sink
	}
}

func (d *VirtualNodeCreationDialog) getFactory() core.VirtualNodeFactory {
	// Determine factory based on type
	nodeType := d.getNodeType()
	switch nodeType {
	case core.VirtualNode_Source:
		return core.Factory_NullAudioSource
	case core.VirtualNode_Loopback:
		return core.Factory_Loopback
	case core.VirtualNode_Filter:
		return core.Factory_FilterChain
	default:
		return core.Factory_NullAudioSink
	}
}

func (d *VirtualNodeCreationDialog) getSampleRate() int {
	active := d.rateCombo.ActiveID()
	switch active {
	case "44100":
		return 44100
	case "96000":
		return 96000
	case "192000":
		return 192000
	default:
		return 48000
	}
}

func (d *VirtualNodeCreationDialog) getBitDepth() int {
	active := d.bitDepthCombo.ActiveID()
	switch active {
	case "16":
		return 16
	case "24":
		return 24
	default:
		return 32
	}
}

func (d *VirtualNodeCreationDialog) getLayout() string {
	return d.layoutCombo.ActiveID()
}

func (d *VirtualNodeCreationDialog) getDescription() string {
	desc, _ := d.descEntry.Text()
	return desc
}

// Show displays the dialog and returns the response
func (d *VirtualNodeCreationDialog) Show() {
	d.dialog.Show()
}
