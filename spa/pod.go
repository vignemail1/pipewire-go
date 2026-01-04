// Package spa - POD (Plain Old Data) type system
// spa/pod.go
// Complete POD marshaling, unmarshaling, and builder implementation

package spa

import (
	"encoding/binary"
	"fmt"
)

// PODValue represents a POD value
type PODValue interface {
	Type() PODType
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

// PODType represents a POD type
type PODType interface {
	ID() uint32
	Name() string
}

// ===== POD Type Implementations =====

// BasePODType implements PODType
type BasePODType struct {
	id   uint32
	name string
}

func (t *BasePODType) ID() uint32   { return t.id }
func (t *BasePODType) Name() string { return t.name }

// ===== POD Values =====

// PODArray represents an array of POD values
type PODArray struct {
	values []PODValue
	typ    PODType
}

func NewPODArray() *PODArray {
	return &PODArray{
		values: make([]PODValue, 0),
		typ:    &BasePODType{id: PODTypeArray, name: "array"},
	}
}

func (a *PODArray) Type() PODType { return a.typ }

func (a *PODArray) Append(val PODValue) error {
	if val == nil {
		return fmt.Errorf("cannot append nil value")
	}
	a.values = append(a.values, val)
	return nil
}

func (a *PODArray) Get(index int) (PODValue, error) {
	if index < 0 || index >= len(a.values) {
		return nil, fmt.Errorf("index out of range: %d", index)
	}
	return a.values[index], nil
}

func (a *PODArray) Length() int { return len(a.values) }

func (a *PODArray) Clear() { a.values = make([]PODValue, 0) }

func (a *PODArray) Marshal() ([]byte, error) {
	w := NewPODWriter()
	w.writeUint32(PODTypeArray)
	w.writeUint32(uint32(len(a.values)))
	for _, v := range a.values {
		data, err := v.Marshal()
		if err != nil {
			return nil, err
		}
		w.buffer = append(w.buffer, data...)
	}
	return w.Bytes(), nil
}

func (a *PODArray) Unmarshal(data []byte) error {
	p := NewPODParser()
	p.data = data
	p.offset = 0

	// Read type ID
	typeID, err := p.readUint32(p.offset)
	if err != nil {
		return err
	}
	if typeID != PODTypeArray {
		return fmt.Errorf("invalid array type: %d", typeID)
	}
	p.offset += 4

	// Read length
	length, err := p.readUint32(p.offset)
	if err != nil {
		return err
	}
	p.offset += 4

	a.values = make([]PODValue, length)
	for i := 0; i < int(length); i++ {
		val, newOffset, err := p.parseValue(p.offset)
		if err != nil {
			return err
		}
		a.values[i] = val
		p.offset = newOffset
	}

	return nil
}

// PODObject represents an object with key-value pairs
type PODObject struct {
	fields map[string]PODValue
	typ    PODType
}

func NewPODObject() *PODObject {
	return &PODObject{
		fields: make(map[string]PODValue),
		typ:    &BasePODType{id: PODTypeObject, name: "object"},
	}
}

func (o *PODObject) Type() PODType { return o.typ }

func (o *PODObject) Set(key string, val PODValue) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	if val == nil {
		return fmt.Errorf("value cannot be nil")
	}
	o.fields[key] = val
	return nil
}

