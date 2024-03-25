package util

import (
	"fmt"
)

// ValidDBSources contains the valid configuration choices for database
var ValidDBSources = []string{"CSV", "SQL"}

// ValidateConfig ensures that all the configuration file numbers are valid, that no
//
//	constraints are violated. If it finds problems it will print them out and
//	return an error. Otherwise, it will return nil.
//
// ---------------------------------------------------------------------------------------
func ValidateConfig(cfg *AppConfig) error {
	if cfg.SingleInvestorMode && cfg.CrucibleMode {
		return fmt.Errorf("SingleInvestorMode and CrucibleMode cannot both be set to true in the same config file")
	}

	if cfg.MaxInfluencers == 0 && cfg.MinInfluencers == 0 {
		return fmt.Errorf("** Configuration Error **  MinInfluencers and MaxInfluencers are either missing or both set to 0")
	}

	//-------------------------------------------------
	// Standard Investment should be <= InitFunds/3
	//-------------------------------------------------
	if cfg.InitFunds/3.0 < cfg.StdInvestment {
		return fmt.Errorf("StdInvestment (%6.2f) cannot be greater than 1/3 of the InitFunds (%6.2f / 3 = %6.2f)", cfg.StdInvestment, cfg.InitFunds, float64(cfg.InitFunds/3.0))
	}

	//-------------------------------------------------
	// Ensure that mutation is in range 1 - 100
	//-------------------------------------------------
	if cfg.MutationRate < 1 || cfg.MutationRate > 100 {
		return fmt.Errorf("mutation rate must be in the range 1 - 100, current value is: %d", cfg.MutationRate)
	}

	//--------------------------------------------------------------------
	// Ensure that DBSource is one of {CSV | Database | OnlineService}
	//--------------------------------------------------------------------
	found := false
	for i := 0; i < len(ValidDBSources) && !found; i++ {
		if cfg.DBSource == ValidDBSources[i] {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("unrecognized DBSource: %s", cfg.DBSource)
	}
	return nil
}
