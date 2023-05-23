package core

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var InfluencerSubclasses = []string{"DRInfluencer", "URInfluencer", "IRInfluencer"}

// NewInfluencer creates and returns a new influencer of the specified subclass.
// It can initialize the objects either randomly or from supplied DNA
//
// INPUTS
//
//		    DNA - a string with configuration information.
//	          It will be parsed for its subclass.
//	          successfully parsed for a subclass, then it is used as the DNA.
//	          DNA is of the form:   {subclass,val1,val2,...}.  If the subclass
//	          is found but no values are present, then it will be crate
//
// RETURNS
//
//	a pointer to the Influencer object (*Influencer)
//	any error encountered
//
// --------------------------------------------------------------------------------
func NewInfluencer(DNA string) (Influencer, error) {
	parsedDNA, err := ParseDNA(DNA)
	if err != nil {
		return nil, err
	}
	rand.Seed(time.Now().UnixNano())

	var Delta1, Delta2, Delta4 int
	if len(parsedDNA) > 1 {
		// If DNA provides deltas, use them
		Delta1 = parsedDNA[1].(int)
		Delta2 = parsedDNA[2].(int)
		Delta4 = parsedDNA[3].(int)
	} else {
		// Otherwise, generate random deltas
		Delta1 = rand.Intn(27) - 30
		Delta2 = rand.Intn(7) - 7
		Delta4 = rand.Intn(14) + 1

		// Ensure Delta1 < Delta2
		for Delta1 >= Delta2 {
			Delta1 = rand.Intn(27) - 30
			Delta2 = rand.Intn(7) - 7
		}
	}

	// Create the appropriate influencer based on the subclass
	switch parsedDNA[0].(string) {
	case "DRInfluencer":
		dri := DRInfluencer{
			Delta1: Delta1,
			Delta2: Delta2,
			Delta4: Delta4,
		}
		return &dri, nil
	case "URInfluencer":
		// Similarly create and return a URInfluencer
	case "IRInfluencer":
		// Similarly create and return an IRInfluencer
	}

	return nil, fmt.Errorf("unknown influencer subclass")
}

// ParseDNA does what you think
//
// RETURNS
//
//	a slice of interface{} values, strings and ints at this point
//	any error found, nil if no errors
//
// --------------------------------------------------------------------------------
func ParseDNA(DNA string) ([]interface{}, error) {
	parts := strings.Split(strings.Trim(DNA, "{}"), ",")

	if len(parts) == 0 {
		return nil, fmt.Errorf("no data provided")
	}

	// Check if the first element is a valid subclass
	isValidSubclass := false
	for _, subclass := range InfluencerSubclasses {
		if parts[0] == subclass {
			isValidSubclass = true
			break
		}
	}

	if !isValidSubclass {
		return nil, fmt.Errorf("first element is not a valid subclass")
	}

	result := make([]interface{}, len(parts))
	for i, part := range parts {
		// subclass?
		if i == 0 {
			result[i] = part
			continue
		}
		// integer?
		if val, err := strconv.Atoi(part); err == nil {
			result[i] = val
			continue
		}
		// float?
		if val, err := strconv.ParseFloat(part, 64); err == nil {
			result[i] = val
			continue
		}
		result[i] = part // gotta be a string
	}

	return result, nil
}
