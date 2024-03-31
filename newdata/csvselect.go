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
// this version took my test run case from 19 sec to 11 sec.
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

	// Binary search the record index
	for left <= right {
		mid := left + (right-left)/2
		recordUnixNano := d.DBRecs[mid].Date.UnixNano()

		// are we within a day?
		if abs(dtUnixNano-recordUnixNano) < dayInNanoseconds {
			// replacing this line brought test run time from 19 sec to 11 sec
			// if d.DBRecs[mid].Date.Year() == dt.Year() && d.DBRecs[mid].Date.Month() == dt.Month() && d.DBRecs[mid].Date.Day() == dt.Day()
			if d.DBRecs[mid].Date.Year() == dty && d.DBRecs[mid].Date.Month() == dtm && d.DBRecs[mid].Date.Day() == dtd { // Further verify that the exact calendar date matches.
				rec := d.mapSubset(&d.DBRecs[mid], fields)
				if len(rec.Fields) == 0 {
					d.IncrementNildata()
					// newrec := d.DBRecs[mid]
					// fmt.Printf("Your request was for  date: %s and the following fields\n", dt.Format("Jan 2, 2006"))
					// for k, v := range fields {
					// 	fmt.Printf("  %d. %s\n", k, v.FQMetric())
					// }
					// fmt.Printf("here are the fields found in the database record:\n")
					// for k, v := range newrec.Fields {
					// 	fmt.Printf("  %s = %f\n", k, v)
					// }
					// fmt.Printf("Check your datafield, validate that the data is present for %s. Make sure that misubclasses is defined correctly\n", dt.Format("Jan 2, 2006"))
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
