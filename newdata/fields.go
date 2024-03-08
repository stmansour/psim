package newdata

// FieldSelector defines how to request a field in a Select call
// ---------------------------------------------------------------
type FieldSelector struct {
	Metric       string
	MID          int
	Locale       string
	LID          int
	Locale2      string
	LID2         int
	Table        string
	BucketNumber int
	FQname       string
}

// FQMetric returns a fully-qualified metric name from the struct data
func (f *FieldSelector) FQMetric() string {
	if len(f.FQname) > 0 {
		return f.FQname
	}
	s := ""
	if len(f.Locale) > 0 {
		s += f.Locale
	}
	if len(f.Locale2) > 0 {
		s += f.Locale2
	}
	s += f.Metric
	f.FQname = s
	return s
}
