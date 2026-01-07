package main

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// buildVirtualNodeDialog creates the GTK4 dialog for creating virtual nodes
func buildVirtualNodeDialog(parent *gtk.ApplicationWindow) *gtk.Dialog {
	dialog := gtk.NewDialog()
	dialog.SetTransientFor(parent)
	dialog.SetModal(true)
	dialog.SetTitle("Create Virtual Node")
	dialog.SetDefaultSize(640, 480)

	content := dialog.ContentArea()
	content.SetMarginTop(12)
	content.SetMarginBottom(12)
	content.SetMarginStart(12)
	content.SetMarginEnd(12)

	// TODO: Build the full GTK4 layout here according to TECHNICAL-IMPLEMENTATION.md
	// - Type selector (Sink, Source, Filter, Loopback)
	// - Name/Description entries
	// - Audio configuration (channels, rate, bit depth, layout)
	// - Advanced options (Passive, Virtual, Exclusive, DontReconnect)
	// - Preset dropdown using getVirtualNodePresets()

	return dialog
}
