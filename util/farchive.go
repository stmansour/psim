package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// CreateTimestampedDir creates a directory with the current timestamp down to nanosecond
// within the specified base directory and returns the path of the created directory.
//
// INPUTS
//
//	baseDir = where to create the archive, an empty string will create it in the
//	          current directory.
//
// ---------------------------------------------------------------------------------------
func CreateTimestampedDir(baseDir string) (string, error) {
	if baseDir == "" {
		baseDir = "."
	}
	// Format the current time to a human-readable form without colons
	// Example: 2024-02-21_15-04-05_123456789
	now := time.Now()
	dirName := now.Format("2006-01-02_15-04-05")
	dirName = fmt.Sprintf("%s_%09d", dirName, now.Nanosecond())
	fullPath := filepath.Join(baseDir, dirName)

	// Create the directory
	err := os.MkdirAll(fullPath, os.ModePerm)
	if err != nil {
		return "", err
	}
	return fullPath, nil
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
