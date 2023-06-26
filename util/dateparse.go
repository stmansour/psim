package util

import (
	"fmt"
	"strings"
	"time"
)

// AcceptedDateFmts is the array of string formats that StringToDate accepts
var AcceptedDateFmts = []string{
	"2006-1-2",
	"1/2/06",
	"1/2/2006",
	"1/2/06",
}

// StringToDate tries to convert the supplied string to a time.Time value. It will use the
// formats called out in dbtypes.go:  RRDATEFMT, RRDATEINPFMT, RRDATEINPFMT2, ...
//
// for further experimentation, try: https://play.golang.org/p/JNUnA5zbMoz
// ----------------------------------------------------------------------------------
func StringToDate(s string) (time.Time, error) {
	// try the ansi std date format first
	var Dt time.Time
	var err error
	s = Stripchars(s, "\"")
	s = strings.TrimSpace(s)
	for i := 0; i < len(AcceptedDateFmts); i++ {
		Dt, err = time.Parse(AcceptedDateFmts[i], s)
		if nil == err {
			return Dt, nil
		}
	}
	return Dt, fmt.Errorf("date could not be decoded")
}
