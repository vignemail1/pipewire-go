// Package spa - POD (Plain Old Data) marshaling and unmarshaling
// spa/pod.go
// Complete implementation with proper alignment and endianness

package spa

import (
	"encoding/binary"
	"fmt"
	"math"
)

// ============================================================================
// POD VALUE INTERFACE
// ============================================================================

// PODValue represents any POD value (marshals/unmarshals itself)
type PODValue interface {
	Type() PODType
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
	String() string
}

// PODType interface for POD types
type PODType interface {
	ID() uint32
	Name() string
}

// BasePODType basic implementation
type BasePODType struct {
	id   uint32
	name string
}

func (t *BasePODType) ID() uint32   { return t.id }
func (t *BasePODType) Name() string { return t.name }

// ============================================================================
// PRIMITIVE POD TYPES
// ============================================================================

// PODInt8 - 8-bit signed integer
type PODInt8 struct{ Value int8 }

func NewPODInt8(val int8) *PODInt8 { return &PODInt8{Value: val} }
func (v *PODInt8) Type() PODType   { return &BasePODType{id: PODTypeInt, name: "int8"} }
func (v *PODInt8) String() string  { return fmt.Sprintf("int8(%d)", v.Value) }
func (v *PODInt8) Marshal() ([]byte, error) {
	return []byte{byte(v.Value)}, nil
}
func (v *PODInt8) Unmarshal(data []byte) error {
	if len(data) < 1 {
		return fmt.Errorf("insufficient data for int8")
	}
	v.Value = int8(data[0])
	return nil
}

// PODInt16 - 16-bit signed integer
type PODInt16 struct{ Value int16 }

func NewPODInt16(val int16) *PODInt16 { return &PODInt16{Value: val} }
func (v *PODInt16) Type() PODType     { return &BasePODType{id: PODTypeInt, name: "int16"} }
func (v *PODInt16) String() string    { return fmt.Sprintf("int16(%d)", v.Value) }
func (v *PODInt16) Marshal() ([]byte, error) {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(v.Value))
	return b, nil
}
func (v *PODInt16) Unmarshal(data []byte) error {
	if len(data) < 2 {
		return fmt.Errorf("insufficient data for int16")
	}
	v.Value = int16(binary.LittleEndian.Uint16(data))
	return nil
}

// PODInt32 - 32-bit signed integer
type PODInt32 struct{ Value int32 }

func NewPODInt32(val int32) *PODInt32 { return &PODInt32{Value: val} }
func (v *PODInt32) Type() PODType     { return &BasePODType{id: PODTypeInt, name: "int32"} }
func (v *PODInt32) String() string    { return fmt.Sprintf("int32(%d)", v.Value) }
func (v *PODInt32) Marshal() ([]byte, error) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(v.Value))
	return b, nil
}
func (v *PODInt32) Unmarshal(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("insufficient data for int32")
	}
	v.Value = int32(binary.LittleEndian.Uint32(data))
	return nil
}

// PODInt64 - 64-bit signed integer
type PODInt64 struct{ Value int64 }

func NewPODInt64(val int64) *PODInt64 { return &PODInt64{Value: val} }
func (v *PODInt64) Type() PODType     { return &BasePODType{id: PODTypeInt64, name: "int64"} }
func (v *PODInt64) String() string    { return fmt.Sprintf("int64(%d)", v.Value) }
func (v *PODInt64) Marshal() ([]byte, error) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(v.Value))
	return b, nil
}
func (v *PODInt64) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("insufficient data for int64")
	}
	v.Value = int64(binary.LittleEndian.Uint64(data))
	return nil
}

// PODUint32 - 32-bit unsigned integer
type PODUint32 struct{ Value uint32 }

