package main

// CopyCsvMISubclassesToSQL copies the CSV data for MInfluencerSubclasses
// into the equivalent SQL table.
func CopyCsvMISubclassesToSQL() error {
	for _, v := range app.csvdb.Mim.MInfluencerSubclasses {
		if err := app.sqldb.InsertMInfluencer(&v); err != nil {
			return err
		}
	}
	return nil
}
