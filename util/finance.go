package util

import (
	"fmt"
	"math"
	"time"
)

// RoundTo rounds a number to a specified number of decimal places.
// This would be used to compare two numbers where the number of decimals
// is not exactly the same. If we're comparing the numbers we don't want
// minor rounding errors to affect the result.
//
// INPUTS:
// num = the number to round
// decimals = the number of decimal places to round to
//
// RETURN
// the rounded number
// -----------------------------------------------------------------------------
func RoundTo(num float64, decimals int) float64 {
	shift := math.Pow(10, float64(decimals))
	return math.Round(num*shift) / shift
}

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
		return 0, nil
	}
	annualizedReturn := math.Pow(endingValue/startingValue, 1/years) - 1

	return annualizedReturn, nil
}
