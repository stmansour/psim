package data

import (
	"testing"
	"time"

	"github.com/stmansour/psim/util"
)

func TestDRCSVLoad(t *testing.T) {
	util.Init()
	Init(util.CreateTestingCFG())

	dt := time.Date(2020, 3, 15, 0, 0, 0, 0, time.UTC)
	drec := DRFindRecord(dt)
	// fmt.Printf("Date = %s -  DRR = %8.4f\n", drec.Date.Format("2006-Jan-02"), drec.Ratio)
	if dt.Year() != 2020 || dt.Day() != 15 {
		t.Errorf("Date expected = 2020-Mar-15, got %s", drec.Date.Format("2006-Jan-02"))
	}
}

func TestBadConfig(t *testing.T) {
	util.Init()
	cfg := util.CreateTestingCFG()
	cfg.C1 = "XYZ"
	err := Init(cfg)
	if err == nil {
		t.Errorf("Data subsystem did not return failure on bad currency configuration.")
	}
}
