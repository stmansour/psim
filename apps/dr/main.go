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
	"psim/data"
	"time"
)

func main() {
	data.Init()
	i := data.DR.DRRecs.Len()
	dt1 := data.DR.DRRecs[0].Date
	dt2 := data.DR.DRRecs[i-1].Date

	fmt.Printf("Discount Rate Info:\n")
	fmt.Printf("   Records:\t%d\n", i)
	fmt.Printf("   Beginning:\t%s\n", dt1.Format("Jan 2, 2006"))
	fmt.Printf("   Ending:\t%s\n", dt2.Format("Jan 2, 2006"))

	//------------------------------------------------------------
	// Just make sure everything looks OK before starting...
	//------------------------------------------------------------
	// rec, err := data.DR.DRRecs.GetRecord(time.Date(2022, 2, 10, 0, 0, 0, 0, time.UTC))
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	rec := data.FindDiscountRateRecord(time.Date(2022, 2, 10, 0, 0, 0, 0, time.UTC))
	if rec == nil {
		fmt.Println("ExchangeRate Record not found.")
		os.Exit(1)
	}

	if !rec.Date.Equal(time.Date(2022, 2, 10, 0, 0, 0, 0, time.UTC)) {
		fmt.Printf("date did not match!\n")
		os.Exit(1)
	}
	if rec.USJPDRRatio != -2.5 {
		fmt.Printf("USJPDRRatio did not match!\n")
		os.Exit(1)
	}
	if rec.USDiscountRate != 0.0025 {
		fmt.Printf("USDiscountRate did not match!\n")
		os.Exit(1)
	}
	if rec.JPDiscountRate != -0.001 {
		fmt.Printf("JPDiscountRate did not match!\n")
		os.Exit(1)
	}

	generateProbabilities()
}

func generateProbabilities() {

}
