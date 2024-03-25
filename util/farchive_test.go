package util_test

import (
	"os"
	"testing"

	"github.com/stmansour/psim/util"
)

func TestFileCopy(t *testing.T) {
	src := "config.json5"
	destDir := "datatest"

	dir, err := util.VerifyOrCreateDirectory(destDir)
	if err != nil {
		t.Errorf("VerifyOrCreateDirectory returned an error: %v", err)
	}

	dest := dir + "/config.json5"

	// Copy the file
	err = util.FileCopy(src, destDir)
	if err != nil {
		t.Errorf("FileCopy returned an error: %v", err)
		return
	}

	// Check if the file was copied
	_, err = os.Stat(dest)
	if err != nil {
		t.Errorf("File was not copied: %v", err)
		return
	}

	// Remove the copied file
	err = os.Remove(dest)
	if err != nil {
		t.Errorf("Error removing copied file: %v", err)
		return
	}

	// Remove the destination directory
	err = os.Remove(destDir)
	if err != nil {
		t.Errorf("Error removing destination directory: %v", err)
		return
	}
}