func (o *PODObject) Get(key string) (PODValue, error) {
	val, exists := o.fields[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	return val, nil
}

func (o *PODObject) Has(key string) bool {
	_, exists := o.fields[key]
	return exists
}

func (o *PODObject) Delete(key string) error {
	if _, exists := o.fields[key]; !exists {
		return fmt.Errorf("key not found: %s", key)
	}
	delete(o.fields, key)
	return nil
}

func (o *PODObject) Keys() []string {
	keys := make([]string, 0, len(o.fields))
	for k := range o.fields {
		keys = append(keys, k)
	}
	return keys
}

func (o *PODObject) Length() int { return len(o.fields) }

func (o *PODObject) Marshal() ([]byte, error) {
	w := NewPODWriter()
	w.writeUint32(PODTypeObject)
	w.writeUint32(uint32(len(o.fields)))
	for k, v := range o.fields {
		w.writeString(k)
		data, err := v.Marshal()
		if err != nil {
			return nil, err
		}
		w.buffer = append(w.buffer, data...)
	}
	return w.Bytes(), nil
}

func (o *PODObject) Unmarshal(data []byte) error {
	p := NewPODParser()
	p.data = data
	p.offset = 0

	typeID, err := p.readUint32(p.offset)
	if err != nil {
		return err
	}
	if typeID != PODTypeObject {
		return fmt.Errorf("invalid object type: %d", typeID)
	}
	p.offset += 4

	length, err := p.readUint32(p.offset)
	if err != nil {
		return err
	}
	p.offset += 4

	o.fields = make(map[string]PODValue)
	for i := 0; i < int(length); i++ {
		key, newOffset, err := p.readString(p.offset)
		if err != nil {
			return err
		}
		p.offset = newOffset

		val, newOffset, err := p.parseValue(p.offset)
		if err != nil {
			return err
		}
		p.offset = newOffset

		o.fields[key] = val
	}

	return nil
}

// ===== POD Scalar Types =====

// PODBool represents a boolean value
type PODBool struct{ value bool }

func NewPODBool(val bool) *PODBool { return &PODBool{value: val} }
func (b *PODBool) Type() PODType   { return &BasePODType{id: PODTypeBool, name: "bool"} }
func (b *PODBool) Marshal() ([]byte, error) {
	w := NewPODWriter()
	w.writeUint32(PODTypeBool)
	if b.value {
		w.writeByte(1)
	} else {
		w.writeByte(0)
	}
	return w.Bytes(), nil
}
func (b *PODBool) Unmarshal(data []byte) error {
	if len(data) < 5 {
		return fmt.Errorf("data too short for bool")
	}
	b.value = data[4] != 0
	return nil
}

// PODInt represents a 32-bit integer
type PODInt struct{ value int32 }

func NewPODInt(val int32) *PODInt { return &PODInt{value: val} }
func (i *PODInt) Type() PODType   { return &BasePODType{id: PODTypeInt, name: "int"} }
func (i *PODInt) Marshal() ([]byte, error) {
	w := NewPODWriter()
	w.writeUint32(PODTypeInt)
	w.writeInt32(i.value)
	return w.Bytes(), nil
}
func (i *PODInt) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("data too short for int")
	}
	i.value = int32(binary.BigEndian.Uint32(data[4:8]))
	return nil
}

// PODInt64 represents a 64-bit integer
type PODInt64 struct{ value int64 }

func NewPODInt64(val int64) *PODInt64 { return &PODInt64{value: val} }
func (i *PODInt64) Type() PODType     { return &BasePODType{id: PODTypeInt64, name: "int64"} }
func (i *PODInt64) Marshal() ([]byte, error) {
	w := NewPODWriter()
	w.writeUint32(PODTypeInt64)
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(i.value))
	w.buffer = append(w.buffer, b...)
	return w.Bytes(), nil
}
func (i *PODInt64) Unmarshal(data []byte) error {
	if len(data) < 12 {
		return fmt.Errorf("data too short for int64")
	}
	i.value = int64(binary.BigEndian.Uint64(data[4:12]))
	return nil
}

// PODUint32 represents a 32-bit unsigned integer
type PODUint32 struct{ value uint32 }

func NewPODUint32(val uint32) *PODUint32 { return &PODUint32{value: val} }
func (u *PODUint32) Type() PODType       { return &BasePODType{id: PODTypeUint32, name: "uint32"} }
func (u *PODUint32) Marshal() ([]byte, error) {
	w := NewPODWriter()
	w.writeUint32(PODTypeUint32)
	w.writeUint32(u.value)
	return w.Bytes(), nil
}
func (u *PODUint32) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("data too short for uint32")
	}
	u.value = binary.BigEndian.Uint32(data[4:8])
	return nil
}

// PODString represents a string value
type PODString struct{ value string }

func NewPODString(val string) *PODString { return &PODString{value: val} }
func (s *PODString) Type() PODType       { return &BasePODType{id: PODTypeString, name: "string"} }
func (s *PODString) Marshal() ([]byte, error) {
	w := NewPODWriter()
	w.writeUint32(PODTypeString)
	w.writeString(s.value)
	return w.Bytes(), nil
}
func (s *PODString) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("data too short for string")
	}
	str, _, err := NewPODParser().readString(4)
	if err != nil {
		return err
	}
	s.value = str
	return nil
}

// PODBytes represents binary data
type PODBytes struct{ value []byte }

func NewPODBytes(val []byte) *PODBytes { return &PODBytes{value: append([]byte{}, val...)} }
func (b *PODBytes) Type() PODType      { return &BasePODType{id: PODTypeBytes, name: "bytes"} }
func (b *PODBytes) Marshal() ([]byte, error) {
	w := NewPODWriter()
	w.writeUint32(PODTypeBytes)
	w.writeUint32(uint32(len(b.value)))
	w.buffer = append(w.buffer, b.value...)
	return w.Bytes(), nil
}
func (b *PODBytes) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("data too short for bytes")
	}
	length := binary.BigEndian.Uint32(data[4:8])
	if len(data) < 8+int(length) {
		return fmt.Errorf("data too short for bytes content")
	}
	b.value = append([]byte{}, data[8:8+length]...)
	return nil
}

