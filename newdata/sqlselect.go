package newdata

import (
	"database/sql"
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

// ShardInfo defines the values necessary to read or write metrics to the correct metrics table
type ShardInfo struct {
	BucketNumber int
	Metric       string
	MID          int
	Currency     string
	LID          int
	Table        string
}

func (p *DatabaseSQL) getShardInfo(Date time.Time, k string) *ShardInfo {
	shard := ShardInfo{}
	shard.BucketNumber = p.GetMetricBucket(k)
	shard.Metric, shard.MID, shard.Currency, shard.LID = p.CSVKeyToSQL(k)
	decade := (Date.Year() / 10) * 10
	shard.Table = fmt.Sprintf("Metrics_%d_%d\n", shard.BucketNumber, decade)
	return &shard
}

// Insert does a sql insert of all the metrics in the supplied record
func (p *DatabaseSQL) Insert(rec *EconometricsRecord) error {
	for k, v := range rec.Fields {
		si := p.getShardInfo(rec.Date, k)
		m := MetricRecord{
			Date:        rec.Date,
			MID:         si.MID,
			LID:         si.LID,
			MetricValue: v,
		}

		query := fmt.Sprintf(`INSERT INTO %s (Date,MID,LID,MetricValue) VALUES (?,?,?,?)`, si.Table)
		_, err := p.DB.Exec(query, m.Date, m.MID, m.LID, m.MetricValue)
		if err != nil {
			return err
		}
	}
	return nil
}

// Select reads the requested fields from the sql database
// --------------------------------------------------------------------
func (p *DatabaseSQL) Select(dt time.Time, ss []string) (*EconometricsRecord, error) {
	rec := EconometricsRecord{
		Date: dt,
	}
	for _, v := range ss {
		si := p.getShardInfo(dt, v)
		var m MetricRecord

		// Prepare the query
		query := `SELECT MEID, Date, MID, LID, MetricValue FROM Metrics WHERE Date=? AND MID=? AND LID=? LIMIT 1`

		// Considering date precision in MySQL DATETIME(6), it's important to format the time in a supported layout
		// MySQL DATETIME(6) format: "2006-01-02 15:04:05.999999"
		dateStr := dt.Format("2006-01-02 15:04:05.999999")

		// Execute the query
		err := p.DB.QueryRow(query, dateStr, si.MID, si.LID).Scan(&m.MEID, &m.Date, &m.MID, &m.LID, &m.MetricValue)
		if err != nil {
			if err == sql.ErrNoRows {
				// Handle the case where no rows are returned
				return nil, fmt.Errorf("no record found matching the criteria")
			}
			// Handle other potential errors
			return nil, err
		}
		rec.Fields[si.Metric] = m.MetricValue
	}
	return &rec, nil
}
