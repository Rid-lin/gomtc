package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type StatType struct {
	// PerMinute       [24][60]uint64
	PerHour         [24]uint64
	Precent         float64
	SizeOfPrecentil uint64
	Average         uint64
	VolumePerDay    uint64
	VolumePerCheck  uint64
	Count           uint32
}

func (l LineOfLogType) ToStatDevice() *StatDevice {
	perHour := [24]uint64{}
	perHour[l.Hour] = l.SizeInBytes
	lineStat := StatDevice{
		Ip:             l.Ipaddress,
		Mac:            l.Login,
		PerHour:        perHour,
		VolumePerDay:   l.SizeInBytes,
		VolumePerCheck: l.SizeInBytes,
		Year:           l.Year,
		Month:          l.Month,
		Day:            l.Day,
		Hour:           l.Hour,
	}
	return &lineStat
}

type Count struct {
	LineParsed,
	LineSkiped,
	LineAdded,
	LineRead,
	LineError,
	TotalLineRead,
	TotalLineAdded,
	TotalLineParsed,
	TotalLineSkiped,
	TotalLineError uint64
}

type LineOfDisplay struct {
	Alias string
	Login string
	InfoType
	StatDeviceType
}

// type lineOfLogType struct {
// 	date        string // squid timestamp 1621969229.000
// 	ipaddress   string
// 	httpstatus  string
// 	method      string
// 	siteName    string
// 	login       string
// 	mime        string
// 	alias       string
// 	hostname    string
// 	comments    string
// 	year        int
// 	day         int
// 	hour        int
// 	minute      int
// 	nsec        int64
// 	timestamp   int64
// 	month       time.Month
// 	sizeInBytes uint64
// }

type InfoType struct {
	InfoName string
	DeviceType
	PersonType
	QuotaType
}

func ParseLineToStruct(line string) (LineOfLogType, error) {
	var l LineOfLogType
	var err error
	valueArray := strings.Fields(line) // split into fields separated by a space to parse into a structure
	if len(valueArray) < 10 {          // check the length of the resulting array to make sure that the string is parsed normally and there are no errors in its format
		return l, fmt.Errorf("Error, string(%v) is not line of Squid-log", line) // If there is an error, then we stop working to avoid unnecessary transformations
	}
	l.Date = valueArray[0]
	l.Timestamp, l.Nsec, err = squidDateToINT64(l.Date)
	if err != nil {
		return LineOfLogType{}, err
	}
	timeUnix := time.Unix(l.Timestamp, 0)
	l.Year = timeUnix.Year()
	l.Month = timeUnix.Month()
	l.Day = timeUnix.Day()
	l.Hour = timeUnix.Hour()
	l.Minute = timeUnix.Minute()
	l.Ipaddress = valueArray[2]
	l.Httpstatus = valueArray[3]
	sizeInBytes, err := strconv.ParseUint(valueArray[4], 10, 64)
	if err != nil {
		sizeInBytes = 0
	}
	l.SizeInBytes = sizeInBytes
	l.Method = valueArray[5]
	l.SiteName = valueArray[6]
	l.Login = valueArray[7]
	l.Mime = valueArray[9]
	if len(valueArray) > 10 {
		l.Hostname = valueArray[10]
	} else {
		l.Hostname = ""
	}
	if len(valueArray) > 11 {
		l.Comments = strings.Join(valueArray[11:], " ")
	} else {
		l.Comments = ""
	}
	return l, nil
}

func DeterminingAlias(value LineOfLogType) string {
	alias := value.Alias
	if alias == "" {
		if value.Login == "" || value.Login == "-" {
			alias = value.Ipaddress
		} else {
			alias = value.Login
		}
	}
	return alias
}

func (count *Count) SumAndReset() {
	count.TotalLineRead = count.TotalLineRead + count.LineRead
	count.TotalLineParsed = count.TotalLineParsed + count.LineParsed
	count.TotalLineAdded = count.TotalLineAdded + count.LineAdded
	count.TotalLineSkiped = count.TotalLineSkiped + count.LineSkiped
	count.TotalLineError = count.TotalLineError + count.LineError
	count.LineParsed = 0
	count.LineSkiped = 0
	count.LineAdded = 0
	count.LineRead = 0
	count.LineError = 0
}

func squidDateToINT64(squidDate string) (timestamp, nsec int64, err error) {
	timestampStr := strings.Split(squidDate, ".")
	timestampStrSec := timestampStr[0]
	if len(timestampStrSec) > 10 {
		timestampStrSec = timestampStrSec[len(timestampStrSec)-10:]
	}
	timestamp, err = strconv.ParseInt(timestampStrSec, 10, 64)
	if err != nil {
		return 0, 0, err
	}
	if len(timestampStr) > 1 {
		nsec, err = strconv.ParseInt(timestampStr[1], 10, 64)
		if err != nil {
			return 0, 0, err
		}
	}
	return
}
