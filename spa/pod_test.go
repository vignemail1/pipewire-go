// Package spa - Tests for POD marshalling/unmarshalling
// spa/pod_test.go

package spa

import (
	"testing"
)

// TestPODInt32 tests 32-bit integer marshalling
func TestPODInt32(t *testing.T) {
	// Create
	val := NewPODInt32(42)
	if val.Value != 42 {
		t.Errorf("expected 42, got %d", val.Value)
	}

	// Marshal
	data, err := val.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if len(data) != 4 {
		t.Errorf("expected 4 bytes, got %d", len(data))
	}

	// Unmarshal
	restored := &PODInt32{}
	err = restored.Unmarshal(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if restored.Value != 42 {
		t.Errorf("expected 42 after unmarshal, got %d", restored.Value)
	}
}

// TestPODFloat tests floating point marshalling
func TestPODFloat(t *testing.T) {
	val := NewPODFloat(3.14)
	data, err := val.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	restored := &PODFloat{}
	err = restored.Unmarshal(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if restored.Value < 3.13 || restored.Value > 3.15 {
		t.Errorf("expected ~3.14, got %f", restored.Value)
	}
}

// TestPODDouble tests 64-bit floating point
func TestPODDouble(t *testing.T) {
	val := NewPODDouble(2.71828)
	data, err := val.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	restored := &PODDouble{}
	err = restored.Unmarshal(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if restored.Value < 2.71 || restored.Value > 2.72 {
		t.Errorf("expected ~2.71828, got %f", restored.Value)
	}
}

// TestPODString tests string marshalling with null terminator
func TestPODString(t *testing.T) {
	val := NewPODString("hello")
	data, err := val.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Verify format: [length(4 bytes) + data + null]
	if len(data) != 4+5+1 { // length + "hello" + null
		t.Errorf("expected 10 bytes, got %d", len(data))
	}

	restored := &PODString{}
	err = restored.Unmarshal(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if restored.Value != "hello" {
		t.Errorf("expected 'hello', got %q", restored.Value)
	}
}

// TestPODUint64 tests 64-bit unsigned integer
func TestPODUint64(t *testing.T) {
	val := NewPODUint64(9223372036854775807) // max int64
	data, err := val.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	restored := &PODUint64{}
	err = restored.Unmarshal(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if restored.Value != 9223372036854775807 {
		t.Errorf("expected 9223372036854775807, got %d", restored.Value)
	}
}

// TestPODArrayBuilder tests fluent array construction
func TestPODArrayBuilder(t *testing.T) {
	arr := NewPODArrayBuilder().
		AddInt32(10).
		AddInt32(20).
		AddString("test").
		AddBool(true).
		Build()

	if len(arr.Values) != 4 {
		t.Errorf("expected 4 values, got %d", len(arr.Values))
	}

	// Verify types
	if _, ok := arr.Values[0].(*PODInt32); !ok {
		t.Errorf("expected PODInt32 at [0]")
	}
	if _, ok := arr.Values[1].(*PODInt32); !ok {
		t.Errorf("expected PODInt32 at [1]")
	}
	if _, ok := arr.Values[2].(*PODString); !ok {
		t.Errorf("expected PODString at [2]")
	}
	if _, ok := arr.Values[3].(*PODBool); !ok {
		t.Errorf("expected PODBool at [3]")
	}
}

// TestPODObjectBuilder tests fluent object construction
func TestPODObjectBuilder(t *testing.T) {
	obj := NewPODObjectBuilder().
		PutString("name", "Alice").
		PutInt32("age", 30).
		PutFloat("height", 1.75).
		PutBool("active", true).
		Build()

	if len(obj.Fields) != 4 {
		t.Errorf("expected 4 fields, got %d", len(obj.Fields))
	}

	// Verify values
	if name, ok := obj.Get("name"); ok {
		if s, ok := name.(*PODString); ok && s.Value != "Alice" {
			t.Errorf("expected name='Alice', got %q", s.Value)
		}
	} else {
		t.Errorf("field 'name' not found")
	}

	if age, ok := obj.Get("age"); ok {
		if i, ok := age.(*PODInt32); ok && i.Value != 30 {
			t.Errorf("expected age=30, got %d", i.Value)
		}
	} else {
		t.Errorf("field 'age' not found")
	}
}

// TestPODFraction tests fraction (num/den)
func TestPODFraction(t *testing.T) {
	val := NewPODFraction(22, 7) // approximation of pi
	data, err := val.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	restored := &PODFraction{}
	err = restored.Unmarshal(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if restored.Num != 22 || restored.Den != 7 {
		t.Errorf("expected 22/7, got %d/%d", restored.Num, restored.Den)
	}

	// Test Value()
	expected := 22.0 / 7.0
	if restored.Value() < 3.14 || restored.Value() > 3.15 {
		t.Errorf("expected ~3.14, got %f", restored.Value())
	}
}

// TestPODRectangle tests rectangle type
func TestPODRectangle(t *testing.T) {
	val := NewPODRectangle(10, 20, 100, 200)
	data, err := val.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	restored := &PODRectangle{}
	err = restored.Unmarshal(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if restored.X != 10 || restored.Y != 20 || restored.W != 100 || restored.H != 200 {
		t.Errorf("expected (10,20,100,200), got (%d,%d,%d,%d)",
			restored.X, restored.Y, restored.W, restored.H)
	}

	// Test Area
	if restored.Area() != 100*200 {
		t.Errorf("expected area=20000, got %d", restored.Area())
	}
}

// TestPODBytes tests binary data
func TestPODBytes(t *testing.T) {
	original := []byte{0xFF, 0xEE, 0xDD, 0xCC}
	val := NewPODBytes(original)
	data, err := val.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	restored := &PODBytes{}
	err = restored.Unmarshal(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(restored.Value) != len(original) {
		t.Errorf("expected %d bytes, got %d", len(original), len(restored.Value))
	}

	for i := range original {
		if restored.Value[i] != original[i] {
			t.Errorf("byte[%d] mismatch: expected 0x%02X, got 0x%02X",
				i, original[i], restored.Value[i])
		}
	}
}

// TestPODWriter tests writer methods
func TestPODWriter(t *testing.T) {
	w := NewPODWriter()

	// Write various types
	w.WriteInt32(42)
	w.WriteFloat(3.14)
	w.WriteString("test")
	w.WriteBool(true)

	data := w.Bytes()
	if w.Len() != len(data) {
		t.Errorf("length mismatch")
	}

	// Verify not empty
	if len(data) == 0 {
		t.Errorf("expected non-empty data")
	}
}

// TestPODParser tests parser methods
func TestPODParser(t *testing.T) {
	// Create some data
	w := NewPODWriter()
	w.WriteInt32(100)
	w.WriteUint32(200)
	data := w.Bytes()

	// Parse it back
	p := NewPODParser(data)

	val1, err := p.ReadInt32()
	if err != nil || val1 != 100 {
		t.Errorf("expected int32=100, got %d", val1)
	}

	val2, err := p.ReadUint32()
	if err != nil || val2 != 200 {
		t.Errorf("expected uint32=200, got %d", val2)
	}

	if p.Offset() != 8 {
		t.Errorf("expected offset=8, got %d", p.Offset())
	}
}

// TestAlignPadding tests alignment helpers
func TestAlignPadding(t *testing.T) {
	tests := []struct {
		offset int
		padding int
	}{
		{0, 0},      // already aligned
		{1, 7},      // need 7 bytes to reach 8
		{7, 1},      // need 1 byte to reach 8
		{8, 0},      // already aligned
		{9, 7},      // need 7 bytes to reach 16
		{16, 0},     // already aligned
	}

	for _, test := range tests {
		if got := AlignPadding(test.offset); got != test.padding {
			t.Errorf("AlignPadding(%d): expected %d, got %d",
				test.offset, test.padding, got)
		}
	}
}

// TestPODTypeSize tests type size function
func TestPODTypeSize(t *testing.T) {
	tests := []struct {
		typeID uint32
		size   int
	}{
		{PODTypeInt, 4},
		{PODTypeInt64, 8},
		{PODTypeFloat, 4},
		{PODTypeDouble, 8},
		{PODTypeFraction, 8},
		{PODTypeRectangle, 16},
		{PODTypeString, -1}, // variable
		{PODTypeArray, -1},  // variable
	}

	for _, test := range tests {
		if got, _ := PODTypeSize(test.typeID); got != test.size {
			t.Errorf("PODTypeSize(%d): expected %d, got %d",
				test.typeID, test.size, got)
		}
	}
}

// TestEndianness validates little-endian encoding
func TestEndianness(t *testing.T) {
	// Write 0x12345678 in little-endian
	val := NewPODUint32(0x12345678)
	data, _ := val.Marshal()

	// Should be: 78 56 34 12 (little-endian)
	expected := []byte{0x78, 0x56, 0x34, 0x12}
	for i, b := range expected {
		if data[i] != b {
			t.Errorf("byte[%d]: expected 0x%02X, got 0x%02X", i, b, data[i])
		}
	}
}

// BenchmarkPODMarshal benchmarks marshalling
func BenchmarkPODMarshal(b *testing.B) {
	val := NewPODString("benchmark test string")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		val.Marshal()
	}
}

// BenchmarkPODUnmarshal benchmarks unmarshalling
func BenchmarkPODUnmarshal(b *testing.B) {
	val := NewPODString("benchmark test string")
	data, _ := val.Marshal()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		restored := &PODString{}
		restored.Unmarshal(data)
	}
}
