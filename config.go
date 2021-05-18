package main

import (
	"time"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigtoml"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	dateLayout     string
	dateTimeLayout string
	LastDate       int64
	LastDay        int64
	LastDayStr     string
	LastDateStr    string

	GonsquidAddr string `default:":3030" usage:"Listen address for HTTP-server"`

	Friends                []string `default:"" usage:"List of aliases, IP addresses, friends' logins"`
	ConfigPath             string   `default:"/etc/gomtc" usage:"folder path to all config files"`
	LogPath                string   `default:"/var/log/gomtc" usage:"folder path to logs-file"`
	ListenAddr             string   `default:":3031" usage:"Listen address for HTTP-server"`
	AssetsPath             string   `default:"/etc/gomtc/assets"  usage:"The path to the assets folder where the template files are located"`
	SubNets                []string `default:"" usage:"List of subnets traffic between which will not be counted"`
	IgnorList              []string `default:"" usage:"List of lines that will be excluded from the final log"`
	LogLevel               string   `default:"info" usage:"Log level: panic, fatal, error, warn, info, debug, trace"`
	FlowAddr               string   `default:"0.0.0.0:2055" usage:"Address and port to listen NetFlow packets"`
	NameFileToLog          string   `default:"" usage:"The file where logs will be written in the format of squid logs"`
	MTAddr                 string   `default:"" usage:"The address of the Mikrotik router, from which the data on the comparison of the MAC address and IP address is taken"`
	MTUser                 string   `default:"" usage:"User of the Mikrotik router, from which the data on the comparison of the MAC address and IP address is taken"`
	MTPass                 string   `default:"" usage:"The password of the user of the Mikrotik router, from which the data on the comparison of the mac-address and IP-address is taken"`
	Loc                    string   `default:"Asia/Yekaterinburg" usage:"Location for time"`
	Interval               string   `default:"10m" usage:"Interval to getting info from Mikrotik"`
	BlockGroup             string   `default:"Block" usage:"The name of the address list in MicrotiK with which access is blocked to users who have exceeded the quota."`
	ReceiveBufferSizeBytes int      `default:"" usage:"Size of RxQueue, i.e. value for SO_RCVBUF in bytes"`
	NumOfTryingConnectToMT int      `default:"10" usage:"The number of attempts to connect to the microtik router"`
	DefaultQuotaHourly     uint64   `default:"0" usage:"Default hourly traffic consumption quota"`
	DefaultQuotaDaily      uint64   `default:"0" usage:"Default daily traffic consumption quota"`
	DefaultQuotaMonthly    uint64   `default:"0" usage:"Default monthly traffic consumption quota"`
	SizeOneKilobyte        uint64   `default:"1024" usage:"The number of bytes in one megabyte"`
	UseTLS                 bool     `default:"false" usage:"Using TLS to connect to a router"`
	CSV                    bool     `default:"false" usage:"Output to csv"`
	NoFlow                 bool     `default:"false" usage:"When this parameter is specified, the netflow packet listener is not launched, therefore, log files are not created, but only parsed."`
	Location               *time.Location
	startTime              time.Time
	endTime                time.Time
	Count
}

var (
	cfg Config
)

func newConfig() *Config {

	var cfg Config
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		// feel free to skip some steps :)
		// SkipEnv:      true,
		SkipFiles:          false,
		AllowUnknownFlags:  true,
		AllowUnknownFields: true,
		SkipDefaults:       false,
		SkipFlags:          false,
		EnvPrefix:          "GOMTC",
		FlagPrefix:         "",
		Files:              []string{"/etc/gomtc/config.toml", "./config/config.toml", "./config.toml"},
		FileDecoders: map[string]aconfig.FileDecoder{
			// from `aconfigyaml` submodule
			// see submodules in repo for more formats
			".toml": aconfigtoml.New(),
		},
	})

	if err := loader.Load(); err != nil {
		panic(err)
	}

	lvl, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Errorf("Error parse the level of logs (%v). Installed by default = Info", cfg.LogLevel)
		lvl, _ = log.ParseLevel("info")
	}
	log.SetLevel(lvl)

	cfg.Location, err = time.LoadLocation(cfg.Loc)
	if err != nil {
		log.Errorf("Error loading Location(%v):%v", cfg.Loc, err)
		cfg.Location = time.UTC
	}

	cfg.dateLayout = "2006-01-02"
	cfg.dateTimeLayout = "2006-01-02 15:04:05"

	log.Debugf("Config %#v:", cfg)

	return &cfg
}
