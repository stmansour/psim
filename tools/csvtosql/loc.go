package main

import (
	"github.com/stmansour/psim/newdata"
)

// CopyCsvLocalesToSQL initilizes the Locale table in sql
func CopyCsvLocalesToSQL() error {
	// Slice of Locale structs with country, currency, and a simple description
	locales := []newdata.Locale{
		{Name: "NON", Currency: "NON", Description: "No locale association"},
		{Name: "USA", Currency: "USD", Description: "United States of America, US Dollar"},
		{Name: "JPN", Currency: "JPY", Description: "Japan, Japanese Yen"},
		{Name: "GBR", Currency: "GBP", Description: "United Kingdom, British Pound"},
		{Name: "EUR", Currency: "EUR", Description: "Eurozone, Euro"},
		{Name: "CAN", Currency: "CAD", Description: "Canada, Canadian Dollar"},
		{Name: "AUS", Currency: "AUD", Description: "Australia, Australian Dollar"},
		{Name: "CHN", Currency: "CNY", Description: "China, Chinese Yuan"},
		{Name: "IND", Currency: "INR", Description: "India, Indian Rupee"},
		{Name: "BRA", Currency: "BRL", Description: "Brazil, Brazilian Real"},
		{Name: "RUS", Currency: "RUB", Description: "Russia, Russian Ruble"},
		{Name: "ZAF", Currency: "ZAR", Description: "South Africa, South African Rand"},
		{Name: "SGP", Currency: "SGD", Description: "Singapore, Singapore Dollar"},
		{Name: "NZL", Currency: "NZD", Description: "New Zealand, New Zealand Dollar"},
		{Name: "MEX", Currency: "MXN", Description: "Mexico, Mexican Peso"},
		{Name: "IDN", Currency: "IDR", Description: "Indonesia, Indonesian Rupiah"},
		{Name: "TUR", Currency: "TRY", Description: "Turkey, Turkish Lira"},
		{Name: "SAU", Currency: "SAR", Description: "Saudi Arabia, Saudi Riyal"},
		{Name: "SWE", Currency: "SEK", Description: "Sweden, Swedish Krona"},
		{Name: "NOR", Currency: "NOK", Description: "Norway, Norwegian Krone"},
		{Name: "DNK", Currency: "DKK", Description: "Denmark, Danish Krone"},
		{Name: "POL", Currency: "PLN", Description: "Poland, Polish Zloty"},
		{Name: "HKG", Currency: "HKD", Description: "Hong Kong, Hong Kong Dollar"},
		{Name: "KOR", Currency: "KRW", Description: "South Korea, South Korean Won"},
		{Name: "THA", Currency: "THB", Description: "Thailand, Thai Baht"},
		{Name: "PHL", Currency: "PHP", Description: "Philippines, Philippine Peso"},
		{Name: "MYS", Currency: "MYR", Description: "Malaysia, Malaysian Ringgit"},
		{Name: "NGA", Currency: "NGN", Description: "Nigeria, Nigerian Naira"},
		{Name: "EGY", Currency: "EGP", Description: "Egypt, Egyptian Pound"},
		{Name: "ISR", Currency: "ILS", Description: "Israel, Israeli New Shekel"},
		{Name: "PAK", Currency: "PKR", Description: "Pakistan, Pakistani Rupee"},
	}

	// Iterate over the locales slice and insert each into the database
	for _, locale := range locales {
		_, err := app.sqldb.InsertLocale(&locale)
		if err != nil {
			return err
		}
	}
	return nil
}
