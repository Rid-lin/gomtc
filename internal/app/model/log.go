package model

type LogLine struct {
	TimeStart string
	TimeEnd   string
	Init      string
	Message   string
}

type DisplayLog struct {
	Header string
	Logs   []*LogLine
}
