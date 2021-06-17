package main

import (
	"net"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type AliasType struct {
	AliasName string
	KeyArr    []KeyDevice
	QuotaType
	PersonType
}

type Transport struct {
	Aliases           map[string]AliasType
	statofYears       map[int]StatOfYearType
	AliasesStrArr     map[string][]string
	change            BlockDevices
	devices           DevicesMapType
	logs              []LogsOfJob
	friends           []string
	AssetsPath        string
	BlockAddressList  string
	ManualAddresList  string
	SizeOneKilobyte   uint64
	DevicesRetryDelay string
	pidfile           string
	// debug               bool
	sshCredentials      SSHCredentials
	fileDestination     *os.File
	csvFiletDestination *os.File
	conn                *net.UDPConn
	timerParse          *time.Timer
	Location            *time.Location
	lastUpdated         time.Time
	lastUpdatedMT       time.Time
	renewOneMac         chan string
	stopReadFromUDP     chan uint8
	exitChan            chan os.Signal
	parseChan           chan *time.Time
	newLogChan          chan os.Signal
	outputChannel       chan decodedRecord
	newCount            newContType
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
}

type ResponseType struct {
	Comments string
	Mac      string
	HostName string
}

type QuotaType struct {
	TimeoutBlock    string
	HourlyQuota     uint64
	DailyQuota      uint64
	MonthlyQuota    uint64
	Disabled        bool
	Dynamic         bool
	Blocked         bool
	ShouldBeBlocked bool
}

type DeviceType struct {
	// From MT
	Id                  string
	ActiveAddress       string // 192.168.65.85
	activeClientId      string // 1:e8:d8:d1:47:55:93
	allowDualStackQueue string
	ActiveMacAddress    string // E8:D8:D1:47:55:93
	activeServer        string // dhcp_lan
	address             string // pool_admin
	AddressLists        string // inet
	clientId            string // 1:e8:d8:d1:47:55:93
	Comment             string // nb=Vlad/com=UTTiST/col=Admin/quotahourly=500000000/quotadaily=50000000000
	dhcpOption          string //
	disabledL           string // false
	dynamic             string // false
	expiresAfter        string // 6m32s
	HostName            string // root-hp
	lastSeen            string // 3m28s
	macAddress          string // E8:D8:D1:47:55:93
	radius              string // false
	server              string // dhcp_lan
	status              string // bound
	insertQueueBefore   string
	rateLimit           string
	useSrcMac           string
	agentCircuitId      string
	blockAccess         string
	leaseTime           string
	agentRemoteId       string
	dhcpOptionSet       string
	srcMacAddress       string
	alwaysBroadcast     string
	//User Defined
	// timeout         time.Time
	Hardware        bool
	Manual          bool
	Blocked         bool
	ShouldBeBlocked bool
	TypeD           string
	// StatOldType
}

type DevicesType []DeviceType
type DevicesMapType map[KeyDevice]DeviceType
type BlockDevices map[KeyDevice]DeviceToBlock

type DeviceToBlock struct {
	Id       string
	Mac      string
	IP       string
	Groups   string
	Disabled bool
}

type lineOfLogType struct {
	date        string // squid timestamp 1621969229.000
	ipaddress   string
	httpstatus  string
	method      string
	siteName    string
	login       string
	mime        string
	alias       string
	hostname    string
	comments    string
	year        int
	day         int
	hour        int
	minute      int
	nsec        int64
	timestamp   int64
	month       time.Month
	sizeInBytes uint64
}

type InfoType struct {
	InfoName string
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
	InfoType
	StatDeviceType
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
	SizeOneKilobyte uint64
	InfoType
}

type RequestForm struct {
	dateFrom string
	dateTo   string
	path     string
	referURL string
	report   string
	// dateFromT time.Time
	// dateToT   time.Time
}

// type StatOldType struct {
// 	VolumePerHour     [24]uint64
// 	Site              string
// 	Precent           float64
// 	VolumeOfPrecentil uint64
// 	Average           uint64
// 	VolumePerDay      uint64
// 	Count             uint32
// }

type StatType struct {
	// PerMinute       [24][60]uint64
	PerHour         [24]uint64
	Precent         float64
	SizeOfPrecentil uint64
	Average         uint64
	VolumePerDay    uint64
	VolumePerCheck  uint64
	Count           uint32
}

type VolumePerType struct {
	PerMinute [60]uint64
	PerHour   uint64
}

type SSHCredentials struct {
	SSHHost       string
	SSHPort       string
	SSHUser       string
	SSHPass       string
	MaxSSHRetries int
	SSHRetryDelay uint16
}

type parseType struct {
	SSHCredentials
	QuotaType
	BlockAddressList string
	Location         *time.Location
}

var (
	fileDestination     *os.File
	csvFiletDestination *os.File
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
