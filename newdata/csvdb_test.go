package newdata_test

import (
	"testing"
	"time"

	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/util"
)

func TestOpenCSVDatabaseAndSelectRecord(t *testing.T) {
	util.Init(-1)
	cfg, err := util.LoadConfig("")
	if err != nil {
		t.Fatalf("failed to read config file: %v\n", err)
	}
	if err := util.ValidateConfig(cfg); err != nil {
		t.Fatalf("ValidateConfig returned error: %s", err)
	}

	d, err := newdata.NewDatabase("CSV", cfg, nil)
	if err != nil {
		t.Fatalf("Error creating CSV database: %s", err)
	}
	d.SetCSVFilename("data/platodb.csv")
	if err := d.Open(); err != nil {
		t.Fatalf("Error opening CSV database: %s", err)
	}

	if err := d.Init(); err != nil {
		t.Fatalf("Error initializing CSV database: %s", err)
	}

	dt := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	ss := []newdata.FieldSelector{
		{Metric: "SP", Locale: "USD"},
		{Metric: "DR", Locale: "JPY"},
		{Metric: "LSNScore_ECON"},
	}

	rec, err := d.Select(dt, ss)
	if err != nil {
		t.Fatalf("Error selecting record for date %s: %s", dt.Format("Jan 2, 2006"), err)
	}

	recString := newdata.DRec2String(rec) // Assuming DRec2String is a function to convert the record to a string
	t.Logf("Record for Jan 1, 2020: %s", recString)
}
