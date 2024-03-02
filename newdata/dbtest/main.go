package main

import (
	"log"

	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/util"
)

func main() {
	ext, err := util.ReadExternalResources()
	if err != nil {
		log.Fatalf("ReadExternalResources returned error: %s\n", err.Error())
	}
	cfg, err := util.LoadConfig("")
	if err != nil {
		log.Fatalf("failed to read config file: %v\n", err)
	}

	db, err := newdata.NewDatabase("SQL", &cfg, ext)
	if err != nil {
		log.Fatalf("Error creating database: %s\n", err.Error())
	}
	if err = db.Open(); err != nil {
		log.Fatalf("db.Open returned error: %s\n", err.Error())
	}

	defer db.SQLDB.DB.Close()

	if err = db.CreateDatabase(); err != nil {
		log.Fatalf("CreateDatabase returned error: %s\n", err.Error())
	}

}
