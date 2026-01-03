// Package spa - Simple Protocol Audio
// spa/types.go
// POD type definitions and implementations
// Phase 3 - Complete type system

package spa

import "fmt"

// PODTypeID represents the type identifier for POD values
// NOTE: PODValue and PODType interfaces are defined in pod.go
type PODTypeID int

// POD Type IDs
const (
	PODTypeIDNone    PODTypeID = 0
	PODTypeIDNull    PODTypeID = 1
	PODTypeIDBool    PODTypeID = 2
	PODTypeIDInt     PODTypeID = 3
	PODTypeIDLong    PODTypeID = 4
	PODTypeIDFloat   PODTypeID = 5
	PODTypeIDDouble  PODTypeID = 6
	PODTypeIDString  PODTypeID = 7
	PODTypeIDBytes   PODTypeID = 8
	PODTypeIDArray   PODTypeID = 9
	PODTypeIDStruct  PODTypeID = 10
	PODTypeIDObject  PODTypeID = 11
	PODTypeIDChoice  PODTypeID = 12
)

// ChoiceType represents choice constraint type
type ChoiceType int

const (
	ChoiceTypeNone  ChoiceType = 0
	ChoiceTypeRange ChoiceType = 1
	ChoiceTypeEnum  ChoiceType = 2
	ChoiceTypeFlags ChoiceType = 3
	ChoiceTypeStep  ChoiceType = 4
)

// ObjectType represents object type identifier
type ObjectType int

const (
	ObjectTypeNone       ObjectType = 0
	ObjectTypeProps      ObjectType = 1
	ObjectTypeFormat     ObjectType = 2
	ObjectTypePortConfig ObjectType = 3
	ObjectTypeRouteSwitch ObjectType = 4
)

// PropType represents property type
type PropType int

const (
	PropTypeDevice   PropType = 0
	PropTypeNode     PropType = 1
	PropTypePort     PropType = 2
	PropTypeLink     PropType = 3
	PropTypeEndpoint PropType = 4
	PropTypeModule   PropType = 5
	PropTypeFactory  PropType = 6
)

// Rectangle represents a rectangular area
type Rectangle struct {
	X      int32
	Y      int32
	Width  uint32
	Height uint32
}

// Fraction represents a fraction
type Fraction struct {
	Numerator   uint32
	Denominator uint32
}

// ===== POD Value Types =====
// These are the concrete implementations of PODValue interface

// PODNull represents a null value
const PODNull = "null"

// PODBool represents a boolean value
type PODBool struct {
	Value bool
}

func (p *PODBool) Marshal() []byte {
	if p.Value {
		return []byte{1}
	}
	return []byte{0}
}

func (p *PODBool) Unmarshal(data []byte) error {
	if len(data) < 1 {
		return fmt.Errorf("insufficient data for bool")
	}
	p.Value = data[0] != 0
	return nil
}

func (p *PODBool) String() string {
	return fmt.Sprintf("%v", p.Value)
}

func (p *PODBool) Type() string {
	return "bool"
}

// PODInt represents an integer value
type PODInt struct {
	Value int32
}

func (p *PODInt) Marshal() []byte {
	return MarshalInt32(p.Value)
}

func (p *PODInt) Unmarshal(data []byte) error {
	val, err := UnmarshalInt32(data)
	if err != nil {
		return err
	}
	p.Value = val
	return nil
}

func (p *PODInt) String() string {
	return fmt.Sprintf("%d", p.Value)
}

func (p *PODInt) Type() string {
	return "int"
}

// PODLong represents a long integer value
type PODLong struct {
	Value int64
}

func (p *PODLong) Marshal() []byte {
	buf := make([]byte, 8)
	for i := 0; i < 8; i++ {
		buf[i] = byte((p.Value >> (uint(i) * 8)) & 0xFF)
	}
	return buf
}

func (p *PODLong) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("insufficient data for long")
	}
	p.Value = 0
	for i := 0; i < 8; i++ {
		p.Value |= int64(data[i]) << (uint(i) * 8)
	}
	return nil
}

func (p *PODLong) String() string {
	return fmt.Sprintf("%d", p.Value)
}

func (p *PODLong) Type() string {
	return "long"
}

// PODFloat represents a float value
type PODFloat struct {
	Value float32
}

func (p *PODFloat) Marshal() []byte {
	return MarshalInt32(int32(p.Value))
}

func (p *PODFloat) Unmarshal(data []byte) error {
	val, err := UnmarshalInt32(data)
	if err != nil {
		return err
	}
	p.Value = float32(val)
	return nil
}

func (p *PODFloat) String() string {
	return fmt.Sprintf("%f", p.Value)
}

func (p *PODFloat) Type() string {
	return "float"
}

// PODDouble represents a double value
type PODDouble struct {
	Value float64
}

