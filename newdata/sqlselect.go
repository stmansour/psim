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
	LID2        int
	MSID        int // metrics source
	MetricValue MetricInfo
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

// GetShardInfo is a function that populates a field selector with the information
// needed for the SQL code to read or write the metric in the SQL database.
// ------------------------------------------------------------------------------------
func (p *DatabaseSQL) GetShardInfo(Date time.Time, f *FieldSelector) {
	f.BucketNumber = p.GetMetricBucket(f.Metric)
	p.FieldSelectorToSQL(f)
	decade := (Date.Year() / 10) * 10
	f.Table = fmt.Sprintf("Metrics_%d_%d", f.BucketNumber, decade)
}

var unrecognized = map[string]bool{}

// Insert does a sql insert of all the metrics in the supplied record
func (p *DatabaseSQL) Insert(rec *EconometricsRecord) error {
	var err error
	noLocale := int(p.LocaleCache["NON"].LID)
	for k, v := range rec.Fields {
		var f FieldSelector
		p.FieldSelectorFromCSVColName(k, &f)
		p.GetShardInfo(rec.Date, &f)
		m := MetricRecord{
			Date:        rec.Date,
			MID:         f.MID,
			LID:         f.LID,
			LID2:        f.LID2,
			MSID:        3, // NOTE:  hardcode - csvfile
			MetricValue: v,
		}
		if f.MID == 0 {
			if _, ok := unrecognized[f.Metric]; !ok {
				fmt.Printf("Skipping all attempts to insert unrecognized metric %s\n", f.Metric)
				unrecognized[f.Metric] = true
			}
			continue
		}
		//--------------------------------------------------------------------
		// for every metric we write, if we have source info then we write it
		//--------------------------------------------------------------------
		if v.MSID > 0 {
			m.MSID = v.MSID
		}
		if f.LID2 != noLocale && f.MID == -1 {
			query := `INSERT INTO ExchangeRate (Date,LID,LID2,MSID,EXClose) VALUES (?,?,?,?,?)`
			if _, err = p.DB.Exec(query, m.Date, m.LID, m.LID2, m.MSID, m.MetricValue.Value); err != nil {
				return err
			}
		} else {
			query := fmt.Sprintf(`INSERT INTO %s (Date,MID,LID,MSID,MetricValue) VALUES (?,?,?,?,?)`, f.Table)
			if _, err = p.DB.Exec(query, m.Date, m.MID, m.LID, m.MSID, m.MetricValue.Value); err != nil {
				return err
			}
		}
		p.InsertCount++

	}
	return nil
}

// Select reads the requested fields from the sql database.
// If ss is nil or zero length then it uses all known metrics
// --------------------------------------------------------------------
func (p *DatabaseSQL) Select(dt time.Time, ss []FieldSelector) (*EconometricsRecord, error) {
	var err error
	rec := EconometricsRecord{
		Date:   dt,
		Fields: map[string]MetricInfo{},
	}

	for _, v := range ss {
		p.FieldSelectorFromCSVColName(v.Metric, &v)
		p.GetShardInfo(dt, &v)
		var m MetricRecord

		// Considering date precision in SQL DATETIME(6), it's important to format the time in a supported layout
		// SQL DATETIME(6) format: "2006-01-02 15:04:05.999999"
		dateStr := dt.Format("2006-01-02")

		if v.Metric == "EXClose" {
			query := `SELECT XID,Date,LID,LID2, EXClose FROM ExchangeRate WHERE Date=? AND LID=? AND LID2=? LIMIT 1`
			err = p.DB.QueryRow(query, dateStr, v.LID, v.LID2).Scan(&m.MEID, &m.Date, &m.LID, &m.LID2, &m.MetricValue.Value)
			if err != nil {
				if err == sql.ErrNoRows {
					continue // nothing to store in Fields
				}
				return nil, err
			}
			rec.Fields[v.FQMetric()] = m.MetricValue
		} else {
			// Prepare the query
			query := fmt.Sprintf(`SELECT MEID,Date,MID,LID,MSID, MetricValue FROM %s WHERE Date=? AND MID=? AND LID=? LIMIT 1`, v.Table)
			var nullint sql.NullInt64
			err = p.DB.QueryRow(query, dateStr, v.MID, v.LID).Scan(&m.MEID, &m.Date, &m.MID, &m.LID, &nullint, &m.MetricValue.Value)
			if err != nil {
				if err == sql.ErrNoRows {
					continue // nothing to store in Fields
				}
				return nil, err
			}
			if nullint.Valid {
				v.MSID = int(nullint.Int64)
				m.MetricValue.MSID = int(nullint.Int64)
			}
			rec.Fields[v.FQMetric()] = m.MetricValue
		}
	}
	return &rec, nil
}
