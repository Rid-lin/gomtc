package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
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
	ip  string
	mac string
}

type StatDeviceType struct {
	mac string
	ip  string
	StatType
}

func (t *Transport) parseAllFilesAndCountingTrafficNew(cfg *Config) {
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
		t.newCount.startTime.In(cfg.Location).Format(cfg.dateTimeLayout),
		t.newCount.endTime.In(cfg.Location).Format(cfg.dateTimeLayout),
		ExTime.Seconds(),
		t.newCount.totalLineParsed/ExTimeInSec)
	fmt.Printf("The parsing started at %v, ended at %v, and lasted %.3v seconds at a rate of %v lines per second.\n",
		t.newCount.startTime.In(cfg.Location).Format(cfg.dateTimeLayout),
		t.newCount.endTime.In(cfg.Location).Format(cfg.dateTimeLayout),
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
	t.Lock()
	delete(t.statofYears[year].monthsStat[month].daysStat, day)
	t.newCount.LastDateNew = time.Date(year, month, day, 0, 0, 0, 1, t.Location).Unix()
	t.newCount.LastDayStrNew = time.Date(year, month, day, 0, 0, 0, 1, t.Location).String()
	t.Unlock()
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
		return fmt.Errorf("don't match to paramert fnStartsWith='%s'", cfg.FnStartsWith)
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
	deviceStat, ok := daysStat.devicesStat[KeyDevice{ip: l.ipaddress, mac: l.login}]
	if !ok {
		deviceStat = StatDeviceType{mac: l.login, ip: l.ipaddress}
		daysStat.devicesStat[KeyDevice{ip: l.ipaddress, mac: l.login}] = deviceStat
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
	daysStat.devicesStat[KeyDevice{ip: l.ipaddress, mac: l.login}] = deviceStat
	month.daysStat[l.day] = daysStat
	year.monthsStat[l.month] = month
	t.statofYears[l.year] = year
	t.Unlock()
}

func (t *Transport) checkQuotasNew() {
	t.RLock()
	p := parseType{
		SSHCredentials:   t.sshCredentials,
		QuotaType:        t.QuotaType,
		BlockAddressList: t.BlockAddressList,
		Location:         t.Location,
	}
	t.RUnlock()
	hour := time.Now().Hour()
	statOfDay := t.getDay(lNow())

	for _, alias := range t.Aliases {
		var VolumePerDay, VolumePerCheck uint64
		var StatPerHour [24]VolumePerType
		for _, key := range alias.KeyArr {
			VolumePerDay += statOfDay.devicesStat[key].VolumePerDay
			VolumePerCheck += statOfDay.devicesStat[key].VolumePerCheck
			for index := range statOfDay.devicesStat[key].StatPerHour {
				StatPerHour[index].PerHour += statOfDay.devicesStat[key].StatPerHour[index].PerHour
			}

			switch {
			case (VolumePerDay >= alias.DailyQuota || StatPerHour[hour].PerHour >= alias.HourlyQuota) && alias.Blocked && alias.ShouldBeBlocked:
				continue
			case VolumePerDay >= alias.DailyQuota && !alias.Blocked:
				alias.ShouldBeBlocked = true
				// alias.TimeoutBlock = setDailyTimeout()
				t.addBlockGroup(alias, p.BlockAddressList)
				alias.Blocked = true
			case StatPerHour[hour].PerHour >= alias.HourlyQuota && !alias.Blocked:
				alias.ShouldBeBlocked = true
				// alias.TimeoutBlock = setHourlyTimeout()
				t.addBlockGroup(alias, p.BlockAddressList)
				alias.Blocked = true
			case alias.Blocked:
				alias.ShouldBeBlocked = false
				t.delBlockGroup(alias, p.BlockAddressList)
				alias.Blocked = false
			}
			t.Lock()
			t.Aliases[alias.AliasName] = alias
			t.Unlock()
		}

	}
	t.Lock()
	for key, device := range t.devices {
		if _, ok := t.change[key]; !ok {
			if device.Blocked {
				device.ShouldBeBlocked = false
				device = device.delBlockGroup(p.BlockAddressList)
				t.change[key] = DeviceToBlock{
					Id:       device.Id,
					Groups:   device.AddressLists,
					Disabled: paramertToBool(device.disabledL),
				}
			}
		}
	}
	t.Unlock()
}

func (t *Transport) getDay(l *lineOfLogType) StatOfDayType {
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
	day, ok := month.daysStat[l.day]
	if !ok {
		day = StatOfDayType{
			devicesStat: map[KeyDevice]StatDeviceType{},
			day:         l.day}
		month.daysStat[l.day] = day
	}
	t.Unlock()
	return day
}
