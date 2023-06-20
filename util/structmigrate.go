package util

import (
	"reflect"
)

// BuildFieldMap creates a map so that we can find
// a field's index using its name as the map index
// --------------------------------------------------------------------
func BuildFieldMap(p interface{}) map[string]int {
	var fmap = map[string]int{}
	v := reflect.ValueOf(p).Elem()
	for j := 0; j < v.NumField(); j++ {
		n := v.Type().Field(j).Name
		fmap[n] = j
	}
	return fmap
}

// MigrateStructVals copies values from pa to pb where the field
// names for the struct pa points to matches the field names in
// the struct pb points to.
// There is a basic assumption that the data will either copy directly
// or convert cleanly from one struct to another.  Where it does not
// it will call XJSONprocess to see if there is a known conversion.
// --------------------------------------------------------------------
func MigrateStructVals(pa interface{}, pb interface{}) error {
	m := BuildFieldMap(pb)
	ar := reflect.ValueOf(pa).Elem()
	for i := 0; i < ar.NumField(); i++ {
		fa := ar.Field(i)
		afldname := ar.Type().Field(i).Name
		if !fa.IsValid() {
			continue
		}
		bdx, ok := m[afldname]
		if !ok {
			continue
		}
		br := reflect.ValueOf(pb).Elem()
		fb := br.Field(bdx)
		if !fb.CanSet() { // BEWARE: if a field name begins with a lowercase letter it cannot be set
			continue
		}
		if fa.Type() == fb.Type() {
			fb.Set(reflect.ValueOf(fa.Interface()))
		}
	}
	return nil
}
