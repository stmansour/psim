package main

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

	rec, err := data.DR.DRRecs.GetRecord(time.Date(2022, 2, 10, 0, 0, 0, 0, time.UTC))
	if err != nil {
		fmt.Println(err)
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
	// fmt.Printf("%#v\n", rec)

}
