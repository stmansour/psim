package newdata

// DNABank defines the table of elite DNA that has been saved because of
// the Investor's high performance
type DNABank struct {
	DNA                      string  // the DNA we're preserving
	C1                       string  // from cfg file: currency 1
	C2                       string  // from cfg file: currency 2
	TxnFeeFactor             float64 // from cfg file: txn fee factor
	TxnFee                   float64 // from cfg file: txn fee
	StopLoss                 float64 // from cfg file: stop loss
	HoldWindowStatsLookBack  int     // from cfg file: hold window
	StdDevFactor             float64 // from cfg file: std dev factor
	AnnualizedReturnAchieved float64 // what did this investor achive
	DtStart                  string  // from cfg file: start date
	DtStop                   string  // from cfg file: stop date
}
