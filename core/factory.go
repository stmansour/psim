package core

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

var InfluencerSubclasses = []string{
	"DRInfluencer",
	"URInfluencer",
	"IRInfluencer",
}

// NewInfluencer creates and returns a new influencer of the specified subclass.
// It can initialize the objects either randomly or from supplied DNA
//
// INPUTS
//
//			    DNA - a string with configuration information.
//		          It will be parsed for its subclass, then it is used as the DNA.
//		          DNA is of the form:   {subclass,var1=val1,var2=val2,...}.
//				  If the subclass is found but no values are present, then it
//	           	  will be created and randomly generated values will be assigned
//
// RETURNS
//
//	a pointer to the Influencer object (*Influencer)
//	any error encountered
//
// --------------------------------------------------------------------------------
func NewInfluencer(DNA string) (Influencer, error) {
	subclass, values, err := ParseDNA(DNA)
	if err != nil {
		return nil, err
	}

	switch subclass {
	case "DRInfluencer":
		dri := DRInfluencer{}
		dri.Delta1, dri.Delta2, dri.Delta4, err = generateDeltas(values) // generates random values if none are provided in the DNA
		if err != nil {
			return nil, err
		}
		return &dri, nil
		// handle other subclasses...
	}

	return nil, fmt.Errorf("unknown subclass: %s", subclass)
}

// ParseDNA does what you think
//
// RETURNS
//
//		a string representing the subclass
//	 a map of interface{} values indexed by variable name
//		any error found, nil if no errors
//
// --------------------------------------------------------------------------------
func ParseDNA(DNA string) (string, map[string]interface{}, error) {
	values := make(map[string]interface{})
	DNA = strings.Trim(DNA, "{}")
	tokens := strings.Split(DNA, ",")
	for i, token := range tokens {
		if i == 0 {
			found := false
			for _, v := range InfluencerSubclasses {
				if v == token {
					found = true
					break
				}
			}
			if !found {
				return "", nil, fmt.Errorf("unknown subclass: %s", token)
			}
		} else {
			pair := strings.Split(token, "=")
			if len(pair) != 2 {
				return "", nil, fmt.Errorf("invalid variable assignment: %s", token)
			}
			if val, err := strconv.ParseInt(pair[1], 10, 64); err == nil {
				values[pair[0]] = int(val)
			} else if val, err := strconv.ParseFloat(pair[1], 64); err == nil {
				values[pair[0]] = float64(val)
			} else {
				values[pair[0]] = pair[1]
			}
		}
	}
	return tokens[0], values, nil
}

// generateDeltas creates values needed for Delta1, Delta2, and Delta4 based
// on what was supplied in the DNA string.
//
// The current conditions on these values are:
//
// Delta1:  must be in the range -30 to -4
// Delta2:  Delta1 < Delta2,  Delta2 < 0, Delta2 > -8
// Delta4:  1 <= Delta4 <= 14
//
// RETURNS
//
//	Delta1, Delta2, and Delta4
//
// --------------------------------------------------------------------------------
func generateDeltas(DNA map[string]interface{}) (Delta1 int, Delta2 int, Delta4 int, err error) {
	// Generate or validate Delta1
	if val, ok := DNA["Delta1"].(int); ok {
		if val >= -30 && val <= -4 {
			Delta1 = val
		} else {
			return 0, 0, 0, fmt.Errorf("invalid Delta1 value: %d, it must be in the range -30 to -4", val)
		}
	} else {
		Delta1 = rand.Intn(27) - 30 // -30 to -4
	}

	// Generate or validate Delta2
	if val, ok := DNA["Delta2"].(int); ok {
		if val > Delta1 && val < 0 && val > -8 {
			Delta2 = val
		} else {
			return 0, 0, 0, fmt.Errorf("invalid Delta2 value: %d, it must be less than 0, greater than Delta1, and greater than -8", val)
		}
	} else {
		for {
			Delta2 = rand.Intn(7) - 8 // -7 to -1
			if Delta2 > Delta1 {
				break
			}
		}
	}

	// Generate or validate Delta4
	if val, ok := DNA["Delta4"].(int); ok {
		if val >= 1 && val <= 14 {
			Delta4 = val
		} else {
			return 0, 0, 0, fmt.Errorf("invalid Delta4 value: %d, it must be in the range 1 to 14", val)
		}
	} else {
		Delta4 = rand.Intn(14) + 1 // 1 to 14
	}

	return Delta1, Delta2, Delta4, nil
}
