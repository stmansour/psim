package data

import (
	"time"
)

// DRInfo is meta information about the discount rate data
type DRInfo struct {
	DtStart time.Time // earliest date with data
	DtStop  time.Time // latest date with data
}

// DR is a structure of data used by all the DR routines
var DR struct {
	DRRecs  RatesAndRatiosRecords // all records... temporary, until we have database
	DtStart time.Time             // earliest date with data
	DtStop  time.Time             // latest date with data
}

// Len returns the length of the supplied RatesAndRatiosRecords array
func (r RatesAndRatiosRecords) Len() int {
	return len(r)
}

// Less is used to sort the records
func (r RatesAndRatiosRecords) Less(i, j int) bool {
	return r[i].Date.Before(r[j].Date)
}

// Swap is used to do exactly what you think it does
func (r RatesAndRatiosRecords) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

// GetRecord returns a RatesAndRatiosRecord based on the supplied date
//
// INPUTS
//
//	date = day for which we want DiscountRate data. Only the day, month, and year are significant.
//
// RETURNS
//
//	the record found if err == nil
//	an empty record and an error if something went wrong.
//
// ----------------------------------------------------------------------------------------------------
// func (r RatesAndRatiosRecords) GetRecord(date time.Time) (RatesAndRatiosRecord, error) {
// 	for _, record := range r {
// 		if record.Date.Equal(date) {
// 			return record, nil
// 		}
// 	}
// 	return RatesAndRatiosRecord{}, fmt.Errorf("record not found for date %s", date.Format("2006-01-02"))
// }

// DRFindRecord returns the record associated with the input date
//
// INPUTS
//
//	dt = date of record to return
//
// RETURNS
//
//	pointer to the record on the supplied date
//	nil - record was not found
//
// ---------------------------------------------------------------------------
func DRFindRecord(dt time.Time) *RatesAndRatiosRecord {
	left := 0
	right := len(DInfo.DBRecs) - 1

	for left <= right {
		mid := left + (right-left)/2
		if DInfo.DBRecs[mid].Date.Year() == dt.Year() && DInfo.DBRecs[mid].Date.Month() == dt.Month() && DInfo.DBRecs[mid].Date.Day() == dt.Day() {
			return &DInfo.DBRecs[mid]
		} else if DInfo.DBRecs[mid].Date.Before(dt) {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return nil
}

// DRGetDataInfo returns meta information about the data
//
// # INPUTS
//
// RETURNS
//
//	{ dtStart, dtStop}
//
// ---------------------------------------------------------------------------
func DRGetDataInfo() DRInfo {
	rec := DRInfo{
		DtStart: DR.DtStart,
		DtStop:  DR.DtStop,
	}
	return rec
}
