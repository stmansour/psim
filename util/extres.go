package util

import (
	"fmt"
	"os"

	json5 "github.com/yosuke-furukawa/json5/encoding/json5"
)

// ExternalResources is used to store sensitive or secret config values
// for gaining access to external resources.
type ExternalResources struct {
	Env    int    `json:"Env"` // 0 = dev, 1 = qa, 2 = production
	DbUser string `json:"Dbuser"`
	DbName string `json:"Dbname"`
	DbPass string `json:"Dbpass"`
	DbHost string `json:"Dbhost"`
	DbPort int    `json:"Dbport"`
	DbType string `json:"Dbtype"`
}

// Define constant variables for DEV, QA, and PROD as per corrected mapping
const (
	DEV  = 0
	QA   = 1
	PROD = 2
)

// EnvironmentToCode maps an environment input string to a corresponding code.
func EnvironmentToCode(env string) int {
	switch env {
	case "DEV":
		return DEV
	case "QA":
		return QA
	case "PROD":
		return PROD
	default:
		return -1 // Return -1 or another value to indicate an unknown environment
	}
}

// GetSQLOpenString builds the string to use for opening an sql database.
// Input string is the name of the database:  "accord" for phonebook, "rentroll" for RentRoll
// Returns:  a string to pass to sql.Open()
// =======================================================================================
func GetSQLOpenString(dbname string, a *ExternalResources) string {
	s := ""
	switch a.Env {
	case DEV: //development
		s = fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True",
			a.DbUser, a.DbPass, dbname)
	case QA:
		s = fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True",
			a.DbUser, a.DbPass, dbname)
	case PROD: //production
		s = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True",
			a.DbUser, a.DbPass, a.DbHost, a.DbPort, dbname)
	default:
		fmt.Printf("Unhandled configuration environment: %d\n", a.Env)
		os.Exit(1)
	}
	return s
}

// ReadExternalResources reads the contents of extres.json5 and fills the ExternalResources struct.
func ReadExternalResources() (*ExternalResources, error) {
	filename := "extres.json5"
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode the JSON5 data into the ExternalResources struct
	var resources ExternalResources
	err = json5.NewDecoder(file).Decode(&resources)
	if err != nil {
		return nil, err
	}

	return &resources, nil
}
