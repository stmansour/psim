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

// WriteMetricsSourcesToCSV takes a slice of MetricsSource and creates a CSV file.
func (d *DatabaseCSV) WriteMetricsSourcesToCSV(locations []MetricsSource) error {
	FullyQualifiedFileName := filepath.Join(d.DBPath, "metricssources.csv")

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
			m.LastUpdate.Format("01/02/2006"),
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
			LocalStringList[subclass.LocaleType],     // fmt.Sprintf("%d", subclass.LocaleType),
			PredictoryStringList[subclass.Predictor], // fmt.Sprintf("%d", subclass.Predictor),
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

// CopySQLRecsToCSV copies SQL data records into a new CSV file called platodb.csv
// that can be used as a database for the simulator
// -----------------------------------------------------------------------------------
func (d *DatabaseCSV) CopySQLRecsToCSV(sqldb *Database) error {
	//----------------------------------------------
	// Create database CSV file:  platodb.csv
	//----------------------------------------------
	FullyQualifiedFileName := filepath.Join(d.DBPath, "platodb.csv")
	file, err := os.Create(FullyQualifiedFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	s := []string{}

	//------------------------------------------------
	// Prepare field selectors...
	//------------------------------------------------
	startDate := time.Date(2015, time.January, 1, 0, 0, 0, 0, time.UTC) // GDELT data starts at 2015
	endDate := time.Date(2023, time.December, 31, 0, 0, 0, 0, time.UTC) // time.Date(2023, time.December, 31, 0, 0, 0, 0, time.UTC)
	loc1 := d.ParentDB.cfg.C1
	loc2 := d.ParentDB.cfg.C2

	f := FieldSelector{
		Metric:  "EXClose",
		Locale:  loc1,
		Locale2: loc2,
	}
	fields := []FieldSelector{
		f,
	}

	//----------------------------
	// Write the header row
	//----------------------------
	fmt.Fprintf(file, "%q", "Date")        // special case 1:  date
	fmt.Fprintf(file, ",%q", f.FQMetric()) // special case 2: EXClose
	for _, v := range sqldb.Mim.MInfluencerSubclasses {
		switch v.LocaleType {
		case LocaleNone:
			field := FieldSelector{
				Metric: v.Metric, // no prepending needed
			}
			fmt.Fprintf(file, ",%q", v.Metric)
			fields = append(fields, field)
			s = append(s, v.Metric)
		case LocaleC1C2: // create 2 columns... one with C1 + metric, the other with C2 + metric
			// field 1
			field1 := FieldSelector{
				Metric: v.Metric, // no prepending needed
				Locale: d.ParentDB.cfg.C1,
			}
			fields = append(fields, field1)
			fld := field1.Locale + v.Metric
			fmt.Fprintf(file, ",%q", fld)
			s = append(s, fld)

			// field 2
			field2 := FieldSelector{
				Metric: v.Metric, // no prepending needed
				Locale: d.ParentDB.cfg.C2,
			}
			fields = append(fields, field2)
			fld = field2.Locale + v.Metric
			fmt.Fprintf(file, ",%q", fld)
			s = append(s, fld)
		default:
			return fmt.Errorf("unrecognized LocaleType on metric %s: %d", v.Metric, v.LocaleType)
		}
	}
	fmt.Fprintf(file, "\n") // end the line

	//----------------------------
	// Write the data rows
	//----------------------------
	year := 0
	month := 0
	for dt := startDate; dt.Before(endDate) || dt.Equal(endDate); dt = dt.AddDate(0, 0, 1) {
		rec, err := sqldb.Select(dt, fields)
		if err != nil {
			return err
		}
		fmt.Fprintf(file, "%q", rec.Date.Format("1/2/2006")) // special case 1: Date
		fmt.Fprintf(file, ",%.6f", rec.Fields[f.FQMetric()]) // special case 2: EXClose

		// Now the remaining metrics
		for i := 0; i < len(s); i++ {
			fld := s[i]
			val := rec.Fields[fld]
			fmt.Fprintf(file, ",%.6f", val)
		}
		fmt.Fprintf(file, "\n")
		if dt.Year() != year || dt.Month() != time.Month(month) {
			year = dt.Year()
			month = int(dt.Month())
			fmt.Printf("%4d-%02d\r", year, month)
		}
	}
	return nil
}