func NewPODUint32(val uint32) *PODUint32 { return &PODUint32{Value: val} }
func (v *PODUint32) Type() PODType       { return &BasePODType{id: PODTypeUint32, name: "uint32"} }
func (v *PODUint32) String() string      { return fmt.Sprintf("uint32(%d)", v.Value) }
func (v *PODUint32) Marshal() ([]byte, error) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v.Value)
	return b, nil
}
func (v *PODUint32) Unmarshal(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("insufficient data for uint32")
	}
	v.Value = binary.LittleEndian.Uint32(data)
	return nil
}

// PODUint64 - 64-bit unsigned integer
type PODUint64 struct{ Value uint64 }

func NewPODUint64(val uint64) *PODUint64 { return &PODUint64{Value: val} }
func (v *PODUint64) Type() PODType       { return &BasePODType{id: PODTypeInt64, name: "uint64"} }
func (v *PODUint64) String() string      { return fmt.Sprintf("uint64(%d)", v.Value) }
func (v *PODUint64) Marshal() ([]byte, error) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v.Value)
	return b, nil
}
func (v *PODUint64) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("insufficient data for uint64")
	}
	v.Value = binary.LittleEndian.Uint64(data)
	return nil
}

// PODFloat - 32-bit floating point
type PODFloat struct{ Value float32 }

func NewPODFloat(val float32) *PODFloat { return &PODFloat{Value: val} }
func (v *PODFloat) Type() PODType       { return &BasePODType{id: PODTypeFloat, name: "float"} }
func (v *PODFloat) String() string      { return fmt.Sprintf("float(%f)", v.Value) }
func (v *PODFloat) Marshal() ([]byte, error) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, math.Float32bits(v.Value))
	return b, nil
}
func (v *PODFloat) Unmarshal(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("insufficient data for float")
	}
	v.Value = math.Float32frombits(binary.LittleEndian.Uint32(data))
	return nil
}

// PODDouble - 64-bit floating point
type PODDouble struct{ Value float64 }

func NewPODDouble(val float64) *PODDouble { return &PODDouble{Value: val} }
func (v *PODDouble) Type() PODType        { return &BasePODType{id: PODTypeDouble, name: "double"} }
func (v *PODDouble) String() string       { return fmt.Sprintf("double(%f)", v.Value) }
func (v *PODDouble) Marshal() ([]byte, error) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, math.Float64bits(v.Value))
	return b, nil
}
func (v *PODDouble) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("insufficient data for double")
	}
	v.Value = math.Float64frombits(binary.LittleEndian.Uint64(data))
	return nil
}

// PODBool - Boolean value
type PODBool struct{ Value bool }

func NewPODBool(val bool) *PODBool { return &PODBool{Value: val} }
func (v *PODBool) Type() PODType   { return &BasePODType{id: PODTypeBool, name: "bool"} }
func (v *PODBool) String() string {
	if v.Value {
		return "bool(true)"
	}
	return "bool(false)"
}
func (v *PODBool) Marshal() ([]byte, error) {
	if v.Value {
		return []byte{1}, nil
	}
	return []byte{0}, nil
}
func (v *PODBool) Unmarshal(data []byte) error {
	if len(data) < 1 {
		return fmt.Errorf("insufficient data for bool")
	}
	v.Value = data[0] != 0
	return nil
}

// PODId - Object ID reference
type PODId struct{ Value uint32 }

func NewPODId(val uint32) *PODId { return &PODId{Value: val} }
func (v *PODId) Type() PODType   { return &BasePODType{id: PODTypeID, name: "id"} }
func (v *PODId) String() string  { return fmt.Sprintf("id(%d)", v.Value) }
func (v *PODId) Marshal() ([]byte, error) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v.Value)
	return b, nil
}
func (v *PODId) Unmarshal(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("insufficient data for id")
	}
	v.Value = binary.LittleEndian.Uint32(data)
	return nil
}

// PODString - String value
type PODString struct{ Value string }

