package newdata

import (
	"time"
)

// func (d *DatabaseCSV) oldSelect(dt time.Time, fields []FieldSelector) (*EconometricsRecord, error) {
// 	left := 0
// 	right := len(d.DBRecs) - 1
// 	dty := dt.Year()
// 	dtm := dt.Month()
// 	dtd := dt.Day()

// 	for left <= right {
// 		mid := left + (right-left)/2
// 		// if d.DBRecs[mid].Date.Year() == dt.Year() && d.DBRecs[mid].Date.Month() == dt.Month() && d.DBRecs[mid].Date.Day() == dt.Day() {
// 		if d.DBRecs[mid].Date.Year() == dty && d.DBRecs[mid].Date.Month() == dtm && d.DBRecs[mid].Date.Day() == dtd { // this drupped the run time from 19 sec to 16 sec
// 			rec := d.mapSubset(&d.DBRecs[mid], fields)
// 			return rec, nil
// 		} else if d.DBRecs[mid].Date.Before(dt) {
// 			left = mid + 1
// 		} else {
// 			right = mid - 1
// 		}
// 	}
// 	return nil, nil
// }

// Select does the select function for CSV databases
// this version took my test run case from 16 sec to 11 sec.
// so, going from 19sec run time to 11 sec   ===>  42.1% improvement!!
// ----------------------------------------------------------------------------
func (d *DatabaseCSV) Select(dt time.Time, fields []FieldSelector) (*EconometricsRecord, error) {
	const dayInNanoseconds = 24 * 60 * 60 * 1000 * 1000 * 1000 // 24 hours in nanoseconds
	left := 0
	right := len(d.DBRecs) - 1

	// Pre-calculate the UnixNano representation of dt for comparison.
	dtUnixNano := dt.UnixNano()
	dty := dt.Year()
	dtm := dt.Month()
	dtd := dt.Day()

	for left <= right {
		mid := left + (right-left)/2
		recordUnixNano := d.DBRecs[mid].Date.UnixNano()

		// are we within 24 hours?
		if abs(dtUnixNano-recordUnixNano) < dayInNanoseconds {
			if d.DBRecs[mid].Date.Year() == dty && d.DBRecs[mid].Date.Month() == dtm && d.DBRecs[mid].Date.Day() == dtd { // Further verify that the exact calendar date matches.
				rec := d.mapSubset(&d.DBRecs[mid], fields)
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
// ----------------------------------------------------------------------------
func (d *DatabaseCSV) mapSubset(rec *EconometricsRecord, fs []FieldSelector) *EconometricsRecord {
	var nr EconometricsRecord
	nr.Date = rec.Date
	if len(fs) > 0 {
		nr.Fields = make(map[string]float64, len(fs))
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

// func (d *DatabaseCSV) mapSubset(rec *EconometricsRecord, fs []FieldSelector) *EconometricsRecord {
// 	var nr EconometricsRecord
// 	nr.Date = rec.Date
// 	ss := []string{}
// 	for _, v := range fs {
// 		ss = append(ss, v.FQMetric())
// 	}

// 	if len(ss) > 0 {
// 		nr.Fields = make(map[string]float64, len(ss))
// 		for _, key := range ss {
// 			if value, exists := rec.Fields[key]; exists {
// 				nr.Fields[key] = value
// 			} else {
// 				d.Nildata++
// 			}
// 			// Optionally, handle the "not exists" case here, if needed.
// 		}
// 	} else {
// 		fields := map[string]float64{}
// 		for k, v := range rec.Fields {
// 			fields[k] = v
// 		}
// 		nr.Fields = fields
// 	}
// 	return &nr
// }
