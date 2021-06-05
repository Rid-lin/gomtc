package main

import (
	"net"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Transport struct {
	AllDates map[string]AliasesType
	// aliases  AliasesType
	// stats    []StatDayType
	// statYears           []statYearType
	statofYears         map[int]StatOfYearType
	AliasesStrArr       map[string][]string
	Infos               map[string]InfoType
	dataOld             MapOfReports
	dataCasheOld        MapOfReports
	change              AliasesOldType
	devices             DevicesType
	logs                []LogsOfJob
	friends             []string
	AssetsPath          string
	BlockAddressList    string
	SizeOneKilobyte     uint64
	DevicesRetryDelay   string
	pidfile             string
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
	DeviceOldType
}

type QuotaType struct {
	TimeoutBlock    string
	HourlyQuota     uint64
	DailyQuota      uint64
	MonthlyQuota    uint64
	Disabled        bool
	Dynamic         bool
	Blocked         bool
	Manual          bool
	ShouldBeBlocked bool
}

type DeviceOldType struct {
	Id       string
	IP       string
	Mac      string
	AMac     string
	HostName string
	Groups   string
	timeout  time.Time
}

type DeviceType struct {
	// From MT
	Id                  string
	activeAddress       string // 192.168.65.85
	activeClientId      string // 1:e8:d8:d1:47:55:93
	allowDualStackQueue string
	activeMacAddress    string // E8:D8:D1:47:55:93
	activeServer        string // dhcp_lan
	address             string // pool_admin
	addressLists        string // inet
	blocked             string // false
	clientId            string // 1:e8:d8:d1:47:55:93
	comment             string // nb=Vlad/com=UTTiST/col=Admin/quotahourly=500000000/quotadaily=50000000000
	dhcpOption          string //
	disabledL           string // false
	dynamic             string // false
	expiresAfter        string // 6m32s
	hostName            string // root-hp
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
	Hardware        bool
	Manual          bool
	ShouldBeBlocked bool
	timeout         time.Time
	StatOldType
}

type DevicesType []DeviceType
type AliasesType []AliasType
type AliasesOldType []AliasOld

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

type MapOfReports map[KeyMapOfReports]AliasOld

type KeyMapOfReports struct {
	DateStr string
	Alias   string
}

type AliasOld struct {
	Alias   string
	DateStr string
	Hits    uint32
	InfoType
	StatOldType
}

type InfoType struct {
	DeviceOldType
	PersonType
	QuotaType
}

type AliasType struct {
	Alias string
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
	InfoType
	StatOldType
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
	InfoType
}

type RequestForm struct {
	dateFrom string
	dateTo   string
	path     string
	referURL string
	report   string
}

type StatOldType struct {
	VolumePerHour     [24]uint64
	Site              string
	Precent           float64
	VolumeOfPrecentil uint64
	Average           uint64
	VolumePerDay      uint64
	Count             uint32
}

type StatType struct {
	StatPerHour     [24]VolumePerType
	Precent         float64
	SizeOfPrecentil uint64
	Average         uint64
	SpeedPerDay     uint64
	SpeedPerCheck   uint64
	VolumePerDay    uint64
	VolumePerCheck  uint64
	Count           uint32
}

type VolumePerType struct {
	Minute      [60]uint64
	SpeedMinute [60]uint64
	Hour        uint64
	SpeedHour   uint64
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
		log.Errorf("Error loading Location(%v):%v", cfg.Loc, err)
		// Location = time.UTC
		Location = time.FixedZone("Custom timezone", int(cfg.Timezone*60*60))
	}

	return &Transport{
		dataOld:             map[KeyMapOfReports]AliasOld{},
		dataCasheOld:        map[KeyMapOfReports]AliasOld{},
		devices:             DevicesType{},
		AliasesStrArr:       make(map[string][]string),
		Location:            Location,
		statofYears:         make(map[int]StatOfYearType),
		pidfile:             cfg.Pidfile,
		DevicesRetryDelay:   cfg.DevicesRetryDelay,
		BlockAddressList:    cfg.BlockGroup,
		fileDestination:     fileDestination,
		csvFiletDestination: csvFiletDestination,
		logs:                []LogsOfJob{},
		friends:             cfg.Friends,
		AssetsPath:          cfg.AssetsPath,
		SizeOneKilobyte:     cfg.SizeOneKilobyte,
		stopReadFromUDP:     make(chan uint8, 2),
		parseChan:           make(chan *time.Time),
		outputChannel:       make(chan decodedRecord, 100),
		renewOneMac:         make(chan string, 100),
		newLogChan:          getNewLogSignalsChannel(),
		exitChan:            getExitSignalsChannel(),
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
