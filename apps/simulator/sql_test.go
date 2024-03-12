package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/util"
)

type AppTest struct {
	extres      *util.ExternalResources
	cfg         *util.AppConfig
	csvdb       *newdata.Database
	sqldb       *newdata.Database
	BucketCount int
}

func TestSQLFuncs(t *testing.T) {
	if os.Getenv("MYSQL_AVAILABLE") != "1" {
		t.Skip("SQL not available, skipping this test")
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
	app.cfg = cfg

	//----------------------------------------------------------------------
	// open the CSV database from which we'll be pulling data
	//----------------------------------------------------------------------
	app.csvdb, err = newdata.NewDatabase("CSV", app.cfg, nil)
	if err != nil {
		t.Errorf("*** PANIC ERROR ***  NewDatabase returned error: %s\n", err)
		return
	}
	if err := app.csvdb.Open(); err != nil {
		t.Errorf("*** PANIC ERROR ***  db.Init returned error: %s\n", err)
		return
	}
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Could not get current working directory: %s", err.Error())
	}
	app.csvdb.SetCSVFilename(dir + "/data/platodb.csv")
	if err := app.csvdb.Init(); err != nil {
		t.Errorf("*** PANIC ERROR ***  db.Init returned error: %s\n", err)
		return
	}

	//---------------------------------------------------------------------
	// open the SQL database
	//---------------------------------------------------------------------
	app.sqldb, err = newdata.NewDatabase("SQL", app.cfg, app.extres)
	if err != nil {
		t.Errorf("Error creating database: %s\n", err.Error())
		return
	}
	if err = app.sqldb.Open(); err != nil {
		t.Errorf("db.Open returned error: %s\n", err.Error())
		return
	}

	defer app.sqldb.SQLDB.DB.Close()

	if err = app.sqldb.DropDatabase(); err != nil {
		t.Errorf("DropDatabase returned error: %s\n", err.Error())
		return
	}
	app.sqldb.SQLDB.ParentDB = app.sqldb // we will need this even before we call Init()

	if err = app.sqldb.CreateDatabasePart1(); err != nil {
		t.Errorf("CreateDatabasePart1 returned error: %s\n", err.Error())
		return
	}

	//----------------------------------------------------------------------
	// We have a new sql database now. Tables are defined, but contain
	// no data at this point. First thing to do is populate the ancillary
	// SQL tables.
	//----------------------------------------------------------------------
	if err = PopulateLocales(&app); err != nil {
		t.Errorf("Error from PopulateLocales: %s\n", err.Error())
		return
	}
	// now we need to load the sqldb's locale cache. It's needed by MigrateTimeSeriesData
	if err = app.sqldb.SQLDB.LoadLocaleCache(); err != nil {
		t.Errorf("Error from LoadLocalCache: %s\n", err.Error())
		return
	}
	if err = CopyCsvMISubclassesToSQL(&app); err != nil {
		t.Errorf("Error from CopyCsvMISubclassesToSql: %s\n", err.Error())
		return
	}
	// now that the MISubclasses table has been loaded, we'll need to cache it for use in MigrateTimeSeriesData
	app.sqldb.Mim.ParentDB = app.sqldb
	if err = app.sqldb.Mim.LoadMInfluencerSubclasses(); err != nil {
		t.Errorf("Error from LoadMInfluencerSubclasses: %s\n", err.Error())
		return
	}

	// grab a record from the csv file, something that's fully populated...
	dt := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)
	fields := []newdata.FieldSelector{}      // an empty slice
	rec, err := app.csvdb.Select(dt, fields) // empty slice gets all fields
	if err != nil {
		return
	}

	i := 0
	for k := range rec.Fields {
		f := newdata.FieldSelector{}
		app.sqldb.SQLDB.FieldSelectorFromCSVColName(k, &f) // populate the components of a field selector from the CSV Column name
		fields = append(fields, f)
		i++
	}

	// Write it to the database...
	if err = app.sqldb.Insert(rec); err != nil {
		return
	}

	// select the fields we want (all of them)

	// now read it back and make sure we have the same values...
	rec1, err := app.sqldb.Select(dt, fields)
	if err != nil {
		return
	}

	// compare...
	count := 0
	if !rec1.Date.Equal(rec.Date) {
		fmt.Printf("Dates miscompare!\n")
		count++
	}
	for k, v := range rec1.Fields {
		if rec.Fields[k] != v {
			fmt.Printf("Metric Values for %s miscompare!\n", k)
			count++
		}
	}
	fmt.Printf("miscompares found: %d\n", count)
}

