package core

import (
	"fmt"
	"time"

	"github.com/stmansour/psim/data"
)

// RatioFunc is a type used by the GetPrediction method of Influencer subclasses.
// It returns the metric ratio for the subclass's specific metrics.
// -----------------------------------------------------------------------------------
type RatioFunc func(*data.RatesAndRatiosRecord, *data.RatesAndRatiosRecord) float64

// getPrediction - using the supplied date, it researches data and makes
// a prediction on whther to "buy" or "hold"
//
// RETURNS
//
//	action     -  "buy" or "hold"
//	prediction - probability of correctness - most valid for "buy" action
//	error      - nil on success, error encountered otherwise
//
// ---------------------------------------------------------------------------
func getPrediction(t3 time.Time, delta1 int, delta2 int, f RatioFunc) (string, float64, error) {
	t1 := t3.AddDate(0, 0, delta1)
	t2 := t3.AddDate(0, 0, delta2)

	rec1 := data.CSVDBFindRecord(t1)
	if rec1 == nil {
		err := fmt.Errorf("data.RatesAndRatiosRecord for %s not found", t1.Format("1/2/2006"))
		return "hold", 0, err
	}
	rec2 := data.CSVDBFindRecord(t2)
	if rec2 == nil {
		err := fmt.Errorf("data.RatesAndRatiosRecord for %s not found", t2.Format("1/2/2006"))
		return "hold", 0, err
	}

	dRR := f(rec1, rec2)

	prediction := "hold"
	if dRR > 0 {
		prediction = "buy"
	}

	// todo - return proper probability
	return prediction, 0.5, nil
}
