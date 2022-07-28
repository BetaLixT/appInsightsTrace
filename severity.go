package appinsightstrace

type SeverityLevel int

const (
	Verbose     SeverityLevel = 0
	Information SeverityLevel = 1
	Warning     SeverityLevel = 2
	Error       SeverityLevel = 3
	Critical    SeverityLevel = 4
)
