package util

import (
	"fmt"
	"testing"
	"time"
)

func TestDateParse(t *testing.T) {
	d := []string{
		"3/7/2023",
		"03/7/2023",
		"03/07/2023",
		"3/7/23",
		"2023-03-07",
	}

	Mar7 := time.Date(2023, time.March, 7, 0, 0, 0, 0, time.UTC)
	day := Mar7.Day()
	month := Mar7.Month()
	year := Mar7.Year()

	for i := 0; i < len(d); i++ {
		dt, err := StringToDate(d[i])
		if err != nil {
			t.Errorf("error from StringToDate: %s\n", err)
		}
		if dt.Year() != year {
			t.Errorf("error: %s - year mismatch: translated to year = %d\n", d[i], dt.Year())
		}
		if dt.Month() != month {
			t.Errorf("error: %s - month mismatch: translated to month = %v\n", d[i], dt.Month())
		}
		if dt.Day() != day {
			t.Errorf("error: %s - day mismatch: translated to day = %d\n", d[i], dt.Day())
		}
	}

	dt, err := StringToDate("2018-01-31")
	if err != nil {
		t.Errorf("error from StringToDate: %s\n", err)
	}
	fmt.Printf("Year = %d, Month = %v, day = %d\n", dt.Year(), dt.Month(), dt.Day())

	dt, err = StringToDate("2020-12-01T00:00:00")
	if err != nil {
		t.Errorf("error from StringToDate: %s\n", err)
	}
	if dt.Year() != 2020 || dt.Month() != time.December || dt.Day() != 1 {
		t.Errorf("Expected D M Y = 2020 December 1, got Year = %d, Month = %v, day = %d\n", dt.Year(), dt.Month(), dt.Day())
	}

}
