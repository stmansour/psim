package newdata

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"

	// Register the MySQL driver
	_ "github.com/go-sql-driver/mysql"
)

// DatabaseSQL is the struct for the SQL database object
type DatabaseSQL struct {
	DB            *sql.DB
	Name          string         // database name
	BucketCount   int            // number of shards by metric name
	MetricIDCache map[string]int // metric name to bucket number
	LocaleIDCache map[string]int // locale name to LID
	ParentDB      *Database      // the database that contains me
}

// GetBucketForString returns the modulo number for the supplied
// metric string. The modulo number indicates which table shard
// the metric is kept in.
func (p *DatabaseSQL) GetBucketForString(s string) int {
	// Check the MetricIDCache first...
	if bucketNumber, found := p.MetricIDCache[s]; found {
		return bucketNumber
	}

	// If not in MetricIDCache, calculate the bucket number
	hash := sha256.Sum256([]byte(s))
	hashInt := 0
	for _, b := range hash[:] {
		hashInt += int(b)
	}
	bucketNumber := hashInt % p.BucketCount

	// Cache the calculated bucket number for future lookups
	p.MetricIDCache[s] = bucketNumber

	return bucketNumber
}

// CreateDatabasePart1 drops the current 'plato' database if it exists then
// creates a new one.
// ---------------------------------------------------------------------
func (p *DatabaseSQL) CreateDatabasePart1() error {
	cmds := []string{
		"CREATE DATABASE IF NOT EXISTS plato",
		"USE plato",
		"DROP TABLE IF EXISTS Locales",
		`CREATE TABLE Locales (
			LID INT AUTO_INCREMENT PRIMARY KEY,
			Name VARCHAR(80) NOT NULL,
			Currency VARCHAR(80) NOT NULL,
			Description TEXT
		);`,
		`CREATE TABLE MISubclasses (
			MID INT AUTO_INCREMENT PRIMARY KEY,
			Name VARCHAR(80) NOT NULL,
			Metric VARCHAR(80) NOT NULL,
			Subclass VARCHAR(80) NOT NULL,
			LocaleType TINYINT NOT NULL,
			Predictor TINYINT NOT NULL,
			MinDelta1 INT NOT NULL,
			MaxDelta1 INT NOT NULL,
			MinDelta2 INT NOT NULL,
			MaxDelta2 INT NOT NULL,
			FitnessW1 DECIMAL(13,6) NOT NULL,
			FitnessW2 DECIMAL(13,6) NOT NULL,
			HoldWindowPos DECIMAL(13,6) NOT NULL,
			HoldWindowNeg DECIMAL(13,6) NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS ExchangeRate (
			XID INT AUTO_INCREMENT PRIMARY KEY,
			Date DATETIME(6) NOT NULL,
			LID INT NOT NULL,
			LID2 INT,
			EXClose FLOAT,
			INDEX(Date)
		);`,
	}

	// Execute the SQL statement to create the table
	for i := 0; i < len(cmds); i++ {
		if _, err := p.DB.Exec(cmds[i]); err != nil {
			return err
		}
	}
	if err := p.createShardedTables(GlobalSQLSettings.BucketCount, true); err != nil {
		log.Fatalf("Failed to process sharded tables: %v", err)
	}

	return nil
}

// createShardedTables creates or prints SQL statements for table creation based on the executeSQL flag.
func (p *DatabaseSQL) createShardedTables(numShards int, executeSQL bool) error {
	for decade := 2000; decade <= 2020; decade += 10 {
		for shardIndex := 0; shardIndex < numShards; shardIndex++ {
			tableName := fmt.Sprintf("Metrics_%d_%d", shardIndex, decade)
			createTableSQL := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
    MEID INT AUTO_INCREMENT PRIMARY KEY,
    Date DATETIME(6) NOT NULL,
	MID INT NOT NULL,
	LID INT NOT NULL,
	LID2 INT,
    MetricValue FLOAT,
    INDEX(Date)
);`, tableName)

			// Depending on the executeSQL flag, print or execute the SQL statement
			if executeSQL {
				// Execute the SQL statement to create the table
				if _, err := p.DB.Exec(createTableSQL); err != nil {
					return fmt.Errorf("failed to create table %s: %v", tableName, err)
				}
				fmt.Printf("Table %s created successfully.\n", tableName)
			}
		}
	}
	return nil
}

// GrantReadAccess grants read access to all tables in the specified database
// for a list of usernames.
// -----------------------------------------------------------------------------
func (p *DatabaseSQL) GrantReadAccess(usernames []string) error {
	for _, username := range usernames {
		grantStmt := fmt.Sprintf("GRANT SELECT ON %s.* TO '%s'@'localhost'", p.Name, username)
		if _, err := p.DB.Exec(grantStmt); err != nil {
			return fmt.Errorf("failed to grant read access to %s: %v", username, err)
		}
	}
	if _, err := p.DB.Exec("FLUSH PRIVILEGES"); err != nil {
		return fmt.Errorf("failed to flush privileges: %v", err)
	}
	return nil
}

// GrantFullAccess grants full access to the 'plato' database for a list of
// usernames.
// --------------------------------------------------------------------------
func (p *DatabaseSQL) GrantFullAccess(usernames []string) error {
	for _, username := range usernames {
		grantStmt := fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO '%s'@'localhost'", p.Name, username)
		if _, err := p.DB.Exec(grantStmt); err != nil {
			return fmt.Errorf("failed to grant privileges to %s: %v", username, err)
		}
	}
	if _, err := p.DB.Exec("FLUSH PRIVILEGES"); err != nil {
		return fmt.Errorf("failed to flush privileges: %v", err)
	}
	return nil
}

// FieldSelectorsFromRecord creates an array of field selectors based
// on the supplied record.
// Improved version of FieldSelectorsFromRecord
func (p *DatabaseSQL) FieldSelectorsFromRecord(rec *EconometricsRecord) []FieldSelector {
	var ff []FieldSelector
	for k := range rec.Fields {
		var f FieldSelector
		p.FieldSelectorFromCSVColName(k, &f)
		ff = append(ff, f)
	}
	return ff
}

// FieldSelectorFromCSVColName updates f with the fields derived from k, a CSV column name
func (p *DatabaseSQL) FieldSelectorFromCSVColName(k string, f *FieldSelector) {
	// Attempt to extract up to two locales from the prefix of the key
	for i := 0; i < 2; i++ {
		if len(k) >= 3 {
			s := k[:3]
			if _, ok := p.LocaleIDCache[s]; ok {
				// If a locale is found, assign it and update the key to remove the found locale
				if i == 0 {
					f.Locale = s
				} else if i == 1 {
					f.Locale2 = s
				}
				k = k[3:]
				continue
			}
		}
		break // Break when no locale is found, or the remaining key is shorter than 3 characters
	}
	// what's left in k is the Metric
	f.Metric = k
}
