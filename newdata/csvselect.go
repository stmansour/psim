package newdata

import (
	"time"
)

// Select does the select function for CSV databases
// ----------------------------------------------------------------------------
func (d *DatabaseCSV) Select(dt time.Time, fields []FieldSelector) (*EconometricsRecord, error) {
	left := 0
	right := len(d.DBRecs) - 1

	for left <= right {
		mid := left + (right-left)/2
		if d.DBRecs[mid].Date.Year() == dt.Year() && d.DBRecs[mid].Date.Month() == dt.Month() && d.DBRecs[mid].Date.Day() == dt.Day() {
			rec := d.mapSubset(&d.DBRecs[mid], fields)
			return rec, nil
		} else if d.DBRecs[mid].Date.Before(dt) {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return nil, nil
}

// mapSubset is a utility function to populate a map with only the fields
// the caller requested.
// ----------------------------------------------------------------------------
func (d *DatabaseCSV) mapSubset(rec *EconometricsRecord, fs []FieldSelector) *EconometricsRecord {
	var nr EconometricsRecord
	nr.Date = rec.Date
	ss := []string{}
	for _, v := range fs {
		ss = append(ss, v.FQMetric())
	}

	if len(ss) > 0 {
		nr.Fields = make(map[string]float64, len(ss))
		for _, key := range ss {
			if value, exists := rec.Fields[key]; exists {
				nr.Fields[key] = value
			} else {
				d.Nildata++
			}
			// Optionally, handle the "not exists" case here, if needed.
		}
	} else {
		fields := map[string]float64{}
		for k, v := range rec.Fields {
			fields[k] = v
		}
		nr.Fields = fields
	}
	return &nr
}
