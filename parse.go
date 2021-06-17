package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type StatOfYearType struct {
	monthsStat map[time.Month]StatOfMonthType // Statistics of all devices that were seen that year
	year       int
}

type StatOfMonthType struct {
	daysStat map[int]StatOfDayType // Statistics of all devices that were seen that month
	month    time.Month
}

type StatOfDayType struct {
	devicesStat map[KeyDevice]StatDeviceType // Statistics of all devices that were seen that day
	day         int
	StatType    // General statistics of the day, to speed up access
}

type KeyDevice struct {
	// ip  string
	mac string
}

type StatDeviceType struct {
	mac string
	ip  string
	StatType
}

func (t *Transport) parseAllFilesAndCountingTraffic(cfg *Config) {
	// Getting the current time to calculate the running time
	t.newCount.startTime = time.Now()
	fmt.Printf("Parsing has started.\r")
	err := t.parseDirToMapNew(cfg)
	if err != nil {
		log.Error(err)
	}
	ExTime := time.Since(t.newCount.startTime)
	ExTimeInSec := uint64(ExTime.Seconds())
	if ExTimeInSec == 0 {
		ExTimeInSec = 1
	}
	t.newCount.endTime = time.Now() // Saves the current time to be inserted into the log table
	t.newCount.lastUpdated = time.Now()
	log.Infof("The parsing started at %v, ended at %v, and lasted %.3v seconds at a rate of %v lines per second.",
		t.newCount.startTime.In(cfg.Location).Format(DateTimeLayout),
		t.newCount.endTime.In(cfg.Location).Format(DateTimeLayout),
		ExTime.Seconds(),
		t.newCount.totalLineParsed/ExTimeInSec)
	fmt.Printf("The parsing started at %v, ended at %v, and lasted %.3v seconds at a rate of %v lines per second.\n",
		t.newCount.startTime.In(cfg.Location).Format(DateTimeLayout),
		t.newCount.endTime.In(cfg.Location).Format(DateTimeLayout),
		ExTime.Seconds(),
		t.newCount.totalLineParsed/ExTimeInSec)
}

func (t *Transport) delOldData(timestamp int64, Location *time.Location) {
	tn := time.Unix(timestamp, 0).In(Location)
	year := tn.Year()
	month := tn.Month()
	day := tn.Day()
	t.Lock()
	defer t.Unlock()
	yearStat, ok := t.statofYears[year]
	if !ok {
		return
	}
	monthStat, ok := yearStat.monthsStat[month]
	if !ok {
		return
	}
	_, ok = monthStat.daysStat[day]
	if !ok {
		return
	}
	delete(t.statofYears[year].monthsStat[month].daysStat, day)
	t.newCount.LastDateNew = time.Date(year, month, day, 0, 0, 0, 1, t.Location).Unix()
	t.newCount.LastDayStrNew = time.Date(year, month, day, 0, 0, 0, 1, t.Location).String()
}

func (t *Transport) parseDirToMapNew(cfg *Config) error {
	// iteration over all files in a folder
	files, err := ioutil.ReadDir(cfg.LogPath)
	if err != nil {
		return err
	}
	SortFileByModTime(files)
	for _, file := range files {
		if err := t.parseFileToMapNew(file, cfg); err != nil {
			log.Error(err)
			continue
		}
		fmt.Printf("From file %v lines Read:%v/Parsed:%v/Added:%v/Skiped:%v/Error:%v\n",
			file.Name(),
			t.newCount.LineRead,
			t.newCount.LineParsed,
			t.newCount.LineAdded,
			t.newCount.LineSkiped,
			t.newCount.LineError)
		t.newCount.SumAndReset()
	}
	fmt.Printf("From all files lines Read:%v/Parsed:%v/Added:%v/Skiped:%v/Error:%v\n",
		// Lines read:%v, parsed:%v, lines added:%v lines skiped:%v lines error:%v",
		t.newCount.totalLineRead,
		t.newCount.totalLineParsed,
		t.newCount.totalLineAdded,
		t.newCount.totalLineSkiped,
		t.newCount.totalLineError)
	return nil
}

