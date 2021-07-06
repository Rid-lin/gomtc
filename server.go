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

	"git.vegner.org/vsvegner/gomtc/internal/app/model"
	v "git.vegner.org/vsvegner/gomtc/internal/app/validation"
	. "git.vegner.org/vsvegner/gomtc/internal/config"
	. "git.vegner.org/vsvegner/gomtc/internal/gzip"
	"git.vegner.org/vsvegner/gomtc/internal/store/memorystore"
	. "git.vegner.org/vsvegner/gomtc/pkg/gsshutdown"
	log "github.com/sirupsen/logrus"
)

func NewTransport(cfg *Config) *Transport {
	gss := NewGSS(
		Exit(cfg), cfg,
		GetSIGHUP(cfg), cfg,
	)

	return &Transport{
		store:         memorystore.New(),
		devices:       make(map[model.KeyDevice]model.DeviceType),
		AliasesStrArr: make(map[string][]string),
		Aliases:       make(map[string]model.AliasType),
		statofYears:   make(map[int]model.StatOfYearType),
		change:        make(BlockDevices),
		ConfigPath:    cfg.ConfigPath,
		// DevicesRetryDelay: cfg.DevicesRetryDelay,
		gomtcSshHost:     cfg.GomtcSshHost,
		BlockAddressList: cfg.BlockGroup,
		ManualAddresList: cfg.ManualGroup,
		friends:          cfg.Friends,
		AssetsPath:       cfg.AssetsPath,
		SizeOneKilobyte:  cfg.SizeOneKilobyte,
		// debug:               cfg.Debug,
		stopReadFromUDP: make(chan uint8, 2),
		parseChan:       make(chan *time.Time),
		renewOneMac:     make(chan string, 100),
		newLogChan:      gss.LogChan,
		exitChan:        gss.ExitChan,
		// sshCredentials: SSHCredentials{
		// 	SSHHost:       cfg.MTAddr,
		// 	SSHPort:       cfg.SSHPort,
		// 	SSHUser:       cfg.MTUser,
		// 	SSHPass:       cfg.MTPass,
		// 	MaxSSHRetries: cfg.MaxSSHRetries,
		// 	SSHRetryDelay: cfg.SSHRetryDelay,
		// },
		QuotaType: model.QuotaType{
			HourlyQuota:  cfg.DefaultQuotaHourly * cfg.SizeOneKilobyte * cfg.SizeOneKilobyte,
			DailyQuota:   cfg.DefaultQuotaDaily * cfg.SizeOneKilobyte * cfg.SizeOneKilobyte,
			MonthlyQuota: cfg.DefaultQuotaMonthly * cfg.SizeOneKilobyte * cfg.SizeOneKilobyte,
		},
		Author: model.Author{
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
	p := model.ParseType{}
	t.RLock()
	// p.SSHCredentials = t.sshCredentials
	p.BlockAddressList = t.BlockAddressList
	p.QuotaType = t.QuotaType
	t.RUnlock()

	// t.readLog(cfg)

	t.getDevices()
	t.parseLog(cfg)
	t.updateAliases(p)
	t.SaveStatisticswithBuffer(path.Join(cfg.ConfigPath, "sqlite.db"), 1024*64)
	t.Lock()
	t.statofYears = map[int]model.StatOfYearType{}
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
	devices := GetDevicesFromRemote(model.ParseType{
		// SSHCredentials:   t.sshCredentials,
		QuotaType:        t.QuotaType,
		BlockAddressList: t.BlockAddressList,
	})
	for _, device := range devices {
		device.Manual = v.InAddressList(device.AddressLists, t.ManualAddresList)
		device.Blocked = v.InAddressList(device.AddressLists, t.BlockAddressList)
		t.devices[model.KeyDevice{
			// ip: device.ActiveAddress,
			Mac: device.ActiveMacAddress}] = device
	}
	t.lastUpdatedMT = time.Now()
	t.Unlock()
}

func (t *Transport) SendGroupStatus(NoControl bool) {
	if NoControl {
		return
	}
	t.RLock()
	p := model.ParseType{
		// SSHCredentials:   t.sshCredentials,
		QuotaType:        t.QuotaType,
		BlockAddressList: t.BlockAddressList,
	}
	t.RUnlock()
	t.change.BlockDevice(p)
	t.Lock()
	t.change = BlockDevices{}
	t.Unlock()
}

func (t *Transport) updateAliases(p model.ParseType) {
	t.Lock()
	var contains bool
	for _, year := range t.statofYears {
		for _, month := range year.MonthsStat {
			for _, day := range month.DaysStat {
				for key, deviceStat := range day.DevicesStat {
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
						t.Aliases[deviceStat.Mac] = model.AliasType{
							AliasName: deviceStat.Mac,
							KeyArr: []model.KeyDevice{{
								// ip: deviceStat.ip,
								Mac: deviceStat.Mac}},
							QuotaType:  model.CheckNULLQuotas(device.ToQuota(), p.QuotaType),
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

func (t *Transport) BlockAlias(a model.AliasType, group string) {
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
			Disabled: v.ParameterToBool(device.DisabledL),
		}
	}
	t.Unlock()
}

func (t *Transport) UnBlockAlias(a model.AliasType, group string) {
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
		t.TotalLineRead,
		t.TotalLineParsed,
		t.TotalLineAdded,
		t.TotalLineSkiped,
		t.TotalLineError)
	return nil
}

func (t *Transport) parseFileToMap(info os.FileInfo, cfg *Config) error {
	if !strings.HasPrefix(info.Name(), cfg.FnStartsWith) {
		return fmt.Errorf("%s don't match to paramert fnStartsWith='%s'", info.Name(), cfg.FnStartsWith)
	}
	FullFileName := filepath.Join(cfg.LogPath, info.Name())
	file, err := os.Open(FullFileName)
	if err != nil {
		//file.Close()
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
		l, err := model.ParseLineToStruct(line)
		if err != nil {
			t.LineError++
			log.Warningf("%v", err)
			continue
		}
		t.LineParsed++
		if t.LastDate > l.Timestamp {
			log.Tracef("line(%v) too old\r", line)
			t.LineSkiped++
			continue
		} else if t.LastDate < l.Timestamp {
			t.LastDate = l.Timestamp
		}

		// The main function of filling the database
		// Adding a row to the database for counting traffic
		t.store.DeviceStat().AddLine(l.ToStatDevice())
		t.addLineOutToMapOfReports(&l)
		t.LineAdded++
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		log.Errorf("Error scanner :%v", err)
		return err
	}
	return nil
}

func (t *Transport) addLineOutToMapOfReports(l *model.LineOfLogType) {
	l.Alias = model.DeterminingAlias(*l)
	t.Lock()
	year, ok := t.statofYears[l.Year]
	if !ok {
		year = model.StatOfYearType{
			Year:       l.Year,
			MonthsStat: map[time.Month]model.StatOfMonthType{},
		}
		t.statofYears[l.Year] = year
	}
	month, ok := year.MonthsStat[l.Month]
	if !ok {
		month = model.StatOfMonthType{
			DaysStat: map[int]model.StatOfDayType{},
			Month:    l.Month}
		year.MonthsStat[l.Month] = month
	}
	daysStat, ok := month.DaysStat[l.Day]
	if !ok {
		daysStat = model.StatOfDayType{
			DevicesStat: map[model.KeyDevice]model.StatDeviceType{},
			Day:         l.Day}
		month.DaysStat[l.Day] = daysStat
	}
	deviceStat, ok := daysStat.DevicesStat[model.KeyDevice{
		// ip: l.ipaddress,
		Mac: l.Login}]
	if !ok {
		deviceStat = model.StatDeviceType{Mac: l.Login, Ip: l.Ipaddress}
		daysStat.DevicesStat[model.KeyDevice{
			// ip: l.ipaddress,
			Mac: l.Login}] = deviceStat
	}
	// Расчет суммы трафика для устройства для дальшейшего отображения
	deviceStat.VolumePerDay = deviceStat.VolumePerDay + l.SizeInBytes
	deviceStat.VolumePerCheck = deviceStat.VolumePerCheck + l.SizeInBytes
	deviceStat.PerHour[l.Hour] = deviceStat.PerHour[l.Hour] + l.SizeInBytes
	// deviceStat.PerMinute[l.hour][l.minute] = deviceStat.PerMinute[l.hour][l.minute] + l.sizeInBytes
	// Расчет суммы трафика для дня для дальшейшего отображения
	// daysStat.VolumePerDay = daysStat.VolumePerDay + l.sizeInBytes
	// daysStat.VolumePerCheck = daysStat.VolumePerCheck + l.sizeInBytes
	// daysStat.PerHour[l.hour] = daysStat.PerHour[l.hour] + l.sizeInBytes
	// daysStat.PerMinute[l.hour][l.minute] = daysStat.PerMinute[l.hour][l.minute] + l.sizeInBytes
	// Возвращаем данные обратно
	daysStat.DevicesStat[model.KeyDevice{
		// ip: l.ipaddress,
		Mac: l.Login}] = deviceStat
	month.DaysStat[l.Day] = daysStat
	year.MonthsStat[l.Month] = month
	t.statofYears[l.Year] = year
	t.Unlock()
}

func (t *Transport) checkQuotas(cfg *Config) {
	t.Lock()
	tn := time.Now().In(Location)
	hour := tn.Hour()
	tns := tn.Format(DateLayout)
	devicesStat := GetDayStat(tns, tns, path.Join(cfg.ConfigPath, "sqlite.db"))
	for key := range t.devices {
		alias := t.Aliases[key.Mac]
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
		key := model.KeyDevice{}
		switch {
		case d.ActiveMacAddress != "":
			key = model.KeyDevice{Mac: d.ActiveMacAddress}
		case d.MacAddress != "":
			key = model.KeyDevice{Mac: d.MacAddress}
		}
		d.ShouldBeBlocked = t.Aliases[key.Mac].ShouldBeBlocked
		switch {
		case (d.Blocked && t.Aliases[key.Mac].ShouldBeBlocked) || (!d.Blocked && !t.Aliases[key.Mac].ShouldBeBlocked):
			continue
		case d.Blocked && !t.Aliases[key.Mac].ShouldBeBlocked:
			d = d.UnBlock(t.BlockAddressList, key)
			t.change[key] = DeviceToBlock{
				Id:       d.Id,
				Mac:      key.Mac,
				IP:       d.ActiveAddress,
				Groups:   d.AddressLists,
				Disabled: v.ParameterToBool(d.DisabledL),
			}
		case !d.Blocked && t.Aliases[key.Mac].ShouldBeBlocked:
			d = d.Block(t.BlockAddressList, key)
			t.change[key] = DeviceToBlock{
				Id:       d.Id,
				Mac:      key.Mac,
				IP:       d.ActiveAddress,
				Groups:   d.AddressLists,
				Disabled: v.ParameterToBool(d.DisabledL),
			}
		}
		t.devices[key] = d
	}
	t.Unlock()
}

func (t *Transport) parseOneFilesAndCountingTraffic(FullFileName string, cfg *Config) error {
	file, err := os.Open(FullFileName)
	if err != nil {
		//file.Close()
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
	ExTime := time.Since(t.StartTime)
	ExTimeInSec := uint64(ExTime.Seconds())
	if ExTimeInSec == 0 {
		ExTimeInSec = 1
	}
	t.EndTime = time.Now() // Saves the current time to be inserted into the log table
	t.LastUpdated = time.Now()
	log.Infof("The parsing started at %v, ended at %v, lasted %.3v seconds at a rate of %v lines per second.",
		t.StartTime.In(Location).Format(DateTimeLayout),
		t.EndTime.In(Location).Format(DateTimeLayout),
		ExTime.Seconds(),
		t.TotalLineParsed/ExTimeInSec)
	fmt.Printf("The parsing started at %v, ended at %v, lasted %.3v seconds at a rate of %v lines per second.\n",
		t.StartTime.In(Location).Format(DateTimeLayout),
		t.EndTime.In(Location).Format(DateTimeLayout),
		ExTime.Seconds(),
		t.TotalLineParsed/ExTimeInSec)
}

func Exit(ve interface{}) func(ve interface{}) {
	return func(ve interface{}) {
	}
}

func GetSIGHUP(vr interface{}) func(vr interface{}) {
	return func(ve interface{}) {
	}
}
