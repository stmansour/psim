package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gonum.org/v1/gonum/stat"
)

func main() {
	// Load CSV data
	file, err := os.Open("data/platodbsmall.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // variable number of fields per record
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// Get headers and data indices
	headers := rawCSVdata[0]
	baseIndex := -1
	for i, header := range headers {
		if header == "USDJPYEXClose" {
			baseIndex = i
			break
		}
	}
	if baseIndex == -1 {
		log.Fatal("Base metric 'USDJPYEXClose' not found in headers")
	}

	// Analysis parameters
	windowSize := 30
	correlationThreshold := 0.80

	// Calculate correlations, skipping the first column (Date) and the baseIndex
	for i := 1; i < len(headers); i++ { // Start from 1 to skip the Date column
		if i != baseIndex {
			metricName := headers[i] // Retrieve the metric name
			if metricName == "" {
				log.Printf("Warning: Metric name for column %d is empty", i)
				continue // Skip this iteration if the metric name is empty
			}
			for start := 0; start < len(rawCSVdata)-windowSize; start++ {
				end := start + windowSize
				baseData, compareData := getFloats(rawCSVdata, baseIndex, i, start, end)
				corr := stat.Correlation(baseData, compareData, nil)
				if abs(corr) >= correlationThreshold {
					fmt.Printf("There is a %.2f correlation between %s and %s from %s to %s.\n",
						corr, headers[baseIndex], metricName, rawCSVdata[start+1][0], rawCSVdata[end][0])
				}
			}
		}
	}
}

// getFloats retrieves slices of floats from the data for correlation calculation.
func getFloats(data [][]string, baseIndex, compareIndex, start, end int) ([]float64, []float64) {
	baseData := make([]float64, end-start)
	compareData := make([]float64, end-start)
	dataIndex := 0                       // Initialize dataIndex to track position within baseData and compareData
	for j := start + 1; j < end+1; j++ { // Note: Adjusted loop to directly use j for iterating over data
		row := data[j]
		cleanedBase := cleanNumberString(row[baseIndex])
		cleanedCompare := cleanNumberString(row[compareIndex])
		if len(cleanedCompare) == 0 || len(cleanedBase) == 0 {
			continue
		}

		base, err := strconv.ParseFloat(cleanedBase, 64)
		if err != nil {
			log.Printf("Error converting base data to float: %s, at row: %d\n", err, j)
			continue
		}
		compare, err := strconv.ParseFloat(cleanedCompare, 64)
		if err != nil {
			log.Printf("Error converting comparison data to float: %s, at row: %d\n", err, j)
			continue
		}
		baseData[dataIndex] = base
		compareData[dataIndex] = compare
		dataIndex++ // Increment dataIndex for each successful conversion
	}
	return baseData[:dataIndex], compareData[:dataIndex] // Use slicing to handle skipped rows
}

// cleanNumberString removes commas from the string to facilitate float parsing.
func cleanNumberString(s string) string {
	return strings.Replace(s, ",", "", -1)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
