package main

// This app is used to calculate the correctness probabilities for all the
// Discount Rate data.  Within the Influencer for DiscountRate there are
// 3 variables.  Based on a purchase date, t3, they are:
//
//		dt1 = number of days prior to t3 where data research begins.
//            As of this writing its range is [2,30].
//      dt2 = number of days prior to t3 where data research stops
//            As of this writing its range is [1,5].
//		dt4 = number of days after t3 when the purchased currency is sold.
//            As of this writing its range is [1,5].
//---------------------------------------------------------------------------
import (
	"fmt"
	"os"
	"psim/core"
	"psim/data"
	"time"
)

type probInfo struct {
	count   int64
	correct int64
	prob    float64
}

var app struct {
	dri     data.DRInfo
	eri     data.ERInfo
	probMap map[string]probInfo
}

func displayStats() (data.DRInfo, data.ERInfo) {
	drinfo := data.DRGetDataInfo()
	fmt.Printf("Discount Rate Info:\n")
	fmt.Printf("   Records:\t%d\n", data.DR.DRRecs.Len())
	fmt.Printf("   Beginning:\t%s\n", drinfo.DtStart.Format("Jan 2, 2006"))
	fmt.Printf("   Ending:\t%s\n", drinfo.DtStop.Format("Jan 2, 2006"))

	erinfo := data.ERGetDataInfo()
	fmt.Printf("Exchange Rate Info:\n")
	fmt.Printf("   Records:\t%d\n", data.ER.ERRecs.Len())
	fmt.Printf("   Beginning:\t%s\n", erinfo.DtStart.Format("Jan 2, 2006"))
	fmt.Printf("   Ending:\t%s\n", erinfo.DtStop.Format("Jan 2, 2006"))

	return drinfo, erinfo
}

func checkDR(t3 time.Time) {
	rec := data.DRFindRecord(t3)
	if rec == nil {
		fmt.Println("DiscountRate Record not found.")
		os.Exit(1)
	}

	if !rec.Date.Equal(t3) {
		fmt.Printf("date did not match!\n")
		os.Exit(1)
	}
	if rec.USJPDRRatio != -15.0 {
		fmt.Printf("USJPDRRatio did not match!  Read %7.4f, looking for: -15 \n", rec.USJPDRRatio)
		os.Exit(1)
	}
	if rec.USDiscountRate != 0.015 {
		fmt.Printf("USDiscountRate did not match!  Read %7.4f, looking for: 0.015 \n", rec.USDiscountRate)
		os.Exit(1)
	}
	if rec.JPDiscountRate != -0.001 {
		fmt.Printf("JPDiscountRate did not match!  Read %7.4f, looking for: -0.001 \n", rec.JPDiscountRate)
		os.Exit(1)
	}
}

func main() {
	app.probMap = map[string]probInfo{}

	data.Init()
	checkDR(time.Date(2018, 2, 14, 0, 0, 0, 0, time.UTC)) // Just make sure everything looks OK before starting...

	//-------------------------------------
	// Now set up the boundaries...
	//-------------------------------------
	drinfo, erinfo := displayStats()
	app.dri = drinfo
	app.eri = erinfo

	//--------------------------------------------------------------------------
	// We must insure that the date range for which we calculate probabilities
	// is such that data exists.  Now that we know how much data we have, make
	// any adjustments necessary.
	//--------------------------------------------------------------------------
	dtStart := erinfo.DtStart.AddDate(0, 0, -core.DR.T1min)
	dtStop := erinfo.DtStop.AddDate(0, 0, -core.DR.T4max-1)

	//--------------------------------------------------------------------------
	// Adjust these dates if the DR data does not yet exist...
	//--------------------------------------------------------------------------
	if dtStop.After(drinfo.DtStop) {
		dtStop = drinfo.DtStop
	}
	if drinfo.DtStart.After(dtStart) {
		dtStart = drinfo.DtStart
	}

	dtEnd := dtStop.AddDate(0, 0, 1)
	fmt.Printf("Probs from %s to %s\n", dtStart.Format("Jan 2, 2006"), dtEnd.Format("Jan 2, 2006"))

	fmt.Printf("t1,t2,t3,t4, dt1, dt2, dt4, dDRR, dERR, prediction, actual\n")
	for dt := dtStart; dtEnd.After(dt); dt = dt.AddDate(0, 0, 1) {
		fmt.Printf("%s\n", dt.Format("Jan 2, 2006"))
		generateProbabilities(dt)
	}

	for i := core.DR.T1min; i <= core.DR.T1max; i++ {
		for j := core.DR.T2min; j <= core.DR.T2max; j++ {
			for k := core.DR.T4min; k <= core.DR.T4max; k++ {
				s := fmt.Sprintf("%d,%d,%d", i, j, k)
				v := app.probMap[s]
				fmt.Printf("%s :  %d  ->  %7.4f\n", s, v.count, v.prob)
			}
		}
	}
}