// CopyCsvMISubclassesToSQL copies the CSV data for MInfluencerSubclasses
// into the equivalent SQL table.
func CopyCsvMISubclassesToSQL(app *AppTest) error {
	for _, v := range app.csvdb.Mim.MInfluencerSubclasses {
		if err := app.sqldb.InsertMInfluencer(&v); err != nil {
			return err
		}
	}
	return nil
}

// PopulateLocales initilizes the Locale table in sql
func PopulateLocales(app *AppTest) error {
	// Slice of Locale structs with country, currency, and a simple description
	locales := []newdata.Locale{
		{Name: "NON", Currency: "NON", Description: "No locale association"},
		{Name: "USA", Currency: "USD", Description: "United States of America, US Dollar"},
		{Name: "JPN", Currency: "JPY", Description: "Japan, Japanese Yen"},
		{Name: "GBR", Currency: "GBP", Description: "United Kingdom, British Pound"},
		{Name: "EUR", Currency: "EUR", Description: "Eurozone, Euro"},
		{Name: "CAN", Currency: "CAD", Description: "Canada, Canadian Dollar"},
		{Name: "AUS", Currency: "AUD", Description: "Australia, Australian Dollar"},
		{Name: "CHN", Currency: "CNY", Description: "China, Chinese Yuan"},
		{Name: "IND", Currency: "INR", Description: "India, Indian Rupee"},
		{Name: "BRA", Currency: "BRL", Description: "Brazil, Brazilian Real"},
		{Name: "RUS", Currency: "RUB", Description: "Russia, Russian Ruble"},
		{Name: "ZAF", Currency: "ZAR", Description: "South Africa, South African Rand"},
		{Name: "SGP", Currency: "SGD", Description: "Singapore, Singapore Dollar"},
		{Name: "NZL", Currency: "NZD", Description: "New Zealand, New Zealand Dollar"},
		{Name: "MEX", Currency: "MXN", Description: "Mexico, Mexican Peso"},
		{Name: "IDN", Currency: "IDR", Description: "Indonesia, Indonesian Rupiah"},
		{Name: "TUR", Currency: "TRY", Description: "Turkey, Turkish Lira"},
		{Name: "SAU", Currency: "SAR", Description: "Saudi Arabia, Saudi Riyal"},
		{Name: "SWE", Currency: "SEK", Description: "Sweden, Swedish Krona"},
		{Name: "NOR", Currency: "NOK", Description: "Norway, Norwegian Krone"},
		{Name: "DNK", Currency: "DKK", Description: "Denmark, Danish Krone"},
		{Name: "POL", Currency: "PLN", Description: "Poland, Polish Zloty"},
		{Name: "HKG", Currency: "HKD", Description: "Hong Kong, Hong Kong Dollar"},
		{Name: "KOR", Currency: "KRW", Description: "South Korea, South Korean Won"},
		{Name: "THA", Currency: "THB", Description: "Thailand, Thai Baht"},
		{Name: "PHL", Currency: "PHP", Description: "Philippines, Philippine Peso"},
		{Name: "MYS", Currency: "MYR", Description: "Malaysia, Malaysian Ringgit"},
		{Name: "NGA", Currency: "NGN", Description: "Nigeria, Nigerian Naira"},
		{Name: "EGY", Currency: "EGP", Description: "Egypt, Egyptian Pound"},
		{Name: "ISR", Currency: "ILS", Description: "Israel, Israeli New Shekel"},
		{Name: "PAK", Currency: "PKR", Description: "Pakistan, Pakistani Rupee"},
	}

	// Iterate over the locales slice and insert each into the database
	for _, locale := range locales {
		lastInsertID, err := app.sqldb.InsertLocale(&locale)
		if err != nil {
			return err
		}
		fmt.Printf("Locale inserted with ID: %d\n", lastInsertID)
	}
	return nil
}
