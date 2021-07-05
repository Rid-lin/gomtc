package config

import (
	"os"
	"strings"
	"time"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	LastDate    int64
	LastDay     int64
	LastDayStr  string
	LastDateStr string
	Location    *time.Location

	Loc                    string   `default:"Asia/Yekaterinburg" usage:"Location for time"`
	Timezone               float32  `default:"5" usage:"Timezone east of UTC"`
	ConfigPath             string   `default:"/etc/gomtc" usage:"folder path to all config files"`
	LogPath                string   `default:"/var/log/gomtc" usage:"folder path to logs-file"`
	AssetsPath             string   `default:"/etc/gomtc/assets"  usage:"The path to the assets folder where the template files are located"`
	NameFileToLog          string   `default:"" usage:"The file where logs will be written in the format of squid logs"`
	FnStartsWith           string   `default:"access.log" usage:"Specifies where the names of the files to be parsed begin"`
	Friends                []string `default:"" usage:"List of aliases, IP addresses, friends' logins"`
	SubNets                []string `default:"" usage:"List of subnets traffic between which will not be counted"`
	IgnorList              []string `default:"" usage:"List of lines that will be excluded from the final log"`
	FlowAddr               string   `default:"0.0.0.0:2055" usage:"Address and port to listen NetFlow packets"`
	ListenAddr             string   `default:":3031" usage:"Listen address for HTTP-server"`
	LogLevel               string   `default:"info" usage:"Log level: panic, fatal, error, warn, info, debug, trace"`
	MTAddr                 string   `default:"127.0.0.1" usage:"The address of the Mikrotik router, from which the data on the comparison of the MAC address and IP address is taken"`
	SSHPort                string   `default:"22" usage:"The port of the Mikrotik router for SSH connection"`
	MTUser                 string   `default:"" usage:"User of the Mikrotik router, from which the data on the comparison of the MAC address and IP address is taken"`
	MTPass                 string   `default:"" usage:"The password of the user of the Mikrotik router, from which the data on the comparison of the mac-address and IP-address is taken"`
	ParseDelay             string   `default:"10m" usage:"Delay parsing logs"`
	BlockGroup             string   `default:"block" usage:"The name of the address list in MikrotiK with which access is blocked to users who have exceeded the quota."`
	ManualGroup            string   `default:"manual" usage:"Name of the address list in MikrotiK, in the presence of which the device is manually controlled."`
	DevicesRetryDelay      string   `default:"1m" usage:"Interval to getting info from Mikrotik"`
	ReceiveBufferSizeBytes int      `default:"" usage:"Size of RxQueue, i.e. value for SO_RCVBUF in bytes"`
	MaxSSHRetries          int      `default:"0" usage:"The number of attempts to connect to the microtik router"`
	SSHRetryDelay          uint16   `default:"0" usage:"Interval of attempts to connect to MT"`
	DefaultQuotaHourly     uint64   `default:"0" usage:"Default hourly traffic consumption quota"`
	DefaultQuotaDaily      uint64   `default:"0" usage:"Default daily traffic consumption quota"`
	DefaultQuotaMonthly    uint64   `default:"0" usage:"Default monthly traffic consumption quota"`
	SizeOneKilobyte        uint64   `default:"1024" usage:"The number of bytes in one megabyte"`
	UseTLS                 bool     `default:"false" usage:"Using TLS to connect to a router"`
	CSV                    bool     `default:"false" usage:"Output to csv"`
	// NoFlow                 bool     `default:"true" usage:"When this parameter is specified, the netflow packet listener is not launched, therefore, log files are not created, but only parsed."`
	NoControl     bool `default:"true" usage:"No need to control the Mikrotik, just read."`
	ParseAllFiles bool `default:"false" usage:"Scans all files in the folder where access.lоg is located once, deleting all data from the database"`
	// Count
}

func NewConfig() *Config {
	// fix for https://github.com/cristalhq/aconfig/issues/82
	args := []string{}
	for _, a := range os.Args {
		if !strings.HasPrefix(a, "-test.") {
			args = append(args, a)
		}
	}
	// fix for https://github.com/cristalhq/aconfig/issues/82

	var cfg Config
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		// feel free to skip some steps :)
		// SkipEnv:      true,
		// MergeFiles: true,
		SkipFiles:          false,
		AllowUnknownFlags:  true,
		AllowUnknownFields: true,
		SkipDefaults:       false,
		SkipFlags:          false,
		FailOnFileNotFound: false,
		EnvPrefix:          "GOMTC",
		FlagPrefix:         "",
		Files: []string{
			"./config.yaml",
			"./config/config.yaml",
			"/etc/gomtc/config.yaml",
			"/etc/gomtc/config/config.yaml",
			"/usr/local/gomtc/config.yaml",
			"/usr/local/gomtc/config/config.yaml",
			"/opt/gomtc/config.yaml",
			"/opt/gomtc/config/config.yaml",
			// "./config.toml",
			// "./config/config.toml",
			// "/etc/gomtc/config.toml",
			// "/etc/gomtc/config/config.toml",
		},
		FileDecoders: map[string]aconfig.FileDecoder{
			// from `aconfigyaml` submodule
			// see submodules in repo for more formats
			".yaml": aconfigyaml.New(),
			// ".toml": aconfigtoml.New(),
		},
		Args: args[1:], // [1:] важно, см. доку к FlagSet.Parse
	})

	if err := loader.Load(); err != nil {
		log.Error(err)
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
	log.Debugf("Config %#v:", cfg)

	return &cfg
}
