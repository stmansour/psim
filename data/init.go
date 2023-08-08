package data

import (
	"bytes"
	"fmt"
	"time"

	"github.com/stmansour/psim/util"
)

// DInfo maintains data needed by the data subsystem.
// The primary need is the two currencies C1 & C2
var DInfo struct {
	cfg     *util.AppConfig
	DBRecs  RatesAndRatiosRecords // all records... temporary, until we have database
	DtStart time.Time             // earliest date with data
	DtStop  time.Time             // latest date with data
	DTypes  []string              // the list of Influencers, each has their own data type
	CSVMap  map[string]int        // which columns are where? Map the data type to a CSV column
}

// RatesAndRatiosRecord is the basic structure of discount rate data
type RatesAndRatiosRecord struct {
	Date    time.Time
	CCRatio float64 // valid if FLAGS & 0 is != 0
	DRRatio float64 // valid if FLAGS & 1 is != 0
	GDRatio float64 // valid if FLAGS & 2 is != 0
	IRRatio float64 // valid if FLAGS & 3 is != 0
	MSRatio float64 // valid if FLAGS & 4 is != 0
	URRatio float64 // valid if FLAGS & 5 is != 0
	EXClose float64 // valid if FLAGS & 6 is != 0
	FLAGS   uint64  // can hold flags for the first 64 values associated with the Date
}

// PLATODB is the csv data file that is used for Discount Rate information
var PLATODB = string("data/platodb.csv")

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

// Init calls the initialize routine for all data types
// ------------------------------------------------------------
func Init(cfg *util.AppConfig) error {
	DInfo.cfg = cfg
	switch DInfo.cfg.DBSource {
	case "CSV":
		if err := LoadCsvDB(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unimplemented DBSource %s", DInfo.cfg.DBSource)
	}
	// ERInit()
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
