package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// GenerationDuration is the resulting structure of a parsed
// Generation Duration Spec.  It contains the duration of a
// generation in terms of how it was defined.
// ------------------------------------------------------------
type GenerationDuration struct {
	Years  int
	Months int
	Weeks  int
	Days   int
}

// ParseGenerationDuration parses a Generation Duration Spec.  The spec
// is a string contain [ number unit ] pairs.  The number can be any
// integer.  The unit can be one of 4 values:
//
//	Y = years
//	M = months
//	W = weeks
//	D = days
//
// There can only be one pair for a particular unit.  That is,
// "1 Y" is valid, "1 Y 2 Y" is not (use "3 Y" instead).  Some
// examples:  "1 Y 6 M" meaning 1 year, 6 months.
//
//	"18 M" 18 months, same as 1 year, 6 months.
//	"90 D" 90 days
//
// INPUTS
//
//	s = Generation Duration Spec string.
//
// RETURNS
//
//	pointer to the GenerationDuration struct created from the string
//	any error encountered.
//
// ------------------------------------------------------------
func ParseGenerationDuration(s string) (*GenerationDuration, error) {
	parts := strings.Fields(s)
	if len(parts)%2 != 0 {
		return nil, errors.New("invalid input format")
	}

	duration := &GenerationDuration{}
	seen := map[string]bool{}

	for i := 0; i < len(parts); i += 2 {
		value, err := strconv.Atoi(parts[i])
		if err != nil {
			return nil, fmt.Errorf("error parsing number: %v", err)
		}

		unit := parts[i+1]
		if seen[unit] {
			return nil, fmt.Errorf("unit %s is repeated", unit)
		}
		seen[unit] = true

		switch unit {
		case "Y":
			duration.Years = value
		case "M":
			duration.Months = value
		case "W":
			duration.Weeks = value
		case "D":
			duration.Days = value
		default:
			return nil, fmt.Errorf("unknown unit: %s", unit)
		}
	}

	return duration, nil
}
