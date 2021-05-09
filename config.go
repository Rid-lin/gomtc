package main

import (
	"time"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigtoml"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	// ConfigFilename         string   `default:"" usage:`
	SubNets                []string `default:"" usage:"List of subnets traffic between which will not be counted"`
	IgnorList              []string `default:"" usage:"List of lines that will be excluded from the final log"`
	LogLevel               string   `default:"info" usage:"Log level: panic, fatal, error, warn, info, debug, trace"`
	FlowAddr               string   `default:"0.0.0.0:2055" usage:"Address and port to listen NetFlow packets"`
	NameFileToLog          string   `default:"" usage:"The file where logs will be written in the format of squid logs"`
	BindAddr               string   `default:":3030" usage:"Listen address for response mac-address from mikrotik"`
	MTAddr                 string   `default:"" usage:"The address of the Mikrotik router, from which the data on the comparison of the MAC address and IP address is taken"`
	MTUser                 string   `default:"" usage:"User of the Mikrotik router, from which the data on the comparison of the MAC address and IP address is taken"`
	MTPass                 string   `default:"" usage:"The password of the user of the Mikrotik router, from which the data on the comparison of the mac-address and IP-address is taken"`
	Loc                    string   `default:"Asia/Yekaterinburg" usage:"Location for time"`
	Interval               string   `default:"10m" usage:"Interval to getting info from Mikrotik"`
	ReceiveBufferSizeBytes int      `default:"" usage:"Size of RxQueue, i.e. value for SO_RCVBUF in bytes"`
	NumOfTryingConnectToMT int      `default:"10" usage:"The number of attempts to connect to the microtik router"`
	DefaultQuotaHourly     uint     `default:"0" usage:"Default hourly traffic consumption quota"`
	DefaultQuotaDaily      uint     `default:"0" usage:"Default daily traffic consumption quota"`
	DefaultQuotaMonthly    uint     `default:"0" usage:"Default monthly traffic consumption quota"`
	SizeOneMegabyte        uint     `default:"1048576" usage:"The number of bytes in one megabyte"`
	UseTLS                 bool     `default:"false" usage:"Using TLS to connect to a router"`
	CSV                    bool     `default:"false" usage:"Output to csv"`
	Location               *time.Location
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
		AllowUnknownFields: true,
		SkipDefaults:       false,
		SkipFlags:          false,
		EnvPrefix:          "GOMTC",
		FlagPrefix:         "",
		Files:              []string{"/etc/gomtc/config.toml", "./config.toml", "./config/config.toml"},
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

	log.Debugf("Config %#v:", cfg)

	return &cfg
}
