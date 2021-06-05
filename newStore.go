package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

// type statYearType struct {
// 	monthsStat []StatMonthType // Statistics of all devices that were seen that year
// 	year       int
// }

// type StatMonthType struct {
// 	daysStat []StatDayType // Statistics of all devices that were seen that month
// 	month    time.Month
// }

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
	ip  string
	mac string
}

// type StatDayType struct {
// 	devicesStat []StatDeviceType // Statistics of all devices that were seen that day
// 	date        string           // date in format YYYY-Month-DD
// 	day         int
// 	StatType    // General statistics of the day, to speed up access
// }

type StatDeviceType struct {
	Mac string
	IP  string
	StatType
}

func (t *Transport) parseAllFilesAndCountingTrafficNew(cfg *Config) {
	// Getting the current time to calculate the running time
	t.newCount.startTime = time.Now()
	fmt.Printf("Parsing has started.\n")
	t.delOldData(t.newCount.LastDateNew, t.Location)
	// fmt.Printf("After Delete DAta\n")
	// t.Print()
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
		t.newCount.startTime.In(cfg.Location).Format(cfg.dateTimeLayout),
		t.newCount.endTime.In(cfg.Location).Format(cfg.dateTimeLayout),
		ExTime.Seconds(),
		t.newCount.totalLineParsed/ExTimeInSec)
	fmt.Printf("The parsing started at %v, ended at %v, and lasted %.3v seconds at a rate of %v lines per second.\n",
		t.newCount.startTime.In(cfg.Location).Format(cfg.dateTimeLayout),
		t.newCount.endTime.In(cfg.Location).Format(cfg.dateTimeLayout),
		ExTime.Seconds(),
		t.newCount.totalLineParsed/ExTimeInSec)
	t.newCount.Count = Count{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

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
		line = filtredMessage(line, cfg.IgnorList)

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
	deviceStat, ok := daysStat.devicesStat[KeyDevice{ip: l.ipaddress, mac: l.login}]
	if !ok {
		deviceStat = StatDeviceType{Mac: l.login, IP: l.ipaddress}
		daysStat.devicesStat[KeyDevice{ip: l.ipaddress, mac: l.login}] = deviceStat
	}
	// Расчет суммы трафика для устройства для дальшейшего отображения
	deviceStat.VolumePerDay = deviceStat.VolumePerDay + l.sizeInBytes
	deviceStat.VolumePerCheck = deviceStat.VolumePerCheck + l.sizeInBytes
	deviceStat.StatPerHour[l.hour].Hour = deviceStat.StatPerHour[l.hour].Hour + l.sizeInBytes
	deviceStat.StatPerHour[l.hour].Minute[l.minute] = deviceStat.StatPerHour[l.hour].Minute[l.minute] + l.sizeInBytes
	// Расчет суммы трафика для дня для дальшейшего отображения
	daysStat.VolumePerDay = daysStat.VolumePerDay + l.sizeInBytes
	daysStat.VolumePerCheck = daysStat.VolumePerCheck + l.sizeInBytes
	daysStat.StatPerHour[l.hour].Hour = daysStat.StatPerHour[l.hour].Hour + l.sizeInBytes
	daysStat.StatPerHour[l.hour].Minute[l.minute] = daysStat.StatPerHour[l.hour].Minute[l.minute] + l.sizeInBytes
	// Возвращаем данные обратно
	daysStat.devicesStat[KeyDevice{ip: l.ipaddress, mac: l.login}] = deviceStat
	month.daysStat[l.day] = daysStat
	year.monthsStat[l.month] = month
	t.statofYears[l.year] = year
	t.Unlock()
}

// func (t *Transport) addLineOutToMapOfReportsSuperNew(l *lineOfLogType) {
// 	l.alias = determiningAlias(*l)
// 	// Идея такая.
// 	// посчитать статистику для каждого отдельного случая, когда:
// 	// есть и мак и айпи, есть только айпи, есть только мак
// 	// записать это в слайс и привязать к отдельному оборудованию.
// 	// при чём по привязка только айпи адресу идёт только в течении сегодняшнего дня, потом не учитывается
// 	// привязка по маку и мак+айпи идёт всегда, т.к. устройство опознано.
// 	indexYear, indexMonth, indexDay, indexDevice := t.getIndexesOfStat(l)
// 	t.Lock()
// 	devStat := t.statYears[indexYear].monthsStat[indexMonth].daysStat[indexDay].devicesStat[indexDevice]
// 	dayStat := t.statYears[indexYear].monthsStat[indexMonth].daysStat[indexDay]
// 	// Расчет суммы трафика для устройства для дальшейшего отображения
// 	devStat.VolumePerDay = devStat.VolumePerDay + l.sizeInBytes
// 	devStat.VolumePerCheck = devStat.VolumePerCheck + l.sizeInBytes
// 	devStat.StatPerHour[l.hour].Hour = devStat.StatPerHour[l.hour].Hour + l.sizeInBytes
// 	devStat.StatPerHour[l.hour].Minute[l.minute] = devStat.StatPerHour[l.hour].Minute[l.minute] + l.sizeInBytes
// 	// Расчет суммы трафика для дня для дальшейшего отображения
// 	dayStat.VolumePerDay = dayStat.VolumePerDay + l.sizeInBytes
// 	dayStat.VolumePerCheck = dayStat.VolumePerCheck + l.sizeInBytes
// 	dayStat.StatPerHour[l.hour].Hour = dayStat.StatPerHour[l.hour].Hour + l.sizeInBytes
// 	dayStat.StatPerHour[l.hour].Minute[l.minute] = dayStat.StatPerHour[l.hour].Minute[l.minute] + l.sizeInBytes
// 	// Возвращаем данные обратно
// 	dayStat.devicesStat[indexDevice] = devStat
// 	t.statYears[indexYear].monthsStat[indexMonth].daysStat[indexDay] = dayStat
// 	t.Unlock()
// }

