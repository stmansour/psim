package newdata

import (
	"database/sql"
	"fmt"
	"log"

	// Register the MySQL driver
	_ "github.com/go-sql-driver/mysql"
)

// DatabaseSQL is the struct for the SQL database object
type DatabaseSQL struct {
	DB   *sql.DB
	Name string // database name
}

// CreateDatabase drops the current 'plato' database if it exists then
// creates a new one.
// ---------------------------------------------------------------------
func (p *DatabaseSQL) CreateDatabase() error {
	cmds := []string{
		"DROP DATABASE IF EXISTS plato",
		"CREATE DATABASE IF NOT EXISTS plato",
		"USE plato",
		"DROP TABLE IF EXISTS Locales",
		`CREATE TABLE Locales (
			LID INT AUTO_INCREMENT PRIMARY KEY,
			Name VARCHAR(255) NOT NULL,
			Description TEXT
		);`,
	}

	// Execute the SQL statement to create the table
	for i := 0; i < len(cmds); i++ {
		if _, err := p.DB.Exec(cmds[i]); err != nil {
			return err
		}
	}
	numShards := 4
	executeSQL := true
	if err := p.createShardedTables(numShards, executeSQL); err != nil {
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
    EntryID INT AUTO_INCREMENT PRIMARY KEY,
    DateTime DATETIME(6) NOT NULL,
    MetricValue FLOAT,
    INDEX(DateTime)
);`, tableName)

			// Depending on the executeSQL flag, print or execute the SQL statement
			if executeSQL {
				// Execute the SQL statement to create the table
				if _, err := p.DB.Exec(createTableSQL); err != nil {
					return fmt.Errorf("failed to create table %s: %v", tableName, err)
				}
				fmt.Printf("Table %s created successfully.\n", tableName)
			} else {
				// Just print the SQL statement
				fmt.Println(createTableSQL)
			}
		}
	}
	return nil
}
