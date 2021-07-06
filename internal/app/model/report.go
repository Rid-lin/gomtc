package model

import "time"

type LineOfLogType struct {
	Date        string // squid timestamp 1621969229.000
	Ipaddress   string
	Httpstatus  string
	Method      string
	SiteName    string
	Login       string
	Mime        string
	Alias       string
	Hostname    string
	Comments    string
	Year        int
	Day         int
	Hour        int
	Minute      int
	Nsec        int64
	Timestamp   int64
	Month       time.Month
	SizeInBytes uint64
}

type DisplayDataType struct {
	ArrayDisplay   []LineOfDisplay
	Header         string
	DateFrom       string
	DateTo         string
	LastUpdated    string
	LastUpdatedMT  string
	Path           string
	Host           string
	ReferURL       string
	TimeToGenerate time.Duration
	Author
	SizeOneType
	QuotaType
}

type SizeOneType struct {
	SizeOneKilobyte uint64
	SizeOneMegabyte uint64
	SizeOneGigabyte uint64
}
type DisplayDataUserType struct {
	Header          string
	Copyright       string
	Mail            string
	SizeOneKilobyte uint64
	InfoType
}

type RequestForm struct {
	DateFrom string
	DateTo   string
	Path     string
	ReferURL string
	Report   string
}

type Author struct {
	Copyright string
	Mail      string
}

func add(slice []LineOfDisplay, line LineOfDisplay) []LineOfDisplay {
	for index, item := range slice {
		if line.Alias == item.Alias {
			slice[index].PerHour = line.PerHour
			return slice
		}
	}
	return append(slice, line)
}

func (rq *RequestForm) ToLine() *LineOfLogType {
	l := LineOfLogType{}
	tn, err := time.Parse("2006-01-02", rq.DateFrom)
	if err != nil {
		tn = time.Now()
	}
	l.Year = tn.Year()
	l.Month = tn.Month()
	l.Day = tn.Day()
	l.Hour = tn.Hour()
	l.Minute = tn.Minute()
	return &l
}
