package main

import (
	"fmt"
	"strings"
	"time"
)

// PML is a struct containing a metric and its handle for a service
type PML struct {
	Metric string
	Handle string
}

// BuildMetricLists builds a list of indicators and forex rates
func BuildMetricLists(startDate, stopDate time.Time) ([]PML, []PML, error) {
	ind := []PML{}
	fxrs := []PML{}

	//-----------------------------------------------------------------
	// In this program, we're interested in Trading Economics data.
	//-----------------------------------------------------------------
	supplier := ""
	sc := app.SQLDB.SQLDB.MetricSrcCache
	for i := 0; i < len(sc); i++ {
		if strings.Contains(strings.ToLower(sc[i].Name), "trading economics") {
			supplier = sc[i].Name
			break
		}
	}
	if supplier == "" {
		return nil, nil, fmt.Errorf("could not find Trading Economics supplier")
	}

	//-----------------------------------------------------------------
	// Build the lists of metrics:  indicators, forex, forex rates
	//-----------------------------------------------------------------
	msm := app.SQLDB.MSMap[supplier]
	for k, v := range msm {
		m := PML{Metric: k, Handle: v}
		if strings.Contains(v, ":") {
			fxrs = append(fxrs, m)
		} else {
			ind = append(ind, m)
		}
	}
	return ind, fxrs, nil
}
