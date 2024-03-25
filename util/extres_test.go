package util_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stmansour/psim/util"
)

func TestNoExtresFile(t *testing.T) {
	extres, err := util.ReadExternalResources()
	if err != nil {
		t.Errorf("Error from ReadExternalResources: %s\n", err.Error())
	}
	fmt.Printf("username = %s\n", extres.DbUser)
}

func TestExtresFile(t *testing.T) {
	err := CreateDummyExtres()
	if err != nil {
		t.Errorf("error creating file: %v", err)
		return
	}
	extres, err := util.ReadExternalResources()
	if err != nil {
		t.Errorf("Error from ReadExternalResources: %s\n", err.Error())
		return
	}
	fmt.Printf("username = %s\n", extres.DbUser)
	err = os.Remove("extres.json5")
	if err != nil {
		t.Errorf("error deleting file: %v", err)
		return
	}
}

func CreateDummyExtres() error {
	content := `{
    "Env": 0,
    "Dbuser": "abc",
    "Dbname": "dbe",
    "Dbpass": "fgh",
    "Dbhost": "ijk",
    "Dbport": 3306,
    "Dbtype": "lmn",
}`

	file, err := os.Create("extres.json5")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}

func TestGetSQLOpenString(t *testing.T) {
	testCases := []struct {
		env      int
		expected string
	}{
		{env: 99, expected: ""},
		{env: util.DEV, expected: "/plato?charset=utf8&parseTime=True"},
		{env: util.PROD, expected: "/plato?charset=utf8&parseTime=True"},
		{env: util.QA, expected: "/plato?charset=utf8&parseTime=True"},
		// Add more test cases for each switch path
	}
	extres, err := util.ReadExternalResources()
	if err != nil {
		t.Errorf("Error from ReadExternalResources: %s\n", err.Error())
	}

	for _, tc := range testCases {
		extres.Env = tc.env
		result := extres.GetSQLOpenString(extres.DbName)
		idx := strings.Index(result, "/")
		if idx != -1 {
			result = result[idx:]
		}

		if result != tc.expected {
			t.Errorf("GetSQLOpenString(%s) returned %s, expected %v", extres.DbName, result, tc.expected)
		}
	}
}
