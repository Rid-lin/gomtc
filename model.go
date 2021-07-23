package main

import (
	"os"
	"sync"
	"time"

	"git.vegner.org/vsvegner/gomtc/internal/app/model"
	"git.vegner.org/vsvegner/gomtc/internal/store"
)

// type SSHCredentials struct {
// 	SSHHost       string
// 	SSHPort       string
// 	SSHUser       string
// 	SSHPass       string
// 	MaxSSHRetries int
// 	SSHRetryDelay uint16
// }

type Transport struct {
	store             store.Store
	Aliases           map[string]model.AliasType
	statofYears       map[int]model.StatOfYearType
	change            BlockDevices
	devices           DevicesMapType
	friends           []string
	DSN               string
	AssetsPath        string
	BlockAddressList  string
	ManualAddresList  string
	DevicesRetryDelay string
	ConfigPath        string
	gomtcSshHost      string
	SizeOneKilobyte   uint64
	timerParse        *time.Timer
	lastUpdated       time.Time
	lastUpdatedMT     time.Time
	renewOneMac       chan string
	stopReadFromUDP   chan uint8
	exitChan          chan os.Signal
	parseChan         chan *time.Time
	newLogChan        chan os.Signal
	model.NewContType
	model.Author
	model.QuotaType
	sync.RWMutex
}

type ResponseType struct {
	IP       string
	Comments string
	Mac      string
	HostName string
}

// type DevicesType []model.DeviceType
type DevicesMapType map[model.KeyDevice]model.DeviceType
type BlockDevices map[model.KeyDevice]DeviceToBlock

type DeviceToBlock struct {
	Id       string
	Mac      string
	IP       string
	Groups   string
	Delay    string
	Disabled bool
}
