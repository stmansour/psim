package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Record struct {
	Country   string
	Category  string
	Date      time.Time
	Rate      float64
	Frequency string
}

func main() {
	// Open the CSV file
	file, err := os.Open("infl-us.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Read the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading CSV file:", err)
		return
	}

	// Parse the records
	data := []Record{}
	for _, record := range records[1:] {
		date, err := time.Parse("1/02/2006", record[2])
		if err != nil {
			fmt.Println("Error parsing date:", err)
			continue
		}

		rate, err := strconv.ParseFloat(record[3][:len(record[3])-1], 64)
		if err != nil {
			fmt.Println("Error parsing rate:", err)
			continue
		}

		data = append(data, Record{
			Country:   record[0],
			Category:  record[1],
			Date:      date,
			Rate:      rate,
			Frequency: record[4],
		})
	}

	d1 := data[0].Date
	d2 := data[len(data)-1].Date

	fmt.Printf("Start date: %s\n", d1.Format("01/02/2006"))
	fmt.Printf("Stop date: %s\n", d2.Format("01/02/2006"))

	r := data[0]
	i := 0

	for dt := d1; dt.Before(d2); dt = dt.Add(24 * time.Hour) {
		if dt.After(r.Date) {
			i++
			r = data[i]
		}
		fmt.Printf("%s,%s,%s,%4.2f%%,%s\n", r.Country, r.Category, dt.Format("01/02/2006"), r.Rate, r.Frequency)
	}

}
