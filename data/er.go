package data

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"
)

// ExchangeRateRecord represents a single entry in the CSV file
type ExchangeRateRecord struct {
	Symbol string
	Date   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
}

// ExchangeRateRecords is a slice of ExchangeRateRecord structures
type ExchangeRateRecords []ExchangeRateRecord

func (r ExchangeRateRecords) Len() int {
	return len(r)
}

// Less is used to sort ExchangeRateRecords
func (r ExchangeRateRecords) Less(i, j int) bool {
	return r[i].Date.Before(r[j].Date)
}

// Swap is used to sort records
func (r ExchangeRateRecords) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

// ER is a structure of data used by all the ER routines
var ER struct {
	ERRecs ExchangeRateRecords
}

// FindExchangeRateRecord returns the record associated with the input date
//
// INPUTS
//
//	dt = date of record to return
//
// RETURNS
//
//	pointer to the record on the supplied date
//
// ---------------------------------------------------------------------------
func FindExchangeRateRecord(dt time.Time) *ExchangeRateRecord {
	// Perform a binary search to find the record with the specified dt
	index := sort.Search(len(ER.ERRecs), func(i int) bool {
		return ER.ERRecs[i].Date.After(dt) || ER.ERRecs[i].Date.Equal(dt)
	})
	if index == len(ER.ERRecs) || ER.ERRecs[index].Date.After(dt) {
		return nil
	}
	return &ER.ERRecs[index]
}

// ERInit initialize the ExchangeRate data subsystem
// ---------------------------------------------------------------------------
func ERInit() {
	var records ExchangeRateRecords

	// Open the CSV file
	file, err := os.Open("data/er.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Read the CSV file
	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Parse the rows into ExchangeRateRecord structures
	for i, row := range rows {
		if i == 0 {
			continue
		}
		date, err := time.Parse("1/2/2006", row[1])
		if err != nil {
			fmt.Println("Error parsing date:", err)
			return
		}
		open, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			fmt.Println("Error parsing open value:", err)
			return
		}
		high, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			fmt.Println("Error parsing high value:", err)
			return
		}
		low, err := strconv.ParseFloat(row[4], 64)
		if err != nil {
			fmt.Println("Error parsing low value:", err)
			return
		}
		close, err := strconv.ParseFloat(row[5], 64)
		if err != nil {
			fmt.Println("Error parsing close value:", err)
			return
		}
		records = append(records, ExchangeRateRecord{
			Symbol: row[0],
			Date:   date,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
		})
	}

	// Sort the records by date
	sort.Sort(records)
	ER.ERRecs = records
}

// func main() {
// 	ERInit()

// 	// Example usage of FindExchangeRateRecord
// 	record := FindExchangeRateRecord(time.Date(2018, 4, 10, 0, 0, 0, 0, time.UTC))
// 	if record != nil {
// 		fmt.Println("Record found:", record)
// 	} else {
// 		fmt.Println("Record not found.")
// 	}
// }
