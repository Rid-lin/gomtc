package main

import (
	"net"
	"os"
	"sync"
	"time"
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
	ConfigPath        string
	// debug               bool
	sshCredentials      SSHCredentials
	fileDestination     *os.File
	csvFiletDestination *os.File
	conn                *net.UDPConn
	timerParse          *time.Timer
	lastUpdated         time.Time
	lastUpdatedMT       time.Time
	renewOneMac         chan string
	stopReadFromUDP     chan uint8
	exitChan            chan os.Signal
	parseChan           chan *time.Time
	newLogChan          chan os.Signal
	outputChannel       chan decodedRecord
	newContType
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
	// dateFromArr [3]int
	// dateToArr   [3]int
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
}
