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
	util.Init(-1)
	Init(util.CreateTestingCFG()) // loads platodb.csv

	dt := time.Date(2020, time.April, 15, 0, 0, 0, 0, time.UTC)
	drec := CSVDBFindRecord(dt)
	if nil == drec {
		t.Errorf("Nil dbrecord for date %s", dt.Format("2006-Jan-02"))
		return
	}
	// util.DPrintf("Date = %s -  DRR = %8.4f\n", drec.Date.Format("2006-Jan-02"), drec.Ratio)
	if drec.Date.Year() != 2020 || drec.Date.Day() != 15 || drec.Date.Month() != time.April {
		t.Errorf("Date expected = 2020-Apr-15, got %s", drec.Date.Format("2006-Jan-02"))
	}
	util.DPrintf("drec 4/15/2020 = %s\n", DRec2String(drec))

	// TODO - turn on this section of code when the new platodb.csv is nailed down

	//---------------------------------------------------------------------------
	// verify that data flags are working correctly. GD data stops on 1/1/2023
	// GDRatio is valid if FLAGS & DataFlags.GDRatioValid is != 0
	// GDRatio should be valid on 12/31/2022
	//---------------------------------------------------------------------------
	// dt = time.Date(2022, time.December, 31, 0, 0, 0, 0, time.UTC)
	// drec = CSVDBFindRecord(dt)
	// if drec.Date.Year() != 2022 || drec.Date.Month() != time.December || drec.Date.Day() != 31 {
	// 	t.Errorf("Date expected = 2022-Dec-31, got %s", drec.Date.Format("2006-Jan-02"))
	// }
	// if drec.FLAGS&DataFlags.GDRatioValid == 0 {
	// 	t.Errorf("GD data should be valid on 2022-Dec-31, but FLAGS & DataFlags.GDRatioValid = %d", drec.FLAGS&DataFlags.GDRatioValid)
	// }
	// util.DPrintf("drec 12/31/2022 = %s\n", DRec2String(drec))

	// //---------------------------------------------------------------------------
	// // GDRatio should be invalid on 1/1/2023
	// //---------------------------------------------------------------------------
	// dt = time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)
	// drec = CSVDBFindRecord(dt)
	// util.DPrintf("drec 1/1/2023 = %s\n", DRec2String(drec))
	// if drec.Date.Year() != 2023 || drec.Date.Month() != time.January || drec.Date.Day() != 1 {
	// 	t.Errorf("Date expected = 2023-Jan-01, got %s", drec.Date.Format("2006-Jan-02"))
	// }
	// if drec.FLAGS&DataFlags.GDRatioValid != 0 {
	// 	t.Errorf("GD data should NOT be valid on 2023-Jan-1, but FLAGS & DataFlags.GDRatioValid = %d", drec.FLAGS&1<<DataFlags.GDRatioValid)
	// }
}

func TestBadConfig(t *testing.T) {
	util.Init(-1)
	cfg := util.CreateTestingCFG()
	cfg.C1 = "XYZ"
	err := Init(cfg)
	if err == nil {
		t.Errorf("Data subsystem did not return failure on bad currency configuration.")
	}
}
