// Package spa - Type definitions and constants
// spa/types.go
// Complete POD type definitions, enums, and utility functions

package spa

import (
	"encoding/json"
	"fmt"
)

// ===== POD Type Constants =====

const (
	PODTypeInvalid   uint32 = 0
	PODTypeNone      uint32 = 1
	PODTypeBool      uint32 = 2
	PODTypeInt       uint32 = 3
	PODTypeInt64     uint32 = 4
	PODTypeUint32    uint32 = 5
	PODTypeString    uint32 = 6
	PODTypeBytes     uint32 = 7
	PODTypeFd        uint32 = 8
	PODTypeArray     uint32 = 9
	PODTypeObject    uint32 = 10
	PODTypeFraction  uint32 = 11
	PODTypeRectangle uint32 = 12
	PODTypeID        uint32 = 13
)

// ===== Choice Type =====

type ChoiceType int

const (
	ChoiceTypeEnum ChoiceType = iota
	ChoiceTypeRange
	ChoiceTypeStep
)

func (c ChoiceType) String() string {
	switch c {
	case ChoiceTypeEnum:
		return "enum"
	case ChoiceTypeRange:
		return "range"
	case ChoiceTypeStep:
		return "step"
	default:
		return "unknown"
	}
}

func (c ChoiceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *ChoiceType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "enum":
		*c = ChoiceTypeEnum
	case "range":
		*c = ChoiceTypeRange
	case "step":
		*c = ChoiceTypeStep
	default:
		return fmt.Errorf("unknown choice type: %s", s)
	}
	return nil
}

// ===== Object Type =====

type ObjectType int

const (
	ObjectTypeCore ObjectType = iota
	ObjectTypeNode
	ObjectTypePort
	ObjectTypeLink
	ObjectTypeFactory
	ObjectTypeDevice
	ObjectTypeProfile
	ObjectTypeInterfaceNode
	ObjectTypeInterfacePort
	ObjectTypeInterfaceLink
)

func (o ObjectType) String() string {
	switch o {
	case ObjectTypeCore:
		return "core"
	case ObjectTypeNode:
		return "node"
	case ObjectTypePort:
		return "port"
	case ObjectTypeLink:
		return "link"
	case ObjectTypeFactory:
		return "factory"
	case ObjectTypeDevice:
		return "device"
	case ObjectTypeProfile:
		return "profile"
	case ObjectTypeInterfaceNode:
		return "interface_node"
	case ObjectTypeInterfacePort:
		return "interface_port"
	case ObjectTypeInterfaceLink:
		return "interface_link"
	default:
		return "unknown"
	}
}

func (o ObjectType) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.String())
}

func (o *ObjectType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "core":
		*o = ObjectTypeCore
	case "node":
		*o = ObjectTypeNode
	case "port":
		*o = ObjectTypePort
	case "link":
		*o = ObjectTypeLink
	case "factory":
		*o = ObjectTypeFactory
	case "device":
		*o = ObjectTypeDevice
	case "profile":
		*o = ObjectTypeProfile
	case "interface_node":
		*o = ObjectTypeInterfaceNode
	case "interface_port":
		*o = ObjectTypeInterfacePort
	case "interface_link":
		*o = ObjectTypeInterfaceLink
	default:
		return fmt.Errorf("unknown object type: %s", s)
	}
	return nil
}

// ===== Property Type =====

type PropType int

const (
	PropTypeUnknown PropType = iota
	PropTypeInfo
	PropTypeFormat
	PropTypeFilter
	PropTypeChannelMap
	PropTypeRoute
	PropTypeLatency
	PropTypeMedia
	PropTypeProfile
	PropTypeClass
	PropTypeRanges
	PropTypeEnumFormat
)

func (p PropType) String() string {
	switch p {
	case PropTypeInfo:
		return "info"
	case PropTypeFormat:
		return "format"
	case PropTypeFilter:
		return "filter"
	case PropTypeChannelMap:
		return "channel_map"
	case PropTypeRoute:
		return "route"
	case PropTypeLatency:
		return "latency"
	case PropTypeMedia:
		return "media"
	case PropTypeProfile:
		return "profile"
	case PropTypeClass:
		return "class"
	case PropTypeRanges:
		return "ranges"
	case PropTypeEnumFormat:
		return "enum_format"
	default:
		return "unknown"
	}
}

func (p PropType) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *PropType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "info":
		*p = PropTypeInfo
	case "format":
		*p = PropTypeFormat
	case "filter":
		*p = PropTypeFilter
	case "channel_map":
		*p = PropTypeChannelMap
	case "route":
		*p = PropTypeRoute
	case "latency":
		*p = PropTypeLatency
	case "media":
		*p = PropTypeMedia
	case "profile":
		*p = PropTypeProfile
	case "class":
		*p = PropTypeClass
	case "ranges":
		*p = PropTypeRanges
	case "enum_format":
		*p = PropTypeEnumFormat
	default:
		return fmt.Errorf("unknown prop type: %s", s)
	}
	return nil
}