// func (t *Transport) getIndexesOfDay(l *lineOfLogType) (int, int, int) {
// 	t.RLock()
// 	defer t.RUnlock()
// 	var indexYear, indexMonth, indexDay int = -1, -1, -1
// 	for index, year := range t.statYears {
// 		if year.year == l.year {
// 			indexYear = index
// 			break
// 		}
// 	}
// 	if indexYear == -1 {
// 		t.statYears = append(t.statYears, statYearType{year: l.year})
// 		indexYear = len(t.statYears) - 1
// 	}
// 	for index, monthStat := range t.statYears[indexYear].monthsStat {
// 		if monthStat.month == l.month {
// 			indexMonth = index
// 			break
// 		}
// 	}
// 	if indexMonth == -1 {
// 		t.statYears[indexYear].monthsStat = append(t.statYears[indexYear].monthsStat, StatMonthType{month: l.month})
// 		indexMonth = len(t.statYears[indexYear].monthsStat) - 1
// 	}
// 	for index, daysStat := range t.statYears[indexYear].monthsStat[indexMonth].daysStat {
// 		if daysStat.day == l.day {
// 			indexDay = index
// 			break
// 		}
// 	}
// 	if indexDay == -1 {
// 		t.statYears[indexYear].monthsStat[indexMonth].daysStat = append(t.statYears[indexYear].monthsStat[indexMonth].daysStat, StatDayType{day: l.day})
// 		indexDay = len(t.statYears[indexYear].monthsStat[indexMonth].daysStat) - 1
// 	}
// 	return indexYear, indexMonth, indexDay
// }

// func (t *Transport) getIndexesOfStat(l *lineOfLogType) (int, int, int, int) {
// 	t.RLock()
// 	defer t.RUnlock()
// 	var indexYear, indexMonth, indexDay, indexDevice int = -1, -1, -1, -1
// 	for index, year := range t.statYears {
// 		if year.year == l.year {
// 			indexYear = index
// 			break
// 		}
// 	}
// 	if indexYear == -1 {
// 		t.statYears = append(t.statYears, statYearType{year: l.year})
// 		indexYear = len(t.statYears) - 1
// 	}
// 	for index, monthStat := range t.statYears[indexYear].monthsStat {
// 		if monthStat.month == l.month {
// 			indexMonth = index
// 			break
// 		}
// 	}
// 	if indexMonth == -1 {
// 		t.statYears[indexYear].monthsStat = append(t.statYears[indexYear].monthsStat, StatMonthType{month: l.month})
// 		indexMonth = len(t.statYears[indexYear].monthsStat) - 1
// 	}
// 	for index, daysStat := range t.statYears[indexYear].monthsStat[indexMonth].daysStat {
// 		if daysStat.day == l.day {
// 			indexDay = index
// 			break
// 		}
// 	}
// 	if indexDay == -1 {
// 		t.statYears[indexYear].monthsStat[indexMonth].daysStat = append(t.statYears[indexYear].monthsStat[indexMonth].daysStat, StatDayType{day: l.day})
// 		indexDay = len(t.statYears[indexYear].monthsStat[indexMonth].daysStat) - 1
// 	}
// 	for index, deviceStat := range t.statYears[indexYear].monthsStat[indexMonth].daysStat[indexDay].devicesStat {
// 		if deviceStat.Mac == l.login || deviceStat.IP == l.ipaddress {
// 			indexDevice = index
// 			break
// 		}
// 	}
// 	var mac, ip string
// 	if isMac(l.login) {
// 		mac = l.login
// 	}
// 	if isIP(l.ipaddress) {
// 		ip = l.ipaddress
// 	}
// 	if indexDevice == -1 {
// 		t.statYears[indexYear].monthsStat[indexMonth].daysStat[indexDay].devicesStat = append(t.statYears[indexYear].monthsStat[indexMonth].daysStat[indexDay].devicesStat, StatDeviceType{Mac: mac, IP: ip})
// 		indexDevice = len(t.statYears[indexYear].monthsStat[indexMonth].daysStat[indexDay].devicesStat) - 1
// 	}
// 	return indexYear, indexMonth, indexDay, indexDevice
// }

// func (t *Transport) Print() {
// 	t.RLock()
// 	for _, year := range t.statofYears {
// 		for _, month := range year.monthsStat {
// 			for _, day := range month.daysStat {
// 				for _, dev := range day.devicesStat {
// 					fmt.Println(year.year, month.month.String(), day.day, dev.IP, dev.Mac, dev.VolumePerDay)
// 				}
// 			}
// 		}
// 	}
// 	t.RUnlock()
// }
