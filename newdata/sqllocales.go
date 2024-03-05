package newdata

import (
	"fmt"
)

// Locale represents a single record for the Locales table.
type Locale struct {
	LID         int64
	Name        string
	Currency    string
	Description string
}

// InsertLocale inserts a new locale into the database
// --------------------------------------------------------------------------------
func (p *Database) InsertLocale(m *Locale) (int64, error) {
	switch p.Datatype {
	case "CSV":
		return 0, fmt.Errorf("this operation is net yet supported for CSV databases")
	case "SQL":
		return p.SQLDB.InsertLocale(m)
	default:
		return 0, fmt.Errorf("unknown database type: %s", p.Datatype)
	}
}

// InsertLocale inserts a new Locale into the Locales table.
func (p *DatabaseSQL) InsertLocale(loc *Locale) (int64, error) {
	stmt, err := p.DB.Prepare("INSERT INTO Locales(Name, Currency, Description) VALUES(?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(loc.Name, loc.Currency, loc.Description)
	if err != nil {
		return 0, err
	}

	// Get the last inserted ID
	lastInsertID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastInsertID, nil
}

// LoadLocaleCache inserts a new Locale into the Locales table.
func (p *DatabaseSQL) LoadLocaleCache() error {
	var lid int
	var name string
	localesMap := make(map[string]int)
	rows, err := p.DB.Query("SELECT LID, Currency FROM Locales")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&lid, &name); err != nil {
			return err
		}
		localesMap[name] = lid
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return err
	}
	p.LocaleIDCache = localesMap
	return nil
}

// CSVKeyToSQL this function takes the CSV column name and splits it into
// a key and LID for use in a SQL table.
// INPUTS
//
//	s = a CSV Column Name
//
// RETURNS
//
//	SQL metric, MID, locale.Currency  and LID
//
// -----------------------------------------------------------------
func (p *DatabaseSQL) CSVKeyToSQL(s string) (string, int, string, int) {
	// Default locale to "NON"
	locale := "NON"
	metric := ""

	// Check if the first three characters represent a known locale
	prefix := s[:3]
	if _, ok := p.LocaleIDCache[prefix]; ok {
		locale = prefix
		metric = s[3:]
	} else {
		// If not a known locale, the entire string is considered the metric name
		metric = s
		prefix = ""
	}
	LID := p.LocaleIDCache[locale]
	MID := p.ParentDB.Mim.MInfluencerSubclasses[metric].MID

	return metric, MID, locale, LID
}
