package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

func NewTransport(cfg *Config) *Transport {
	var err error

	if !cfg.NoFlow {
		fileDestination, err = os.OpenFile(cfg.NameFileToLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fileDestination.Close()
			log.Fatalf("Error, the '%v' file could not be created (there are not enough premissions or it is busy with another program): %v", cfg.NameFileToLog, err)
		}
	}

	if cfg.CSV {
		csvFiletDestination, err = os.OpenFile(cfg.NameFileToLog+".csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fileDestination.Close()
			log.Fatalf("Error, the '%v' file could not be created (there are not enough premissions or it is busy with another program): %v", cfg.NameFileToLog, err)
		}
	}

	Location, err := time.LoadLocation(cfg.Loc)
	if err != nil {
		Location = time.FixedZone("Custom timezone", int(cfg.Timezone*60*60))
		log.Warningf("Error loading timezone from location(%v):%v. Using a fixed time zone(%v:%v)", cfg.Loc, err, Location, cfg.Timezone*60*60)
		// Location = time.UTC
	}

	return &Transport{
		devices:             make(map[KeyDevice]DeviceType),
		AliasesStrArr:       make(map[string][]string),
		Aliases:             make(map[string]AliasType),
		Location:            Location,
		statofYears:         make(map[int]StatOfYearType),
		change:              make(BlockDevices),
		pidfile:             cfg.Pidfile,
		DevicesRetryDelay:   cfg.DevicesRetryDelay,
		BlockAddressList:    cfg.BlockGroup,
		ManualAddresList:    cfg.ManualGroup,
		fileDestination:     fileDestination,
		csvFiletDestination: csvFiletDestination,
		logs:                []LogsOfJob{},
		friends:             cfg.Friends,
		AssetsPath:          cfg.AssetsPath,
		SizeOneKilobyte:     cfg.SizeOneKilobyte,
		// debug:               cfg.Debug,
		stopReadFromUDP: make(chan uint8, 2),
		parseChan:       make(chan *time.Time),
		outputChannel:   make(chan decodedRecord, 100),
		renewOneMac:     make(chan string, 100),
		newLogChan:      getNewLogSignalsChannel(),
		exitChan:        getExitSignalsChannel(),
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

func (t *Transport) getDevicesToCashe() {
	t.Lock()
	devices := parseInfoFromMTAsValueToSlice(parseType{
		SSHCredentials:   t.sshCredentials,
		QuotaType:        t.QuotaType,
		BlockAddressList: t.BlockAddressList,
		Location:         t.Location,
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

func (t *Transport) checkQuotas() {
	// t.RLock()
	// p := parseType{
	// 	SSHCredentials:   t.sshCredentials,
	// 	QuotaType:        t.QuotaType,
	// 	BlockAddressList: t.BlockAddressList,
	// 	Location:         t.Location,
	// }
	// t.RUnlock()
	hour := time.Now().Hour()
	day := t.getDay(lNow())
	t.Lock()
	for _, alias := range t.Aliases {
		var VolumePerDay, VolumePerCheck uint64
		var StatPerHour [24]VolumePerType
		for _, key := range alias.KeyArr {
			VolumePerDay += day.devicesStat[key].VolumePerDay
			VolumePerCheck += day.devicesStat[key].VolumePerCheck
			for index := range day.devicesStat[key].StatPerHour {
				StatPerHour[index].PerHour += day.devicesStat[key].StatPerHour[index].PerHour
			}
		}
		if VolumePerDay >= alias.DailyQuota || StatPerHour[hour].PerHour >= alias.HourlyQuota {
			alias.ShouldBeBlocked = true
			// t.BlockAlias(alias, p.BlockAddressList)
		} else {
			alias.ShouldBeBlocked = false
			// t.UnBlockAlias(alias, p.BlockAddressList)
		}
		// switch {
		// case (VolumePerDay >= alias.DailyQuota || StatPerHour[hour].PerHour >= alias.HourlyQuota) && alias.Blocked && alias.ShouldBeBlocked:
		// 	continue
		// case VolumePerDay >= alias.DailyQuota && !alias.Blocked:
		// 	alias.ShouldBeBlocked = true
		// 	t.addBlockGroup(alias, p.BlockAddressList)
		// 	alias.Blocked = true
		// case StatPerHour[hour].PerHour >= alias.HourlyQuota && !alias.Blocked:
		// 	alias.ShouldBeBlocked = true
		// 	t.addBlockGroup(alias, p.BlockAddressList)
		// 	alias.Blocked = true
		// case alias.Blocked:
		// 	alias.ShouldBeBlocked = false
		// 	t.delBlockGroup(alias, p.BlockAddressList)
		// 	alias.Blocked = false
		// }
		t.Aliases[alias.AliasName] = alias

	}
	t.Unlock()
	// t.Lock()
	// for key, device := range t.devices {
	// 	if _, ok := t.change[key]; !ok {
	// 		if device.Blocked {
	// 			device = device.UnBlock(p.BlockAddressList)
	// 			t.change[key] = DeviceToBlock{
	// 				Id:       device.Id,
	// 				Groups:   device.AddressLists,
	// 				Disabled: paramertToBool(device.disabledL),
	// 			}
	// 		}
	// 	}
	// }
	// t.Unlock()
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
