package newcore

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/stmansour/psim/util"
	"github.com/xuri/excelize/v2"
	"gonum.org/v1/gonum/stat"
)

// DNALogResult represents the results of a run presented in this log
type DNALogResult struct {
	SuccessCoefficient float64
	Consistency        float64
	AnnualizedReturn   float64
}

// DNALog is the class that creates the DNA log
type DNALog struct {
	cfg       *util.AppConfig
	s         *Simulator
	filename  string         // name of report file
	parent    *Crucible      // crucible object containing this
	row       int            // current row number
	sheetName string         // current sheet we're working on
	f         *excelize.File // excel file
	Results   map[string]*DNALogResult
}

// NewDNALog creates and returns a new DNA log object
func NewDNALog() *DNALog {
	return &DNALog{
		Results: make(map[string]*DNALogResult),
	}
}

// Init initializes the object
func (dl *DNALog) Init(c *Crucible, cfg *util.AppConfig, sim *Simulator) {
	dl.cfg = cfg
	dl.parent = c
	dl.s = sim
}

// setCellNext - adds a line to the excel file
func (dl *DNALog) setCellNext(row int, col string, s string) int {
	srow := strconv.Itoa(row)
	dl.f.SetCellValue(dl.sheetName, col+srow, s)
	return row + 1
}

// ReportHeader - excel version of the standard report header used in simstats,
// finrep, crep, etc
// -------------------------------------------------------------------------------
func (dl *DNALog) ReportHeader(row int) int {
	// et := dl.s.GetSimulationRunTime()
	// a := time.Time(dl.s.Cfg.DtStart)
	// b := time.Time(dl.s.Cfg.DtStop)
	// c := b.AddDate(0, 0, 1)

	row = dl.setCellNext(row, "A", fmt.Sprintf("Program Version:  %s", util.Version()))

	row = dl.setCellNext(row, "A", fmt.Sprintf("Run Date: %s", time.Now().Format("Mon, Jan 2, 2006 - 15:04:05 MST")))

	// row = dl.setCellNext(row, "A", fmt.Sprintf("Available processor cores: %d", runtime.NumCPU()))
	// row = dl.setCellNext(row, "A", fmt.Sprintf("Worker Threads: %d", dl.s.WorkerThreads))

	row = dl.setCellNext(row, "A", fmt.Sprintf("Configuration File:  %s", dl.s.Cfg.ConfigFilename))
	if dl.s.db.Datatype == "CSV" {
		row = dl.setCellNext(row, "A", fmt.Sprintf("Database: %s", dl.s.db.CSVDB.DBFname))
		row = dl.setCellNext(row, "A", fmt.Sprintf("Nil data requests: %d", dl.s.db.CSVDB.Nildata))
	} else {
		row = dl.setCellNext(row, "A", fmt.Sprintf("Database: %s  (SQL)", dl.s.db.SQLDB.Name))
	}
	// row = dl.setCellNext(row, "A", fmt.Sprintf("Simulation Start Date: %s", a.Format("Mon, Jan 2, 2006 - 15:04:05 MST")))
	// row = dl.setCellNext(row, "A", fmt.Sprintf("Simulation Stop Date: %s", b.Format("Mon, Jan 2, 2006 - 15:04:05 MST")))
	// if dl.s.Cfg.SingleInvestorMode {
	// 	row = dl.setCellNext(row, "A", "Single Investor Mode")
	// 	row = dl.setCellNext(row, "A", fmt.Sprintf("DNA: %s", dl.s.Cfg.SingleInvestorDNA))
	// } else {
	// 	row = dl.setCellNext(row, "A", fmt.Sprintf("Generations: %d", dl.s.GensCompleted))
	// 	if len(dl.s.Cfg.GenDurSpec) > 0 {
	// 		row = dl.setCellNext(row, "A", fmt.Sprintf("Generation Lifetime: %s", util.FormatGenDur(dl.s.Cfg.GenDur)))
	// 	}
	// 	row = dl.setCellNext(row, "A", fmt.Sprintf("Simulation Loop Count: %d", dl.s.Cfg.LoopCount))
	// 	row = dl.setCellNext(row, "A", fmt.Sprintf("Simulation Time Duration: %s", util.DateDiffString(a, c)))
	// }
	row = dl.setCellNext(row, "A", fmt.Sprintf("C1: %s", dl.s.Cfg.C1))
	row = dl.setCellNext(row, "A", fmt.Sprintf("C2: %s", dl.s.Cfg.C2))

	// row = dl.setCellNext(row, "A", fmt.Sprintf("Population: %d", dl.s.Cfg.PopulationSize))
	// row = dl.setCellNext(row, "A", fmt.Sprintf("Influencers: min %d,  max %d", dl.s.Cfg.MinInfluencers, dl.s.Cfg.MaxInfluencers))
	row = dl.setCellNext(row, "A", fmt.Sprintf("Initial Funds: %.2f %s", dl.s.Cfg.InitFunds, dl.s.Cfg.C1))
	row = dl.setCellNext(row, "A", fmt.Sprintf("Standard Investment: %.2f %s", dl.s.Cfg.StdInvestment, dl.s.Cfg.C1))
	row = dl.setCellNext(row, "A", fmt.Sprintf("Stop Loss: %.2f%%", dl.s.Cfg.StopLoss*100))
	// row = dl.setCellNext(row, "A", fmt.Sprintf("Preserve Elite: %v  (%5.2f%%)", dl.s.Cfg.PreserveElite, dl.s.Cfg.PreserveElitePct))
	row = dl.setCellNext(row, "A", fmt.Sprintf("Transaction Fee: %.2f (flat rate)  %5.1f bps", dl.s.Cfg.TxnFee, dl.s.Cfg.TxnFeeFactor*10000))
	row = dl.setCellNext(row, "A", fmt.Sprintf("Investor Bonus Plan: %v", dl.s.Cfg.InvestorBonusPlan))
	// row = dl.setCellNext(row, "A", fmt.Sprintf("Gen 0 Elites: %v", dl.s.Cfg.Gen0Elites))
	row = dl.setCellNext(row, "A", fmt.Sprintf("HoldWindowStatsLookback: %d", dl.s.Cfg.HoldWindowStatsLookBack))
	row = dl.setCellNext(row, "A", fmt.Sprintf("StdDevVariationFactor: %.4f", dl.s.Cfg.StdDevVariationFactor))

	// omr := float64(0)
	// if dl.s.factory.MutateCalls > 0 {
	// 	omr = 100.0 * float64(dl.s.factory.Mutations) / float64(dl.s.factory.MutateCalls)
	// }
	// row = dl.setCellNext(row, "A", fmt.Sprintf("Observed Mutation Rate: %6.3f%%", omr))
	// if !dl.s.Cfg.CrucibleMode {
	// 	row = dl.setCellNext(row, "A", fmt.Sprintf("Elapsed Run Time: %s", et))
	// }
	row++

	return row
}

