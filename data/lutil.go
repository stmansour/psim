package data

import "fmt"

// GetLValue returns a particular field from the supplied lrec
func GetLValue(rec *LinguisticDataRecord, cur, droot string) (float64, error) {
	columnName := "L" + cur + droot
	return GetValueByColName(rec, columnName)
}

// GetValueByColName returns a particular field from the supplied lrec
func GetValueByColName(rec *LinguisticDataRecord, columnName string) (float64, error) {
	switch columnName {
	case "LALLLSNScore":
		return rec.LALLLSNScore, nil
	case "LALLLSPScore":
		return rec.LALLLSPScore, nil
	case "LALLWHAScore":
		return rec.LALLWHAScore, nil
	case "LALLWHOScore":
		return rec.LALLWHOScore, nil
	case "LALLWHLScore":
		return rec.LALLWHLScore, nil
	case "LALLWPAScore":
		return rec.LALLWPAScore, nil
	case "LALLWDECount":
		return rec.LALLWDECount, nil
	case "LALLWDFCount":
		return rec.LALLWDFCount, nil
	case "LALLWDPCount":
		return rec.LALLWDPCount, nil
	case "LALLWDMCount":
		return rec.LALLWDMCount, nil
	case "LUSDLSNScore_ECON":
		return rec.LUSDLSNScore_ECON, nil
	case "LUSDLSPScore_ECON":
		return rec.LUSDLSPScore_ECON, nil
	case "LUSDWHAScore_ECON":
		return rec.LUSDWHAScore_ECON, nil
	case "LUSDWHOScore_ECON":
		return rec.LUSDWHOScore_ECON, nil
	case "LUSDWHLScore_ECON":
		return rec.LUSDWHLScore_ECON, nil
	case "LUSDWPAScore_ECON":
		return rec.LUSDWPAScore_ECON, nil
	case "LUSDWDECount_ECON":
		return rec.LUSDWDECount_ECON, nil
	case "LUSDWDFCount_ECON":
		return rec.LUSDWDFCount_ECON, nil
	case "LUSDWDPCount_ECON":
		return rec.LUSDWDPCount_ECON, nil
	case "LUSDLIMCount_ECON":
		return rec.LUSDLIMCount_ECON, nil
	case "LJPYLSNScore_ECON":
		return rec.LJPYLSNScore_ECON, nil
	case "LJPYLSPScore_ECON":
		return rec.LJPYLSPScore_ECON, nil
	case "LJPYWHAScore_ECON":
		return rec.LJPYWHAScore_ECON, nil
	case "LJPYWHOScore_ECON":
		return rec.LJPYWHOScore_ECON, nil
	case "LJPYWHLScore_ECON":
		return rec.LJPYWHLScore_ECON, nil
	case "LJPYWPAScore_ECON":
		return rec.LJPYWPAScore_ECON, nil
	case "LJPYWDECount_ECON":
		return rec.LJPYWDECount_ECON, nil
	case "LJPYWDFCount_ECON":
		return rec.LJPYWDFCount_ECON, nil
	case "LJPYWDPCount_ECON":
		return rec.LJPYWDPCount_ECON, nil
	case "LJPYLIMCount_ECON":
		return rec.LJPYLIMCount_ECON, nil
	case "WTOILClose":
		return rec.WTOILClose, nil
	}
	return 0, fmt.Errorf("field %s not found", columnName)
}
