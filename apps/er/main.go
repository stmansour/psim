package main

// EXCHANGE RATE - data access - test program.
//---------------------------------------------------------------------------
import (
	"fmt"
	"time"

	"github.com/stmansour/psim/data"
)

func main() {
	data.Init()
	i := data.DR.DRRecs.Len()
	dt1 := data.DR.DtStart
	dt2 := data.DR.DtStop

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
	dt1 = data.ER.DtStart
	dt2 = data.ER.DtStop

	fmt.Printf("Exchange Rate Info:\n")
	fmt.Printf("   Records:\t%d\n", i)
	fmt.Printf("   Beginning:\t%s\n", dt1.Format("Jan 2, 2006"))
	fmt.Printf("   Ending:\t%s\n", dt2.Format("Jan 2, 2006"))
}
