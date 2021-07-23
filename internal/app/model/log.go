package model

type LogLine struct {
	TimeStart string
	TimeEnd   string
	Message   string
	Source    string
}

type DisplayLog struct {
	Header string
	Logs   []*LogLine
}
