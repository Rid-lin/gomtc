package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (t *Transport) runOnce(cfg *Config) {
	p := parseType{}
	t.RLock()
	p.SSHCredentials = t.sshCredentials
	p.BlockAddressList = t.BlockAddressList
	p.QuotaType = t.QuotaType
	p.Location = t.Location
	t.RUnlock()

	t.readLog(cfg)

	t.getDevices()
	t.delOldData(t.newCount.LastDateNew, t.Location)
	t.parseAllFilesAndCountingTraffic(cfg)
	t.updateAliases(p)
	t.checkQuotas()
	t.SendGroupStatus()

	t.writeLog(cfg)
	t.newCount.Count = Count{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	t.setTimerParse(cfg.ParseDelay)
}

func (t *Transport) setTimerParse(IntervalStr string) {
	interval, err := time.ParseDuration(IntervalStr)
	if err != nil {
		t.timerParse = time.NewTimer(15 * time.Minute)
	} else {
		t.timerParse = time.NewTimer(interval)
	}
}

func replaceQuotes(lineOld string) string {
	lineNew := strings.ReplaceAll(lineOld, "'", "&quot")
	line := strings.ReplaceAll(lineNew, `"`, "&quot")
	return line
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

func parseLineToStruct(line string, cfg *Config) (lineOfLogType, error) {
	var l lineOfLogType
	var err error
	valueArray := strings.Fields(line) // split into fields separated by a space to parse into a structure
	if len(valueArray) < 10 {          // check the length of the resulting array to make sure that the string is parsed normally and there are no errors in its format
		return l, fmt.Errorf("Error, string(%v) is not line of Squid-log", line) // If there is an error, then we stop working to avoid unnecessary transformations
	}
	l.date = valueArray[0]
	l.timestamp, l.nsec, err = squidDateToINT64(l.date)
	if err != nil {
		return lineOfLogType{}, err
	}
	timeUnix := time.Unix(l.timestamp, 0)
	l.year = timeUnix.Year()
	l.month = timeUnix.Month()
	l.day = timeUnix.Day()
	l.hour = timeUnix.Hour()
	l.minute = timeUnix.Minute()
	l.ipaddress = valueArray[2]
	l.httpstatus = valueArray[3]
	sizeInBytes, err := strconv.ParseUint(valueArray[4], 10, 64)
	if err != nil {
		sizeInBytes = 0
	}
	l.sizeInBytes = sizeInBytes
	l.method = valueArray[5]
	l.siteName = valueArray[6]
	l.login = valueArray[7]
	l.mime = valueArray[9]
	if len(valueArray) > 10 {
		l.hostname = valueArray[10]
	} else {
		l.hostname = ""
	}
	if len(valueArray) > 11 {
		l.comments = strings.Join(valueArray[11:], " ")
	} else {
		l.comments = ""
	}
	return l, nil
}

func determiningAlias(value lineOfLogType) string {
	alias := value.alias
	if alias == "" {
		if value.login == "" || value.login == "-" {
			alias = value.ipaddress
		} else {
			alias = value.login
		}
	}
	return alias
}

func SortFileByModTime(files []os.FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Unix() < files[j].ModTime().Unix()
	})
}

func (count *Count) SumAndReset() {
	count.totalLineRead = count.totalLineRead + count.LineRead
	count.totalLineParsed = count.totalLineParsed + count.LineParsed
	count.totalLineAdded = count.totalLineAdded + count.LineAdded
	count.totalLineSkiped = count.totalLineSkiped + count.LineSkiped
	count.totalLineError = count.totalLineError + count.LineError
	count.LineParsed = 0
	count.LineSkiped = 0
	count.LineAdded = 0
	count.LineRead = 0
	count.LineError = 0
}
