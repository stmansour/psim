package util

import (
	"io"
	"os"
	"path/filepath"
)

// VerifyOrCreateDirectory creates a directory with the current timestamp down to nanosecond
// within the specified base directory and returns the path of the created directory.
//
// INPUTS
//
//	d = where to create the archive directory if it does not yet exist
//
// ---------------------------------------------------------------------------------------
func VerifyOrCreateDirectory(d string) (string, error) {
	err := os.MkdirAll(d, os.ModePerm)
	if err != nil {
		return "", err
	}
	return d, nil
}

// FileCopy copies a file from src to dest directory.
//
// INPUTS
//
//	src     = fully qualified file to copy
//	destDir = destination directoy for the copied file
//
// ---------------------------------------------------------------------------------------
func FileCopy(src, destDir string) error {
	// Open the source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Resolve the destination file path
	destPath := filepath.Join(destDir, filepath.Base(src))

	// Create the destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy the contents
	_, err = io.Copy(destFile, srcFile)
	return err
}

// GetExecutableDir returns the directory containing the executable that started the current process.
func GetExecutableDir() (string, error) {
	// Get the full path of the executable.
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	// Get the directory from the executable path.
	execDir := filepath.Dir(execPath)

	return execDir, nil
}
