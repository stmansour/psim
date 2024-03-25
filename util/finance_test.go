package util_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stmansour/psim/util"
)

func TestAnnualizedReturn(t *testing.T) {
	// Define test cases
	testCases := []struct {
		startingValue float64
		endingValue   float64
		dtStart       time.Time
		dtEnd         time.Time
		expected      float64
	}{
		// Test case 1:  $1 gain in 1 day
		{
			startingValue: 100,
			endingValue:   101,
			dtStart:       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			dtEnd:         time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
			expected:      36.87754075120426,
		},
		// Test case 2: $20 gain in one month
		{
			startingValue: 100,
			endingValue:   120,
			dtStart:       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			dtEnd:         time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
			expected:      7.569073635307722,
		},
		// Test case 3: $50 gain in one year
		{
			startingValue: 100,
			endingValue:   150,
			dtStart:       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			dtEnd:         time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:      0.49875421093199135,
		},
	}

	// Loop through test cases
	for _, tc := range testCases {
		// Call AnnualizedReturn with test case inputs
		result, err := util.AnnualizedReturn(tc.startingValue, tc.endingValue, tc.dtStart, tc.dtEnd)
		fmt.Printf("result: %v\n", result)

		// Check if result matches expected output
		if err != nil {
			t.Errorf("AnnualizedReturn(%v, %v, %v, %v) returned an error: %s", tc.startingValue, tc.endingValue, tc.dtStart, tc.dtEnd, err.Error())
		} else if math.Abs(result-tc.expected) > 0.0001 {
			t.Errorf("AnnualizedReturn(%v, %v, %v, %v) = %v, expected %v", tc.startingValue, tc.endingValue, tc.dtStart, tc.dtEnd, result, tc.expected)
		}
	}
}