func NewPODString(val string) *PODString { return &PODString{Value: val} }
func (v *PODString) Type() PODType       { return &BasePODType{id: PODTypeString, name: "string"} }
func (v *PODString) String() string      { return fmt.Sprintf("string(%q)", v.Value) }
func (v *PODString) Marshal() ([]byte, error) {
	data := []byte(v.Value)
	// String format: length (4 bytes) + data + null terminator
	b := make([]byte, 4+len(data)+1)
	binary.LittleEndian.PutUint32(b, uint32(len(data)+1))
	copy(b[4:], data)
	b[4+len(data)] = 0 // null terminator
	return b, nil
}
func (v *PODString) Unmarshal(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("insufficient data for string length")
	}
	length := binary.LittleEndian.Uint32(data)
	if len(data) < 4+int(length) {
		return fmt.Errorf("insufficient data for string content")
	}
	v.Value = string(data[4 : 4+length-1]) // -1 to exclude null terminator
	return nil
}

// PODBytes - Binary data
type PODBytes struct{ Value []byte }

func NewPODBytes(val []byte) *PODBytes { return &PODBytes{Value: append([]byte{}, val...)} }
func (v *PODBytes) Type() PODType      { return &BasePODType{id: PODTypeBytes, name: "bytes"} }
func (v *PODBytes) String() string     { return fmt.Sprintf("bytes(%d bytes)", len(v.Value)) }
func (v *PODBytes) Marshal() ([]byte, error) {
	b := make([]byte, 4+len(v.Value))
	binary.LittleEndian.PutUint32(b, uint32(len(v.Value)))
	copy(b[4:], v.Value)
	return b, nil
}
func (v *PODBytes) Unmarshal(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("insufficient data for bytes length")
	}
	length := binary.LittleEndian.Uint32(data)
	if len(data) < 4+int(length) {
		return fmt.Errorf("insufficient data for bytes content")
	}
	v.Value = append([]byte{}, data[4:4+length]...)
	return nil
}

// PODFraction - Fraction (num/den)
type PODFraction struct {
	Num uint32
	Den uint32
}

func NewPODFraction(num, den uint32) *PODFraction { return &PODFraction{Num: num, Den: den} }
func (v *PODFraction) Type() PODType              { return &BasePODType{id: PODTypeFraction, name: "fraction"} }
func (v *PODFraction) String() string             { return fmt.Sprintf("fraction(%d/%d)", v.Num, v.Den) }
func (v *PODFraction) Marshal() ([]byte, error) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint32(b[0:4], v.Num)
	binary.LittleEndian.PutUint32(b[4:8], v.Den)
	return b, nil
}
func (v *PODFraction) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("insufficient data for fraction")
	}
	v.Num = binary.LittleEndian.Uint32(data[0:4])
	v.Den = binary.LittleEndian.Uint32(data[4:8])
	return nil
}

// PODRectangle - Rectangle (x, y, width, height)
type PODRectangle struct {
	X int32
	Y int32
	W int32
	H int32
}

func NewPODRectangle(x, y, w, h int32) *PODRectangle { return &PODRectangle{X: x, Y: y, W: w, H: h} }
func (v *PODRectangle) Type() PODType                 { return &BasePODType{id: PODTypeRectangle, name: "rectangle"} }
func (v *PODRectangle) String() string                { return fmt.Sprintf("rect(%d,%d %dx%d)", v.X, v.Y, v.W, v.H) }
func (v *PODRectangle) Marshal() ([]byte, error) {
	b := make([]byte, 16)
	binary.LittleEndian.PutUint32(b[0:4], uint32(v.X))
	binary.LittleEndian.PutUint32(b[4:8], uint32(v.Y))
	binary.LittleEndian.PutUint32(b[8:12], uint32(v.W))
	binary.LittleEndian.PutUint32(b[12:16], uint32(v.H))
	return b, nil
}
func (v *PODRectangle) Unmarshal(data []byte) error {
	if len(data) < 16 {
		return fmt.Errorf("insufficient data for rectangle")
	}
	v.X = int32(binary.LittleEndian.Uint32(data[0:4]))
	v.Y = int32(binary.LittleEndian.Uint32(data[4:8]))
	v.W = int32(binary.LittleEndian.Uint32(data[8:12]))
	v.H = int32(binary.LittleEndian.Uint32(data[12:16]))
	return nil
}

