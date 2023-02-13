package data

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// DiscountRateRecord is the basic structure of discount rate data
type DiscountRateRecord struct {
	Date           time.Time
	USDiscountRate float64
	JPDiscountRate float64
	USJPDRRatio    float64
}

// DiscountRateRecords is a type for an array of DR records
type DiscountRateRecords []DiscountRateRecord

// DRInfo is meta information about the discount rate data
type DRInfo struct {
	DtStart time.Time // earliest date with data
	DtStop  time.Time // latest date with data
}

// DR is a structure of data used by all the DR routines
var DR struct {
	DRRecs  DiscountRateRecords // all records... temporary, until we have database
	DtStart time.Time           // earliest date with data
	DtStop  time.Time           // latest date with data
}

// Len returns the length of the supplied DiscountRateRecords array
func (r DiscountRateRecords) Len() int {
	return len(r)
}

// Less is used to sort the records
func (r DiscountRateRecords) Less(i, j int) bool {
	return r[i].Date.Before(r[j].Date)
}

// Swap is used to do exactly what you think it does
func (r DiscountRateRecords) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

// GetRecord returns a DiscountRateRecord based on the supplied date
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
func (r DiscountRateRecords) GetRecord(date time.Time) (DiscountRateRecord, error) {
	for _, record := range r {
		if record.Date.Equal(date) {
			return record, nil
		}
	}
	return DiscountRateRecord{}, fmt.Errorf("record not found for date %s", date.Format("2006-01-02"))
}

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
func DRFindRecord(dt time.Time) *DiscountRateRecord {
	// Perform a binary search to find the record with the specified dt
	index := sort.Search(len(DR.DRRecs), func(i int) bool {
		return DR.DRRecs[i].Date.After(dt) || DR.DRRecs[i].Date.Equal(dt)
	})
	if index == len(DR.DRRecs) || DR.DRRecs[index].Date.After(dt) {
		return nil
	}
	return &DR.DRRecs[index]
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

// DRInit - initialize this subsystem
// ---------------------------------------------------------------------------
func DRInit() {
	file, err := os.Open("data/dr.csv")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	lines, err := reader.ReadAll()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	records := DiscountRateRecords{}
	for i, line := range lines {
		if i == 0 {
			continue
		}

		date, err := time.Parse("1/2/2006", line[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		usDiscountRate, err := strconv.ParseFloat(strings.TrimSuffix(line[1], "%"), 64)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		usDiscountRate /= 100

		jpDiscountRate, err := strconv.ParseFloat(strings.TrimSuffix(line[2], "%"), 64)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		jpDiscountRate /= 100

		usJpdrRatio, err := strconv.ParseFloat(line[3], 64)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		records = append(records, DiscountRateRecord{
			Date:           date,
			USDiscountRate: usDiscountRate,
			JPDiscountRate: jpDiscountRate,
			USJPDRRatio:    usJpdrRatio,
		})
	}

	DR.DRRecs = records
	sort.Sort(DR.DRRecs)
	l := DR.DRRecs.Len()
	DR.DtStart = DR.DRRecs[0].Date
	DR.DtStop = DR.DRRecs[l-1].Date
}

/*

create a golang program to read a csv file named er.csv, containing these fields:
Symbol string, Date time.Time, Open,High,Low,Close are float64.
The date can be parsed in this format '1/2/2006'.
Store the records in a global array sorted by date.  Provide an access function where the input parameter
is a date (only day, month, and year are relevant), and it returns the associated record. Since the
array is ordered by date, use an efficient algorithm to search for the date.

*/
