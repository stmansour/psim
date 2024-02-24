package util

import (
	"fmt"
	"math"
	"time"
)

// AnnualizedReturn computes the annualized return on an investment.
// INPUTS:
//
//	startingValue initial value of the investment
//	endingValue   final value of the investment
//	dtStart       investment period start date
//	dtEnd         investment period end date
//
// -----------------------------------------------------------------------------
func AnnualizedReturn(startingValue, endingValue float64, dtStart, dtEnd time.Time) (float64, error) {
	if dtStart.After(dtEnd) {
		return 0, fmt.Errorf("start date must be before end date")
	}

	duration := dtEnd.Sub(dtStart).Hours() / 24 // Calculate the total duration in days
	years := duration / 365.25                  // Convert duration from days to years

	if years <= 0 {
		return 0, fmt.Errorf("investment period must be greater than 0")
	}
	annualizedReturn := math.Pow(endingValue/startingValue, 1/years) - 1

	return annualizedReturn, nil
}