// ============================================================================
// COMPOSITE POD TYPES
// ============================================================================

// PODArray - Array of POD values
type PODArray struct {
	Values []PODValue
}

func NewPODArray() *PODArray { return &PODArray{Values: make([]PODValue, 0)} }
func (v *PODArray) Type() PODType { return &BasePODType{id: PODTypeArray, name: "array"} }
func (v *PODArray) String() string { return fmt.Sprintf("array(%d items)", len(v.Values)) }
func (v *PODArray) Append(val PODValue) error {
	if val == nil {
		return fmt.Errorf("cannot append nil value")
	}
	v.Values = append(v.Values, val)
	return nil
}
func (v *PODArray) Marshal() ([]byte, error) {
	w := NewPODWriter()
	if err := w.WriteUint32(uint32(len(v.Values))); err != nil {
		return nil, err
	}
	for _, item := range v.Values {
		data, err := item.Marshal()
		if err != nil {
			return nil, err
		}
		if err := w.WriteBytes(data); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}
func (v *PODArray) Unmarshal(data []byte) error {
	p := NewPODParser(data)
	length, err := p.ReadUint32()
	if err != nil {
		return err
	}
	v.Values = make([]PODValue, 0, length)
	for i := uint32(0); i < length; i++ {
		val, err := p.ParseValue()
		if err != nil {
			return err
		}
		v.Values = append(v.Values, val)
	}
	return nil
}

// PODObject - Object with key-value pairs
type PODObject struct {
	Fields map[string]PODValue
}

func NewPODObject() *PODObject { return &PODObject{Fields: make(map[string]PODValue)} }
func (v *PODObject) Type() PODType { return &BasePODType{id: PODTypeObject, name: "object"} }
func (v *PODObject) String() string { return fmt.Sprintf("object(%d fields)", len(v.Fields)) }
func (v *PODObject) Set(key string, val PODValue) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	if val == nil {
		return fmt.Errorf("value cannot be nil")
	}
	v.Fields[key] = val
	return nil
}
func (v *PODObject) Get(key string) (PODValue, bool) {
	val, ok := v.Fields[key]
	return val, ok
}
func (v *PODObject) Marshal() ([]byte, error) {
	w := NewPODWriter()
	if err := w.WriteUint32(uint32(len(v.Fields))); err != nil {
		return nil, err
	}
	for k, val := range v.Fields {
		if err := w.WriteString(k); err != nil {
			return nil, err
		}
		data, err := val.Marshal()
		if err != nil {
			return nil, err
		}
		if err := w.WriteBytes(data); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}
func (v *PODObject) Unmarshal(data []byte) error {
	p := NewPODParser(data)
	length, err := p.ReadUint32()
	if err != nil {
		return err
	}
	v.Fields = make(map[string]PODValue)
	for i := uint32(0); i < length; i++ {
		key, err := p.ReadString()
		if err != nil {
			return err
		}
		val, err := p.ParseValue()
		if err != nil {
			return err
		}
		v.Fields[key] = val
	}
	return nil
}

// ============================================================================
// BUILDERS - Fluent API for constructing POD values
// ============================================================================

// PODArrayBuilder - Fluent builder for arrays
type PODArrayBuilder struct {
	array *PODArray
}

func NewPODArrayBuilder() *PODArrayBuilder {
	return &PODArrayBuilder{array: NewPODArray()}
}
func (b *PODArrayBuilder) Add(val PODValue) *PODArrayBuilder {
	if val != nil {
		b.array.Append(val)
	}
	return b
}
func (b *PODArrayBuilder) AddInt32(val int32) *PODArrayBuilder {
	return b.Add(NewPODInt32(val))
}
func (b *PODArrayBuilder) AddUint32(val uint32) *PODArrayBuilder {
	return b.Add(NewPODUint32(val))
}
func (b *PODArrayBuilder) AddFloat(val float32) *PODArrayBuilder {
	return b.Add(NewPODFloat(val))
}
func (b *PODArrayBuilder) AddString(val string) *PODArrayBuilder {
	return b.Add(NewPODString(val))
}
func (b *PODArrayBuilder) AddBool(val bool) *PODArrayBuilder {
	return b.Add(NewPODBool(val))
}
func (b *PODArrayBuilder) Build() *PODArray {
	return b.array
}

// PODObjectBuilder - Fluent builder for objects
type PODObjectBuilder struct {
	object *PODObject
}

func NewPODObjectBuilder() *PODObjectBuilder {
	return &PODObjectBuilder{object: NewPODObject()}
}
func (b *PODObjectBuilder) Put(key string, val PODValue) *PODObjectBuilder {
	if key != "" && val != nil {
		b.object.Set(key, val)
	}
	return b
}
func (b *PODObjectBuilder) PutInt32(key string, val int32) *PODObjectBuilder {
	return b.Put(key, NewPODInt32(val))
}
func (b *PODObjectBuilder) PutUint32(key string, val uint32) *PODObjectBuilder {
	return b.Put(key, NewPODUint32(val))
}
func (b *PODObjectBuilder) PutFloat(key string, val float32) *PODObjectBuilder {
	return b.Put(key, NewPODFloat(val))
}
func (b *PODObjectBuilder) PutString(key string, val string) *PODObjectBuilder {
	return b.Put(key, NewPODString(val))
}
func (b *PODObjectBuilder) PutBool(key string, val bool) *PODObjectBuilder {
	return b.Put(key, NewPODBool(val))
}
func (b *PODObjectBuilder) Build() *PODObject {
	return b.object
}

// ============================================================================
// PARSER - Unmarshalling POD binary data
// ============================================================================

// PODParser - Parse POD data from bytes
type PODParser struct {
	data   []byte
	offset int
}

func NewPODParser(data []byte) *PODParser {
	return &PODParser{data: data, offset: 0}
}

func (p *PODParser) Offset() int { return p.offset }

func (p *PODParser) ParseValue() (PODValue, error) {
	if p.offset+4 > len(p.data) {
		return nil, fmt.Errorf("insufficient data for type header")
	}

	// Peek at type ID (not part of payload, inferred from context)
	// For now, try to auto-detect based on alignment heuristics
	// In production, caller must specify type or use framed POD

	// For basic parsing, we'll return common types
	// This is a limitation - proper POD requires type information from protocol
	return nil, fmt.Errorf("ParseValue requires explicit type context - use ParseInt32, ParseString, etc.")
}

func (p *PODParser) ReadByte() (byte, error) {
	if p.offset >= len(p.data) {
		return 0, fmt.Errorf("offset out of range")
	}
	val := p.data[p.offset]
	p.offset++
	return val, nil
}

func (p *PODParser) ReadInt32() (int32, error) {
	if p.offset+4 > len(p.data) {
		return 0, fmt.Errorf("insufficient data for int32")
	}
	val := int32(binary.LittleEndian.Uint32(p.data[p.offset : p.offset+4]))
	p.offset += 4
	return val, nil
}

func (p *PODParser) ReadInt64() (int64, error) {
	if p.offset+8 > len(p.data) {
		return 0, fmt.Errorf("insufficient data for int64")
	}
	val := int64(binary.LittleEndian.Uint64(p.data[p.offset : p.offset+8]))
	p.offset += 8
	return val, nil
}

func (p *PODParser) ReadUint32() (uint32, error) {
	if p.offset+4 > len(p.data) {
		return 0, fmt.Errorf("insufficient data for uint32")
	}
	val := binary.LittleEndian.Uint32(p.data[p.offset : p.offset+4])
	p.offset += 4
	return val, nil
}

func (p *PODParser) ReadUint64() (uint64, error) {
	if p.offset+8 > len(p.data) {
		return 0, fmt.Errorf("insufficient data for uint64")
	}
	val := binary.LittleEndian.Uint64(p.data[p.offset : p.offset+8])
	p.offset += 8
	return val, nil
}

func (p *PODParser) ReadFloat() (float32, error) {
	if p.offset+4 > len(p.data) {
		return 0, fmt.Errorf("insufficient data for float")
	}
	val := math.Float32frombits(binary.LittleEndian.Uint32(p.data[p.offset : p.offset+4]))
	p.offset += 4
	return val, nil
}

func (p *PODParser) ReadDouble() (float64, error) {
	if p.offset+8 > len(p.data) {
		return 0, fmt.Errorf("insufficient data for double")
	}
	val := math.Float64frombits(binary.LittleEndian.Uint64(p.data[p.offset : p.offset+8]))
	p.offset += 8
	return val, nil
}

func (p *PODParser) ReadBool() (bool, error) {
	b, err := p.ReadByte()
	return b != 0, err
}

func (p *PODParser) ReadString() (string, error) {
	length, err := p.ReadUint32()
	if err != nil {
		return "", err
	}
	if p.offset+int(length) > len(p.data) {
		return "", fmt.Errorf("insufficient data for string")
	}
	// Length includes null terminator
	str := string(p.data[p.offset : p.offset+int(length)-1])
	p.offset += int(length)
	return str, nil
}

func (p *PODParser) ReadBytes() ([]byte, error) {
	length, err := p.ReadUint32()
	if err != nil {
		return nil, err
	}
	if p.offset+int(length) > len(p.data) {
		return nil, fmt.Errorf("insufficient data for bytes")
	}
	data := append([]byte{}, p.data[p.offset:p.offset+int(length)]...)
	p.offset += int(length)
	return data, nil
}

// ============================================================================
// WRITER - Marshalling POD data to bytes
// ============================================================================

// PODWriter - Write POD data to bytes
type PODWriter struct {
	buffer []byte
}

func NewPODWriter() *PODWriter {
	return &PODWriter{buffer: make([]byte, 0, 256)}
}

func (w *PODWriter) WriteByte(b byte) error {
	w.buffer = append(w.buffer, b)
	return nil
}

func (w *PODWriter) WriteInt32(val int32) error {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(val))
	w.buffer = append(w.buffer, b...)
	return nil
}

