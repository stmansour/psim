package main

import (
	"github.com/stmansour/psim/newdata"
)

// CopyCsvLocalesToSQL initilizes the Locale table in sql
func CopyCsvLocalesToSQL() error {
	// Slice of Locale structs with country, currency, and a simple description
	locales := []newdata.Locale{
		{Name: "AUS", Currency: "AUD", Country: "Australia", Description: "Australia, Australian Dollar"},
		{Name: "BRA", Currency: "BRL", Country: "Brazil", Description: "Brazil, Brazilian Real"},
		{Name: "CAN", Currency: "CAD", Country: "Canada", Description: "Canada, Canadian Dollar"},
		{Name: "CHN", Currency: "CNY", Country: "China", Description: "China, Chinese Yuan"},
		{Name: "DNK", Currency: "DKK", Country: "Denmark", Description: "Denmark, Danish Krone"},
		{Name: "EGY", Currency: "EGP", Country: "Egypt", Description: "Egypt, Egyptian Pound"},
		{Name: "EUR", Currency: "EUR", Country: "Eurozone", Description: "Eurozone, Euro"},
		{Name: "GBR", Currency: "GBP", Country: "United Kingdom", Description: "United Kingdom, British Pound"},
		{Name: "HKG", Currency: "HKD", Country: "Hong Kong", Description: "Hong Kong, Hong Kong Dollar"},
		{Name: "IDN", Currency: "IDR", Country: "Indonesia", Description: "Indonesia, Indonesian Rupiah"},
		{Name: "IND", Currency: "INR", Country: "India", Description: "India, Indian Rupee"},
		{Name: "ISR", Currency: "ILS", Country: "Israel", Description: "Israel, Israeli New Shekel"},
		{Name: "JPN", Currency: "JPY", Country: "Japan", Description: "Japan, Japanese Yen"},
		{Name: "KOR", Currency: "KRW", Country: "South Korea", Description: "South Korea, South Korean Won"},
		{Name: "MEX", Currency: "MXN", Country: "Mexico", Description: "Mexico, Mexican Peso"},
		{Name: "MYS", Currency: "MYR", Country: "Malaysia", Description: "Malaysia, Malaysian Ringgit"},
		{Name: "NGA", Currency: "NGN", Country: "Nigeria", Description: "Nigeria, Nigerian Naira"},
		{Name: "NON", Currency: "NON", Country: "No locale association", Description: "No locale association"},
		{Name: "NOR", Currency: "NOK", Country: "Norway", Description: "Norway, Norwegian Krone"},
		{Name: "NZL", Currency: "NZD", Country: "New Zealand", Description: "New Zealand, New Zealand Dollar"},
		{Name: "PAK", Currency: "PKR", Country: "Pakistan", Description: "Pakistan, Pakistani Rupee"},
		{Name: "PHL", Currency: "PHP", Country: "Philippines", Description: "Philippines, Philippine Peso"},
		{Name: "POL", Currency: "PLN", Country: "Poland", Description: "Poland, Polish Zloty"},
		{Name: "RUS", Currency: "RUB", Country: "Russia", Description: "Russia, Russian Ruble"},
		{Name: "SAU", Currency: "SAR", Country: "Saudi Arabia", Description: "Saudi Arabia, Saudi Riyal"},
		{Name: "SGP", Currency: "SGD", Country: "Singapore", Description: "Singapore, Singapore Dollar"},
		{Name: "SWE", Currency: "SEK", Country: "Sweden", Description: "Sweden, Swedish Krona"},
		{Name: "THA", Currency: "THB", Country: "Thailand", Description: "Thailand, Thai Baht"},
		{Name: "TUR", Currency: "TRY", Country: "Turkey", Description: "Turkey, Turkish Lira"},
		{Name: "USA", Currency: "USD", Country: "United States", Description: "United States of America, US Dollar"},
		{Name: "ZAF", Currency: "ZAR", Country: "South Africa", Description: "South Africa, South African Rand"},
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
