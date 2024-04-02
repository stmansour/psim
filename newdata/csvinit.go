package newdata

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/stmansour/psim/util"
)

// DatabaseCSV is the database structure definition for csv databases
type DatabaseCSV struct {
	DBFname         string              // the csv file used as a database
	DBPath          string              // path where DB files are kept
	DBRecs          EconometricsRecords // all records... temporary, until we have database
	DtStart         time.Time           // earliest date with data
	DtStop          time.Time           // latest date with data
	DTypes          []string            // the list of Influencers, each has their own data type
	CSVMap          map[string]int      // which columns are where? Map the data type to a CSV column
	MetricSrcCache  []MetricsSource     // known data sources
	ColIdx          []string            // the inverse of CSVMap: supply the index to get the column name
	Nildata         int64               // data at the requested date/time did not exist
	ParentDB        *Database           // the database that contains me
	NumMetricFields int                 // number of metric fields, used to allocate the metric map
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
// func (d *Database) Init() error {
// 	d.Mim = NewInfluencerManager()
// 	switch d.Datatype {
// 	case "CSV":
// 		return d.CSVInit()
// 	default:
// 		return fmt.Errorf("unrecognized database type: %s", d.Datatype)
// 	}
// }

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
func (d *DatabaseCSV) CSVInit() error {
	if err := d.LoadCsvDB(); err != nil {
		return err
	}
	if err := d.ParentDB.Mim.Init(d.ParentDB); err != nil {
		return err
	}
	if err := d.LoadMetricsSourceCache(); err != nil {
		return err
	}

	return nil
}

// LoadMetricsSourceCache reads the contents of a CSV file into a slice of MetricsSource, regardless of the column order.
func (d *DatabaseCSV) LoadMetricsSourceCache() error {
	// Extract the directory from the original file path
	dir := filepath.Dir(d.DBFname)
	filename := "metricssources.csv"
	newFilePath := filepath.Join(dir, filename)
	file, err := os.Open(newFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read the header row
	header, err := reader.Read()
	if err != nil {
		return err
	}

	// Map column names to their indices
	columnIndices := make(map[string]int)
	for i, columnName := range header {
		columnIndices[columnName] = i
	}

	var metrics []MetricsSource

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		var m MetricsSource
		for columnName, i := range columnIndices {
			switch columnName {
			case "MSID":
				m.MSID, err = strconv.Atoi(row[i])
				if err != nil {
					return fmt.Errorf("error parsing MSID: %v", err)
				}
			case "LastUpdate":
				m.LastUpdate, err = util.StringToDate(row[i])
				if err != nil {
					return fmt.Errorf("error parsing LastUpdate: %v", err)
				}
			case "URL":
				m.URL = row[i]
			case "Name":
				m.Name = row[i]
			}
		}

		metrics = append(metrics, m)
	}
	d.MetricSrcCache = metrics

	return nil
}
