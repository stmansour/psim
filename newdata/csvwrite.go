package newdata

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// EnsureDataDirectory ensures that the directory where the CSV database will be
// created exists. It appends "data" to the basepath.  If that directory does not
// exist, create it.
// ---------------------------------------------------------------------------------
func (d *DatabaseCSV) EnsureDataDirectory() (string, error) {
	basePath := d.DBPath
	dataPath := filepath.Join(basePath, "data")
	// Check if the directory exists
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		// The directory does not exist, create it
		if err := os.Mkdir(dataPath, 0755); err != nil {
			return "", err
		}
	}
	return dataPath, nil
}

// InsertMetricsSources takes a slice of MetricsSource and creates a CSV file.
func (d *DatabaseCSV) InsertMetricsSources(locations []MetricsSource) error {
	FullyQualifiedFileName := filepath.Join(d.DBPath, "metricsources.csv")

	file, err := os.Create(FullyQualifiedFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header
	header := []string{"MSID", "LastUpdate", "URL", "Name"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write the data rows
	for _, m := range locations {
		row := []string{
			strconv.Itoa(m.MSID),
			m.LastUpdate.Format(time.RFC3339), // Format time as RFC3339
			m.URL,
			m.Name,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// WriteLocalesToCSV is Function to write the slice of Locale structs into a CSV file
func (d *DatabaseCSV) WriteLocalesToCSV(locales map[string]Locale) error {
	FullyQualifiedFileName := filepath.Join(d.DBPath, "locales.csv")
	file, err := os.Create(FullyQualifiedFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	header := []string{"LID", "Name", "Currency", "Description"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header to CSV file: %v", err)
	}

	// Iterate over locales and write each one
	for _, locale := range locales {
		record := []string{
			fmt.Sprintf("%d", locale.LID),
			locale.Name,
			locale.Currency,
			locale.Description,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writing record to CSV file: %v", err)
		}
	}

	// Flush ensures all buffered data is written to the underlying file
	writer.Flush()

	if err := writer.Error(); err != nil {
		return fmt.Errorf("error flushing data to CSV file: %v", err)
	}

	return nil
}

// WriteMISubclassesToCSV writes the map of MInfluencerSubclass structs into a CSV file
// --------------------------------------------------------------------------------------
func (d *DatabaseCSV) WriteMISubclassesToCSV(subclassesMap map[string]MInfluencerSubclass) error {
	FullyQualifiedFileName := filepath.Join(d.DBPath, "misubclasses.csv")
	file, err := os.Create(FullyQualifiedFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	header := []string{
		"MID", "Name", "Metric", "BlocType", "LocaleType", "Predictor", "Subclass",
		"MinDelta1", "MaxDelta1", "MinDelta2", "MaxDelta2",
		"FitnessW1", "FitnessW2", "HoldWindowPos", "HoldWindowNeg",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header to CSV file: %v", err)
	}

	// Iterate over the map and write each struct
	for _, subclass := range subclassesMap {
		record := []string{
			fmt.Sprintf("%d", subclass.MID),
			subclass.Name,
			subclass.Metric,
			fmt.Sprintf("%d", subclass.BlocType),
			fmt.Sprintf("%d", subclass.LocaleType),
			fmt.Sprintf("%d", subclass.Predictor),
			subclass.Subclass,
			fmt.Sprintf("%d", subclass.MinDelta1),
			fmt.Sprintf("%d", subclass.MaxDelta1),
			fmt.Sprintf("%d", subclass.MinDelta2),
			fmt.Sprintf("%d", subclass.MaxDelta2),
			fmt.Sprintf("%f", subclass.FitnessW1),
			fmt.Sprintf("%f", subclass.FitnessW2),
			fmt.Sprintf("%f", subclass.HoldWindowPos),
			fmt.Sprintf("%f", subclass.HoldWindowNeg),
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writing record to CSV file: %v", err)
		}
	}

	// Flush ensures all buffered data is written to the underlying file
	writer.Flush()

	if err := writer.Error(); err != nil {
		return fmt.Errorf("error flushing data to CSV file: %v", err)
	}

	return nil
}
