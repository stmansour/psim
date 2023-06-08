package data

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// DRCSV is the csv data file that is used for Discount Rate information
var DRCSV = string("data/dr.csv")

// DiscountRateRecord is the basic structure of discount rate data
type DiscountRateRecord struct {
	Date           time.Time
	USDiscountRate float64
	JPDiscountRate float64
	DRRatio        float64
	EXClose        float64
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
// func (r DiscountRateRecords) GetRecord(date time.Time) (DiscountRateRecord, error) {
// 	for _, record := range r {
// 		if record.Date.Equal(date) {
// 			return record, nil
// 		}
// 	}
// 	return DiscountRateRecord{}, fmt.Errorf("record not found for date %s", date.Format("2006-01-02"))
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
func DRFindRecord(dt time.Time) *DiscountRateRecord {
	left := 0
	right := len(DR.DRRecs) - 1

	for left <= right {
		mid := left + (right-left)/2
		if DR.DRRecs[mid].Date.Year() == dt.Year() && DR.DRRecs[mid].Date.Month() == dt.Month() && DR.DRRecs[mid].Date.Day() == dt.Day() {
			return &DR.DRRecs[mid]
		} else if DR.DRRecs[mid].Date.Before(dt) {
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

//****************************************************************************
//
//  Column naming and formatting conventions:
//
// Date,USD_DiscountRate,JPY_DiscountRate,USDJPY_DRRatio
//
//----------------------------------------------------------------------------
//  COLUMN 1  =  DATE
//----------------------------------------------------------------------------
//  Date     - MM/DD/YYYY
//
//----------------------------------------------------------------------------
//  COLUMN 2  =  DATATYPE RATIO
//----------------------------------------------------------------------------
//  The data type for C1 divided by the datatype for C2
//
//  For Currency, use the ISO 4217 naming conventions, 3-letter strings, the
//  first two identify the country, the last is represents the currency name.
//  Examples:  USD = United States Dollar,  JPY Japanese Yen
//
//  Exchange Rate - use the ISO 3-letter strings, list C1 first, followed by
//                  C2.  So for example, for C1 = USD and C2 = JPY, the
//                  ExchangeRate would be USDJPY
//
//  DataType - use a 2 letter identifier:
//             DR = Discount Rate
//             IR = Inflation Rate
//             UR = Unemployment Rate
//             EX = Exchange Rate -- can be appended with "Open", "Low", "High", "Close"
//
//  DataTypeRatio - use the exchange rate,
//                  followed by the 2 letter datatype, followed by an R
//                  For the USD JPY example, Discount Rate Ratio would be:
//                  USDJPYDRR
//
//****************************************************************************

// DRInit - initialize this subsystem.
//  1. Determine whether the data will come from a CSV file, a SQL
//     database, or an online service.  As of this writing we only have
//     CSV file data implemented.
//  2. If data source is CSV read it in and validate that we have the
//     correct information.
//
// ---------------------------------------------------------------------------
func DRInit() error {
	file, err := os.Open(DRCSV)
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

	DRRatioCol := -1
	EXCloseCol := -1
	records := DiscountRateRecords{}
	for i, line := range lines {
		if i == 0 {
			// handle the unicode case...
			line[0] = HandleUTF8FileChars(line[0])

			if line[0] != "Date" {
				log.Panicf("Problem with %s, column 1 is labelled %q, it should be %q\n", DRCSV, line[0], "Date")
			}
			//---------------------------------------------------------
			// Search for the columns of interest.
			//---------------------------------------------------------
			for j := 1; j < len(line); j++ {
				validcpair := validCurrencyPair(line[j]) // do the first 6 chars make a currency pair that matches with the simulation configuation?
				l := len(line[j])
				if l == 13 && strings.HasSuffix(line[j], "DRRatio") && validcpair { // len("USDJPYDRRatio") = 13
					DRRatioCol = j
				}
				if l == 13 && strings.HasSuffix(line[j], "EXClose") && validcpair { // len("USDJPYEXClose") = 13
					EXCloseCol = j
				}
			}
			if DRRatioCol < 0 {
				return fmt.Errorf("No column in %s had label  %s%s%s, which is required for the current simulation configuration",
					DRCSV, DInfo.cfg.C1, DInfo.cfg.C2, "DRRatio")
			}
			if EXCloseCol < 0 {
				return fmt.Errorf("No column in %s had label  %s%s%s, which is required for the current simulation configuration",
					DRCSV, DInfo.cfg.C1, DInfo.cfg.C2, "EXClose")
			}
			continue // continue to the next line now
		}

		date, err := time.Parse("1/2/2006", line[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// usDiscountRate, err := strconv.ParseFloat(strings.TrimSuffix(line[1], "%"), 64)
		// if err != nil {
		// 	fmt.Println(err)
		// 	os.Exit(1)
		// }
		// usDiscountRate /= 100

		// jpDiscountRate, err := strconv.ParseFloat(strings.TrimSuffix(line[2], "%"), 64)
		// if err != nil {
		// 	fmt.Println(err)
		// 	os.Exit(1)
		// }
		// jpDiscountRate /= 100

		USDJPYDRRatio, err := strconv.ParseFloat(line[DRRatioCol], 64)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		records = append(records, DiscountRateRecord{
			Date: date,
			// USDiscountRate: usDiscountRate,
			// JPDiscountRate: jpDiscountRate,
			DRRatio: USDJPYDRRatio,
		})
	}

	DR.DRRecs = records
	sort.Sort(DR.DRRecs)
	l := DR.DRRecs.Len()
	DR.DtStart = DR.DRRecs[0].Date
	DR.DtStop = DR.DRRecs[l-1].Date
	return nil
}

func validCurrencyPair(line string) bool {
	if len(line) < 6 {
		return false
	}
	myC1 := line[0:3]
	myC2 := line[3:6]
	return myC1 == DInfo.cfg.C1 && myC2 == DInfo.cfg.C2
}