func (w *PODWriter) WriteInt64(val int64) error {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(val))
	w.buffer = append(w.buffer, b...)
	return nil
}

func (w *PODWriter) WriteUint32(val uint32) error {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, val)
	w.buffer = append(w.buffer, b...)
	return nil
}

func (w *PODWriter) WriteUint64(val uint64) error {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, val)
	w.buffer = append(w.buffer, b...)
	return nil
}

func (w *PODWriter) WriteFloat(val float32) error {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, math.Float32bits(val))
	w.buffer = append(w.buffer, b...)
	return nil
}

func (w *PODWriter) WriteDouble(val float64) error {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, math.Float64bits(val))
	w.buffer = append(w.buffer, b...)
	return nil
}

func (w *PODWriter) WriteBool(val bool) error {
	if val {
		return w.WriteByte(1)
	}
	return w.WriteByte(0)
}

func (w *PODWriter) WriteString(s string) error {
	data := []byte(s)
	// Length includes null terminator
	if err := w.WriteUint32(uint32(len(data) + 1)); err != nil {
		return err
	}
	w.buffer = append(w.buffer, data...)
	w.buffer = append(w.buffer, 0) // null terminator
	return nil
}

func (w *PODWriter) WriteBytes(data []byte) error {
	if err := w.WriteUint32(uint32(len(data))); err != nil {
		return err
	}
	w.buffer = append(w.buffer, data...)
	return nil
}

