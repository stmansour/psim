package util

import (
	"fmt"
	"os"
	"testing"
)

func TestNoExtresFile(t *testing.T) {
	extres, err := ReadExternalResources()
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
	extres, err := ReadExternalResources()
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
