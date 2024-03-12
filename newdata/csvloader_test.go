package newdata

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stmansour/psim/util"
)

func TestCSVDataAccess(t *testing.T) {
	util.Init(-1)
	cfg := util.CreateTestingCFG()
	if err := util.ValidateConfig(cfg); err != nil {
		log.Panicf("*** PANIC ERROR ***  ValidateConfig returned error: %s\n", err)
	}
	d, err := NewDatabase("CSV", cfg, nil)
	if err != nil {
		t.Errorf("error creating database: %s", err.Error())
		os.Exit(1)
	}

	path, err := os.Getwd()
	if err != nil {
		t.Errorf("error getting current directory: %s", err.Error())
		os.Exit(1)
	}
	d.SetCSVFilename(path + "/data/platodb.csv") // This call is not actually necessary, but this is when you'd set the override filename if you need to
	err = d.Open()                               // opens the database. In the CSV case, loads it into memory
	if err != nil {
		t.Errorf("error opening database: %s", err.Error())
		os.Exit(1)
	}

	err = d.Init() // read it in
	if err != nil {
		t.Errorf("error initializing database: %s", err.Error())
		os.Exit(1)
	}

	dt := time.Date(2023, time.January, 15, 0, 0, 0, 0, time.UTC)
	// ss := []string{"USDSP", "JPYDR", "LSNScore_ECON"}
	ss := []FieldSelector{
		{Metric: "SP", Locale: "USD"},
		{Metric: "DR", Locale: "JPY"},
		{Metric: "LSNScore_ECON"},
	}
	p, err := d.Select(dt, ss)
	if err != nil {
		t.Errorf("error selecting record for date %s: %s", dt.Format("Jan 2, 2006"), err.Error())
	}
	fmt.Printf("%v\n", p)
}
