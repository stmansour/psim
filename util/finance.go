package util

import (
	"fmt"
	"math"
	"time"
)

// AnnualizedReturn computes the annualized return on an investment.
// startingValue and endingValue are the initial and final values of the investment.
// startDate and endDate are of type time.Time representing the start and end dates of the investment period.
func AnnualizedReturn(startingValue, endingValue float64, startDate, endDate time.Time) (float64, error) {
	// Ensure the start date is before the end date
	if startDate.After(endDate) {
		return 0, fmt.Errorf("start date must be before end date")
	}

	duration := endDate.Sub(startDate).Hours() / 24 // Calculate the total duration in days
	years := duration / 365.25                      // Convert duration from days to years

	// Calculate the annualized return
	if years <= 0 {
		return 0, fmt.Errorf("investment period must be greater than 0")
	}
	annualizedReturn := math.Pow(endingValue/startingValue, 1/years) - 1

	return annualizedReturn, nil
}