// PODFraction represents a fraction
type PODFraction struct {
	num uint32
	den uint32
}

func NewPODFraction(num, den uint32) *PODFraction { return &PODFraction{num: num, den: den} }
func (f *PODFraction) Type() PODType              { return &BasePODType{id: PODTypeFraction, name: "fraction"} }
func (f *PODFraction) Marshal() ([]byte, error) {
	w := NewPODWriter()
	w.writeUint32(PODTypeFraction)
	w.writeUint32(f.num)
	w.writeUint32(f.den)
	return w.Bytes(), nil
}
func (f *PODFraction) Unmarshal(data []byte) error {
	if len(data) < 12 {
		return fmt.Errorf("data too short for fraction")
	}
	f.num = binary.BigEndian.Uint32(data[4:8])
	f.den = binary.BigEndian.Uint32(data[8:12])
	return nil
}

// PODRectangle represents a rectangle
type PODRectangle struct {
	x, y, w, h int32
}

func NewPODRectangle(x, y, w, h int32) *PODRectangle {
	return &PODRectangle{x: x, y: y, w: w, h: h}
}
func (r *PODRectangle) Type() PODType { return &BasePODType{id: PODTypeRectangle, name: "rectangle"} }
func (r *PODRectangle) Marshal() ([]byte, error) {
	w := NewPODWriter()
	w.writeUint32(PODTypeRectangle)
	w.writeInt32(r.x)
	w.writeInt32(r.y)
	w.writeInt32(r.w)
	w.writeInt32(r.h)
	return w.Bytes(), nil
}
func (r *PODRectangle) Unmarshal(data []byte) error {
	if len(data) < 20 {
		return fmt.Errorf("data too short for rectangle")
	}
	r.x = int32(binary.BigEndian.Uint32(data[4:8]))
	r.y = int32(binary.BigEndian.Uint32(data[8:12]))
	r.w = int32(binary.BigEndian.Uint32(data[12:16]))
	r.h = int32(binary.BigEndian.Uint32(data[16:20]))
	return nil
}

// ===== Builders =====

// PODArrayBuilder for constructing arrays
type PODArrayBuilder struct {
	values []PODValue
}

func NewPODArrayBuilder() *PODArrayBuilder {
	return &PODArrayBuilder{values: make([]PODValue, 0)}
}

func (b *PODArrayBuilder) Add(val PODValue) *PODArrayBuilder {
	if val != nil {
		b.values = append(b.values, val)
	}
	return b
}

func (b *PODArrayBuilder) AddInt(val int32) *PODArrayBuilder {
	return b.Add(NewPODInt(val))
}

func (b *PODArrayBuilder) AddBool(val bool) *PODArrayBuilder {
	return b.Add(NewPODBool(val))
}

func (b *PODArrayBuilder) AddString(val string) *PODArrayBuilder {
	return b.Add(NewPODString(val))
}

func (b *PODArrayBuilder) AddUint32(val uint32) *PODArrayBuilder {
	return b.Add(NewPODUint32(val))
}

func (b *PODArrayBuilder) Build() *PODArray {
	arr := NewPODArray()
	arr.values = append([]PODValue{}, b.values...)
	return arr
}

// PODObjectBuilder for constructing objects
type PODObjectBuilder struct {
	fields map[string]PODValue
}

func NewPODObjectBuilder() *PODObjectBuilder {
	return &PODObjectBuilder{fields: make(map[string]PODValue)}
}

// NewPODBuilder creates a generic POD builder (wrapper around PODObjectBuilder)
func NewPODBuilder() *PODObjectBuilder {
	return NewPODObjectBuilder()
}

func (b *PODObjectBuilder) Put(key string, val PODValue) *PODObjectBuilder {
	if key != "" && val != nil {
		b.fields[key] = val
	}
	return b
}

func (b *PODObjectBuilder) PutInt(key string, val int32) *PODObjectBuilder {
	return b.Put(key, NewPODInt(val))
}

func (b *PODObjectBuilder) PutBool(key string, val bool) *PODObjectBuilder {
	return b.Put(key, NewPODBool(val))
}

func (b *PODObjectBuilder) PutString(key string, val string) *PODObjectBuilder {
	return b.Put(key, NewPODString(val))
}

func (b *PODObjectBuilder) PutUint32(key string, val uint32) *PODObjectBuilder {
	return b.Put(key, NewPODUint32(val))
}

func (b *PODObjectBuilder) Build() *PODObject {
	obj := NewPODObject()
	for k, v := range b.fields {
		obj.fields[k] = v
	}
	return obj
}

// ===== Parser =====

type PODParser struct {
	data   []byte
	offset int
}

