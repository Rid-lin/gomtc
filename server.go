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

	. "git.vegner.org/vsvegner/gomtc/internal/config"
	. "git.vegner.org/vsvegner/gomtc/internal/gzip"
	. "git.vegner.org/vsvegner/gomtc/pkg/gsshutdown"
	log "github.com/sirupsen/logrus"
)

type newContType struct {
	Count
	startTime   time.Time
	endTime     time.Time
	LastUpdated time.Time
	LastDate    int64
	LastDayStr  string
}

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

func NewTransport(cfg *Config) *Transport {
	var err error

	if cfg.CSV {
		csvFiletDestination, err = os.OpenFile(cfg.NameFileToLog+".csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fileDestination.Close()
			log.Fatalf("Error, the '%v' file could not be created (there are not enough premissions or it is busy with another program): %v", cfg.NameFileToLog, err)
		}
	}

	gss := NewGSS(
		Exit(cfg), cfg,
		GetSIGHUP(cfg), cfg,
	)

	return &Transport{
		devices:             make(map[KeyDevice]DeviceType),
		AliasesStrArr:       make(map[string][]string),
		Aliases:             make(map[string]AliasType),
		statofYears:         make(map[int]StatOfYearType),
		change:              make(BlockDevices),
		pidfile:             cfg.Pidfile,
		ConfigPath:          cfg.ConfigPath,
		DevicesRetryDelay:   cfg.DevicesRetryDelay,
		BlockAddressList:    cfg.BlockGroup,
		ManualAddresList:    cfg.ManualGroup,
		fileDestination:     fileDestination,
		csvFiletDestination: csvFiletDestination,
		friends:             cfg.Friends,
		AssetsPath:          cfg.AssetsPath,
		SizeOneKilobyte:     cfg.SizeOneKilobyte,
		// debug:               cfg.Debug,
		stopReadFromUDP: make(chan uint8, 2),
		parseChan:       make(chan *time.Time),
		renewOneMac:     make(chan string, 100),
		newLogChan:      gss.LogChan,
		exitChan:        gss.ExitChan,
		sshCredentials: SSHCredentials{
			SSHHost:       cfg.MTAddr,
			SSHPort:       "22",
			SSHUser:       cfg.MTUser,
			SSHPass:       cfg.MTPass,
			MaxSSHRetries: cfg.MaxSSHRetries,
			SSHRetryDelay: cfg.SSHRetryDelay,
		},
		QuotaType: QuotaType{
			HourlyQuota:  cfg.DefaultQuotaHourly * cfg.SizeOneKilobyte * cfg.SizeOneKilobyte,
			DailyQuota:   cfg.DefaultQuotaDaily * cfg.SizeOneKilobyte * cfg.SizeOneKilobyte,
			MonthlyQuota: cfg.DefaultQuotaMonthly * cfg.SizeOneKilobyte * cfg.SizeOneKilobyte,
		},
		Author: Author{
			Copyright: "GoSquidLogAnalyzer © 2020-2021 by Vladislav Vegner",
			Mail:      "mailto:vegner.vs@uttist.ru",
		},
	}
}

func (t *Transport) Start(cfg *Config) {
	// Endless file parsing loop
	go func(cfg *Config) {
		t.runOnce(cfg)
		for {
			<-t.timerParse.C
			t.runOnce(cfg)
		}
	}(cfg)
	t.handleRequest(cfg)
}

func (t *Transport) setTimerParse(IntervalStr string) {
	interval, err := time.ParseDuration(IntervalStr)
	if err != nil {
		t.timerParse = time.NewTimer(15 * time.Minute)
	} else {
		t.timerParse = time.NewTimer(interval)
	}
}

func (t *Transport) runOnce(cfg *Config) {
	p := parseType{}
	t.RLock()
	p.SSHCredentials = t.sshCredentials
	p.BlockAddressList = t.BlockAddressList
	p.QuotaType = t.QuotaType
	t.RUnlock()

	// t.readLog(cfg)

	t.getDevices()
	t.parseLog(cfg)
	t.updateAliases(p)
	t.SaveStatisticswithBuffer(path.Join(cfg.ConfigPath, "sqlite.db"), 1024*64)
	t.Lock()
	t.statofYears = map[int]StatOfYearType{}
	t.Unlock()
	t.checkQuotas(cfg)
	t.BlockDevices()
	t.SendGroupStatus(cfg.NoControl)
	t.getDevices()

	// t.writeLog(cfg)
	// t.Count = Count{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	t.setTimerParse(cfg.ParseDelay)
}

func (t *Transport) getDevices() {
	t.Lock()
	devices := parseInfoFromMTAsValueToSlice(parseType{
		SSHCredentials:   t.sshCredentials,
		QuotaType:        t.QuotaType,
		BlockAddressList: t.BlockAddressList,
	})
	for _, device := range devices {
		device.Manual = inAddressList(device.AddressLists, t.ManualAddresList)
		device.Blocked = inAddressList(device.AddressLists, t.BlockAddressList)
		t.devices[KeyDevice{
			// ip: device.ActiveAddress,
			mac: device.ActiveMacAddress}] = device
	}
	t.lastUpdatedMT = time.Now()
	t.Unlock()
}

