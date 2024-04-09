package newdata_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stmansour/psim/newdata"
)

func TestEnsureDataDirectory(t *testing.T) {
	// Specify a test directory path
	testBasePath := "./zztestzz"
	if err := os.Mkdir(testBasePath, 0755); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Create a DatabaseCSV instance with the test directory path
	d := newdata.DatabaseCSV{DBPath: testBasePath}

	// Call the EnsureDataDirectory function
	createdPath, err := d.EnsureDataDirectory()

	// Verify that the directories were created
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify that the created path is correct
	expectedPath := filepath.Join(testBasePath, "data")
	if createdPath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, createdPath)
	}

	// Clean up: remove the test directories
	err = os.RemoveAll(testBasePath)
	if err != nil {
		t.Errorf("Error cleaning up test directories: %v", err)
	}
}
