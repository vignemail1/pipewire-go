// Package core - Protocol method and event types
// core/method.go
// Defines request/response types for PipeWire protocol communication

package core

import (
	"fmt"
	"github.com/vignemail1/pipewire-go/spa"
)

// RegistryBindRequest represents a registry.bind() method call
// Used to bind to a global object and receive its events
type RegistryBindRequest struct {
	ClientObjectID uint32                 // Object ID of the requesting client
	VersionID      uint32                 // Interface version requested
	Type           string                 // Interface type (e.g., "Node", "Link")
	Properties     map[string]interface{} // Additional properties
}

// ToPOD converts RegistryBindRequest to POD object format
func (r *RegistryBindRequest) ToPOD() (*spa.PODObject, error) {
	if r == nil {
		return nil, fmt.Errorf("RegistryBindRequest is nil")
	}

	obj := &spa.PODObject{
		Values: make([]spa.PODValue, 0),
	}

	// Add client object ID
	obj.Values = append(obj.Values, &spa.PODString{Value: "client-object-id"})
	obj.Values = append(obj.Values, &spa.PODUint32{Value: r.ClientObjectID})

	// Add version
	obj.Values = append(obj.Values, &spa.PODString{Value: "version"})
	obj.Values = append(obj.Values, &spa.PODUint32{Value: r.VersionID})

	// Add type
	obj.Values = append(obj.Values, &spa.PODString{Value: "type"})
	obj.Values = append(obj.Values, &spa.PODString{Value: r.Type})

	// Add properties if present
	if len(r.Properties) > 0 {
		obj.Values = append(obj.Values, &spa.PODString{Value: "properties"})
		propsObj := &spa.PODObject{Values: make([]spa.PODValue, 0)}
		for k, v := range r.Properties {
			propsObj.Values = append(propsObj.Values, &spa.PODString{Value: k})
			switch val := v.(type) {
			case string:
				propsObj.Values = append(propsObj.Values, &spa.PODString{Value: val})
			case uint32:
				propsObj.Values = append(propsObj.Values, &spa.PODUint32{Value: val})
			case bool:
				propsObj.Values = append(propsObj.Values, &spa.PODBool{Value: val})
			default:
				propsObj.Values = append(propsObj.Values, &spa.PODString{Value: fmt.Sprintf("%v", val)})
			}
		}
		obj.Values = append(obj.Values, propsObj)
	}

	return obj, nil
}

// LinkCreateRequest represents a link creation request
// Used to create a connection between output and input ports
type LinkCreateRequest struct {
	OutputPortID uint32            // Source port ID
	InputPortID  uint32            // Destination port ID
	Properties   map[string]string // Link properties (e.g., "passive")
	Passive      bool              // Whether link is passive
}

// ToPOD converts LinkCreateRequest to POD object format
func (r *LinkCreateRequest) ToPOD() (*spa.PODObject, error) {
	if r == nil {
		return nil, fmt.Errorf("LinkCreateRequest is nil")
	}

	obj := &spa.PODObject{
		Values: make([]spa.PODValue, 0),
	}

	// Add port IDs
	obj.Values = append(obj.Values, &spa.PODString{Value: "output-port-id"})
	obj.Values = append(obj.Values, &spa.PODUint32{Value: r.OutputPortID})

	obj.Values = append(obj.Values, &spa.PODString{Value: "input-port-id"})
	obj.Values = append(obj.Values, &spa.PODUint32{Value: r.InputPortID})

	// Add passive flag
	obj.Values = append(obj.Values, &spa.PODString{Value: "passive"})
	obj.Values = append(obj.Values, &spa.PODBool{Value: r.Passive})

	// Add properties
	if len(r.Properties) > 0 {
		obj.Values = append(obj.Values, &spa.PODString{Value: "properties"})
		propsObj := &spa.PODObject{Values: make([]spa.PODValue, 0)}
		for k, v := range r.Properties {
			propsObj.Values = append(propsObj.Values, &spa.PODString{Value: k})
			propsObj.Values = append(propsObj.Values, &spa.PODString{Value: v})
		}
		obj.Values = append(obj.Values, propsObj)
	}

	return obj, nil
}

