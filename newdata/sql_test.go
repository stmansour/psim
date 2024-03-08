package newdata

import (
	"os"
	"testing"
	"time"

	"github.com/stmansour/psim/util"
)

type AppTest struct {
	extres      *util.ExternalResources
	cfg         *util.AppConfig
	csvdb       *Database
	sqldb       *Database
	BucketCount int
}

func TestFieldSelectorBuilder(t *testing.T) {
	if os.Getenv("MYSQL_AVAILABLE") != "1" {
		t.Skip("MySQL not available, skipping this test")
	}

	var err error
	var app AppTest

	//----------------------------------------------------------------------
	// Now get any other info we need for the databases
	//----------------------------------------------------------------------
	app.extres, err = util.ReadExternalResources()
	if err != nil {
		t.Errorf("ReadExternalResources returned error: %s\n", err.Error())
		return
	}
	cfg, err := util.LoadConfig("")
	if err != nil {
		t.Errorf("failed to read config file: %v\n", err)
		return
	}
	app.cfg = &cfg

	//----------------------------------------------------------------------
	// open the CSV database from which we'll be pulling data
	//----------------------------------------------------------------------
	app.csvdb, err = NewDatabase("CSV", app.cfg, nil)
	if err != nil {
		t.Errorf("*** PANIC ERROR ***  NewDatabase returned error: %s\n", err)
		return
	}
	if err := app.csvdb.Open(); err != nil {
		t.Errorf("*** PANIC ERROR ***  db.Init returned error: %s\n", err)
		return
	}
	if err := app.csvdb.Init(); err != nil {
		t.Errorf("*** PANIC ERROR ***  db.Init returned error: %s\n", err)
		return
	}

	//---------------------------------------------------------------------
	// open the MySQL database
	//---------------------------------------------------------------------
	app.sqldb, err = NewDatabase("SQL", app.cfg, app.extres)
	if err != nil {
		t.Errorf("Error creating database: %s\n", err.Error())
		return
	}
	if err = app.sqldb.Open(); err != nil {
		t.Errorf("db.Open returned error: %s\n", err.Error())
		return
	}

	if err = app.sqldb.Init(); err != nil {
		t.Errorf("db.Init returned error: %s\n", err.Error())
		return
	}

	defer app.sqldb.SQLDB.DB.Close()

	fields := map[string]float64{
		"USDJPYEXClose": 0.1,
		"JPYSNSNScore":  0.1,
		"USDBP":         0.1,
		"SNSPScore":     0.1,
		"CC":            0.1,
	}

	rec := EconometricsRecord{
		Date:   time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
		Fields: fields,
	}

	// expected results
	res := []FieldSelector{
		{Metric: "EXClose", Locale: "USD", Locale2: "JPY"},
		{Metric: "SNSNScore", Locale: "JPY", Locale2: ""},
		{Metric: "BP", Locale: "USD", Locale2: ""},
		{Metric: "SNSPScore", Locale: "", Locale2: ""},
		{Metric: "CC", Locale: "", Locale2: ""},
	}

	f := app.sqldb.SQLDB.FieldSelectorsFromRecord(&rec)
	for i := 0; i < len(f); i++ {
		// Search for the metric in res...
		k := -1
		for j := 0; j < len(res); j++ {
			if res[j].Metric == f[i].Metric {
				k = j
				break
			}
		}
		if k < 0 {
			t.Errorf("could not find metric %s in res", f[i].Metric)
			continue
		}
		if res[k].Metric != f[i].Metric || res[k].Locale != f[i].Locale || res[k].Locale2 != f[i].Locale2 {
			t.Errorf("FieldSelector mismatch: res (expected): %v , found (f): %v", res[k], f[i])
		}
	}
}
