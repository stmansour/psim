package newdata

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/stmansour/psim/util"
)

//****************************************************************************
//
//  For Currency, use the ISO 4217 naming conventions, 3-letter strings, the
//  first two identify the country, the last is represents the currency name.
//
//  	Examples:
//              USD = United States Dollar
//              AUD = Australian dollar
//              JPY Japanese Yen
//
//  DataTypes:
//		column names / field names
//
//  Qualifier
//      Ratio - indicates that the value is a ratio
//      Close - indicates that this is the "Close" value for the date.  Currently,
//              it applies only the Exchange Rate (EX) info
//
//      Examples:
//              USDJPYDRRatio - USD / JPY Discount Rate Ratio
//              USDJPYEXClose = USD / JPY Exchange Rate Closing value
//
//****************************************************************************

// LoadCsvDB - Read in the data from the CSV file
// -----------------------------------------------------------------------------
func (d *DatabaseCSV) LoadCsvDB() error {
	fname := PLATODB // this is the default: data/platodb.csv
	if len(d.DBFname) > 0 {
		fname = d.DBFname
	}

	file, err := os.Open(fname)
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
	//----------------------------------------------------------------------
	// Keep track of the column with the data needed for each ratio.  This
	// is based on the two currencies in the simulation.
	//----------------------------------------------------------------------
	d.CSVMap = map[string]int{}
	for k := 0; k < len(d.DTypes); k++ {
		d.CSVMap[d.DTypes[k]] = -1 // haven't located this column yet
	}
	records := EconometricsRecords{}

	for i, line := range lines {
		if i == 0 {
			// handle the unicode case...
			line[0] = HandleUTF8FileChars(line[0])
			if line[0] != "Date" {
				log.Panicf("Problem with %s, column 1 is labelled %q, it should be %q\n", PLATODB, line[0], "Date")
			}
			d.ColIdx = append(d.ColIdx, "Date")
			//----------------------------------------------------------------------------
			// Save the column names for multiple ways to index
			//----------------------------------------------------------------------------
			for j := 1; j < len(line); j++ {
				s := util.Stripchars(line[j], " ")
				d.CSVMap[s] = j
				d.ColIdx = append(d.ColIdx, s)
			}
			continue // remaining rows are data, code below handles data, continue to the next line now
		}

		rec := EconometricsRecord{
			Fields: map[string]float64{},
		}
		rec.Date, err = util.StringToDate(line[0])
		if err != nil {
			fmt.Printf("*** ERROR *** on line %d, date = %q", i, line[0])
			fmt.Println(err)
			os.Exit(1)
		}
		for j := 1; j < len(line); j++ {
			if len(line[j]) == 0 {
				continue
			}
			x, err := strconv.ParseFloat(util.Stripchars(line[j], ","), 64)
			if err != nil {
				log.Panicf("invalid float value: %q, err = %s\n", line[j], err)
			}
			rec.Fields[d.ColIdx[j]] = x
		}
		records = append(records, rec)
	}

	d.DBRecs = records
	sort.Sort(d.DBRecs)
	l := d.DBRecs.Len()
	d.DtStart = d.DBRecs[0].Date
	d.DtStop = d.DBRecs[l-1].Date
	return nil
}

// Len returns the length of the supplied EconometricsRecords array
func (r EconometricsRecords) Len() int {
	return len(r)
}

// Less is used to sort the records
func (r EconometricsRecords) Less(i, j int) bool {
	return r[i].Date.Before(r[j].Date)
}

// Swap is used to do exactly what you think it does
func (r EconometricsRecords) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

// CSVDBFindRecord returns the record associated with the input date
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
func (d *DatabaseCSV) CSVDBFindRecord(dt time.Time) *EconometricsRecord {
	left := 0
	right := len(d.DBRecs) - 1

	for left <= right {
		mid := left + (right-left)/2
		if d.DBRecs[mid].Date.Year() == dt.Year() && d.DBRecs[mid].Date.Month() == dt.Month() && d.DBRecs[mid].Date.Day() == dt.Day() {
			return &d.DBRecs[mid]
		} else if d.DBRecs[mid].Date.Before(dt) {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return nil
}
