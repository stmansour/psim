package newcore

import (
	"fmt"
	"log"
	"strconv"

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
	filename  string    // name of report file
	parent    *Crucible // crucible object containing this
	row       int       // current row number
	sheetName string    // current sheet we're working on
	Results   map[string]*DNALogResult
}

// NewDNALog creates and returns a new DNA log object
func NewDNALog() *DNALog {
	return &DNALog{
		Results: make(map[string]*DNALogResult),
	}
}

// Init initializes the object
func (dl *DNALog) Init(c *Crucible, cfg *util.AppConfig) {
	dl.cfg = cfg
	dl.parent = c
}

// WriteHeader writes the DNA log as an Excel spreadsheet file
func (dl *DNALog) WriteHeader() error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

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
	// COLUMN HEADERS
	// Set complex headers and merge cells
	// Row 1 headers
	//===========================================================================
	f.MergeCell(dl.sheetName, "B5", "K5")
	f.SetCellValue(dl.sheetName, "B5", "Rolling Success Coefficients")

	f.MergeCell(dl.sheetName, "L5", "N5")
	f.SetCellValue(dl.sheetName, "L5", "Rolling 1 Week")

	f.MergeCell(dl.sheetName, "O5", "Q5")
	f.SetCellValue(dl.sheetName, "O5", "Rolling 2 Weeks")

	f.MergeCell(dl.sheetName, "R5", "T5")
	f.SetCellValue(dl.sheetName, "R5", "Rolling 3 Weeks")

	f.MergeCell(dl.sheetName, "U5", "W5")
	f.SetCellValue(dl.sheetName, "u5", "Rolling 1 Month")

	f.MergeCell(dl.sheetName, "x5", "z5")
	f.SetCellValue(dl.sheetName, "z5", "Rolling 3 Month")

	f.MergeCell(dl.sheetName, "aa5", "ac5")
	f.SetCellValue(dl.sheetName, "aa5", "Rolling 6 Month")

	f.MergeCell(dl.sheetName, "ad5", "af5")
	f.SetCellValue(dl.sheetName, "ad5", "Rolling 9 Month")

	f.MergeCell(dl.sheetName, "ag5", "ai5")
	f.SetCellValue(dl.sheetName, "ag5", "Rolling 12 Months")

	f.MergeCell(dl.sheetName, "aj5", "al5")
	f.SetCellValue(dl.sheetName, "aj5", "2024 YTD")

	f.MergeCell(dl.sheetName, "am5", "ao5")
	f.SetCellValue(dl.sheetName, "am5", "2023")

	f.MergeCell(dl.sheetName, "ap5", "ar5")
	f.SetCellValue(dl.sheetName, "ap5", "2022")

	f.MergeCell(dl.sheetName, "as5", "au5")
	f.SetCellValue(dl.sheetName, "as5", "2021")

	f.MergeCell(dl.sheetName, "av5", "ax5")
	f.SetCellValue(dl.sheetName, "av5", "2020")

	f.MergeCell(dl.sheetName, "ay5", "ba5")
	f.SetCellValue(dl.sheetName, "ay5", "2019")

	f.SetCellValue(dl.sheetName, "A6", "INVESTOR")
	f.SetCellValue(dl.sheetName, "B6", "Weeks 1-2 Success Coefficient")
	f.SetCellValue(dl.sheetName, "C6", "Weeks 1-3 Success Coefficient")
	f.SetCellValue(dl.sheetName, "D6", "1,3 Month Success Coefficient")
	f.SetCellValue(dl.sheetName, "E6", "1,3,6 Month Success Coefficient")
	f.SetCellValue(dl.sheetName, "F6", "1,3,6,9 Month Success Coefficient")
	f.SetCellValue(dl.sheetName, "G6", "2 Year Success Coefficient 2024 YTD, 2023")
	f.SetCellValue(dl.sheetName, "H6", "3 Year Success Coefficient 2024 YTD, 2023, 2022")
	f.SetCellValue(dl.sheetName, "I6", "4 Year Success Coefficient 2024 YTD, 2023, 2022, 2021")
	f.SetCellValue(dl.sheetName, "J6", "5 Year Success Coefficient 2024 YTD, 2023, 2022, 2021, 2020")
	f.SetCellValue(dl.sheetName, "K6", "6 Year Success Coefficient 2024 YTD, 2023, 2022, 2021, 2020, 2019")

	f.SetCellValue(dl.sheetName, "l6", "Consistency")
	f.SetCellValue(dl.sheetName, "m6", "Annualized Return")
	f.SetCellValue(dl.sheetName, "n6", "Success Coefficient")

	f.SetCellValue(dl.sheetName, "o6", "Consistency")
	f.SetCellValue(dl.sheetName, "p6", "Annualized Return")
	f.SetCellValue(dl.sheetName, "q6", "Success Coefficient")

	f.SetCellValue(dl.sheetName, "r6", "Consistency")
	f.SetCellValue(dl.sheetName, "s6", "Annualized Return")
	f.SetCellValue(dl.sheetName, "t6", "Success Coefficient")

	f.SetCellValue(dl.sheetName, "u6", "Consistency")
	f.SetCellValue(dl.sheetName, "v6", "Annualized Return")
	f.SetCellValue(dl.sheetName, "w6", "Success Coefficient")

	f.SetCellValue(dl.sheetName, "x6", "Consistency")
	f.SetCellValue(dl.sheetName, "y6", "Annualized Return")
	f.SetCellValue(dl.sheetName, "z6", "Success Coefficient")

	f.SetCellValue(dl.sheetName, "aa6", "Consistency")
	f.SetCellValue(dl.sheetName, "ab6", "Annualized Return")
	f.SetCellValue(dl.sheetName, "ac6", "Success Coefficient")

	f.SetCellValue(dl.sheetName, "ad6", "Consistency")
	f.SetCellValue(dl.sheetName, "ae6", "Annualized Return")
	f.SetCellValue(dl.sheetName, "af6", "Success Coefficient")

	f.SetCellValue(dl.sheetName, "ag6", "Consistency")
	f.SetCellValue(dl.sheetName, "ah6", "Annualized Return")
	f.SetCellValue(dl.sheetName, "ai6", "Success Coefficient")

	f.SetCellValue(dl.sheetName, "aj6", "Consistency")
	f.SetCellValue(dl.sheetName, "ak6", "Annualized Return")
	f.SetCellValue(dl.sheetName, "al6", "Success Coefficient")

	f.SetCellValue(dl.sheetName, "am6", "Consistency")
	f.SetCellValue(dl.sheetName, "an6", "Annualized Return")
	f.SetCellValue(dl.sheetName, "ao6", "Success Coefficient")

	f.SetCellValue(dl.sheetName, "ap6", "Consistency")
	f.SetCellValue(dl.sheetName, "aq6", "Annualized Return")
	f.SetCellValue(dl.sheetName, "ar6", "Success Coefficient")

	f.SetCellValue(dl.sheetName, "as6", "Consistency")
	f.SetCellValue(dl.sheetName, "at6", "Annualized Return")
	f.SetCellValue(dl.sheetName, "au6", "Success Coefficient")

	f.SetCellValue(dl.sheetName, "av6", "Consistency")
	f.SetCellValue(dl.sheetName, "aw6", "Annualized Return")
	f.SetCellValue(dl.sheetName, "ax6", "Success Coefficient")

	f.SetCellValue(dl.sheetName, "ay6", "Consistency")
	f.SetCellValue(dl.sheetName, "az6", "Annualized Return")
	f.SetCellValue(dl.sheetName, "ba6", "Success Coefficient")

	f.SetCellValue(dl.sheetName, "bb6", "Investor ID")
	f.SetCellValue(dl.sheetName, "bc6", "Investor DNA")

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

	f.SetCellStyle(dl.sheetName, "A5", "BC6", colHdrStyle)

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
	dl.row = 7 // start on row 7

	// Save the Excel file
	dl.filename = "dnalog.xlsx"
	if err := f.SaveAs(dl.filename); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Excel file created successfully.")
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
	} else {
		invname += " (" + v.ShortID() + ")"
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
		consistency, sc = dl.ConsistencySC(dl.parent.InvestorHistory[i]) // 1 week
		f.SetCellValue(dl.sheetName, dl.getCell(c, r), consistency)
		f.SetCellValue(dl.sheetName, dl.getCell(c+2, r), sc)
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
