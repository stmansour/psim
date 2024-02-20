package newdata

import (
	"bytes"
	"fmt"
)

// DRec2String provides a visual representation of what data is
// available in the data record.
// ------------------------------------------------------------------
func DRec2String(drec *EconometricsRecord) string {
	s := fmt.Sprintf("    Date = %s\n", drec.Date.Format("Jan 2, 2006"))
	for k, v := range drec.Fields {
		s += fmt.Sprintf("    %s: %9.3f\n", k, v)
	}
	return s + "\n"
}

// LRecToString returns a human-readable string for the supplied lrec
func LRecToString(lrec *LinguisticDataRecord) string {
	s := fmt.Sprintf(`    Date = %s
	LALLLSNScore      = %9.3f
	LALLLSPScore      = %9.3f
	LALLWHAScore      = %9.3f
	LALLWHOScore      = %9.3f
	LALLWHLScore      = %9.3f
	LALLWPAScore      = %9.3f
	LALLWDECount      = %9.3f
	LALLWDFCount      = %9.3f
	LALLWDPCount      = %9.3f
	LALLWDMCount      = %9.3f
	LUSDLSNScore_ECON = %9.3f
	LUSDLSPScore_ECON = %9.3f
	LUSDWHAScore_ECON = %9.3f
	LUSDWHOScore_ECON = %9.3f
	LUSDWHLScore_ECON = %9.3f
	LUSDWPAScore_ECON = %9.3f
	LUSDWDECount_ECON = %9.3f
	LUSDWDFCount_ECON = %9.3f
	LUSDWDPCount_ECON = %9.3f
	LUSDLIMCount_ECON = %9.3f
	LJPYLSNScore_ECON = %9.3f
	LJPYLSPScore_ECON = %9.3f
	LJPYWHAScore_ECON = %9.3f
	LJPYWHOScore_ECON = %9.3f
	LJPYWHLScore_ECON = %9.3f
	LJPYWPAScore_ECON = %9.3f
	LJPYWDECount_ECON = %9.3f
	LJPYWDFCount_ECON = %9.3f
	LJPYWDPCount_ECON = %9.3f
	LJPYLIMCount_ECON = %9.3f
	`,
		lrec.Date.Format("Jan 2, 2006"),
		lrec.LALLLSNScore,
		lrec.LALLLSPScore,
		lrec.LALLWHAScore,
		lrec.LALLWHOScore,
		lrec.LALLWHLScore,
		lrec.LALLWPAScore,
		lrec.LALLWDECount,
		lrec.LALLWDFCount,
		lrec.LALLWDPCount,
		lrec.LALLWDMCount,
		lrec.LUSDLSNScore_ECON,
		lrec.LUSDLSPScore_ECON,
		lrec.LUSDWHAScore_ECON,
		lrec.LUSDWHOScore_ECON,
		lrec.LUSDWHLScore_ECON,
		lrec.LUSDWPAScore_ECON,
		lrec.LUSDWDECount_ECON,
		lrec.LUSDWDFCount_ECON,
		lrec.LUSDWDPCount_ECON,
		lrec.LUSDLIMCount_ECON,
		lrec.LJPYLSNScore_ECON,
		lrec.LJPYLSPScore_ECON,
		lrec.LJPYWHAScore_ECON,
		lrec.LJPYWHOScore_ECON,
		lrec.LJPYWHLScore_ECON,
		lrec.LJPYWPAScore_ECON,
		lrec.LJPYWDECount_ECON,
		lrec.LJPYWDFCount_ECON,
		lrec.LJPYWDPCount_ECON,
		lrec.LJPYLIMCount_ECON,
	)
	return s
}

// HandleUTF8FileChars returns the first line of the file with
// utf8 markers removed if they were there. Otherwise, it just
// returns the input string.
// ----------------------------------------------------------------
func HandleUTF8FileChars(line string) string {
	bom := []byte{0xEF, 0xBB, 0xBF}
	strBytes := []byte(line)

	if len(strBytes) >= len(bom) && bytes.Equal(strBytes[:len(bom)], bom) {
		// If the line starts with BOM, remove it.
		line = string(strBytes[len(bom):])
	}
	return line
}
