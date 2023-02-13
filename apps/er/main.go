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

	// Example usage of ERFindRecord
	record := data.ERFindRecord(time.Date(2018, 4, 10, 0, 0, 0, 0, time.UTC))
	if record == nil {
		fmt.Println("ExchangeRate Record not found.")
	}

	i = data.ER.ERRecs.Len()
	dt1 = data.ER.ERRecs[0].Date
	dt2 = data.ER.ERRecs[i-1].Date

	fmt.Printf("Exchange Rate Info:\n")
	fmt.Printf("   Records:\t%d\n", i)
	fmt.Printf("   Beginning:\t%s\n", dt1.Format("Jan 2, 2006"))
	fmt.Printf("   Ending:\t%s\n", dt2.Format("Jan 2, 2006"))
}