// LinkDestroyRequest represents a link destruction request
type LinkDestroyRequest struct {
	LinkID uint32 // Link object ID to destroy
}

// ToPOD converts LinkDestroyRequest to POD object format
func (r *LinkDestroyRequest) ToPOD() (*spa.PODObject, error) {
	if r == nil {
		return nil, fmt.Errorf("LinkDestroyRequest is nil")
	}

	obj := &spa.PODObject{
		Values: make([]spa.PODValue, 0),
	}

	obj.Values = append(obj.Values, &spa.PODString{Value: "link-id"})
	obj.Values = append(obj.Values, &spa.PODUint32{Value: r.LinkID})

	return obj, nil
}

// RegistryGlobalEvent represents a registry global event
// Sent by daemon when new object is available
type RegistryGlobalEvent struct {
	ObjectID    uint32            // ID of the new object
	Type        string            // Interface type
	Version     uint32            // Interface version
	Properties  map[string]string // Object properties
}

// FromPOD parses RegistryGlobalEvent from POD object
func (e *RegistryGlobalEvent) FromPOD(obj *spa.PODObject) error {
	if e == nil {
		return fmt.Errorf("RegistryGlobalEvent is nil")
	}
	if obj == nil {
		return fmt.Errorf("POD object is nil")
	}

	e.Properties = make(map[string]string)

	// Extract fields from POD
	objID, err := ExtractUint32FromPOD(obj, "object-id")
	if err == nil {
		e.ObjectID = objID
	}

	t, err := ExtractStringFromPOD(obj, "type")
	if err == nil {
		e.Type = t
	}

	v, err := ExtractUint32FromPOD(obj, "version")
	if err == nil {
		e.Version = v
	}

	// Extract properties
	if propsVal, ok := obj.Get("properties"); ok {
		if propsObj, ok := propsVal.(*spa.PODObject); ok {
			for i := 0; i < len(propsObj.Values)-1; i += 2 {
				if keyVal, ok := propsObj.Values[i].(*spa.PODString); ok {
					if valVal, ok := propsObj.Values[i+1].(*spa.PODString); ok {
						e.Properties[keyVal.Value] = valVal.Value
					}
				}
			}
		}
	}

	return nil
}

// LinkInfoEvent represents a link info event
// Sent by daemon when link state changes
type LinkInfoEvent struct {
	LinkID       uint32            // Link object ID
	InputPortID  uint32            // Connected input port
	OutputPortID uint32            // Connected output port
	State        uint32            // Link state (0=error, 1=freewheeling, 2=active)
	Properties   map[string]string // Link properties
}

// FromPOD parses LinkInfoEvent from POD object
func (e *LinkInfoEvent) FromPOD(obj *spa.PODObject) error {
	if e == nil {
		return fmt.Errorf("LinkInfoEvent is nil")
	}
	if obj == nil {
		return fmt.Errorf("POD object is nil")
	}

	e.Properties = make(map[string]string)

	// Extract IDs
	linkID, err := ExtractUint32FromPOD(obj, "link-id")
	if err == nil {
		e.LinkID = linkID
	}

	inputID, err := ExtractUint32FromPOD(obj, "input-port-id")
	if err == nil {
		e.InputPortID = inputID
	}

	outputID, err := ExtractUint32FromPOD(obj, "output-port-id")
	if err == nil {
		e.OutputPortID = outputID
	}

	state, err := ExtractUint32FromPOD(obj, "state")
	if err == nil {
		e.State = state
	}

	// Extract properties
	if propsVal, ok := obj.Get("properties"); ok {
		if propsObj, ok := propsVal.(*spa.PODObject); ok {
			for i := 0; i < len(propsObj.Values)-1; i += 2 {
				if keyVal, ok := propsObj.Values[i].(*spa.PODString); ok {
					if valVal, ok := propsObj.Values[i+1].(*spa.PODString); ok {
						e.Properties[keyVal.Value] = valVal.Value
					}
				}
			}
		}
	}

	return nil
}
