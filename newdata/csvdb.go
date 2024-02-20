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
func (d *CSVDatasource) LoadCsvDB() error {
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

	if err = d.LoadLinguistics(lines); err != nil {
		fmt.Printf("error from LoadLingustics: %s\n", err.Error())
	}
	// util.DPrintf("Loaded %d records.   %s - %s\n", l, d.DtStart.Format("jan 2, 2006"), d.DtStop.Format("jan 2, 2006"))
	return nil
}

// LoadLinguistics loads the linguistic stats from the CSV file
//
// RETURNS
//
//	any error encountered
//
// -----------------------------------------------------------------------------
func (d *CSVDatasource) LoadLinguistics(lines [][]string) error {
	var records []LinguisticDataRecord
	var err error
	cols := make(map[string]int, 100)

	for i, line := range lines {
		if i == 0 {
			for j := 0; j < len(line); j++ {
				cols[line[j]] = j
				// fmt.Printf("col %d = %s\n", j, lines[0][j])
			}
			continue // we've done all we need to do with lines[0]
		}

		var rec LinguisticDataRecord
		rec.Date, err = util.StringToDate(line[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for columnName, index := range cols {
			if len(line[index]) == 0 {
				continue
			}
			// Assuming all fields are float64 as per your struct
			value, err := strconv.ParseFloat(line[index], 64)
			switch columnName {
			case "LALLLSNScore":
				rec.LALLLSNScore = value
			case "LALLLSPScore":
				rec.LALLLSPScore = value
			case "LALLWHAScore":
				rec.LALLWHAScore = value
			case "LALLWHOScore":
				rec.LALLWHOScore = value
			case "LALLWHLScore":
				rec.LALLWHLScore = value
			case "LALLWPAScore":
				rec.LALLWPAScore = value
			case "LALLWDECount":
				rec.LALLWDECount = value
			case "LALLWDFCount":
				rec.LALLWDFCount = value
			case "LALLWDPCount":
				rec.LALLWDPCount = value
			case "LALLWDMCount":
				rec.LALLWDMCount = value
			case "LUSDLSNScore_ECON":
				rec.LUSDLSNScore_ECON = value
			case "LUSDLSPScore_ECON":
				rec.LUSDLSPScore_ECON = value
			case "LUSDWHAScore_ECON":
				rec.LUSDWHAScore_ECON = value
			case "LUSDWHOScore_ECON":
				rec.LUSDWHOScore_ECON = value
			case "LUSDWHLScore_ECON":
				rec.LUSDWHLScore_ECON = value
			case "LUSDWPAScore_ECON":
				rec.LUSDWPAScore_ECON = value
			case "LUSDWDECount_ECON":
				rec.LUSDWDECount_ECON = value
			case "LUSDWDFCount_ECON":
				rec.LUSDWDFCount_ECON = value
			case "LUSDWDPCount_ECON":
				rec.LUSDWDPCount_ECON = value
			case "LUSDLIMCount_ECON":
				rec.LUSDLIMCount_ECON = value
			case "LJPYLSNScore_ECON":
				rec.LJPYLSNScore_ECON = value
			case "LJPYLSPScore_ECON":
				rec.LJPYLSPScore_ECON = value
			case "LJPYWHAScore_ECON":
				rec.LJPYWHAScore_ECON = value
			case "LJPYWHOScore_ECON":
				rec.LJPYWHOScore_ECON = value
			case "LJPYWHLScore_ECON":
				rec.LJPYWHLScore_ECON = value
			case "LJPYWPAScore_ECON":
				rec.LJPYWPAScore_ECON = value
			case "LJPYWDECount_ECON":
				rec.LJPYWDECount_ECON = value
			case "LJPYWDFCount_ECON":
				rec.LJPYWDFCount_ECON = value
			case "LJPYWDPCount_ECON":
				rec.LJPYWDPCount_ECON = value
			case "LJPYLIMCount_ECON":
				rec.LJPYLIMCount_ECON = value
			case "WTOILClose":
				rec.WTOILClose = value
			default:
				err = nil // in this code, we should only be getting float64s.  If there was an error on a column we don't care about, ignore it
			}
			if err != nil {
				fmt.Printf("Data error in csv file, col = %s.  Err = %s\n", columnName, err.Error())
			}
		}
		records = append(records, rec)
	}

	d.LRecs = records
	return nil
}

/*
// getNamedFloat - centralize a bunch of lines that would need to be
//
//	repeated for every column of data without this func.
//
// INPUTS
//
//	val = name of data column excluding C1C2
//	line = array of strings -- parsed csv input line
//
// RETURNS
// float64 = the ratio if it exists, value is only valid if bool is true
// uint64  = a flag in bitpos - if 1 it means that the value is valid, 0
//
//	means the value was not supplied.
//
// --------------------------------------------------------------------------
func (d *CSVDatasource) getNamedFloat(val string, line []string) float64 {

	// util.DPrintf("bitpos = %d, find %s val... ", bitpos, val)

	key, exists := d.CSVMap[val]
	if !exists || key < 0 {
		// util.DPrintf("failed! A\n")
		return 0
	}
	s := line[key]
	if s == "" {
		// util.DPrintf("failed! B\n")
		return 0
	}
	ratio, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Panicf("getNamedFloat: invalid value: %q, err = %s\n", val, err)
	}

	return ratio
}

func (d *CSVDatasource) validCurrencyPair(line string) bool {
	if len(line) < 6 {
		return false
	}
	myC1 := line[0:3]
	myC2 := line[3:6]
	return myC1 == d.cfg.C1 && myC2 == d.cfg.C2
}
*/

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
func (d *CSVDatasource) CSVDBFindRecord(dt time.Time) *EconometricsRecord {
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

// CSVDBFindLRecord returns the lingustics record associated with the input date
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
func (d *CSVDatasource) CSVDBFindLRecord(dt time.Time) *LinguisticDataRecord {
	left := 0
	right := len(d.DBRecs) - 1

	for left <= right {
		mid := left + (right-left)/2
		if d.DBRecs[mid].Date.Year() == dt.Year() && d.DBRecs[mid].Date.Month() == dt.Month() && d.DBRecs[mid].Date.Day() == dt.Day() {
			return &d.LRecs[mid]
		} else if d.DBRecs[mid].Date.Before(dt) {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return nil
}
