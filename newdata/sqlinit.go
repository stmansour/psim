package newdata

import (
	"fmt"
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
	return nil
}

// GetMinMaxDates fetches the earliest and latest dates from the given table.
// ---------------------------------------------------------------------------
func (p *DatabaseSQL) GetMinMaxDates() (err error) {
	queryMin := "SELECT MIN(Date) FROM Metrics_0_2020"
	queryMax := "SELECT MAX(Date) FROM Metrics_0_2020"

	err = p.DB.QueryRow(queryMin).Scan(&p.DtStart)
	if err != nil {
		return fmt.Errorf("error getting minimum date: %w", err)
	}

	err = p.DB.QueryRow(queryMax).Scan(&p.DtStop)
	if err != nil {
		return fmt.Errorf("error getting maximum date: %w", err)
	}

	return nil
}
