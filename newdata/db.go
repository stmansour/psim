package newdata

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/stmansour/psim/util"
)

// EconometricsRecords is a type for an array of DR records
type EconometricsRecords []EconometricsRecord

// Database is the abstraction for the data source
type Database struct {
	cfg      *util.AppConfig          // application configuration info
	extres   *util.ExternalResources  // the db may require secrets
	Datatype string                   // "CSV", "MYSQL"
	CSVDB    *DatabaseCSV             // valid when Datatype is "CSV"
	SQLDB    *DatabaseSQL             // valid when Datatype is "SQL"
	Mim      *MetricInfluencerManager // metrics manager
}

// EconometricsRecord is the basic structure of discount rate data
type EconometricsRecord struct {
	Date   time.Time
	Fields map[string]float64
}

// GetMetricBucket calculates or retrieves from MetricIDCache the bucket
// for the supplied metric name
// ------------------------------------------------------------------
func (p *DatabaseSQL) GetMetricBucket(s string) int {
	// Get it from the MetricIDCache if possible
	if bucketNumber, found := p.MetricIDCache[s]; found {
		return bucketNumber
	}

	//-----------------------------------------------
	// If not in MetricIDCache, calculate the bucket number
	// and store it in the hash
	//-----------------------------------------------
	hash := sha256.Sum256([]byte(s))
	hashInt := 0
	for _, b := range hash[:] {
		hashInt += int(b)
	}
	bucketNumber := hashInt % p.BucketCount
	p.MetricIDCache[s] = bucketNumber

	return bucketNumber
}

// NewDatabase creates a new database structure
// dtype: "CSV", "MYSQL"
// ------------------------------------------------------------
func NewDatabase(dtype string, cfg *util.AppConfig, ex *util.ExternalResources) (*Database, error) {
	switch dtype {
	case "CSV":
		db := Database{
			cfg:      cfg,
			Datatype: "CSV",
		}
		db.CSVDB = &DatabaseCSV{}
		return &db, nil

	case "SQL":
		db := Database{
			cfg:      cfg,
			Datatype: "SQL",
			extres:   ex,
		}
		db.SQLDB = &DatabaseSQL{
			Name:        "plato",
			BucketCount: 4, // we will adjust as needed
		}
		db.SQLDB.MetricIDCache = make(map[string]int)
		return &db, nil

	default:
		return nil, fmt.Errorf("unrecognized database type: %s", dtype)
	}
}

// Select reads and returns data from the database.
// ----------------------------------------------------------------------------
func (p *Database) Select(dt time.Time, fields []string) (*EconometricsRecord, error) {
	var err error
	switch p.Datatype {
	case "CSV":
		return p.CSVDB.Select(dt, fields)
	case "SQL":
		return p.SQLDB.Select(dt, fields)
	default:
		err = fmt.Errorf("unrecognized data source: %s", p.cfg.DBSource)
		return nil, err
	}
}

// DropDatabase deletes the sql database.  Use this with caution
// ---------------------------------------------------------------------------------
func (p *Database) DropDatabase() error {
	if p.Datatype == "CSV" {
		return nil
	}
	if p.Datatype != "SQL" {
		return fmt.Errorf("unknown database type: %s", p.Datatype)
	}
	// Construct DSN for the initial MySQL connection without specifying a database
	host := "tcp(127.0.0.1:3306)"
	dsnWithoutDB := fmt.Sprintf("%s:%s@%s/", p.extres.DbUser, p.extres.DbPass, host)

	// Connect to MySQL without specifying a database
	db, err := sql.Open("mysql", dsnWithoutDB)
	if err != nil {
		return err
	}
	defer db.Close()

	// delete it!
	_, err = db.Exec("DROP DATABASE IF EXISTS " + p.extres.DbName)
	if err != nil {
		return err
	}
	return nil
}

func (p *Database) ensureDatabase() error {
	if p.Datatype == "CSV" {
		return fmt.Errorf("this function is not valid for database type: %s", p.Datatype)
	}
	if p.Datatype != "SQL" {
		return fmt.Errorf("unknown database type: %s", p.Datatype)
	}

	// Construct DSN for the initial MySQL connection without specifying a database
	host := "tcp(127.0.0.1:3306)"
	dsnWithoutDB := fmt.Sprintf("%s:%s@%s/", p.extres.DbUser, p.extres.DbPass, host)

	// Connect to MySQL without specifying a database
	db, err := sql.Open("mysql", dsnWithoutDB)
	if err != nil {
		return err
	}
	defer db.Close()

	// Attempt to create the database if it doesn't exist
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + p.extres.DbName)
	if err != nil {
		return err
	}
	return nil
}

// Open opens the database for use. It creates the SQL DATABASE if needed,
// but it does not create any TABLES.
// ------------------------------------------------------------------------------
func (p *Database) Open() error {
	var err error
	p.Mim = NewInfluencerManager()
	switch p.Datatype {
	case "CSV":
		p.CSVDB.ParentDB = p
		return nil // nothing to do here at this point
	case "SQL":
		// Open a connection to MySQL without specifying a database
		if err = p.ensureDatabase(); err != nil {
			return err
		}
		dsn := util.GetSQLOpenString("plato", p.extres)
		p.SQLDB.DB, err = sql.Open("mysql", dsn)
		if err != nil {
			return err
		}
		if err = p.SQLDB.DB.Ping(); err == nil {
			return nil
		}
		if strings.Contains(err.Error(), "Unknown database") {
			if _, err = p.SQLDB.DB.Exec("DROP DATABASE IF EXISTS plato;"); err != nil {
				return err
			}
			if _, err = p.SQLDB.DB.Exec("CREATE DATABASE IF NOT EXISTS plato;"); err != nil {
				return err
			}
		}
		return err
	default:
		return fmt.Errorf("unknown database type: %s", p.Datatype)
	}
}

// CreateDatabasePart1 opens the database for use. If you're going to create a database, call it
// after Open() but before Init() to ensure that internal caches are properly loaded.
// -----------------------------------------------------------------------------------------
func (p *Database) CreateDatabasePart1() error {
	switch p.Datatype {
	case "CSV":
		return nil
	case "SQL":
		return p.SQLDB.CreateDatabasePart1()
	default:
		return fmt.Errorf("unknown database type: %s", p.Datatype)
	}
}

// Init loads the databases internal caches and other initialization. Make this call
// after CreateDatabasePart1 or MigrateCSVtoSQL so that caches are loaded with the copied data.
// -------------------------------------------------------------------------------------------
func (p *Database) Init() error {
	switch p.Datatype {
	case "CSV":
		return p.CSVDB.CSVInit()
	case "SQL":
		return p.SQLDB.SQLInit()
	default:
		return fmt.Errorf("unknown database type: %s", p.Datatype)
	}
}

// Insert inserts the Econometrics record into the database
// after CreateDatabasePart1 or MigrateCSVtoSQL so that caches are loaded with the copied data.
// -------------------------------------------------------------------------------------------
func (p *Database) Insert(rec *EconometricsRecord) error {
	switch p.Datatype {
	case "CSV":
		return fmt.Errorf("this function is not valid for database type: %s", p.Datatype)
	case "SQL":
		return p.SQLDB.Insert(rec)
	default:
		return fmt.Errorf("unknown database type: %s", p.Datatype)
	}
}
