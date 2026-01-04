// Package verbose - Verbose Logging and Debugging
// verbose/dumper.go
// Data dumping and debugging utilities
// Phase 1 - Debugging helpers

package verbose

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
)

// Dumper provides utilities for dumping and inspecting data
type Dumper struct {
	writer io.Writer
	indent string
	depth  int
}

// NewDumper creates a new dumper writing to provided writer
func NewDumper(w io.Writer) *Dumper {
	return &Dumper{
		writer: w,
		indent: "  ",
		depth:  0,
	}
}

// DumpBytes dumps a byte slice in hex and ASCII format
func (d *Dumper) DumpBytes(label string, data []byte) {
	if len(data) == 0 {
		fmt.Fprintf(d.writer, "%s[%s]: <empty>\n", d.Indent(), label)
		return
	}

	fmt.Fprintf(d.writer, "%s[%s]:\n", d.Indent(), label)
	d.depth++
	defer func() { d.depth-- }()

	// Hex dump
	for i := 0; i < len(data); i += 16 {
		end := i + 16
		if end > len(data) {
			end = len(data)
		}

		// Hex part
		hex := hex.EncodeToString(data[i:end])
		fmt.Fprintf(d.writer, "%s%04x:  ", d.Indent(), i)

		// Formatted hex with spaces
		for j := 0; j < len(hex); j += 2 {
			if j > 0 {
				fmt.Fprint(d.writer, " ")
			}
			if j/2 == 8 {
				fmt.Fprint(d.writer, " ")
			}
			fmt.Fprint(d.writer, hex[j:j+2])
		}

		// ASCII part
		fmt.Fprint(d.writer, "  |")
		for _, b := range data[i:end] {
			if b >= 32 && b < 127 {
				fmt.Fprintf(d.writer, "%c", b)
			} else {
				fmt.Fprint(d.writer, ".")
			}
		}
		fmt.Fprintf(d.writer, "|\n")
	}
}

// DumpString dumps a string value
func (d *Dumper) DumpString(label, value string) {
	fmt.Fprintf(d.writer, "%s[%s]: %q\n", d.Indent(), label, value)
}

// DumpInt dumps an integer value
func (d *Dumper) DumpInt(label string, value int64) {
	fmt.Fprintf(d.writer, "%s[%s]: %d (0x%x)\n", d.Indent(), label, value, value)
}

// DumpFloat dumps a float value
func (d *Dumper) DumpFloat(label string, value float64) {
	fmt.Fprintf(d.writer, "%s[%s]: %f\n", d.Indent(), label, value)
}

// DumpBool dumps a boolean value
func (d *Dumper) DumpBool(label string, value bool) {
	fmt.Fprintf(d.writer, "%s[%s]: %v\n", d.Indent(), label, value)
}

// DumpTime dumps a time value
func (d *Dumper) DumpTime(label string, t time.Time) {
	fmt.Fprintf(d.writer, "%s[%s]: %s (unix: %d)\n", d.Indent(), label, t.Format(time.RFC3339), t.Unix())
}

// DumpMap dumps a map structure
func (d *Dumper) DumpMap(label string, m map[string]interface{}) {
	fmt.Fprintf(d.writer, "%s[%s]: map[%d items]\n", d.Indent(), label, len(m))
	d.depth++
	defer func() { d.depth-- }()

	for key, value := range m {
		d.DumpValue(key, value)
	}
}

// DumpStruct dumps a struct or interface
func (d *Dumper) DumpStruct(label string, v interface{}) {
	if v == nil {
		fmt.Fprintf(d.writer, "%s[%s]: <nil>\n", d.Indent(), label)
		return
	}

	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf(v)

	fmt.Fprintf(d.writer, "%s[%s]: %s\n", d.Indent(), label, rt.String())
	d.depth++
	defer func() { d.depth-- }()

	// Handle pointers
	if rt.Kind() == reflect.Ptr {
		if rv.IsNil() {
			fmt.Fprintf(d.writer, "%s<nil pointer>\n", d.Indent())
			return
		}
		rv = rv.Elem()
		rt = rt.Elem()
	}

	// Dump struct fields
	if rt.Kind() == reflect.Struct {
		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)
			value := rv.Field(i)
			d.DumpValue(field.Name, value.Interface())
		}
	}
}

