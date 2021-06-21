package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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
	err := t.parseDirToMap(cfg)
	if err != nil {
		log.Error(err)
	}
}

// func (t *Transport) delOldData(timestamp int64) {
// 	tn := time.Unix(timestamp, 0).In(Location)
// 	year := tn.Year()
// 	month := tn.Month()
// 	day := tn.Day()
// 	t.Lock()
// 	defer t.Unlock()
// 	yearStat, ok := t.statofYears[year]
// 	if !ok {
// 		return
// 	}
// 	monthStat, ok := yearStat.monthsStat[month]
// 	if !ok {
// 		return
// 	}
// 	_, ok = monthStat.daysStat[day]
// 	if !ok {
// 		return
// 	}
// 	delete(t.statofYears[year].monthsStat[month].daysStat, day)
// 	t.LastDate = time.Date(year, month, day, 0, 0, 0, 1, Location).Unix()
// 	t.LastDayStr = time.Date(year, month, day, 0, 0, 0, 1, Location).String()
// }

func (t *Transport) parseDirToMap(cfg *Config) error {
	// iteration over all files in a folder
	files, err := ioutil.ReadDir(cfg.LogPath)
	if err != nil {
		return err
	}
	SortFileByModTime(files)
	for _, file := range files {
		if err := t.parseFileToMap(file, cfg); err != nil {
			log.Error(err)
			continue
		}
		fmt.Printf("From file %v lines Read:%v/Parsed:%v/Added:%v/Skiped:%v/Error:%v\n",
			file.Name(),
			t.LineRead,
			t.LineParsed,
			t.LineAdded,
			t.LineSkiped,
			t.LineError)
		t.SumAndReset()
	}
	fmt.Printf("From all files lines Read:%v/Parsed:%v/Added:%v/Skiped:%v/Error:%v\n",
		// Lines read:%v, parsed:%v, lines added:%v lines skiped:%v lines error:%v",
		t.totalLineRead,
		t.totalLineParsed,
		t.totalLineAdded,
		t.totalLineSkiped,
		t.totalLineError)
	return nil
}

func (t *Transport) parseFileToMap(info os.FileInfo, cfg *Config) error {
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
	if err := t.parseOneFilesAndCountingTraffic(FullFileName, cfg); err != nil {
		return err
	}
	return nil
}