func NewPODParser() *PODParser {
	return &PODParser{data: nil, offset: 0}
}

func (p *PODParser) Parse(data []byte) (PODValue, error) {
	p.data = data
	p.offset = 0
	val, _, err := p.parseValue(0)
	return val, err
}

func (p *PODParser) parseValue(offset int) (PODValue, int, error) {
	if offset+4 > len(p.data) {
		return nil, offset, fmt.Errorf("not enough data for type ID")
	}

	typeID := binary.BigEndian.Uint32(p.data[offset : offset+4])
	newOffset := offset + 4

	switch typeID {
	case PODTypeBool:
		if newOffset+1 > len(p.data) {
			return nil, newOffset, fmt.Errorf("not enough data for bool")
		}
		return NewPODBool(p.data[newOffset] != 0), newOffset + 1, nil

	case PODTypeInt:
		if newOffset+4 > len(p.data) {
			return nil, newOffset, fmt.Errorf("not enough data for int")
		}
		val := int32(binary.BigEndian.Uint32(p.data[newOffset : newOffset+4]))
		return NewPODInt(val), newOffset + 4, nil

	case PODTypeUint32:
		if newOffset+4 > len(p.data) {
			return nil, newOffset, fmt.Errorf("not enough data for uint32")
		}
		val := binary.BigEndian.Uint32(p.data[newOffset : newOffset+4])
		return NewPODUint32(val), newOffset + 4, nil

	case PODTypeString:
		if newOffset+4 > len(p.data) {
			return nil, newOffset, fmt.Errorf("not enough data for string length")
		}
		strlen := binary.BigEndian.Uint32(p.data[newOffset : newOffset+4])
		newOffset += 4
		if newOffset+int(strlen) > len(p.data) {
			return nil, newOffset, fmt.Errorf("not enough data for string content")
		}
		str := string(p.data[newOffset : newOffset+int(strlen)])
		return NewPODString(str), newOffset + int(strlen), nil

	case PODTypeArray:
		arr := NewPODArray()
		if newOffset+4 > len(p.data) {
			return nil, newOffset, fmt.Errorf("not enough data for array length")
		}
		length := binary.BigEndian.Uint32(p.data[newOffset : newOffset+4])
		newOffset += 4

		for i := 0; i < int(length); i++ {
			val, offset, err := p.parseValue(newOffset)
			if err != nil {
				return nil, offset, err
			}
			arr.values = append(arr.values, val)
			newOffset = offset
		}
		return arr, newOffset, nil

	default:
		return nil, newOffset, fmt.Errorf("unsupported POD type: %d", typeID)
	}
}

func (p *PODParser) readByte(offset int) (byte, error) {
	if offset >= len(p.data) {
		return 0, fmt.Errorf("offset out of range")
	}
	return p.data[offset], nil
}

func (p *PODParser) readInt32(offset int) (int32, error) {
	if offset+4 > len(p.data) {
		return 0, fmt.Errorf("not enough data")
	}
	return int32(binary.BigEndian.Uint32(p.data[offset : offset+4])), nil
}

func (p *PODParser) readUint32(offset int) (uint32, error) {
	if offset+4 > len(p.data) {
		return 0, fmt.Errorf("not enough data")
	}
	return binary.BigEndian.Uint32(p.data[offset : offset+4]), nil
}

func (p *PODParser) readString(offset int) (string, int, error) {
	if offset+4 > len(p.data) {
		return "", offset, fmt.Errorf("not enough data for string length")
	}
	strlen := binary.BigEndian.Uint32(p.data[offset : offset+4])
	offset += 4
	if offset+int(strlen) > len(p.data) {
		return "", offset, fmt.Errorf("not enough data for string content")
	}
	str := string(p.data[offset : offset+int(strlen)])
	return str, offset + int(strlen), nil
}

// ===== Writer =====

type PODWriter struct {
	buffer []byte
}

func NewPODWriter() *PODWriter {
	return &PODWriter{buffer: make([]byte, 0)}
}

func (w *PODWriter) writeByte(b byte) {
	w.buffer = append(w.buffer, b)
}

func (w *PODWriter) writeInt32(val int32) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(val))
	w.buffer = append(w.buffer, b...)
}

func (w *PODWriter) writeUint32(val uint32) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, val)
	w.buffer = append(w.buffer, b...)
}

func (w *PODWriter) writeString(s string) {
	data := []byte(s)
	w.writeUint32(uint32(len(data)))
	w.buffer = append(w.buffer, data...)
}

func (w *PODWriter) Bytes() []byte {
	return append([]byte{}, w.buffer...)
}

func (w *PODWriter) Write(val PODValue) ([]byte, error) {
	if val == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	return val.Marshal()
}