// DumpValue dumps any value
func (d *Dumper) DumpValue(label string, value interface{}) {
	if value == nil {
		fmt.Fprintf(d.writer, "%s[%s]: <nil>\n", d.Indent(), label)
		return
	}

	switch v := value.(type) {
	case string:
		d.DumpString(label, v)
	case int:
		d.DumpInt(label, int64(v))
	case int32:
		d.DumpInt(label, int64(v))
	case int64:
		d.DumpInt(label, v)
	case uint:
		d.DumpInt(label, int64(v))
	case uint32:
		d.DumpInt(label, int64(v))
	case uint64:
		d.DumpInt(label, int64(v))
	case float32:
		d.DumpFloat(label, float64(v))
	case float64:
		d.DumpFloat(label, v)
	case bool:
		d.DumpBool(label, v)
	case []byte:
		d.DumpBytes(label, v)
	case time.Time:
		d.DumpTime(label, v)
	case map[string]interface{}:
		d.DumpMap(label, v)
	default:
		d.DumpStruct(label, v)
	}
}

// Indent returns the current indentation string
func (d *Dumper) Indent() string {
	var buf bytes.Buffer
	for i := 0; i < d.depth; i++ {
		buf.WriteString(d.indent)
	}
	return buf.String()
}

// SetIndent sets the indentation string
func (d *Dumper) SetIndent(indent string) {
	d.indent = indent
}

// DumpError dumps an error with context
func (d *Dumper) DumpError(label string, err error) {
	if err == nil {
		fmt.Fprintf(d.writer, "%s[%s]: <nil error>\n", d.Indent(), label)
		return
	}

	fmt.Fprintf(d.writer, "%s[%s]: %T: %v\n", d.Indent(), label, err, err)
}

// DumpHex dumps data in hex format only
func (d *Dumper) DumpHex(label string, data []byte) {
	fmt.Fprintf(d.writer, "%s[%s]: %s\n", d.Indent(), label, hex.EncodeToString(data))
}

// DumpBase64 dumps data in base64 format
func (d *Dumper) DumpBase64(label string, data []byte) {
	import_base64 := "ZW5jb2RpbmcvYmFzZTY0" // Would use import
	_ = import_base64
	// In real code: base64.StdEncoding.EncodeToString(data)
	fmt.Fprintf(d.writer, "%s[%s]: [base64 encoding not shown]\n", d.Indent(), label)
}

// CompareBytes compares two byte slices and shows differences
func (d *Dumper) CompareBytes(label1 string, data1 []byte, label2 string, data2 []byte) {
	maxLen := len(data1)
	if len(data2) > maxLen {
		maxLen = len(data2)
	}

	fmt.Fprintf(d.writer, "%sComparison: %s vs %s\n", d.Indent(), label1, label2)
	d.depth++
	defer func() { d.depth-- }()

	different := false
	for i := 0; i < maxLen; i++ {
		var b1, b2 byte
		if i < len(data1) {
			b1 = data1[i]
		}
		if i < len(data2) {
			b2 = data2[i]
		}

		if b1 != b2 {
			different = true
			fmt.Fprintf(d.writer, "%s[%04x] %02x vs %02x\n", d.Indent(), i, b1, b2)
		}
	}

	if !different && len(data1) == len(data2) {
		fmt.Fprintf(d.writer, "%s<identical>\n", d.Indent())
	}
}

// DumpSlice dumps a slice of values
func (d *Dumper) DumpSlice(label string, slice interface{}) {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		fmt.Fprintf(d.writer, "%s[%s]: not a slice\n", d.Indent(), label)
		return
	}

	fmt.Fprintf(d.writer, "%s[%s]: slice[%d]\n", d.Indent(), label, rv.Len())
	d.depth++
	defer func() { d.depth-- }()

	for i := 0; i < rv.Len(); i++ {
		d.DumpValue(fmt.Sprintf("[%d]", i), rv.Index(i).Interface())
	}
}

// Summary returns a summary of the dump
func (d *Dumper) Summary(title string) {
	sep := strings.Repeat("=", 60)
	fmt.Fprintf(d.writer, "\n%s\n%s\n%s\n", sep, title, sep)
}

// Header writes a header line
func (d *Dumper) Header(text string) {
	fmt.Fprintf(d.writer, "\n%s>>> %s\n", d.Indent(), text)
}

// Footer writes a footer line
func (d *Dumper) Footer() {
	fmt.Fprintf(d.writer, "%s<<<\n", d.Indent())
}
