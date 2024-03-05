package main

import (
	"fmt"
	"time"
)

// MetricRecord defines the structure for entries into the metric tables
type MetricRecord struct {
	MEID        int
	Date        time.Time
	MID         int
	LID         int
	MetricValue float64
}

// MigrateTimeSeriesData migrates the CSV time-series data into sql tables
func MigrateTimeSeriesData() error {

	b := []map[string]float64{}
	for i := 0; i < app.sqldb.SQLDB.BucketCount; i++ {
		c := map[string]float64{}
		b = append(b, c)
	}

	// grab a record from the csv file, something that's fully populated...
	dt := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)
	fields := []string{}                     // an empty slice
	rec, err := app.csvdb.Select(dt, fields) // empty slice gets all fields
	if err != nil {
		return err
	}
	// sort the datavalues
	for k, v := range rec.Fields {
		bucketNumber := app.sqldb.SQLDB.GetMetricBucket(k)
		key, MID, locale, LID := app.sqldb.SQLDB.CSVKeyToSQL(k)
		decade := (rec.Date.Year() / 10) * 10
		fmt.Printf("%s:  metric = %s, %s, MID = %d, decade = %d, LID = %d, bucket = %d, val = %8.2f\n", k, key, locale, MID, decade, LID, bucketNumber, v)
		table := fmt.Sprintf("Metrics_%d_%d\n", bucketNumber, decade)
		fmt.Printf("     TABLE = %s\n", table)
		b[bucketNumber][k] = v // store the key and value in the appropriate bucket
		m := MetricRecord{
			Date:        rec.Date,
			MID:         MID,
			LID:         LID,
			MetricValue: v,
		}

		query := fmt.Sprintf(`INSERT INTO %s (Date,MID,LID,MetricValue) VALUES (?,?,?,?)`, table)
		_, err := app.sqldb.SQLDB.DB.Exec(query, m.Date, m.MID, m.LID, m.MetricValue)
		if err != nil {
			return err
		}
	}
	return nil
}
