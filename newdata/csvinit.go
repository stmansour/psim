package newdata

import (
	"fmt"
	"time"
)

// DatasourceCSV is the database structure definition for csv databases
type DatasourceCSV struct {
	DBFname string              // the csv file used as a database
	DBRecs  EconometricsRecords // all records... temporary, until we have database
	DtStart time.Time           // earliest date with data
	DtStop  time.Time           // latest date with data
	DTypes  []string            // the list of Influencers, each has their own data type
	CSVMap  map[string]int      // which columns are where? Map the data type to a CSV column
	ColIdx  []string            // the inverse of CSVMap: supply the index to get the column name
	Nildata int64               // data at the requested date/time did not exist
}

// PLATODB is the csv data file that is used for Discount Rate information
var PLATODB = string("data/platodb.csv")

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

// Init initializes the database
// ---------------------------------------------------------------------------------
func (d *Database) Init() error {
	switch d.Datatype {
	case "CSV":
		return d.CSVInit()
	default:
		return fmt.Errorf("unrecognized database type: %s", d.Datatype)
	}
}

// SetCSVFilename sets the CSV filename
// ---------------------------------------------------------------------------------
func (d *Database) SetCSVFilename(f string) {
	d.CSVDB.DBFname = f
}

// CSVInit calls the initialize routine for all data types
// INPUTS
// cfg      - pointer to the AppConfig struct
// dbfname - db file name override.  If nil or len() == 0 then it uses the default
// --------------------------------------------------------------------------------
func (d *Database) CSVInit() error {
	if err := d.CSVDB.LoadCsvDB(); err != nil {
		return err
	}
	return nil
}
