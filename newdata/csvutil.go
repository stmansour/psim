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
		s += fmt.Sprintf("    %s: %9.3f\n", k, v.Value)
	}
	return s + "\n"
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