// ===== Rectangle =====

type Rectangle struct {
	X, Y, W, H int32
}

func NewRectangle(x, y, w, h int32) *Rectangle {
	return &Rectangle{X: x, Y: y, W: w, H: h}
}

func (r *Rectangle) String() string {
	return fmt.Sprintf("Rectangle(%d, %d, %dx%d)", r.X, r.Y, r.W, r.H)
}

func (r *Rectangle) Area() int64 {
	return int64(r.W) * int64(r.H)
}

func (r *Rectangle) Marshal() ([]byte, error) {
	w := NewPODWriter()
	w.writeUint32(PODTypeRectangle)
	w.writeInt32(r.X)
	w.writeInt32(r.Y)
	w.writeInt32(r.W)
	w.writeInt32(r.H)
	return w.Bytes(), nil
}

func (r *Rectangle) Unmarshal(data []byte) error {
	if len(data) < 20 {
		return fmt.Errorf("data too short for rectangle")
	}
	r.X = int32(readUint32BE(data[4:8]))
	r.Y = int32(readUint32BE(data[8:12]))
	r.W = int32(readUint32BE(data[12:16]))
	r.H = int32(readUint32BE(data[16:20]))
	return nil
}

// ===== Fraction =====

type Fraction struct {
	Num, Den uint32
}

func NewFraction(num, den uint32) *Fraction {
	return &Fraction{Num: num, Den: den}
}

func (f *Fraction) String() string {
	return fmt.Sprintf("%d/%d", f.Num, f.Den)
}

func (f *Fraction) Value() float64 {
	if f.Den == 0 {
		return 0
	}
	return float64(f.Num) / float64(f.Den)
}

func (f *Fraction) Marshal() ([]byte, error) {
	w := NewPODWriter()
	w.writeUint32(PODTypeFraction)
	w.writeUint32(f.Num)
	w.writeUint32(f.Den)
	return w.Bytes(), nil
}

func (f *Fraction) Unmarshal(data []byte) error {
	if len(data) < 12 {
		return fmt.Errorf("data too short for fraction")
	}
	f.Num = readUint32BE(data[4:8])
	f.Den = readUint32BE(data[8:12])
	return nil
}

// ===== Helper Functions =====

// readUint32BE reads a big-endian uint32
func readUint32BE(data []byte) uint32 {
	if len(data) < 4 {
		return 0
	}
	return uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
}

// writeUint32BE writes a big-endian uint32
func writeUint32BE(data []byte, val uint32) {
	if len(data) < 4 {
		return
	}
	data[0] = byte((val >> 24) & 0xFF)
	data[1] = byte((val >> 16) & 0xFF)
	data[2] = byte((val >> 8) & 0xFF)
	data[3] = byte(val & 0xFF)
}

// PODTypeFromID converts a type ID to string
func PODTypeFromID(id uint32) string {
	switch id {
	case PODTypeInvalid:
		return "invalid"
	case PODTypeNone:
		return "none"
	case PODTypeBool:
		return "bool"
	case PODTypeInt:
		return "int"
	case PODTypeInt64:
		return "int64"
	case PODTypeUint32:
		return "uint32"
	case PODTypeString:
		return "string"
	case PODTypeBytes:
		return "bytes"
	case PODTypeFd:
		return "fd"
	case PODTypeArray:
		return "array"
	case PODTypeObject:
		return "object"
	case PODTypeFraction:
		return "fraction"
	case PODTypeRectangle:
		return "rectangle"
	case PODTypeID:
		return "id"
	default:
		return fmt.Sprintf("unknown(%d)", id)
	}
}

// PODTypeIDFromString converts a string to type ID
func PODTypeIDFromString(s string) uint32 {
	switch s {
	case "invalid":
		return PODTypeInvalid
	case "none":
		return PODTypeNone
	case "bool":
		return PODTypeBool
	case "int":
		return PODTypeInt
	case "int64":
		return PODTypeInt64
	case "uint32":
		return PODTypeUint32
	case "string":
		return PODTypeString
	case "bytes":
		return PODTypeBytes
	case "fd":
		return PODTypeFd
	case "array":
		return PODTypeArray
	case "object":
		return PODTypeObject
	case "fraction":
		return PODTypeFraction
	case "rectangle":
		return PODTypeRectangle
	case "id":
		return PODTypeID
	default:
		return PODTypeInvalid
	}
}
