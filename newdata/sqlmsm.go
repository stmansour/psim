package newdata

import (
	"fmt"
)

// MSMapping is a 3-tuple (MSID, MID, MetricName) that defines how a metric suppliere
type MSMapping struct {
	MSID       int    // metrics source
	MID        int    // metric
	MetricName string // use this string with the metric source api to get metric
}

// WriteMetricSourceMapToSQL writes a 3-tuple (MSID, MID, MetricName) to the database
func (p *DatabaseSQL) WriteMetricSourceMapToSQL(MSID, MID int, mapping string) error {
	query := `INSERT INTO MetricSourcesMapping (MSID, MID, MetricName) VALUES (?,?,?)`
	_, err := p.DB.Exec(query, MSID, MID, mapping)
	if err != nil {
		return fmt.Errorf("SaveMSMapping failed: %w", err)
	}
	return nil
}

// LoadMetricSourceMapFromSQL initializes an internal cache of metrics sources
// int the DatabaseSQL struct.
// ---------------------------------------------------------------------------
func (p *DatabaseSQL) LoadMetricSourceMapFromSQL() error {
	for i := 0; i < len(p.MetricSrcCache); i++ {
		sourceName := p.MetricSrcCache[i].Name // this name will be the key in the cache map

		// we get back one of these:
		// type MetricsSource struct {
		// 	MSID       int
		// 	LastUpdate time.Time
		// 	URL        string
		// 	Name       string
		// }
		//
		// type MetricSourceMap map[string]string
		ms := p.MetricSrcCache[i]        // search for all MetricSourcesMapping entries with this ms.MSID
		msm := make(MetricSourceMap, 10) // start off with 10 of them

		query := fmt.Sprintf("SELECT MSID, MID, MetricName FROM MetricSourcesMapping WHERE MSID = %d", ms.MSID)
		rows, err := p.DB.Query(query)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var m MSMapping
			var mis MInfluencerSubclass
			err := rows.Scan(&m.MSID, &m.MID, &m.MetricName) // the mapping from sourceName's m.MetricName to our internal m.MID
			if err != nil {
				return err
			}

			//-----------------------------------------------------------------------
			// search for this m.MID in parentDB.Mim to get the name of the metric
			//-----------------------------------------------------------------------
			for _, v := range p.ParentDB.Mim.MInfluencerSubclasses {
				if v.MID == m.MID {
					mis = v
				}
			}
			if mis.MID == 0 {
				return fmt.Errorf("could not find MID %d in parentDB.Mim.MInfluencerSubclasses", m.MID)
			}

			// now we have enough information to create a map between our internal MID (mis.Metric) and m.MetricName
			msm[mis.Metric] = m.MetricName
		}

		// Check for errors from iterating over rows
		if err = rows.Err(); err != nil {
			return err
		}
		p.ParentDB.MSMap[sourceName] = msm

	}

	return nil
}
