package newdata

import (
	"sync/atomic"
	"time"
)

// IncrementNildata safely increments the nil data counter
// in multithreaded environments
// ----------------------------------------------------------------------------
func (d *DatabaseCSV) IncrementNildata() {
	atomic.AddInt64(&d.Nildata, 1)
}

// Select does the select function for CSV databases
// ----------------------------------------------------------------------------
func (d *DatabaseCSV) Select(dt time.Time, fields []FieldSelector) (*EconometricsRecord, error) {
	dayInNanoseconds := int64(24 * 60 * 60 * 1000 * 1000 * 1000) // 36 hours in nanoseconds
	left := 0
	right := len(d.DBRecs) - 1

	// get the date with no tz offset...
	dty := dt.Year()
	dtm := dt.Month()
	dtd := dt.Day()
	dtt := time.Date(dty, dtm, dtd, 0, 0, 0, 0, time.UTC)

	// Pre-calculate the UnixNano representation of dt for comparison.
	dtUnixNano := dtt.UnixNano()

	// Binary search the record index
	for left <= right {
		mid := left + (right-left)/2
		recordUnixNano := d.DBRecs[mid].Date.UnixNano()

		// are we within a day?
		if abs(dtUnixNano-recordUnixNano) < dayInNanoseconds {
			if d.DBRecs[mid].Date.Year() == dty && d.DBRecs[mid].Date.Month() == dtm && d.DBRecs[mid].Date.Day() == dtd { // Further verify that the exact calendar date matches.
				rec := d.mapSubset(&d.DBRecs[mid], fields)
				if len(rec.Fields) == 0 {
					d.IncrementNildata()
				}
				return rec, nil
			} else if d.DBRecs[mid].Date.Before(dt) {
				left = mid + 1
			} else {
				right = mid - 1
			}
		} else if recordUnixNano < dtUnixNano {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return nil, nil
}

// Helper function for absolute value calculation
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// mapSubset is a utility function to populate a map with only the fields
// the caller requested.
// When len(fs) == 0, the entire record is returned. The assumption here is
// that the caller always uses this data in a read-only manner. So it is safe
// to return the entire record.
// ----------------------------------------------------------------------------
func (d *DatabaseCSV) mapSubset(rec *EconometricsRecord, fs []FieldSelector) *EconometricsRecord {
	nr := EconometricsRecord{
		Date: rec.Date,
	}
	if len(fs) > 0 {
		nr.Fields = make(map[string]MetricInfo, len(fs))
		for _, selector := range fs {
			key := selector.FQMetric()
			if value, exists := rec.Fields[key]; exists {
				nr.Fields[key] = value
			} else {
				d.Nildata++
			}
		}
	} else {
		nr.Fields = rec.Fields
	}
	return &nr
}
