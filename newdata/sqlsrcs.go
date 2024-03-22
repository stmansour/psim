package newdata

import (
	"fmt"
	"log"
	"time"
)

// MetricsSource defines a site that we use as a source for metrics within the database.
// The LastUpdate value is assumed to applie to all values retrieved from this source.
// ---------------------------------------------------------------------------------------
type MetricsSource struct {
	MSID       int
	LastUpdate time.Time
	URL        string
	Name       string
}

// InsertMetricsSource inserts a new MetricsSource into the database
// --------------------------------------------------------------------------------
func (p *Database) InsertMetricsSource(m *MetricsSource) (int64, error) {
	switch p.Datatype {
	case "CSV":
		return 0, fmt.Errorf("this operation is net yet supported for CSV databases")
	case "SQL":
		return p.SQLDB.InsertMetricsSource(m)
	default:
		return 0, fmt.Errorf("unknown database type: %s", p.Datatype)
	}
}

// InsertMetricsSource inserts a new MetricsSource into the MetricsSources table.
func (p *DatabaseSQL) InsertMetricsSource(loc *MetricsSource) (int64, error) {
	now := time.Now()
	stmt, err := p.DB.Prepare("INSERT INTO MetricsSources(URL, Name, LastUpdate) VALUES(?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(loc.URL, loc.Name, now)
	if err != nil {
		return 0, err
	}

	// Get the last inserted ID
	lastInsertID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastInsertID, nil
}

// CopyCsvMetricsSourcesToSQL takes a slice of MetricsSource and inserts them into the database.
func (p *DatabaseSQL) CopyCsvMetricsSourcesToSQL(locations []MetricsSource) error {
	for _, loc := range locations {
		_, err := p.InsertMetricsSource(&loc)
		if err != nil {
			// Decide how to handle the error; either continue with the next or return the error
			// For this example, we'll log the error and continue with the next entry
			log.Printf("Error inserting MetricsSource: %v", err)
			continue
		}
	}
	return nil
}

// LoadMetricsSourceCache initializes an internal cache of metrics sources
// int the DatabaseSQL struct.
// ---------------------------------------------------------------------------
func (p *DatabaseSQL) LoadMetricsSourceCache() error {
	query := "SELECT MSID, LastUpdate, URL, Name FROM MetricsSources"
	rows, err := p.DB.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var metrics []MetricsSource

	for rows.Next() {
		var m MetricsSource
		err := rows.Scan(&m.MSID, &m.LastUpdate, &m.URL, &m.Name)
		if err != nil {
			return err
		}
		metrics = append(metrics, m)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		return err
	}
	p.MetricSrcCache = metrics

	return nil
}
