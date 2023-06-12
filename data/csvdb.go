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

// DRInfo is meta information about the discount rate data
type DRInfo struct {
	DtStart time.Time // earliest date with data
	DtStop  time.Time // latest date with data
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

// LoadCsvDB - Read in the data from the CSV file
//  1. Determine whether the data will come from a CSV file, a SQL
//     database, or an online service.  As of this writing we only have
//     CSV file data implemented.
//  2. If data source is CSV read it in and validate that we have the
//     correct information.
//
// ---------------------------------------------------------------------------
func LoadCsvDB() error {
	file, err := os.Open(PLATODB)
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
	records := RatesAndRatiosRecords{}
	for i, line := range lines {
		if i == 0 {
			// handle the unicode case...
			line[0] = HandleUTF8FileChars(line[0])

			if line[0] != "Date" {
				log.Panicf("Problem with %s, column 1 is labelled %q, it should be %q\n", PLATODB, line[0], "Date")
			}
			//---------------------------------------------------------
			// Search for the columns of interest.
			//---------------------------------------------------------
			for j := 1; j < len(line); j++ {
				validcpair := validCurrencyPair(line[j]) // do the first 6 chars make a currency pair that matches with the simulation configuation?
				l := len(line[j])
				if l == 13 && strings.HasSuffix(line[j], "DRRatio") && validcpair { // len("USDJPYDRRatio") = 13
					DRRatioCol = j
				} else if l == 13 && strings.HasSuffix(line[j], "EXClose") && validcpair { // len("USDJPYEXClose") = 13
					EXCloseCol = j
				}
			}
			if DRRatioCol < 0 {
				return fmt.Errorf("No column in %s had label  %s%s%s, which is required for the current simulation configuration",
					PLATODB, DInfo.cfg.C1, DInfo.cfg.C2, "DRRatio")
			}
			if EXCloseCol < 0 {
				return fmt.Errorf("No column in %s had label  %s%s%s, which is required for the current simulation configuration",
					PLATODB, DInfo.cfg.C1, DInfo.cfg.C2, "EXClose")
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

		DRRatio, err := strconv.ParseFloat(line[DRRatioCol], 64)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		EXClose, err := strconv.ParseFloat(line[EXCloseCol], 64)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		records = append(records, RatesAndRatiosRecord{
			Date: date,
			// USDiscountRate: usDiscountRate,
			// JPDiscountRate: jpDiscountRate,
			DRRatio: DRRatio,
			EXClose: EXClose,
		})
	}

	DInfo.DBRecs = records
	sort.Sort(DInfo.DBRecs)
	l := DInfo.DBRecs.Len()
	DInfo.DtStart = DInfo.DBRecs[0].Date
	DInfo.DtStop = DInfo.DBRecs[l-1].Date
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
func CSVDBFindRecord(dt time.Time) *RatesAndRatiosRecord {
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