func (p *PODDouble) Marshal() []byte {
	buf := make([]byte, 8)
	for i := 0; i < 8; i++ {
		buf[i] = byte((int64(p.Value) >> (uint(i) * 8)) & 0xFF)
	}
	return buf
}

func (p *PODDouble) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("insufficient data for double")
	}
	var val int64
	for i := 0; i < 8; i++ {
		val |= int64(data[i]) << (uint(i) * 8)
	}
	p.Value = float64(val)
	return nil
}

func (p *PODDouble) String() string {
	return fmt.Sprintf("%f", p.Value)
}

func (p *PODDouble) Type() string {
	return "double"
}

// PODString represents a string value
type PODString struct {
	Value string
}

func (p *PODString) Marshal() []byte {
	return MarshalString(p.Value)
}

func (p *PODString) Unmarshal(data []byte) error {
	str, err := UnmarshalString(data)
	if err != nil {
		return err
	}
	p.Value = str
	return nil
}

func (p *PODString) String() string {
	return p.Value
}

func (p *PODString) Type() string {
	return "string"
}

// PODBytes represents raw bytes value
type PODBytes struct {
	Data []byte
}

func (p *PODBytes) Marshal() []byte {
	return p.Data
}

func (p *PODBytes) Unmarshal(data []byte) error {
	p.Data = make([]byte, len(data))
	copy(p.Data, data)
	return nil
}

func (p *PODBytes) String() string {
	return fmt.Sprintf("%x", p.Data)
}

func (p *PODBytes) Type() string {
	return "bytes"
}

// PODArray represents an array of values
type PODArray struct {
	Items []PODValue
	Type  PODTypeID
}

func (p *PODArray) Marshal() []byte {
	// Marshal array: type + count + items
	result := MarshalUint32(uint32(p.Type))
	result = append(result, MarshalUint32(uint32(len(p.Items)))...)
	for _, item := range p.Items {
		result = append(result, item.Marshal()...)
	}
	return result
}

func (p *PODArray) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("insufficient data for array header")
	}
	// Parse type and count
	typeVal, _ := UnmarshalUint32(data[:4])
	count, _ := UnmarshalUint32(data[4:8])
	p.Type = PODTypeID(typeVal)
	p.Items = make([]PODValue, count)
	return nil
}

func (p *PODArray) String() string {
	return fmt.Sprintf("array[%d]", len(p.Items))
}

func (p *PODArray) Type() string {
	return "array"
}

// PODStruct represents a structured value
type PODStruct struct {
	Members map[string]PODValue
}

func (p *PODStruct) Marshal() []byte {
	result := []byte{}
	for key, val := range p.Members {
		result = append(result, MarshalString(key)...)
		result = append(result, val.Marshal()...)
	}
	return result
}

func (p *PODStruct) Unmarshal(data []byte) error {
	p.Members = make(map[string]PODValue)
	return nil
}

func (p *PODStruct) String() string {
	return fmt.Sprintf("struct{%d}", len(p.Members))
}

func (p *PODStruct) Type() string {
	return "struct"
}

// PODObject represents an object value
type PODObject struct {
	Type       ObjectType
	ID         uint32
	Properties map[string]PODValue
}

func (p *PODObject) Marshal() []byte {
	result := MarshalUint32(uint32(p.Type))
	result = append(result, MarshalUint32(p.ID)...)
	for key, val := range p.Properties {
		result = append(result, MarshalString(key)...)
		result = append(result, val.Marshal()...)
	}
	return result
}

func (p *PODObject) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("insufficient data for object header")
	}
	typeVal, _ := UnmarshalUint32(data[:4])
	p.Type = ObjectType(typeVal)
	p.ID, _ = UnmarshalUint32(data[4:8])
	p.Properties = make(map[string]PODValue)
	return nil
}

func (p *PODObject) String() string {
	return fmt.Sprintf("object{type=%d,id=%d}", p.Type, p.ID)
}

func (p *PODObject) Type() string {
	return "object"
}

// PODChoice represents a choice value
type PODChoice struct {
	ChoiceType ChoiceType
	Value      PODValue
	Min        PODValue
	Max        PODValue
}

func (p *PODChoice) Marshal() []byte {
	result := MarshalUint32(uint32(p.ChoiceType))
	result = append(result, p.Value.Marshal()...)
	if p.Min != nil {
		result = append(result, p.Min.Marshal()...)
	}
	if p.Max != nil {
		result = append(result, p.Max.Marshal()...)
	}
	return result
}

func (p *PODChoice) Unmarshal(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("insufficient data for choice type")
	}
	choiceType, _ := UnmarshalUint32(data[:4])
	p.ChoiceType = ChoiceType(choiceType)
	return nil
}

func (p *PODChoice) String() string {
	return fmt.Sprintf("choice{%v}", p.ChoiceType)
}

func (p *PODChoice) Type() string {
	return "choice"
}
