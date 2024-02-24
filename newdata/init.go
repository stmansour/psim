package newdata

import (
	"fmt"
	"time"

	"github.com/stmansour/psim/util"
)

// EconometricsRecords is a type for an array of DR records
type EconometricsRecords []EconometricsRecord

// Database is the abstraction for the data source
type Database struct {
	cfg      *util.AppConfig // application configuration info
	Datatype string          // "CSV", "MYSQL"
	CSVDB    *CSVDatasource  // valid when Datatype is "CSV"
}

// CSVDatasource is the database structure definition for csv databases
type CSVDatasource struct {
	DBFname string                 // the csv file used as a database
	DBRecs  EconometricsRecords    // all records... temporary, until we have database
	LRecs   []LinguisticDataRecord // all lingustic records
	DtStart time.Time              // earliest date with data
	DtStop  time.Time              // latest date with data
	DTypes  []string               // the list of Influencers, each has their own data type
	CSVMap  map[string]int         // which columns are where? Map the data type to a CSV column
	ColIdx  []string               // the inverse of CSVMap: supply the index to get the column name
	Nildata int64                  // data at the requested date/time did not exist
}

// EconometricsRecord is the basic structure of discount rate data
type EconometricsRecord struct {
	Date   time.Time
	Fields map[string]float64
}

// LinguisticDataRecord is a temporary structure of data for linguistic metrics
type LinguisticDataRecord struct {
	Date              time.Time
	LALLLSNScore      float64
	LALLLSPScore      float64
	LALLWHAScore      float64
	LALLWHOScore      float64
	LALLWHLScore      float64
	LALLWPAScore      float64
	LALLWDECount      float64
	LALLWDFCount      float64
	LALLWDPCount      float64
	LALLWDMCount      float64
	LUSDLSNScore_ECON float64
	LUSDLSPScore_ECON float64
	LUSDWHAScore_ECON float64
	LUSDWHOScore_ECON float64
	LUSDWHLScore_ECON float64
	LUSDWPAScore_ECON float64
	LUSDWDECount_ECON float64
	LUSDWDFCount_ECON float64
	LUSDWDPCount_ECON float64
	LUSDLIMCount_ECON float64
	LJPYLSNScore_ECON float64
	LJPYLSPScore_ECON float64
	LJPYWHAScore_ECON float64
	LJPYWHOScore_ECON float64
	LJPYWHLScore_ECON float64
	LJPYWPAScore_ECON float64
	LJPYWDECount_ECON float64
	LJPYWDFCount_ECON float64
	LJPYWDPCount_ECON float64
	LJPYLIMCount_ECON float64
	WTOILClose        float64
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

// NewDatabase creates a new database structure
// dtype: "CSV", "MYSQL"
// ------------------------------------------------------------
func NewDatabase(dtype string, cfg *util.AppConfig) (*Database, error) {
	switch dtype {
	case "CSV":
		d := Database{
			cfg:      cfg,
			Datatype: "CSV",
		}
		d.CSVDB = &CSVDatasource{}
		return &d, nil

	default:
		return nil, fmt.Errorf("unrecognized database type: %s", dtype)
	}
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
