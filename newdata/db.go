package newdata

import (
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
	cfg      *util.AppConfig         // application configuration info
	extres   *util.ExternalResources // the db may require secrets
	Datatype string                  // "CSV", "MYSQL"
	CSVDB    *DatasourceCSV          // valid when Datatype is "CSV"
	SQLDB    *DatabaseSQL            // valid when Datatype is "SQL"
}

// EconometricsRecord is the basic structure of discount rate data
type EconometricsRecord struct {
	Date   time.Time
	Fields map[string]float64
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
		db.CSVDB = &DatasourceCSV{}
		return &db, nil

	case "SQL":
		db := Database{
			cfg:      cfg,
			Datatype: "SQL",
			extres:   ex,
		}
		db.SQLDB = &DatabaseSQL{
			Name: "plato",
		}
		return &db, nil

	default:
		return nil, fmt.Errorf("unrecognized database type: %s", dtype)
	}
}

func (p *Database) ensureDatabase() error {
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

// Open opens the database for use
func (p *Database) Open() error {
	var err error
	switch p.Datatype {
	case "CSV":
		return p.CSVDB.LoadCsvDB()
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
			_, err = p.SQLDB.DB.Exec("DROP DATABASE IF EXISTS plato;CREATE DATABASE IF NOT EXISTS plato;")
		}
		return err
	default:
		return fmt.Errorf("unknown database type: %s", p.Datatype)
	}
}

// CreateDatabase opens the database for use
func (p *Database) CreateDatabase() error {
	switch p.Datatype {
	case "CSV":
		return nil
	case "SQL":
		return p.SQLDB.CreateDatabase()
	default:
		return fmt.Errorf("unknown database type: %s", p.Datatype)
	}
}
