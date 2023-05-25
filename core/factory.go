package core

import (
	"errors"
	"fmt"
	"math/rand"
	"psim/util"
	"strconv"
	"strings"
)

var InfluencerSubclasses = []string{
	"DRInfluencer",
	"URInfluencer",
	"IRInfluencer",
}

// Factory contains methods to create objects based on a DNA string

type Factory struct {
	cfg *util.AppConfig // system-wide configuration info
}

// Init - initializes the factory
//
// --------------------------------------------------------------------------------
func (f *Factory) Init(cfg *util.AppConfig) {
	f.cfg = cfg
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
func (f *Factory) NewInfluencer(DNA string) (Influencer, error) {
	subclassName, DNAmap, err := f.ParseDNA(DNA)
	if err != nil {
		return nil, err
	}

	Delta1, Delta2, Delta4, err := f.GenerateDeltas(DNAmap)
	if err != nil {
		return nil, err
	}

	switch subclassName {
	case "DRInfluencer":
		dri := DRInfluencer{
			Delta1: Delta1,
			Delta2: Delta2,
			Delta4: Delta4,
			// more fields here...
		}
		return &dri, nil
	//... other cases here
	default:
		return nil, errors.New("unknown subclass")
	}
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
func (f *Factory) ParseDNA(DNA string) (string, map[string]interface{}, error) {
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
// func (f *Factory) GenerateDeltas(DNA map[string]interface{}) (Delta1 int, Delta2 int, Delta4 int, err error) {
// 	var ok bool
// 	Delta1, ok = DNA["Delta1"].(int)
// 	if !ok {
// 		util.DPrintf("f.cfg.MaxDelta1 = %d, f.cfg.MinDelta1 = %d\n", f.cfg.MaxDelta1, f.cfg.MinDelta1)
// 		Delta1 = util.RandomInRange(f.cfg.MaxDelta1, f.cfg.MinDelta1)
// 	}
// 	Delta2, ok = DNA["Delta2"].(int)
// 	if !ok {
// 		Delta2 = util.RandomInRange(f.cfg.MaxDelta2, f.cfg.MinDelta2)
// 		for Delta2 <= Delta1 {
// 			Delta2 = util.RandomInRange(f.cfg.MaxDelta2, f.cfg.MinDelta2)
// 		}
// 	} else if Delta2 <= Delta1 || Delta2 > 0 {
// 		return 0, 0, 0, errors.New("invalid DNA: Delta2 <= Delta1 or Delta2 > 0")
// 	}
// 	Delta4, ok = DNA["Delta4"].(int)
// 	if !ok {
// 		Delta4 = util.RandomInRange(f.cfg.MaxDelta4, f.cfg.MinDelta4)
// 		for Delta4 >= Delta2 {
// 			Delta4 = util.RandomInRange(f.cfg.MaxDelta4, f.cfg.MinDelta4)
// 		}
// 	} else if Delta4 >= Delta2 {
// 		return 0, 0, 0, errors.New("invalid DNA: Delta4 >= Delta2")
// 	}
// 	return Delta1, Delta2, Delta4, nil
// }

func (f *Factory) GenerateDeltas(DNA map[string]interface{}) (Delta1 int, Delta2 int, Delta4 int, err error) {
	// Generate or validate Delta1
	if val, ok := DNA["Delta1"].(int); ok {

		if val >= f.cfg.MinDelta1 && val <= f.cfg.MaxDelta1 {
			Delta1 = val
		} else {
			return 0, 0, 0, fmt.Errorf("invalid Delta1 value: %d, it must be in the range %d to %d", val, f.cfg.MinDelta1, f.cfg.MaxDelta2)
		}
	} else {
		Delta1 = rand.Intn(27) - 30 // -30 to -4
	}

	// Generate or validate Delta2
	if val, ok := DNA["Delta2"].(int); ok {
		if val > Delta1 && val <= f.cfg.MaxDelta2 && val >= f.cfg.MinDelta2 {
			Delta2 = val
		} else {
			mn := f.cfg.MinDelta2 // assume the min value is the configured lower limit
			if Delta1 > mn {      // if this is true, then Delta1's range can overlap Delta2
				mn = Delta1
			}
			return 0, 0, 0, fmt.Errorf("invalid Delta2 value: %d, it must be in the range %d to %d", val, mn, f.cfg.MaxDelta2)
		}
	} else {
		for {
			Delta2 = util.RandomInRange(f.cfg.MinDelta2, f.cfg.MaxDelta2)
			if Delta2 > Delta1 {
				break // if Delta2 is after Delta1, we're done. Otherwise we just keep trying
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
		Delta4 = util.RandomInRange(f.cfg.MinDelta4, f.cfg.MaxDelta4)
	}

	return Delta1, Delta2, Delta4, nil
}