// WriteHeader writes the DNA log as an Excel spreadsheet file
func (dl *DNALog) WriteHeader() error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	dl.f = f

	//===========================================================================
	// Create the main sheet
	//===========================================================================
	dl.sheetName = "DNA Log"
	index, err := f.NewSheet(dl.sheetName)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	// sheet, _ := f.NewSheet(dl.sheetName)
	f.SetActiveSheet(index)

	//===========================================================================
	//  TITLE
	//===========================================================================
	var titleStyle int
	var blackBGStyle int
	titleStyle, err = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 24,
		},
	})
	if err != nil {
		fmt.Println(err)
		return err
	}
	f.SetCellValue(dl.sheetName, "A1", "PLATO DNA LOG")
	f.SetCellStyle(dl.sheetName, "A1", "A1", titleStyle)

	blackBGStyle, err = f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#000000"},
			Pattern: 1,
		},
	})
	if err != nil {
		fmt.Println(err)
		return err
	}
	f.SetCellStyle(dl.sheetName, "A2", "BA2", blackBGStyle)

	//===========================================================================
	// STANDARD REPORT HEADER
	//===========================================================================
	stdReportStartRow := 4
	hdrRow1 := dl.ReportHeader(stdReportStartRow) // start the header info on row 4
	hdr1 := fmt.Sprintf("%d", hdrRow1)
	stdReportStopRow := hdrRow1 - 1
	stdReportHeaderStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			// Bold: true,
			Size: 14,
		},
	})
	if err != nil {
		fmt.Println(err)
		return err
	}
	f.SetCellStyle(dl.sheetName, "A"+strconv.Itoa(stdReportStartRow), "A"+strconv.Itoa(stdReportStopRow), stdReportHeaderStyle)

	//===========================================================================
	// COLUMN HEADERS
	// Set complex headers and merge cells
	// Row 1 headers
	//===========================================================================
	f.MergeCell(dl.sheetName, "B"+hdr1, "K"+hdr1)
	f.SetCellValue(dl.sheetName, "B"+hdr1, "Rolling Success Coefficients")

	f.MergeCell(dl.sheetName, "L"+hdr1, "N"+hdr1)
	f.SetCellValue(dl.sheetName, "L"+hdr1, "Rolling 1 Week")

	f.MergeCell(dl.sheetName, "O"+hdr1, "Q"+hdr1)
	f.SetCellValue(dl.sheetName, "O"+hdr1, "Rolling 2 Weeks")

	f.MergeCell(dl.sheetName, "R"+hdr1, "T"+hdr1)
	f.SetCellValue(dl.sheetName, "R"+hdr1, "Rolling 3 Weeks")

	f.MergeCell(dl.sheetName, "U"+hdr1, "W"+hdr1)
	f.SetCellValue(dl.sheetName, "u"+hdr1, "Rolling 1 Month")

	f.MergeCell(dl.sheetName, "x"+hdr1, "z"+hdr1)
	f.SetCellValue(dl.sheetName, "z"+hdr1, "Rolling 3 Month")

	f.MergeCell(dl.sheetName, "aa"+hdr1, "ac"+hdr1)
	f.SetCellValue(dl.sheetName, "aa"+hdr1, "Rolling 6 Month")

	f.MergeCell(dl.sheetName, "ad"+hdr1, "af"+hdr1)
	f.SetCellValue(dl.sheetName, "ad"+hdr1, "Rolling 9 Month")

	f.MergeCell(dl.sheetName, "ag"+hdr1, "ai"+hdr1)
	f.SetCellValue(dl.sheetName, "ag"+hdr1, "Rolling 12 Months")

	f.MergeCell(dl.sheetName, "aj"+hdr1, "al"+hdr1)
	f.SetCellValue(dl.sheetName, "aj"+hdr1, "2024 YTD")

	f.MergeCell(dl.sheetName, "am"+hdr1, "ao"+hdr1)
	f.SetCellValue(dl.sheetName, "am"+hdr1, "2023")

	f.MergeCell(dl.sheetName, "ap"+hdr1, "ar"+hdr1)
	f.SetCellValue(dl.sheetName, "ap"+hdr1, "2022")

	f.MergeCell(dl.sheetName, "as"+hdr1, "au"+hdr1)
	f.SetCellValue(dl.sheetName, "as"+hdr1, "2021")

	f.MergeCell(dl.sheetName, "av"+hdr1, "ax"+hdr1)
	f.SetCellValue(dl.sheetName, "av"+hdr1, "2020")

	f.MergeCell(dl.sheetName, "ay"+hdr1, "ba"+hdr1)
	f.SetCellValue(dl.sheetName, "ay"+hdr1, "2019")

	hdrRow2 := hdrRow1 + 1
	hdr2 := fmt.Sprintf("%d", hdrRow2)

	f.SetCellValue(dl.sheetName, "A"+hdr2, "INVESTOR")
	f.SetCellValue(dl.sheetName, "B"+hdr2, "Weeks 1-2 Success Coefficient")
	f.SetCellValue(dl.sheetName, "C"+hdr2, "Weeks 1-3 Success Coefficient")
	f.SetCellValue(dl.sheetName, "D"+hdr2, "1,3 Month Success Coefficient")
	f.SetCellValue(dl.sheetName, "E"+hdr2, "1,3,6 Month Success Coefficient")
	f.SetCellValue(dl.sheetName, "F"+hdr2, "1,3,6,9 Month Success Coefficient")
	f.SetCellValue(dl.sheetName, "G"+hdr2, "2 Year Success Coefficient 2024 YTD, 2023")
	f.SetCellValue(dl.sheetName, "H"+hdr2, "3 Year Success Coefficient 2024 YTD, 2023, 2022")
	f.SetCellValue(dl.sheetName, "I"+hdr2, "4 Year Success Coefficient 2024 YTD, 2023, 2022, 2021")
	f.SetCellValue(dl.sheetName, "J"+hdr2, "5 Year Success Coefficient 2024 YTD, 2023, 2022, 2021, 2020")
	f.SetCellValue(dl.sheetName, "K"+hdr2, "6 Year Success Coefficient 2024 YTD, 2023, 2022, 2021, 2020, 2019")

	f.SetCellValue(dl.sheetName, "l"+hdr2, "Consistency")
	f.SetCellValue(dl.sheetName, "m"+hdr2, "Annualized Return")
	f.SetCellValue(dl.sheetName, "n"+hdr2, "Success Coefficient")

	f.SetCellValue(dl.sheetName, "o"+hdr2, "Consistency")
	f.SetCellValue(dl.sheetName, "p"+hdr2, "Annualized Return")
	f.SetCellValue(dl.sheetName, "q"+hdr2, "Success Coefficient")

	f.SetCellValue(dl.sheetName, "r"+hdr2, "Consistency")
	f.SetCellValue(dl.sheetName, "s"+hdr2, "Annualized Return")
	f.SetCellValue(dl.sheetName, "t"+hdr2, "Success Coefficient")

	f.SetCellValue(dl.sheetName, "u"+hdr2, "Consistency")
	f.SetCellValue(dl.sheetName, "v"+hdr2, "Annualized Return")
	f.SetCellValue(dl.sheetName, "w"+hdr2, "Success Coefficient")

	f.SetCellValue(dl.sheetName, "x"+hdr2, "Consistency")
	f.SetCellValue(dl.sheetName, "y"+hdr2, "Annualized Return")
	f.SetCellValue(dl.sheetName, "z"+hdr2, "Success Coefficient")

	f.SetCellValue(dl.sheetName, "aa"+hdr2, "Consistency")
	f.SetCellValue(dl.sheetName, "ab"+hdr2, "Annualized Return")
	f.SetCellValue(dl.sheetName, "ac"+hdr2, "Success Coefficient")

	f.SetCellValue(dl.sheetName, "ad"+hdr2, "Consistency")
	f.SetCellValue(dl.sheetName, "ae"+hdr2, "Annualized Return")
	f.SetCellValue(dl.sheetName, "af"+hdr2, "Success Coefficient")

	f.SetCellValue(dl.sheetName, "ag"+hdr2, "Consistency")
	f.SetCellValue(dl.sheetName, "ah"+hdr2, "Annualized Return")
	f.SetCellValue(dl.sheetName, "ai"+hdr2, "Success Coefficient")

	f.SetCellValue(dl.sheetName, "aj"+hdr2, "Consistency")
	f.SetCellValue(dl.sheetName, "ak"+hdr2, "Annualized Return")
	f.SetCellValue(dl.sheetName, "al"+hdr2, "Success Coefficient")

	f.SetCellValue(dl.sheetName, "am"+hdr2, "Consistency")
	f.SetCellValue(dl.sheetName, "an"+hdr2, "Annualized Return")
	f.SetCellValue(dl.sheetName, "ao"+hdr2, "Success Coefficient")

	f.SetCellValue(dl.sheetName, "ap"+hdr2, "Consistency")
	f.SetCellValue(dl.sheetName, "aq"+hdr2, "Annualized Return")
	f.SetCellValue(dl.sheetName, "ar"+hdr2, "Success Coefficient")

	f.SetCellValue(dl.sheetName, "as"+hdr2, "Consistency")
	f.SetCellValue(dl.sheetName, "at"+hdr2, "Annualized Return")
	f.SetCellValue(dl.sheetName, "au"+hdr2, "Success Coefficient")

	f.SetCellValue(dl.sheetName, "av"+hdr2, "Consistency")
	f.SetCellValue(dl.sheetName, "aw"+hdr2, "Annualized Return")
	f.SetCellValue(dl.sheetName, "ax"+hdr2, "Success Coefficient")

	f.SetCellValue(dl.sheetName, "ay"+hdr2, "Consistency")
	f.SetCellValue(dl.sheetName, "az"+hdr2, "Annualized Return")
	f.SetCellValue(dl.sheetName, "ba"+hdr2, "Success Coefficient")

	f.SetCellValue(dl.sheetName, "bb"+hdr2, "Investor ID")
	f.SetCellValue(dl.sheetName, "bc"+hdr2, "Investor DNA")

	// Define a new style with bold font
	colHdrStyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			WrapText:   true,
			Horizontal: "center",
		},
		Font: &excelize.Font{
			// Bold: true,
			Size: 14,
		},
		Border: []excelize.Border{
			{Type: "top", Color: "#000000", Style: 1},    // Left border
			{Type: "left", Color: "#000000", Style: 1},   // Top border
			{Type: "right", Color: "#000000", Style: 1},  // Right border
			{Type: "bottom", Color: "#000000", Style: 1}, // Bottom border
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#DDDDDD"},
			Pattern: 1,
		},
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	f.SetCellStyle(dl.sheetName, "A"+hdr1, "BC"+hdr2, colHdrStyle)

	// Set column widths - I think the units are
	if err = f.SetColWidth(dl.sheetName, "A", "A", 20); err != nil {
		fmt.Println(err)
		return err
	}

	if err = f.SetColWidth(dl.sheetName, "B", "bc", 15); err != nil {
		fmt.Println(err)
		return err
	}

	// Set the active sheet to the created sheet
	f.SetActiveSheet(index)
	dl.row = 1 + hdrRow2

	// Save the Excel file
	dl.filename = "dnalog.xlsx"
	if err := f.SaveAs(dl.filename); err != nil {
		fmt.Println(err)
	}
	return nil
}

