package util

import (
	"fmt"
)

// ValidDBSources contains the valid configuration choices for database
var ValidDBSources = []string{"CSV", "Database", "OnlineService"}

// ValidateConfig ensures that all the configuration file numbers are valid, that no
//
//	constraints are violated. If it finds problems it will print them out and
//	return an error. Otherwise, it will return nil.
//
// ---------------------------------------------------------------------------------------
func ValidateConfig(cfg *AppConfig) error {
	var err error
	err = nil // assume everything is fine.  It will be set if any error conditions are hit

	//-----------------------------------------------------------------------------------
	// Validate Influencer research time.
	//
	//  T3 = current date, the date on which the Investor is checking to see whether
	//       or not it should use some of its remaining C1 to purchase C2.
	//  T1 = Start time for Influencer research -- expressed as a negative number, days prior to T3
	//  T2 = Stop time for Influencer research -- expressed as a negative number, days prior to T3
	//  T4 = Date to convert C2 back to C1 if a "buy" was performed -- expressed as number of days after T3
	//
	//  MinDelta1      MaxDelta1  MinDelta2  MaxDelta2
	//   |                     |  |            |
	//   |  T1                 |  |T2          | T3         T4
	//   +---|-----------------+--+-|----------+-|----------|
	//       |                      |<- Delta2 ->|<-Delta4->|
	//       |<-----------  Delta1  ------------>|
	//-----------------------------------------------------------------------------------
	for k := range cfg.SCInfo {
		if cfg.SCInfo[k].MinDelta2 >= 0 {
			err = fmt.Errorf("MinDelta2 (%d) must be less than 0", cfg.SCInfo[k].MinDelta2)
			fmt.Printf("** Configuration Error **  %s\n", err)
		}
		if !(cfg.SCInfo[k].MaxDelta1 < cfg.SCInfo[k].MinDelta2) {
			err = fmt.Errorf("MaxDelta1 (%d) must be less than MinDelta2 (%d)", cfg.SCInfo[k].MaxDelta1, cfg.SCInfo[k].MinDelta2)
			fmt.Printf("** Configuration Error **  %s\n", err)
		}
		if !(cfg.SCInfo[k].MinDelta1 < cfg.SCInfo[k].MaxDelta1) {
			fmt.Printf("cfg[%s]:\n", k)
			fmt.Printf("MinDelta1 = %d, MaxDelta1 = %d\n", cfg.SCInfo[k].MinDelta1, cfg.SCInfo[k].MaxDelta1)
			err = fmt.Errorf("MaxDelta1 (%d) must be greater than MinDelta1 (%d)", cfg.SCInfo[k].MaxDelta1, cfg.SCInfo[k].MinDelta1)
			fmt.Printf("** Configuration Error **  %s\n", err)
		}
		if !(cfg.SCInfo[k].MinDelta2 < cfg.SCInfo[k].MaxDelta2) {
			fmt.Printf("cfg[%s]:\n", k)
			fmt.Printf("MinDelta2 = %d, MaxDelta2 = %d\n", cfg.SCInfo[k].MinDelta2, cfg.SCInfo[k].MaxDelta2)
			err = fmt.Errorf("MaxDelta2 (%d) must be greater than MinDelta2 (%d)", cfg.SCInfo[k].MaxDelta2, cfg.SCInfo[k].MinDelta2)
			fmt.Printf("** Configuration Error **  %s\n", err)
		}

		if cfg.MinDelta4 < 1 {
			err = fmt.Errorf("MinDelta4 (%d) must be > 0", cfg.MinDelta4)
			fmt.Printf("** Configuration Error **  %s\n", err)
		}
		if cfg.MinDelta4 >= cfg.MaxDelta4 {
			err = fmt.Errorf("MinDelta4 (%d) must be less than MaxDelta4 (%d)", cfg.MinDelta4, cfg.MaxDelta4)
			fmt.Printf("** Configuration Error **  %s\n", err)
		}
	}

	//-------------------------------------------------
	// Standard Investment should be <= InitFunds/3
	//-------------------------------------------------
	if cfg.InitFunds/3.0 < cfg.StdInvestment {
		err = fmt.Errorf("StdInvestment (%6.2f) cannot be greater than 1/3 of the InitFunds (%6.2f / 3 = %6.2f)", cfg.StdInvestment, cfg.InitFunds, float64(cfg.InitFunds/3.0))
		fmt.Printf("** Configuration Error **  %s\n", err)
	}

	//-------------------------------------------------
	// Weighting validation
	//-------------------------------------------------
	if cfg.DRW1+cfg.DRW2 != float64(1.0) {
		err = fmt.Errorf("DRW1 (%4.2f) plus DRW2 (%4.2f) must equal 1.0", cfg.DRW1, cfg.DRW2)
		fmt.Printf("** Configuration Error **  %s\n", err)
	}

	//-------------------------------------------------
	// Ensure that mutation is in range 1 - 100
	//-------------------------------------------------
	if cfg.MutationRate < 1 || cfg.MutationRate > 100 {
		err = fmt.Errorf("mutation rate must be in the range 1 - 100, current value is: %d", cfg.MutationRate)
		fmt.Printf("** Configuration Error **  %s\n", err)
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
		err = fmt.Errorf("unrecognized DBSource: %s", cfg.DBSource)
		fmt.Printf("** Configuration Error **  %s\n", err)
	}

	// Create a map for quick lookup of valid influencer subclasses
	validMap := make(map[string]bool)
	for _, subclass := range ValidInfluencerSubclasses {
		validMap[subclass] = true
	}

	// Validate the influencer subclasses
	for _, subclass := range cfg.InfluencerSubclasses {
		if !validMap[subclass] {
			err = fmt.Errorf("invalid Influencer subclass: %q", subclass)
			fmt.Printf("** Configuration Error **  %s\n", err)
		}
	}

	return err
}
