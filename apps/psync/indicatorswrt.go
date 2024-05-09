package main

import (
	"fmt"
	"strings"

	"github.com/stmansour/psim/newdata"
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

		//--------------------------------------------------------------
		// we need to find our internal metric name for this indicator
		// The service handle is indicators[i].Category... we need to
		// map this back to the internal metric name...  Unfortunately,
		// the mapping is the inverse of what we want. We'll need to
		// search the map...
		//--------------------------------------------------------------
		mtc := ""
		for k, v := range app.SQLDB.MSMap["Trading Economics"] {
			if v == indicators[i].Category {
				mtc = k
				break
			}
		}
		if mtc == "" {
			return fmt.Errorf("could not find internal metric name for %s", indicators[i].Category)
		}
		//-------------------------------------------------------
		// Set the field selector with the internal metric name
		//-------------------------------------------------------
		var f = newdata.FieldSelector{
			Metric: mtc,
		}
		//----------------------------------------------------
		// find the locale for this indicator... Look for the
		// country name in our locale cache
		//----------------------------------------------------
		country := strings.ToLower(indicators[i].Country)
		for k, v := range app.SQLDB.SQLDB.LocaleCache {
			if strings.ToLower(v.Country) == country {
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
		if app.Verbose {
			fmt.Printf("%s - %s (%s): %.4f", indicators[i].DateTime.Format("2006-01-02"), indicators[i].Category, indicators[i].Country, indicators[i].Value)
		}
		if len(rec.Fields) == 0 {
			//------------------------------------------------------
			// This is case 1. We do not have this data. Add it...
			//------------------------------------------------------
			if app.Verbose {
				fmt.Printf(" || SQL Record not found, adding\n")
			}
			fld := newdata.MetricInfo{
				Value: indicators[i].Value,
				MSID:  app.MSID,
			}
			flds := make(map[string]newdata.MetricInfo, 1)
			flds[fqmetric] = fld
			rec.Fields = flds
			rec.Date = indicators[i].DateTime
			if err = app.SQLDB.Insert(rec); err != nil {
				return err
			}
		} else if toleranceMiscompare(rec.Fields[fqmetric].Value, indicators[i].Value) {
			//------------------------------------------------------------
			// This is case 2. We have it, but it miscompares.
			//------------------------------------------------------------
			if app.Verbose {
				fmt.Printf(" || SQL Record miscompare\n")
			}
			fmt.Printf("*** MISCOMPARE - INDICATOR VALUE ***\n")
			fmt.Printf("    Rec:  Date = %s, metric = %s, rec.Fields[metric] = %v\n", rec.Date.Format("2006-01-02"), fqmetric, rec.Fields[fqmetric].Value)
			fmt.Printf("    API:  Date = %s, i = %d, fxrs[i].Close = %.2f\n", indicators[i].DateTime.Format("2006-01-02"), i, indicators[i].Value)
			if app.APIFixMiscompares {
				var newrec newdata.EconometricsRecord
				newrec.Date = rec.Date
				newfld := rec.Fields[fqmetric]     // start with the original information
				newfld.Value = indicators[i].Value // here's the new value
				newfld.MSID = app.MSID             // we're now updating the value from the Metric Source
				newrec.Fields = make(map[string]newdata.MetricInfo, 1)
				newrec.Fields[fqmetric] = newfld
				if err = app.SQLDB.Update(&newrec); err != nil {
					return err
				}
				app.Corrected++
			}
			app.Miscompared++
		} else {
			//-----------------------------------------------------------------------
			// This is case 3. We have it, and it compares. Update verified count...
			//-----------------------------------------------------------------------
			if app.Verbose {
				fmt.Printf(" || SQL Record values matched, verified\n")
			}
			app.Verified++
		}
	}
	return nil
}

// toleranceMiscompare looks at the difference between v1 and v2. If the
// difference is greater than the global app.Tolerance it returns true
// which indicates that the values MISCOMPARE.
// ------------------------------------------------------------------------------
func toleranceMiscompare(v1, v2 float64) bool {
	x := v1 - v2
	if x < 0 {
		x = -x
	}
	return x > app.Tolerance
}