func (w *PODWriter) Bytes() []byte {
	return append([]byte{}, w.buffer...)
}

func (w *PODWriter) Len() int {
	return len(w.buffer)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// ParseObject extracts a map[string]interface{} from POD Object
func ParseObject(obj *PODObject) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range obj.Fields {
		result[k] = podValueToInterface(v)
	}
	return result
}

// podValueToInterface converts a PODValue to a Go interface{}
func podValueToInterface(v PODValue) interface{} {
	switch val := v.(type) {
	case *PODInt32:
		return val.Value
	case *PODInt64:
		return val.Value
	case *PODUint32:
		return val.Value
	case *PODUint64:
		return val.Value
	case *PODFloat:
		return val.Value
	case *PODDouble:
		return val.Value
	case *PODBool:
		return val.Value
	case *PODString:
		return val.Value
	case *PODBytes:
		return val.Value
	case *PODArray:
		arr := make([]interface{}, len(val.Values))
		for i, item := range val.Values {
			arr[i] = podValueToInterface(item)
		}
		return arr
	case *PODObject:
		return ParseObject(val)
	default:
		return nil
	}
}

// BuildObject creates a PODObject from map[string]interface{}
func BuildObject(data map[string]interface{}) (*PODObject, error) {
	obj := NewPODObject()
	for k, v := range data {
		podVal, err := interfaceToPODValue(v)
		if err != nil {
			return nil, fmt.Errorf("invalid value for key %q: %w", k, err)
		}
		obj.Set(k, podVal)
	}
	return obj, nil
}

