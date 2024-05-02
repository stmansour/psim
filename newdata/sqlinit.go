package newdata

import (
	"database/sql"
	"fmt"
	"time"
)

// SQLInit performs initialization such as loading caches
func (p *DatabaseSQL) SQLInit() error {
	err := p.ParentDB.Mim.Init(p.ParentDB)
	if err != nil {
		return err
	}
	if err = p.LoadLocaleCache(); err != nil {
		return err
	}
	p.MetricIDCache = make(map[string]int, 10) // enough to get it started
	if err = p.GetMinMaxDates(); err != nil {
		return err
	}
	if err = p.LoadMetricsSourceCache(); err != nil {
		return err
	}
	if err = p.LoadMetricSourceMapFromSQL(); err != nil {
		return err
	}
	return nil
}

// GetMinMaxDates fetches the earliest and latest dates from the given table.
// ---------------------------------------------------------------------------
func (p *DatabaseSQL) GetMinMaxDates() (err error) {
	queryMin := "SELECT MIN(Date) FROM Metrics_0_2020"
	queryMax := "SELECT MAX(Date) FROM Metrics_0_2020"

	var dt sql.NullTime

	err = p.DB.QueryRow(queryMin).Scan(&dt)
	if err != nil {
		return fmt.Errorf("error getting minimum date: %w", err)
	}
	if dt.Valid {
		// Use minDate.Time
		p.DtStart = dt.Time
	} else {
		p.DtStart = time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	}

	err = p.DB.QueryRow(queryMax).Scan(&dt)
	if err != nil {
		return fmt.Errorf("error getting maximum date: %w", err)
	}
	if dt.Valid {
		// Use minDate.Time
		p.DtStop = dt.Time
	} else {
		p.DtStop = time.Date(2023, time.December, 31, 23, 59, 59, 0, time.UTC)
	}

	return nil
}
