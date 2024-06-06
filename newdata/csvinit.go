package newdata

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	if err := d.LoadMetricSourceMapFromCSV(); err != nil {
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

// LoadMetricSourceMapFromCSV reads a CSV file and maps internal metric names to their corresponding names or symbols
// in various metric supplier systems. For example, internally we refer to the gold commodity as "Gold", but to
// fetch the gold price from the TradingEconomics API, we need to request "XAUUSD:CUR". This file provides the
// necessary mappings to retrieve data from different metric suppliers.
//
// The CSV file should have the following structure:
//   - The first column contains the internal metric names.
//   - The remaining column headers represent different metric sources (e.g., TradingEconomics, GDELT).
//   - Each subsequent row maps an internal metric name (first column) to its corresponding name in the API
//     for each metric source.
//   - If a cell is empty, it means that the metric source API does not have that particular metric.
//
// Returns any error encountered, or nil on success.
// --------------------------------------------------------------------------------------------------------------------------
func (d *DatabaseCSV) LoadMetricSourceMapFromCSV() error {
	// Extract the directory from the original file path
	dir := filepath.Dir(d.DBFname)
	filename := "msm.csv"
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

	//-------------------------------------------------------------------------------------------------------------
	// Deal with the Byte Order Mark issue.  This is where a program like Excel creates a CSV file with a BOM
	// character at the beginning of the file. \ufeff is the BOM character. We just want to ignore the BOM
	//character if it is present. We do this using the strings.TrimPrefix() function.
	//-------------------------------------------------------------------------------------------------------------
	header[0] = strings.TrimPrefix(header[0], "\ufeff")

	//---------------------------------------------------------------------------------------------
	// Next, verify that the first column is "Metric" and that the remaining columns
	// are Metrics Sources. We must be able to match all the metric sources. If not,
	// return an error...
	//---------------------------------------------------------------------------------------------
	if header[0] != "Metric" {
		return fmt.Errorf("first column must be 'Metric'")
	}
	for i := 1; i < len(header); i++ {
		found := false
		for j := 0; j < len(d.MetricSrcCache); j++ {
			if strings.Contains(d.MetricSrcCache[j].Name, header[i]) {
				found = true
				d.ParentDB.MSMap[header[i]] = make(MetricSourceMap, 20)
				break
			}
		}
		if !found {
			return fmt.Errorf("could not find metric source: %s", header[i])
		}
	}

	//---------------------------------------------------------------------------------------------
	// Finally, we can read the data rows and map each metric name to its corresponding
	//---------------------------------------------------------------------------------------------
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		for i := 1; i < len(header); i++ {
			if len(row[i]) == 0 {
				continue
			}
			//               met src    metric    metric src name
			d.ParentDB.MSMap[header[i]][row[0]] = row[i]
		}
	}

	return nil
}
