package core

import (
	"fmt"

	"github.com/stmansour/psim/util"
)

// ValidateConfig ensures that all the configuration file numbers are valid, that no
//
//	constraints are violated. If it finds problems it will print them out and
//	return an error. Otherwise, it will return nil.
//
// ---------------------------------------------------------------------------------------
func ValidateConfig(cfg *util.AppConfig) error {
	var err error
	err = nil // assume everything is fine

	//-----------------------------------------------------------------------------------
	// Validate Influencer research time.
	//
	//  T3 = current date, the date on which the Investor is checking to see whether
	//       or not it should use some of its remaining C1 to purchase C2.
	//  T1 = Start time for Influencer research -- expressed as a negative number, days prior to T3
	//  T2 = Stop time for Influencer research -- expressed as a negative number, days prior to T3
	//  T4 = Date to convert C2 back to C1 if a "buy" was performed -- expressed as number of days after T3
	//
	//      T1                     T2          T3         T4
	//    ---+----------------------+------------+----------+
	//       |                      |<- Delta2 ->|<-Delta4->|
	//       |<-----------  Delta1  ------------>|
	//-----------------------------------------------------------------------------------
	if cfg.MinDelta2 >= 0 {
		err = fmt.Errorf("MinDelta2 (%d) must be less than 0", cfg.MinDelta2)
		fmt.Printf("** Configuration Error **  %s\n", err.Error())
	}
	if cfg.MaxDelta2 <= cfg.MaxDelta1 {
		err = fmt.Errorf("MaxDelta2 (%d) must be greater than MaxDelta1 (%d)", cfg.MaxDelta2, cfg.MaxDelta1)
		fmt.Printf("** Configuration Error **  %s\n", err.Error())
	}
	if cfg.MaxDelta1 < cfg.MinDelta1 {
		err = fmt.Errorf("MaxDelta1 (%d) must be less than MinDelta1 (%d)", cfg.MaxDelta1, cfg.MinDelta1)
		fmt.Printf("** Configuration Error **  %s\n", err.Error())
	}
	if cfg.MaxDelta2 < cfg.MinDelta2 {
		err = fmt.Errorf("MaxDelta2 (%d) must be less than MinDelta2 (%d)", cfg.MaxDelta2, cfg.MinDelta2)
		fmt.Printf("** Configuration Error **  %s\n", err.Error())
	}
	if cfg.MinDelta4 < 1 {
		err = fmt.Errorf("MinDelta4 (%d) must be > 0", cfg.MinDelta4)
		fmt.Printf("** Configuration Error **  %s\n", err.Error())
	}
	if cfg.MinDelta4 >= cfg.MaxDelta4 {
		err = fmt.Errorf("MinDelta4 (%d) must be less than MaxDelta4 (%d)", cfg.MinDelta4, cfg.MaxDelta4)
		fmt.Printf("** Configuration Error **  %s\n", err.Error())
	}

	//-------------------------------------------------
	// Standard Investment should be <= InitFunds/3
	//-------------------------------------------------
	if cfg.InitFunds/3.0 < cfg.StdInvestment {
		err = fmt.Errorf("StdInvestment (%6.2f) cannot be greater than 1/3 of the InitFunds (%6.2f / 3 = %6.2f)", cfg.StdInvestment, cfg.InitFunds, float64(cfg.InitFunds/3.0))
		fmt.Printf("** Configuration Error **  %s\n", err.Error())
	}

	return err
}
