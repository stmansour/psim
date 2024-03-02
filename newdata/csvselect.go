package newdata

import (
	"fmt"
	"time"
)

// Select reads and returns data from the database.
// ----------------------------------------------------------------------------
func (d *Database) Select(dt time.Time, fields []string) (*EconometricsRecord, error) {
	var err error
	switch d.cfg.DBSource {
	case "CSV":
		return d.CSVDB.Select(dt, fields)
	default:
		err = fmt.Errorf("unrecognized data source: %s", d.cfg.DBSource)
		return nil, err
	}
}

// Select does the select function for CSV databases
// ----------------------------------------------------------------------------
func (d *DatasourceCSV) Select(dt time.Time, fields []string) (*EconometricsRecord, error) {
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
func (d *DatasourceCSV) mapSubset(rec *EconometricsRecord, ss []string) *EconometricsRecord {
	var nr EconometricsRecord
	nr.Date = rec.Date
	nr.Fields = make(map[string]float64, len(ss))
	for _, key := range ss {
		if value, exists := rec.Fields[key]; exists {
			nr.Fields[key] = value
		} else {
			d.Nildata++
		}
		// Optionally, handle the "not exists" case here, if needed.
	}
	return &nr
}
