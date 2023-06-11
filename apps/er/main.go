package main

// EXCHANGE RATE - data access - test program.
//---------------------------------------------------------------------------
import (
	"fmt"
	"log"
	"time"

	"github.com/stmansour/psim/data"
	"github.com/stmansour/psim/util"
)

func main() {
	util.Init()
	cfg, err := util.LoadConfig()
	if err != nil {
		log.Fatalf("could not load config file:  %s\n", err)
	}
	if err = data.Init(&cfg); err != nil {
		log.Fatalf("Error initilizing data subsystem: %s\n", err)
	}
	i := data.DR.DRRecs.Len()
	dt1 := data.DR.DtStart
	dt2 := data.DR.DtStop

	fmt.Printf("Discount Rate & Exchange Rate Info:\n")
	fmt.Printf("   Records:\t%d\n", i)
	fmt.Printf("   Beginning:\t%s\n", dt1.Format("Jan 2, 2006"))
	fmt.Printf("   Ending:\t%s\n", dt2.Format("Jan 2, 2006"))

	// Example usage of CSVDBFindRecord
	record := data.CSVDBFindRecord(time.Date(2018, 4, 10, 0, 0, 0, 0, time.UTC))
	if record == nil {
		fmt.Println("ExchangeRate Record not found.")
	}

	// i = data.ER.ERRecs.Len()
	// dt1 = data.ER.DtStart
	// dt2 = data.ER.DtStop

	// fmt.Printf("Exchange Rate Info:\n")
	// fmt.Printf("   Records:\t%d\n", i)
	// fmt.Printf("   Beginning:\t%s\n", dt1.Format("Jan 2, 2006"))
	// fmt.Printf("   Ending:\t%s\n", dt2.Format("Jan 2, 2006"))
}