// WriteRow adds information about the current Crucible Investor to the report
func (dl *DNALog) WriteRow() {
	f, err := excelize.OpenFile(dl.filename)
	if err != nil {
		log.Fatal(err)
	}
	row := strconv.Itoa(dl.row)

	//---------------------------------------------------------------------------
	// NAME
	//---------------------------------------------------------------------------
	v := dl.parent.sim.factory.NewInvestorFromDNA(dl.cfg.TopInvestors[dl.parent.idx].DNA)
	invname := dl.cfg.TopInvestors[dl.parent.idx].Name
	if len(invname) == 0 {
		invname = v.ShortID()
	}
	f.SetCellValue(dl.sheetName, "A"+row, invname)

	//---------------------------------------------------------------------------
	// Success Coefficients
	//---------------------------------------------------------------------------
	m := dl.parent.AnnualizedReturnList[:2] // 1 week, 2 week
	mean, stddev := stat.MeanStdDev(m, nil)
	consistency := 1 - stddev
	sc := mean * consistency
	f.SetCellValue(dl.sheetName, "B"+row, sc)

	m = dl.parent.AnnualizedReturnList[:3] // 1 week, 2 week, 3 week
	mean, stddev = stat.MeanStdDev(m, nil)
	consistency = 1 - stddev
	sc = mean * consistency
	f.SetCellValue(dl.sheetName, "C"+row, sc)

	m = dl.parent.AnnualizedReturnList[3:5] // 1month, 3month
	mean, stddev = stat.MeanStdDev(m, nil)
	consistency = 1 - stddev
	sc = mean * consistency
	f.SetCellValue(dl.sheetName, "D"+row, sc)

	m = dl.parent.AnnualizedReturnList[3:6] // 1month, 3month, 6month
	mean, stddev = stat.MeanStdDev(m, nil)
	consistency = 1 - stddev
	sc = mean * consistency
	f.SetCellValue(dl.sheetName, "E"+row, sc)

	m = dl.parent.AnnualizedReturnList[3:7] // 1month, 3month, 6month, 9month
	mean, stddev = stat.MeanStdDev(m, nil)
	consistency = 1 - stddev
	sc = mean * consistency
	f.SetCellValue(dl.sheetName, "F"+row, sc)

	m = dl.parent.AnnualizedReturnList[8:10] // 2024ytd,2023
	mean, stddev = stat.MeanStdDev(m, nil)
	consistency = 1 - stddev
	sc = mean * consistency
	f.SetCellValue(dl.sheetName, "G"+row, sc)

	m = dl.parent.AnnualizedReturnList[8:11] // 2024ytd,2023, 2022
	mean, stddev = stat.MeanStdDev(m, nil)
	consistency = 1 - stddev
	sc = mean * consistency
	f.SetCellValue(dl.sheetName, "H"+row, sc)

	m = dl.parent.AnnualizedReturnList[8:12] // 2024ytd,2023, 2022, 2021
	mean, stddev = stat.MeanStdDev(m, nil)
	consistency = 1 - stddev
	sc = mean * consistency
	f.SetCellValue(dl.sheetName, "I"+row, sc)

	m = dl.parent.AnnualizedReturnList[8:13] // 2024ytd,2023, 2022, 2021, 2020
	mean, stddev = stat.MeanStdDev(m, nil)
	consistency = 1 - stddev
	sc = mean * consistency
	f.SetCellValue(dl.sheetName, "J"+row, sc)

	m = dl.parent.AnnualizedReturnList[8:14] // 2024ytd,2023, 2022, 2021, 2020, 2019
	mean, stddev = stat.MeanStdDev(m, nil)
	consistency = 1 - stddev
	sc = mean * consistency
	f.SetCellValue(dl.sheetName, "K"+row, sc)

	//---------------------------------------------------------------------------
	// ANNUALIZED RETURNS BY PERIOD
	//---------------------------------------------------------------------------
	f.SetCellValue(dl.sheetName, "M"+row, dl.parent.AnnualizedReturnList[0])
	f.SetCellValue(dl.sheetName, "P"+row, dl.parent.AnnualizedReturnList[1])
	f.SetCellValue(dl.sheetName, "S"+row, dl.parent.AnnualizedReturnList[2])
	f.SetCellValue(dl.sheetName, "V"+row, dl.parent.AnnualizedReturnList[3])
	f.SetCellValue(dl.sheetName, "Y"+row, dl.parent.AnnualizedReturnList[4])
	f.SetCellValue(dl.sheetName, "AB"+row, dl.parent.AnnualizedReturnList[5])
	f.SetCellValue(dl.sheetName, "AE"+row, dl.parent.AnnualizedReturnList[6])
	f.SetCellValue(dl.sheetName, "AH"+row, dl.parent.AnnualizedReturnList[7])
	f.SetCellValue(dl.sheetName, "AK"+row, dl.parent.AnnualizedReturnList[8])
	f.SetCellValue(dl.sheetName, "AN"+row, dl.parent.AnnualizedReturnList[9])
	f.SetCellValue(dl.sheetName, "AQ"+row, dl.parent.AnnualizedReturnList[10])
	f.SetCellValue(dl.sheetName, "AT"+row, dl.parent.AnnualizedReturnList[11])
	f.SetCellValue(dl.sheetName, "AW"+row, dl.parent.AnnualizedReturnList[12])
	f.SetCellValue(dl.sheetName, "AZ"+row, dl.parent.AnnualizedReturnList[13])

	//---------------------------------------------------------------------------
	// DAILY CONSISTENCY AND SUCCESS COEFFICIENT BY PERIOD
	//---------------------------------------------------------------------------
	var r, c int
	if c, r, err = excelize.CellNameToCoordinates("L" + row); err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < len(dl.parent.cfg.CrucibleSpans); i++ {
		consistency, _ = dl.ConsistencySC(dl.parent.InvestorHistory[i]) // 1 week
		f.SetCellValue(dl.sheetName, dl.getCell(c, r), consistency)
		f.SetCellValue(dl.sheetName, dl.getCell(c+2, r), consistency*dl.parent.AnnualizedReturnList[i])
		c += 3
	}

	//---------------------------------------------------------------------------
	// INVESTOR ID and DNA
	//---------------------------------------------------------------------------
	row = strconv.Itoa(dl.row)
	f.SetCellValue(dl.sheetName, "BB"+row, v.ID)
	f.SetCellValue(dl.sheetName, "BC"+row, dl.cfg.TopInvestors[dl.parent.idx].DNA)

	percentStyle, err := f.NewStyle(&excelize.Style{
		NumFmt: 10, // Number format: 10 corresponds to "0.00%". https://xuri.me/excelize/en/style.html#number_format
		Alignment: &excelize.Alignment{
			WrapText:   true,
			Horizontal: "center",
		},
		Font: &excelize.Font{
			Size: 14,
		},
	})
	if err != nil {
		fmt.Println("Error creating style:", err)
		return
	}

	// Apply the style to a range of cells
	if err := f.SetCellStyle(dl.sheetName, "A"+row, "BA"+row, percentStyle); err != nil {
		fmt.Println("Error applying style:", err)
		return
	}

	if err = f.Save(); err != nil {
		log.Fatal(err)
	}

	dl.row++ // move to the next row
}

// ConsistencySC returns the consistency and success coefficient
func (dl *DNALog) ConsistencySC(m []float64) (float64, float64) {
	mean, stddev := stat.MeanStdDev(m, nil)
	consistency := 1 - stddev
	sc := mean * consistency
	return consistency, sc
}

// getCell returns the cell name without having to deal with the error
func (dl *DNALog) getCell(c, r int) string {
	newCell, err := excelize.CoordinatesToCellName(c, r)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return newCell
}