// generateProbablilities is the routine used to generate the probabilities for
// the DiscountRate Influencer.
// --------------------------------------------------------------------------------
func generateProbabilities(t3 time.Time) {

	//--------------------------------------------------------
	// Start the simulation on core.DR.T1min days after t3
	//--------------------------------------------------------
	dtT1min := t3.AddDate(0, 0, core.DR.T1min)   // t1 starts here
	dtT1max := t3.AddDate(0, 0, 1+core.DR.T1max) // t1 goes up to but does not include this date

	dtT2min := t3.AddDate(0, 0, core.DR.T2min)   // t2 starts here
	dtT2max := t3.AddDate(0, 0, 1+core.DR.T2max) // t2 goes up to but does not include this date

	dtT4min := t3.AddDate(0, 0, core.DR.T4min)   // t2 starts here
	dtT4max := t3.AddDate(0, 0, 1+core.DR.T4max) // t2 goes up to but does not include this date

	t1a := core.DR.T1min // range for t1 relative to t3
	// t1b := core.DR.T1max

	// t2a := core.DR.T2min // range for t2 relative to t3
	// t2b := core.DR.T2max

	// t4a := core.DR.T4min // range for t4 relative to t3
	// t4b := core.DR.T4max

	dt2 := 0 // define the deltas for each of the moving dates
	dt4 := 0
	dt1 := t1a

	// fmt.Printf("t3 range:  [ %s , %s )\n", dtT1min.Format("Jan 2, 2006"), dtT1max.Format("Jan 2, 2006"))

	n := 0
	//------------------------------------------------------------------------------
	// t1 defines how far back to examine Exchange Rates prior to t3
	//------------------------------------------------------------------------------
	for t1 := dtT1min; t1.Before(dtT1max); t1 = t1.AddDate(0, 0, 1) {
		dt2 = core.DR.T2min // restart each time through
		//------------------------------------------------------------------------------
		// t2 defines how far back to examine Exchange Rates prior to t3
		//------------------------------------------------------------------------------
		for t2 := dtT2min; t2.Before(dtT2max); t2 = t2.AddDate(0, 0, 1) {
			if t2.After(t1.AddDate(0, 0, 1)) {
				dt4 = core.DR.T4min // restart each time throu
				for t4 := dtT4min; t4.Before(dtT4max); t4 = t4.AddDate(0, 0, 1) {
					//---------------------------------------------------------------------------
					// Determine dDRR = (DiscountRateRatio at t1) - (DiscountRateRatio at t2)
					//---------------------------------------------------------------------------
					rec1 := data.DRFindRecord(t1)
					if rec1 == nil {
						fmt.Printf("ExchangeRate Record for %s not found.\n", t1.Format("1/2/2006"))
						os.Exit(1)
					}
					rec2 := data.DRFindRecord(t2)
					if rec2 == nil {
						fmt.Printf("ExchangeRate Record for %s not found.\n", t2.Format("1/2/2006"))
						os.Exit(1)
					}
					dDRR := rec1.USJPDRRatio - rec2.USJPDRRatio

					//-------------------------------------------------------------------------------
					// Prediction formula (based on the change in DiscountRateRatios):
					//     dDRR > 0:   buy on t3, sell on t4
					//     dDRR <= 0:  take no action
					//-------------------------------------------------------------------------------
					prediction := "hold"
					if dDRR > 0 {
						prediction = "buy"
					}

					//-------------------------------------------------------------------------------
					// Determine deltaERR (dERR) = (ExchangeRateRatio at t1) - (ExchangeRateRatio at t2)
					//-------------------------------------------------------------------------------
					er1 := data.ERFindRecord(t1)
					if er1 == nil {
						fmt.Printf("ExchangeRate Record for %s not found.\n", t1.Format("1/2/2006"))
						os.Exit(1)
					}
					er2 := data.ERFindRecord(t2)
					if er2 == nil {
						fmt.Printf("ExchangeRate Record for %s not found.\n", t2.Format("1/2/2006"))
						os.Exit(1)
					}
					dERR := er1.Close - er2.Close

					//-------------------------------------------------------------------------------
					// Check to see if the prediction is correct. If dERR > 0 AND dDRR > 0 then
					// then the prediction was correct.
					//-------------------------------------------------------------------------------
					predictionResult := false
					if dERR > 0 && dDRR > 0 {
						predictionResult = true
					}

					addToProbabilities(dt1, dt2, dt4, predictionResult)

					fmt.Printf("%s,%s,%s,%s,%d,%d,%d,%6.2f,%6.2f,%s,%t\n",
						t1.Format("01/02/2006"), t2.Format("01/02/2006"),
						t3.Format("01/02/2006"), t4.Format("01/02/2006"),
						dt1, dt2, dt4,
						dDRR, dERR,
						prediction,
						predictionResult,
					)
					dt4++
					n++
				}
			}
			dt2++
		}
		dt1++
	}
	fmt.Printf("total possibilities: %d\n", n)
}

func addToProbabilities(i, j, k int, correct bool) {
	s := fmt.Sprintf("%d,%d,%d", i, j, k)
	v, ok := app.probMap[s]
	if ok {
		v.count++
		if correct {
			v.correct++
		}
		v.prob = float64(v.correct) / float64(v.count)
		app.probMap[s] = v
	} else {
		v := probInfo{
			count:   1,
			correct: 0,
			prob:    0.0,
		}
		if correct {
			v.correct = 1
		}
		v.prob = float64(v.correct) / float64(v.count)
		app.probMap[s] = v
	}
}
