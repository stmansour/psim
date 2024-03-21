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
	var loc Locale

	localesMap := make(map[string]Locale)
	rows, err := p.DB.Query("SELECT LID, Name, Currency, Description FROM Locales")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&loc.LID, &loc.Name, &loc.Currency, &loc.Description); err != nil {
			return err
		}
		localesMap[loc.Currency] = loc
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return err
	}
	p.LocaleCache = localesMap
	return nil
}

// FieldSelectorToSQL this function takes the CSV column name and splits it into
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
func (p *DatabaseSQL) FieldSelectorToSQL(f *FieldSelector) {
	f.LID = 1  // assume no locale
	f.LID2 = 1 // assume no locale2
	if len(f.Locale) > 0 {
		f.LID = int(p.LocaleCache[f.Locale].LID)
	}
	if len(f.Locale2) > 0 {
		f.LID2 = int(p.LocaleCache[f.Locale2].LID)
	}

	// Handle special cases, and the general case as the default
	switch f.Metric {
	case "EXClose":
		f.MID = -1
	default:
		f.MID = p.ParentDB.Mim.MInfluencerSubclasses[f.Metric].MID
	}
}