// interfaceToPODValue converts a Go interface{} to PODValue
func interfaceToPODValue(v interface{}) (PODValue, error) {
	switch val := v.(type) {
	case int:
		return NewPODInt32(int32(val)), nil
	case int32:
		return NewPODInt32(val), nil
	case int64:
		return NewPODInt64(val), nil
	case uint32:
		return NewPODUint32(val), nil
	case uint64:
		return NewPODUint64(val), nil
	case float32:
		return NewPODFloat(val), nil
	case float64:
		return NewPODDouble(val), nil
	case bool:
		return NewPODBool(val), nil
	case string:
		return NewPODString(val), nil
	case []byte:
		return NewPODBytes(val), nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}

// FormatArray parses audio format from POD array
// Format: [sampleType, sampleRate, numChannels, ...]
func FormatArray(arr *PODArray) (format string, sampleRate uint32, channels uint32, err error) {
	if len(arr.Values) < 3 {
		err = fmt.Errorf("format array must have at least 3 elements")
		return
	}

	// First element: sample type (e.g., "S16LE")
	if st, ok := arr.Values[0].(*PODString); ok {
		format = st.Value
	} else {
		err = fmt.Errorf("format[0] must be string")
		return
	}

	// Second element: sample rate
	if sr, ok := arr.Values[1].(*PODUint32); ok {
		sampleRate = sr.Value
	} else {
		err = fmt.Errorf("format[1] must be uint32")
		return
	}

	// Third element: channels
	if ch, ok := arr.Values[2].(*PODUint32); ok {
		channels = ch.Value
	} else {
		err = fmt.Errorf("format[2] must be uint32")
		return
	}

	return
}

// PropertyArray parses key-value properties from POD array
// Array format: [key1, value1, key2, value2, ...]
func PropertyArray(arr *PODArray) (map[string]string, error) {
	if len(arr.Values)%2 != 0 {
		return nil, fmt.Errorf("property array must have even number of elements")
	}

	props := make(map[string]string)
	for i := 0; i < len(arr.Values); i += 2 {
		key, ok := arr.Values[i].(*PODString)
		if !ok {
			return nil, fmt.Errorf("property key[%d] must be string", i)
		}

		val, ok := arr.Values[i+1].(*PODString)
		if !ok {
			return nil, fmt.Errorf("property value[%d] must be string", i+1)
		}

		props[key.Value] = val.Value
	}

	return props, nil
}
