package main

import (
	"fmt"

	core "github.com/vignemail1/pipewire-go/core"
)

// VirtualNodePreset represents a GUI-friendly preset description
type VirtualNodePreset struct {
	ID          string
	Name        string
	Description string
}

// getVirtualNodePresets returns the list of presets for the GUI dropdown
func getVirtualNodePresets() []VirtualNodePreset {
	names := core.GetVirtualNodePresetNames()
	presets := make([]VirtualNodePreset, 0, len(names))

	for _, id := range names {
		config := core.GetVirtualNodePreset(id)
		preset := VirtualNodePreset{
			ID:          id,
			Name:        config.Name,
			Description: config.Description,
		}
		presets = append(presets, preset)
	}

	return presets
}

// formatPresetLabel formats a preset for display in the dropdown
func formatPresetLabel(p VirtualNodePreset) string {
	if p.Description == "" {
		return p.Name
	}
	return fmt.Sprintf("%s - %s", p.Name, p.Description)
}
