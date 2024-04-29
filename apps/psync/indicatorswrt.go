package main

import (
	"log"
	"strings"

	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/util"
)

// TEIndcatorInfo contains information about the indicators
type TEIndcatorInfo struct {
	APIName     string
	DBFieldName string
	Frequency   string
}

// TRIndicators contains information about the indicators
var TRIndicators = []TEIndcatorInfo{
	{APIName: "Housing Starts MoM", DBFieldName: "HS", Frequency: "Monthly"},
	{APIName: "Inflation Rate MoM", DBFieldName: "IR", Frequency: "Monthly"},
}

// UpdateIndicators updates the indicators in the SQL database
func UpdateIndicators(indicators []Indicator) error {
	for i := 0; i < len(indicators); i++ {
		fields := []newdata.FieldSelector{}
		var f = newdata.FieldSelector{
			Metric: util.Stripchars(indicators[i].Category, " "), // not sure if this is a good mapping.
		}
		//----------------------------------------------------
		// find the locale for this indicator... Look for the
		// country name in our locale cache
		//----------------------------------------------------
		country := strings.ToLower(indicators[i].Country)
		loc := newdata.Locale{}
		for k, v := range app.SQLDB.SQLDB.LocaleCache {
			if strings.ToLower(v.Country) == country {
				loc = v
				f.Locale = k
				break
			}
		}
		fields = append(fields, f)

		//---------------------------------------------------------
		// read the indicator for the indicator's date to see if
		// we have any data for it.
		//---------------------------------------------------------
		rec, err := app.SQLDB.Select(indicators[i].DateTime, fields)
		if err != nil {
			return err
		}
		fqmetric := f.FQMetric()
		if len(rec.Fields) == 0 {
			// currently we do not have data for this indicator.  Add it...
			fld := newdata.MetricInfo{
				Value: indicators[i].Value,
				MSID:  app.MSID,
			}
			flds := make(map[string]newdata.MetricInfo, 1)
			metric := util.Stripchars(indicators[i].Category, " ")
			flds[loc.Currency+metric] = fld
			rec.Fields = flds
			rec.Date = indicators[i].DateTime
			if err = app.SQLDB.Insert(rec); err != nil {
				return err
			}
		} else if rec.Fields[fqmetric].Value != indicators[i].Value {
			// currently we have data for this indicator. But the values miscompare
			log.Printf("Current record for: %s:%s:%s has miscompare: %.2f != %.2f\n",
				indicators[i].Category, indicators[i].Country, indicators[i].DateTime.Format("2006-01-02"),
				rec.Fields[fqmetric].Value, indicators[i].Value)
		}

	}

	return nil
}