func (t *Transport) parseFileToMapNew(info os.FileInfo, cfg *Config) error {
	if !strings.HasPrefix(info.Name(), cfg.FnStartsWith) {
		return fmt.Errorf("%s don't match to paramert fnStartsWith='%s'", info.Name(), cfg.FnStartsWith)
	}
	FullFileName := filepath.Join(cfg.LogPath, info.Name())
	file, err := os.Open(FullFileName)
	if err != nil {
		file.Close()
		return fmt.Errorf("Error opening squid log file(FullFileName):%v", err)
	}
	defer file.Close()
	// Проверить является ли файл архивом
	_, errGZip := gzip.NewReader(file)
	// Если файл читается с ошибками и это не ошибка gzip.ErrHeader, то возвращаем ошибку
	if errGZip != nil && errGZip != gzip.ErrHeader {
		return errGZip
		// Если нет ошибок, то извлекаем файл во временную папку и передаём в след.шаг имя файла
	} else if errGZip == nil {
		dir, err := ioutil.TempDir("", "gomtc")
		if err != nil {
			log.Errorf("Error extracting file to temporary folder:%v", err)
		}
		defer os.RemoveAll(dir) // очистка
		FileName, err := unGzip(FullFileName, dir)
		if err != nil {
			return err
		}
		FullFileName = FileName
		// FullFileName = filepath.Join(path, FileName)
	}
	// Если ошибка gzip.ErrHeader, то обрабатываем как текстовый файл.
	file.Close()
	file, err = os.Open(FullFileName)
	if err != nil {
		file.Close()
		return fmt.Errorf("Error opening squid log file(FullFileName):%v", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if err := t.parseLogToArrayByLineNew(scanner, cfg); err != nil {
		return err
	}
	if scanner.Err() != nil {
		return scanner.Err()
	}
	return nil
}

func (t *Transport) parseLogToArrayByLineNew(scanner *bufio.Scanner, cfg *Config) error {
	for scanner.Scan() { // We go through the entire file to the end
		t.newCount.LineRead++
		line := scanner.Text() // get the text from the line, for simplicity
		line = filtredMessage(line, cfg.IgnorList)
		if line == "" {
			t.newCount.LineSkiped++
			continue
		}
		line = replaceQuotes(line)
		l, err := parseLineToStruct(line, cfg)
		if err != nil {
			t.newCount.LineError++
			log.Warningf("%v", err)
			continue
		}
		t.newCount.LineParsed++
		if t.newCount.LastDateNew > l.timestamp {
			log.Tracef("line(%v) too old\r", line)
			t.newCount.LineSkiped++
			continue
		} else if t.newCount.LastDateNew < l.timestamp {
			t.newCount.LastDateNew = l.timestamp
		}

		// The main function of filling the database
		// Adding a row to the database for counting traffic
		t.addLineOutToMapOfReportsSuperNew(&l)
		t.newCount.LineAdded++
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		log.Errorf("Error scanner :%v", err)
		return err
	}
	return nil
}

func (t *Transport) addLineOutToMapOfReportsSuperNew(l *lineOfLogType) {
	l.alias = determiningAlias(*l)
	// Идея такая.
	// посчитать статистику для каждого отдельного случая, когда:
	// есть и мак и айпи, есть только айпи, есть только мак
	// записать это в слайс и привязать к отдельному оборудованию.
	// при чём по привязка только айпи адресу идёт только в течении сегодняшнего дня, потом не учитывается
	// привязка по маку и мак+айпи идёт всегда, т.к. устройство опознано.
	t.Lock()
	year, ok := t.statofYears[l.year]
	if !ok {
		year = StatOfYearType{
			year:       l.year,
			monthsStat: map[time.Month]StatOfMonthType{},
		}
		t.statofYears[l.year] = year
	}
	month, ok := year.monthsStat[l.month]
	if !ok {
		month = StatOfMonthType{
			daysStat: map[int]StatOfDayType{},
			month:    l.month}
		year.monthsStat[l.month] = month
	}
	daysStat, ok := month.daysStat[l.day]
	if !ok {
		daysStat = StatOfDayType{
			devicesStat: map[KeyDevice]StatDeviceType{},
			day:         l.day}
		month.daysStat[l.day] = daysStat
	}
	deviceStat, ok := daysStat.devicesStat[KeyDevice{
		// ip: l.ipaddress,
		mac: l.login}]
	if !ok {
		deviceStat = StatDeviceType{mac: l.login, ip: l.ipaddress}
		daysStat.devicesStat[KeyDevice{
			// ip: l.ipaddress,
			mac: l.login}] = deviceStat
	}
	// Расчет суммы трафика для устройства для дальшейшего отображения
	deviceStat.VolumePerDay = deviceStat.VolumePerDay + l.sizeInBytes
	deviceStat.VolumePerCheck = deviceStat.VolumePerCheck + l.sizeInBytes
	deviceStat.StatPerHour[l.hour].PerHour = deviceStat.StatPerHour[l.hour].PerHour + l.sizeInBytes
	deviceStat.StatPerHour[l.hour].PerMinute[l.minute] = deviceStat.StatPerHour[l.hour].PerMinute[l.minute] + l.sizeInBytes
	// Расчет суммы трафика для дня для дальшейшего отображения
	daysStat.VolumePerDay = daysStat.VolumePerDay + l.sizeInBytes
	daysStat.VolumePerCheck = daysStat.VolumePerCheck + l.sizeInBytes
	daysStat.StatPerHour[l.hour].PerHour = daysStat.StatPerHour[l.hour].PerHour + l.sizeInBytes
	daysStat.StatPerHour[l.hour].PerMinute[l.minute] = daysStat.StatPerHour[l.hour].PerMinute[l.minute] + l.sizeInBytes
	// Возвращаем данные обратно
	daysStat.devicesStat[KeyDevice{
		// ip: l.ipaddress,
		mac: l.login}] = deviceStat
	month.daysStat[l.day] = daysStat
	year.monthsStat[l.month] = month
	t.statofYears[l.year] = year
	t.Unlock()
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
