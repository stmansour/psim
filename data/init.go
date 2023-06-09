package data

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/stmansour/psim/util"
)

// DInfo maintains data needed by the data subsystem.
// The primary need is the two currencies C1 & C2
var DInfo struct {
	cfg    *util.AppConfig
	DBRecs RatesAndRatiosRecords // all records... temporary, until we have database
}

// RatesAndRatiosRecords is a type for an array of DR records
type RatesAndRatiosRecords []RatesAndRatiosRecord

// CurrencyInfo contains information about currencies used in this program
type CurrencyInfo struct {
	Country      string // name of the issuing country
	CountryCode  string // two-letter designator for country
	Currency     string // name of the currency
	CurrencyCode string // typically the first char of the currency name
}

// Currencies is a list a CurrencyInfo for all the currencies supported by this program
var Currencies = []CurrencyInfo{
	{
		Country:      "United States",
		CountryCode:  "US",
		Currency:     "Dollar",
		CurrencyCode: "D",
	},
	{
		Country:      "Japan",
		CountryCode:  "JP",
		Currency:     "Yen",
		CurrencyCode: "Y",
	},
	{
		Country:      "Great Britain",
		CountryCode:  "GB",
		Currency:     "Pound",
		CurrencyCode: "P",
	},
	{
		Country:      "Australia",
		CountryCode:  "AU",
		Currency:     "Dollar",
		CurrencyCode: "D",
	},
}

// PLATODB is the csv data file that is used for Discount Rate information
var PLATODB = string("data/platodb.csv")

// RatesAndRatiosRecord is the basic structure of discount rate data
type RatesAndRatiosRecord struct {
	Date time.Time
	// USDiscountRate float64
	// JPDiscountRate float64
	DRRatio float64
	EXClose float64
}

// Init calls the initialize routine for all data types
// ------------------------------------------------------------
func Init(cfg *util.AppConfig) error {
	DInfo.cfg = cfg
	if err := LoadCsvDB(); err != nil {
		return err
	}
	ERInit()
	return nil
}

// HandleUTF8FileChars returns the first line of the file with
// utf8 markers removed if they were there. Otherwise, it just
// returns the input string.
// ----------------------------------------------------------------
func HandleUTF8FileChars(line string) string {
	bom := []byte{0xEF, 0xBB, 0xBF}
	strBytes := []byte(line)

	if len(strBytes) >= len(bom) && bytes.Equal(strBytes[:len(bom)], bom) {
		// If the line starts with BOM, remove it.
		line = string(strBytes[len(bom):])
	}
	return line
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
				}
				if l == 13 && strings.HasSuffix(line[j], "EXClose") && validcpair { // len("USDJPYEXClose") = 13
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

		EXClose, err := strconv.ParseFloat(line[DRRatioCol], 64)
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
	DR.DtStart = DInfo.DBRecs[0].Date
	DR.DtStop = DInfo.DBRecs[l-1].Date
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
