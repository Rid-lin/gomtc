package main

import (
	"os"
	"path"
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

	// Location, err := time.LoadLocation(cfg.Loc)
	// if err != nil {
	// 	Location = time.FixedZone("Custom timezone", int(cfg.Timezone*60*60))
	// 	log.Warningf("Error loading timezone from location(%v):%v. Using a fixed time zone(%v:%v)", cfg.Loc, err, Location, cfg.Timezone*60*60)
	// 	// Location = time.UTC
	// }

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
	t.checkQuotas(cfg)
	t.BlockDevices()
	// t.SendGroupStatus(cfg.NoControl)
	t.getDevices()

	t.SaveStatisticswithBuffer(path.Join(cfg.ConfigPath, "sqlite.db"), 1024*64)
	t.Lock()
	t.statofYears = map[int]StatOfYearType{}
	t.Unlock()
	// t.writeLog(cfg)
	// t.Count = Count{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	t.setTimerParse(cfg.ParseDelay)
}
