package main

import (
	"fmt"
	"strings"
)

// CopyCsvMISubclassesToSQL copies the CSV data for MInfluencerSubclasses
// into the equivalent SQL table.
func CopyCsvMISubclassesToSQL() error {
	for _, v := range app.csvdb.Mim.MInfluencerSubclasses {
		if err := app.sqldb.InsertMInfluencer(&v); err != nil {
			return err
		}
	}
	return nil
}

// CopyMetricsSourceMapToSQL copies the CSV MSMap into the equivalent SQL table.
//
// ------------------------------------------------------------------------------
func CopyMetricsSourceMapToSQL() error {
	for k, ms := range app.csvdb.MSMap {
		if k == "Metric" {
			continue //
		}

		//-----------------------------------------------------------------
		// First thing to do is look up the MSID. This can be found
		// by matching the supplier name, k, with a name in the csvdb
		// MetricSrcCache.
		//-----------------------------------------------------------------
		MID := -1
		MSID := -1
		for i := 0; i < len(app.csvdb.CSVDB.MetricSrcCache); i++ {
			if strings.Contains(app.csvdb.CSVDB.MetricSrcCache[i].Name, k) {
				MSID = app.csvdb.CSVDB.MetricSrcCache[i].MSID
				break
			}
		}
		if MSID == -1 {
			return fmt.Errorf("could not find metric source: %s", k)
		}

		//--------------------------------------------------------------------
		// Now that we know the Metric Source (the company identified by MSID)
		// we can spin through ms and insert the Metric Source's codes for
		// the metrics in the MetricSourcesMapping table. The ms map is of the
		// form:     map[MetricName]MetricSourceMapping
		// example:  ms["Gold"] = "XAUUSD:CUR"
		//--------------------------------------------------------------------
		for metric, mapping := range ms {
			// find the MID for metric
			MID = app.sqldb.Mim.MInfluencerSubclasses[metric].MID

			// now we have all we need. Insert it
			if err := app.sqldb.SQLDB.WriteMetricSourceMapToSQL(MSID, MID, mapping); err != nil {
				return err
			}
		}
	}
	return nil
}
