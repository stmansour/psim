package sqlt

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// GenerateDBFileName generates a unique SQLite database filename for the running instance.
func GenerateDBFileName() (string, error) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %v", err)
	}

	// Get the process ID (PID)
	pid := os.Getpid()

	// Use a timestamp to ensure uniqueness even if the same PID is reused quickly
	timestamp := time.Now().UnixNano()

	// Construct the filename
	filename := fmt.Sprintf("plato_sim_%d_%d.db", pid, timestamp)

	// Combine the directory and filename
	filepath := filepath.Join(cwd, filename)

	return filepath, nil
}

// CreateSchema creates the schema in the SQLite database.
func CreateSchema(db *sql.DB) error {
	createTableSQL := `CREATE TABLE IF NOT EXISTS hashes (
		hash TEXT PRIMARY KEY
	);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		return err
	}
	return nil
}

// CheckAndInsertHash checks for the existence of a hash and inserts it if it does not exist.
func CheckAndInsertHash(db *sql.DB, hash string, elite bool) (bool, error) {
	if elite {
		return false, nil // ok to insert an elite
	}
	// Check if the hash exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM hashes WHERE hash = ?)", hash).Scan(&exists)
	if err != nil {
		return false, err
	}

	// If the hash exists, return true
	if exists {
		return true, nil
	}

	// If the hash does not exist, insert it
	_, err = db.Exec("INSERT INTO hashes (hash) VALUES (?)", hash)
	if err != nil {
		return false, err
	}

	// Return false to indicate the hash was not found and has been inserted
	return false, nil
}
