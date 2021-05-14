package main

import (
	"net"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/routeros.v2"
)

type Transport struct {
	infoOfDevices       map[string]InfoOfDeviceType
	data                MapOfReports
	dataCashe           MapOfReports
	Location            *time.Location
	fileDestination     *os.File
	csvFiletDestination *os.File
	conn                *net.UDPConn
	clientROS           *routeros.Client
	logs                []LogsOfJob
	lastUpdated         time.Time
	lastUpdatedMT       time.Time
	friends             []string
	AssetsPath          string
	BlockAddressList    string
	SizeOneKilobyte     uint64
	timer               *time.Timer
	renewOneMac         chan string
	stopReadFromUDP     chan uint8
	exitChan            chan os.Signal
	parseChan           chan *time.Time
	newLogChan          chan os.Signal
	outputChannel       chan decodedRecord
	Aliases             map[string][]string
	Author
	QuotaType
	sync.RWMutex
}

type Author struct {
	Copyright string
	Mail      string
}
type request struct {
	Time,
	IP string
	// timeInt int64
}

type ResponseType struct {
	// IP       string `JSON:"IP"`
	// Mac      string `JSON:"Mac"`
	// Hostname string `JSON:"Hostname"`
	Comments string `JSON:"Comment"`
	DeviceType
}

type QuotaType struct {
	HourlyQuota     uint64
	DailyQuota      uint64
	MonthlyQuota    uint64
	Disabled        bool
	Blocked         bool
	Manual          bool
	ShouldBeBlocked bool
}

type DeviceType struct {
	Id           string
	IP           string
	Mac          string
	AMac         string
	HostName     string
	Groups       string
	AddressLists []string
}

type lineOfLogType struct {
	date,
	ipaddress,
	httpstatus,
	method,
	siteName,
	login,
	mime,
	alias,
	hostname,
	comments string
	year,
	day,
	hour,
	minute int
	nsec,
	timestamp int64
	month       time.Month
	sizeInBytes uint64
	// port int
	// sitehashe   string
}

type MapOfReports map[KeyMapOfReports]ValueMapOfReportsType

type KeyMapOfReports struct {
	DateStr string
	Alias   string
}

type ValueMapOfReportsType struct {
	// SizeOfHourU [24]uint64
	Alias   string
	DateStr string
	// SizeU uint64
	Hits uint32
	InfoOfDeviceType
	StatType
}

type InfoOfDeviceType struct {
	DeviceType
	PersonType
	QuotaType
}

type PersonType struct {
	Comment  string
	Comments string
	Name     string
	Position string
	Company  string
	IDUser   string
	TypeD    string
}

type Count struct {
	LineParsed,
	LineSkiped,
	LineAdded,
	LineRead,
	LineError,
	totalLineRead,
	totalLineAdded,
	totalLineParsed,
	totalLineSkiped,
	totalLineError uint64
}

type LineOfDisplay struct {
	Alias string
	Login string
	InfoOfDeviceType
	StatType
}

type DisplayDataType struct {
	ArrayDisplay   []LineOfDisplay
	Logs           []LogsOfJob
	Header         string
	DateFrom       string
	DateTo         string
	LastUpdated    string
	LastUpdatedMT  string
	Path           string
	Host           string
	ReferURL       string
	TimeToGenerate time.Duration
	Author
	SizeOneType
	QuotaType
}

type SizeOneType struct {
	SizeOneKilobyte uint64
	SizeOneMegabyte uint64
	SizeOneGigabyte uint64
}
type DisplayDataUserType struct {
	Header          string
	Copyright       string
	Mail            string
	Alias           string
	SizeOneKilobyte uint64
	InfoOfDeviceType
}

type RequestForm struct {
	dateFrom,
	dateTo,
	path,
	// host,
	referURL,
	report string
}

type StatType struct {
	// SizeOfHourF        [24]float64
	// SizeOfHourStr      [24]string
	// SizeStr            string
	// SizeOfPrecentilStr string
	// PrecentStr         string
	// AverageStr         string
	// SizeF              float64
	SizeOfHour      [24]uint64
	Site            string
	Precent         float64
	SizeOfPrecentil uint64
	Average         uint64
	Size            uint64
	Count           uint32
}

func NewTransport(cfg *Config) *Transport {

	clientROS, err := dial(cfg)
	if err != nil {
		log.Errorf("Error connect to %v:%v", cfg.MTAddr, err)
		clientROS = tryingToReconnectToMokrotik(cfg)
	}

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
		log.Errorf("Error loading Location(%v):%v", cfg.Loc, err)
		Location = time.UTC
	}

	return &Transport{
		data:                map[KeyMapOfReports]ValueMapOfReportsType{},
		dataCashe:           map[KeyMapOfReports]ValueMapOfReportsType{},
		infoOfDevices:       make(map[string]InfoOfDeviceType),
		Aliases:             make(map[string][]string),
		Location:            Location,
		BlockAddressList:    cfg.BlockGroup,
		clientROS:           clientROS,
		fileDestination:     fileDestination,
		csvFiletDestination: csvFiletDestination,
		logs:                []LogsOfJob{},
		friends:             cfg.Friends,
		AssetsPath:          cfg.AssetsPath,
		SizeOneKilobyte:     cfg.SizeOneKilobyte,
		stopReadFromUDP:     make(chan uint8),
		parseChan:           make(chan *time.Time),
		outputChannel:       make(chan decodedRecord, 100),
		renewOneMac:         make(chan string, 100),
		newLogChan:          getNewLogSignalsChannel(),
		exitChan:            getExitSignalsChannel(),
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
