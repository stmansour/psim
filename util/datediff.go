package util

import (
	"fmt"
	"strings"
	"time"
)

// DateDiff returns the number of years, months, days, hours, mins, and secs
// between the two provided dates.
// --------------------------------------------------------------------------
func DateDiff(a, b time.Time) (year, month, day, hour, min, sec int) {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	h1, m1, s1 := a.Clock()
	h2, m2, s2 := b.Clock()

	year = int(y2 - y1)
	month = int(M2 - M1)
	day = int(d2 - d1)
	hour = int(h2 - h1)
	min = int(m2 - m1)
	sec = int(s2 - s1)

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}
	if min < 0 {
		min += 60
		hour--
	}
	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	return
}

func addDurStr(x int, n string) string {
	s := fmt.Sprintf("%d %s", x, n)
	if x != 1 {
		s += "s"
	}
	return s
}

// DateDiffString returns a stringified number of years, months, and days
// from the result of DateDiff(a,b)
// ------------------------------------------------------------------------
func DateDiffString(a, b time.Time) string {
	year, month, day, _, _, _ := DateDiff(a, b)
	s := addDurStr(year, "year")
	s += " "
	s += addDurStr(month, "month")
	s += " "
	s += addDurStr(day, "day")
	return s
}

// ElapsedTime returns the elapsed time as a string
func ElapsedTime(dtStart, dtStop time.Time) string {
	elapsed := dtStop.Sub(dtStart) // calculate elapsed time
	s := ""
	hrs := int(elapsed.Hours())
	mins := int(elapsed.Minutes()) % 60
	secs := int(elapsed.Seconds()) % 60
	msec := int(elapsed.Milliseconds()) % 1000
	if hrs > 0 {
		s += fmt.Sprintf(" %d hr", hrs)
	} else if mins > 0 {
		s += fmt.Sprintf(" %d min", mins)
	} else if secs > 0 {
		s += fmt.Sprintf(" %d sec", secs)
	} else if msec > 0 {
		s += fmt.Sprintf(" %d msec", msec)
	}
	return strings.TrimLeft(s, " ") // remove leading spaces
}
