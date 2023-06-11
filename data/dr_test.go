package data

import (
	"testing"
	"time"

	"github.com/stmansour/psim/util"
)

// func TestCVSDBFind(t *testing.T) {
// 	util.Init()
// 	cfg := util.CreateTestingCFG()
// 	Init(util.CreateTestingCFG())

// 	dt := time.Time(cfg.DtStart)
// 	dtStop := time.Time(cfg.DtStop)

// 	for dtStop.After(dt) || dtStop.Equal(dt) {
// 		r1 := CSVDBFindRecord(dt)
// 		if r1 == nil {
// 			t.Fail()
// 			fmt.Printf("NIL:  dt = %s : r1 = nil\n", dt.Format("2006-Jan-02"))
// 			return
// 		}
// 		r2 := ERFindRecord(dt)
// 		if r2 == nil {
// 			t.Fail()
// 			fmt.Printf("NIL:  dt = %s : r2 = nil\n", dt.Format("2006-Jan-02"))
// 			return
// 		}
// 		if r1.EXClose != r2.EXClose {
// 			t.Fail()
// 			fmt.Printf("MISMATCH:  dt = %s : r1.EXClose = %8.4f, r2.EXClose = %8.4f\n", dt.Format("2006-Jan-02"), r1.EXClose, r2.EXClose)
// 			return
// 		}
// 		dt = dt.AddDate(0, 0, 1)
// 	}
// }

func TestPLATODBLoad(t *testing.T) {
	util.Init()
	Init(util.CreateTestingCFG())

	dt := time.Date(2020, 3, 15, 0, 0, 0, 0, time.UTC)
	drec := CSVDBFindRecord(dt)
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
