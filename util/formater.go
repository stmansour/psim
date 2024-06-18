package util

import (
	"fmt"
	"strings"
)

// FormatField is a pair of a label and a format
type FormatField struct {
	Label        string // value in the header row
	Fmt          string // %q for strings, %.4f for floats, %d for ints
	ColumnHeader string // if blank, use Label
}

// FormatPos is a pair of a column index and a format
type FormatPos struct {
	Index int
	Fmt   string
}

// FormatTriplet is a pair - a format string and a value
// type FormatTriplet struct {
// 	Sfmt string
// 	Val  interface{}
// }

// ColData is the column name and its data
type ColData struct {
	Label string
	Val   interface{}
}

// Formater is a collection of FormatFields that can be printed in
// a row of a CSV file
type Formater struct {
	Def   []FormatField
	Field map[string]FormatPos
}

// NewFormater createas a new Formater and returns a pointer
func NewFormater(cvals []FormatField) *Formater {
	Formater := Formater{
		Def: cvals,
	}

	Formater.Field = make(map[string]FormatPos, len(cvals))

	for i := 0; i < len(cvals); i++ {
		cv := cvals[i]
		if _, ok := Formater.Field[cv.Label]; ok {
			fmt.Printf("*** FORMATER WARNING *** Duplicate label %s\n", cv.Label)
		}
		Formater.Field[cv.Label] = FormatPos{Index: i, Fmt: cv.Fmt}
	}
	return &Formater
}

// func (f *Formater) FormatTriplet(fp FormatTriplet) string {
// 	// FormatTriplet formats a value based on its format string.
// 	if fp.Sfmt == "" {
// 		return ""
// 	}
// 	return fmt.Sprintf(fp.Sfmt, fp.Val)
// }

// FormatHeader formats the header row
func (f *Formater) Header() string {
	l := len(f.Def)
	flds := make([]string, l) // this is an array with a slot for each column in the row
	for i := 0; i < l; i++ {
		if len(f.Def[i].ColumnHeader) > 0 {
			flds[i] = f.Def[i].ColumnHeader
			continue
		}
		flds[i] = f.Def[i].Label
	}
	s := strings.Join(flds, ",")
	return s
}

// FormatRow formats a row of data
func (f *Formater) Row(data []ColData) string {
	l := len(f.Def)
	flds := make([]string, l) // this is an array with a slot for each column in the row
	for i := 0; i < len(data); i++ {
		if fp, ok := f.Field[data[i].Label]; ok {
			flds[fp.Index] = fmt.Sprintf(fp.Fmt, data[i].Val)
		} else {
			fmt.Printf("*** WARNING *** Field %s not found\n", data[i].Label)
		}
	}
	s := strings.Join(flds, ",")
	return s
}
