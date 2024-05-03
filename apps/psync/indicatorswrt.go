package main

import (
	"fmt"
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
			Metric: util.Stripchars(indicators[i].Category, " "), // not sure if this is a good mapping. It will work for now.
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

		//------------------------------------------------------------------------------
		// one of 3 things happens now:
		//    1. if this data was not found in the database, we insert it
		//    2. if the value in the database is different, we report the difference
		//    3. if it matches, we mark a successful validation of the value
		//------------------------------------------------------------------------------
		if len(rec.Fields) == 0 {
			//------------------------------------------------------
			// This is case 1. We do not have this data. Add it...
			//------------------------------------------------------
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
			//------------------------------------------------------------
			// This is case 2. We have it, but it miscompares.  Flag it!
			//------------------------------------------------------------
			fmt.Printf("*** MISCOMPARE - INDICATOR VALUE ***\n")
			fmt.Printf("    Rec:  Date = %s, metric = %s, rec.Fields[metric] = %v\n", rec.Date.Format("2006-01-02"), fqmetric, rec.Fields[fqmetric].Value)
			fmt.Printf("    API:  Date = %s, i = %d, fxrs[i].Close = %.2f\n", indicators[i].DateTime.Format("2006-01-02"), i, indicators[i].Value)
			app.Miscompared++
		} else {
			//-----------------------------------------------------------------------
			// This is case 3. We have it, and it compares. Update verified count...
			//-----------------------------------------------------------------------
			app.Verified++
		}
	}
	return nil
}