func (t *Transport) parseLogToArrayByLine(scanner *bufio.Scanner, cfg *Config) error {
	for scanner.Scan() { // We go through the entire file to the end
		t.LineRead++
		line := scanner.Text() // get the text from the line, for simplicity
		line = filtredMessage(line, cfg.IgnorList)
		if line == "" {
			t.LineSkiped++
			continue
		}
		line = replaceQuotes(line)
		l, err := parseLineToStruct(line, cfg)
		if err != nil {
			t.LineError++
			log.Warningf("%v", err)
			continue
		}
		t.LineParsed++
		if t.LastDate > l.timestamp {
			log.Tracef("line(%v) too old\r", line)
			t.LineSkiped++
			continue
		} else if t.LastDate < l.timestamp {
			t.LastDate = l.timestamp
		}

		// The main function of filling the database
		// Adding a row to the database for counting traffic
		t.addLineOutToMapOfReportsSuperNew(&l)
		t.LineAdded++
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
	deviceStat.PerHour[l.hour] = deviceStat.PerHour[l.hour] + l.sizeInBytes
	// deviceStat.PerMinute[l.hour][l.minute] = deviceStat.PerMinute[l.hour][l.minute] + l.sizeInBytes
	// Расчет суммы трафика для дня для дальшейшего отображения
	// daysStat.VolumePerDay = daysStat.VolumePerDay + l.sizeInBytes
	// daysStat.VolumePerCheck = daysStat.VolumePerCheck + l.sizeInBytes
	// daysStat.PerHour[l.hour] = daysStat.PerHour[l.hour] + l.sizeInBytes
	// daysStat.PerMinute[l.hour][l.minute] = daysStat.PerMinute[l.hour][l.minute] + l.sizeInBytes
	// Возвращаем данные обратно
	daysStat.devicesStat[KeyDevice{
		// ip: l.ipaddress,
		mac: l.login}] = deviceStat
	month.daysStat[l.day] = daysStat
	year.monthsStat[l.month] = month
	t.statofYears[l.year] = year
	t.Unlock()
}

func (t *Transport) checkQuotas(cfg *Config) {
	t.Lock()
	tn := time.Now().In(Location)
	hour := tn.Hour()
	tns := tn.Format(DateLayout)
	devicesStat := GetDayStat(tns, tns, path.Join(cfg.ConfigPath, "sqlite.db"))
	for key, d := range devicesStat {
		alias := t.Aliases[key.mac]
		if d.VolumePerDay >= alias.DailyQuota || d.PerHour[hour] >= alias.HourlyQuota {
			alias.ShouldBeBlocked = true
		} else {
			alias.ShouldBeBlocked = false
		}
		t.Aliases[alias.AliasName] = alias
	}
	t.Unlock()
}

// func (t *Transport) checkQuotas(cfg *Config) {
// 	t.Lock()
// 	tn := time.Now().In(Location)
// 	hour := tn.Hour()
// 	tns := tn.Format(DateLayout)
// 	devicesStat := GetDayStat(tns, tns, path.Join(cfg.ConfigPath, "sqlite.db"))
// 	for _, alias := range t.Aliases {
// 		var VolumePerDay, VolumePerCheck uint64
// 		var StatPerHour [24]VolumePerType
// 		for _, key := range alias.KeyArr {
// 			VolumePerDay += devicesStat[key].VolumePerDay
// 			VolumePerCheck += devicesStat[key].VolumePerCheck
// 			for index := range devicesStat[key].PerHour {
// 				StatPerHour[index].PerHour += devicesStat[key].PerHour[index]
// 			}
// 		}
// 		if VolumePerDay >= alias.DailyQuota || StatPerHour[hour].PerHour >= alias.HourlyQuota {
// 			alias.ShouldBeBlocked = true
// 		} else {
// 			alias.ShouldBeBlocked = false
// 		}
// 		t.Aliases[alias.AliasName] = alias
// 	}
// 	t.Unlock()
// }

// func (t *Transport) GetDayStat(l *lineOfLogType) map[KeyDevice]StatDeviceType {
// 	t.Lock()
// 	year, ok := t.statofYears[l.year]
// 	if !ok {
// 		year = StatOfYearType{
// 			year:       l.year,
// 			monthsStat: map[time.Month]StatOfMonthType{},
// 		}
// 		t.statofYears[l.year] = year
// 	}
// 	month, ok := year.monthsStat[l.month]
// 	if !ok {
// 		month = StatOfMonthType{
// 			daysStat: map[int]StatOfDayType{},
// 			month:    l.month}
// 		year.monthsStat[l.month] = month
// 	}
// 	day, ok := month.daysStat[l.day]
// 	if !ok {
// 		day = StatOfDayType{
// 			devicesStat: map[KeyDevice]StatDeviceType{},
// 			day:         l.day}
// 		month.daysStat[l.day] = day
// 	}
// 	t.Unlock()
// 	return day.devicesStat
// }

func (t *Transport) BlockDevices() {
	t.Lock()
	for _, d := range t.devices {
		if d.Manual {
			continue
		}
		key := KeyDevice{}
		switch {
		case d.ActiveMacAddress != "":
			key = KeyDevice{mac: d.ActiveMacAddress}
		case d.macAddress != "":
			key = KeyDevice{mac: d.macAddress}
		}
		// TODO подумать над преобразованием ClientID в мак адрес
		switch {
		case (d.Blocked && t.Aliases[key.mac].ShouldBeBlocked) || (!d.Blocked && !t.Aliases[key.mac].ShouldBeBlocked):
			continue
		case d.Blocked && !t.Aliases[key.mac].ShouldBeBlocked:
			d = d.UnBlock(t.BlockAddressList, key)
			t.change[key] = DeviceToBlock{
				Id:       d.Id,
				Mac:      key.mac,
				IP:       d.ActiveAddress,
				Groups:   d.AddressLists,
				Disabled: paramertToBool(d.disabledL),
			}
		case !d.Blocked && t.Aliases[key.mac].ShouldBeBlocked:
			d = d.Block(t.BlockAddressList, key)
			t.change[key] = DeviceToBlock{
				Id:       d.Id,
				Mac:      key.mac,
				IP:       d.ActiveAddress,
				Groups:   d.AddressLists,
				Disabled: paramertToBool(d.disabledL),
			}
		}
	}
	t.Unlock()
}

func (t *Transport) parseOneFilesAndCountingTraffic(FullFileName string, cfg *Config) error {
	file, err := os.Open(FullFileName)
	if err != nil {
		file.Close()
		return fmt.Errorf("Error opening squid log file(FullFileName):%v", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if err := t.parseLogToArrayByLine(scanner, cfg); err != nil {
		return err
	}
	if scanner.Err() != nil {
		return scanner.Err()
	}
	return nil
}

func (t *Transport) timeCalculationAndPrinting() {
	ExTime := time.Since(t.startTime)
	ExTimeInSec := uint64(ExTime.Seconds())
	if ExTimeInSec == 0 {
		ExTimeInSec = 1
	}
	t.endTime = time.Now() // Saves the current time to be inserted into the log table
	t.LastUpdated = time.Now()
	log.Infof("The parsing started at %v, ended at %v, lasted %.3v seconds at a rate of %v lines per second.",
		t.startTime.In(Location).Format(DateTimeLayout),
		t.endTime.In(Location).Format(DateTimeLayout),
		ExTime.Seconds(),
		t.totalLineParsed/ExTimeInSec)
	fmt.Printf("The parsing started at %v, ended at %v, lasted %.3v seconds at a rate of %v lines per second.\n",
		t.startTime.In(Location).Format(DateTimeLayout),
		t.endTime.In(Location).Format(DateTimeLayout),
		ExTime.Seconds(),
		t.totalLineParsed/ExTimeInSec)
}