func (t *Transport) SendGroupStatus(NoControl bool) {
	if NoControl {
		return
	}
	t.RLock()
	p := parseType{
		SSHCredentials:   t.sshCredentials,
		QuotaType:        t.QuotaType,
		BlockAddressList: t.BlockAddressList,
	}
	t.RUnlock()
	t.change.sendLeaseSet(p)
	t.Lock()
	t.change = BlockDevices{}
	t.Unlock()
}

func (t *Transport) updateAliases(p parseType) {
	t.Lock()
	var contains bool
	for _, year := range t.statofYears {
		for _, month := range year.monthsStat {
			for _, day := range month.daysStat {
				for key, deviceStat := range day.devicesStat {
					device := t.devices[key]
					for _, alias := range t.Aliases {
						switch {
						case alias.DeviceInAlias(key):
							alias.UpdateQuota(device.ToQuota())
							alias.UpdatePerson(device.ToPerson())
							goto gotAlias
						// case alias.IPOnlyInAlias(key) && day.day == time.Now().Day():
						// 	alias.UpdateQuota(device.ToQuota())
						// 	alias.UpdatePerson(device.ToPerson())
						// 	goto gotAlias
						case alias.MacInAlias(key):
							alias.UpdateQuota(device.ToQuota())
							alias.UpdatePerson(device.ToPerson())
							goto gotAlias
						}
					}
					if !contains {
						device := t.devices[key]
						t.Aliases[deviceStat.mac] = AliasType{
							AliasName: deviceStat.mac,
							KeyArr: []KeyDevice{{
								// ip: deviceStat.ip,
								mac: deviceStat.mac}},
							QuotaType:  checkNULLQuotas(device.ToQuota(), p.QuotaType),
							PersonType: device.ToPerson(),
						}
					}
				gotAlias:
				}
			}
		}
	}
	t.Unlock()
}

func (t *Transport) BlockAlias(a AliasType, group string) {
	t.Lock()
	for _, key := range a.KeyArr {
		device := t.devices[key]
		if device.IsNULL() {
			delete(t.devices, key)
			continue
		}
		if device.Manual || device.Blocked {
			continue
		}
		device = device.Block(group, key)
		t.change[key] = DeviceToBlock{
			Id:       device.Id,
			Groups:   device.AddressLists,
			Disabled: paramertToBool(device.disabledL),
		}
	}
	t.Unlock()
}

func (t *Transport) UnBlockAlias(a AliasType, group string) {
	t.Lock()
	for _, key := range a.KeyArr {
		device := t.devices[key]
		if device.Manual || !device.Blocked {
			continue
		}
		device = device.UnBlock(group, key)
	}
	t.Unlock()
}

func (t *Transport) parseAllFilesAndCountingTraffic(cfg *Config) {
	err := t.parseDirToMap(cfg)
	if err != nil {
		log.Error(err)
	}
}

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
		FileName, err := UnGzip(FullFileName, dir)
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
	for key := range t.devices {
		alias := t.Aliases[key.mac]
		ds := devicesStat[key]
		if ds.VolumePerDay >= alias.DailyQuota || ds.PerHour[hour] >= alias.HourlyQuota {
			alias.ShouldBeBlocked = true
		} else {
			alias.ShouldBeBlocked = false
		}
		t.Aliases[alias.AliasName] = alias
	}
	// for key, d := range devicesStat {
	// 	alias := t.Aliases[key.mac]
	// 	if d.VolumePerDay >= alias.DailyQuota || d.PerHour[hour] >= alias.HourlyQuota {
	// 		alias.ShouldBeBlocked = true
	// 	} else {
	// 		alias.ShouldBeBlocked = false
	// 	}
	// 	t.Aliases[alias.AliasName] = alias
	// }

	t.Unlock()
}

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
		d.ShouldBeBlocked = t.Aliases[key.mac].ShouldBeBlocked
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
		t.devices[key] = d
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

func Exit(ve interface{}) func(ve interface{}) {
	return func(ve interface{}) {
		cfg, ok := ve.(*Config)
		if ok {
			if err := os.Remove(cfg.Pidfile); err != nil {
				log.Errorf("File (%v) deletion error:%v", cfg.Pidfile, err)
			}
		}
	}
}

func GetSIGHUP(vr interface{}) func(vr interface{}) {
	return func(ve interface{}) {
		_, ok := ve.(*Config)
		if ok {
			log.Println("Received a signal from logrotate, close the file.")
			if err := fileDestination.Sync(); err != nil {
				log.Errorf("File(%v) sync error:%v", fileDestination.Name(), err)
			}
			if err := fileDestination.Close(); err != nil {
				log.Errorf("File(%v) close error:%v", fileDestination.Name(), err)
			}
		}
	}
}
