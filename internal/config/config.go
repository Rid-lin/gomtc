package config

import (
	"os"
	"path"
	"strings"
	"time"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"

	log "github.com/sirupsen/logrus"
)

var Location *time.Location

type Config struct {
	DSN                 string   `default:"" usage:"Database server name (default:$confg_patch/sqlite.db)"`
	Loc                 string   `default:"Asia/Yekaterinburg" usage:"Location for time"`
	Timezone            float32  `default:"5" usage:"Timezone east of UTC"`
	ConfigPath          string   `default:"/etc/gomtc" usage:"folder path to all config files"`
	LogPath             string   `default:"/var/log/gomtc" usage:"folder path to logs-file"`
	AssetsPath          string   `default:"/etc/gomtc/assets"  usage:"The path to the assets folder where the template files are located"`
	FnStartsWith        string   `default:"access.log" usage:"Specifies where the names of the files to be parsed begin"`
	Friends             []string `default:"" usage:"List of aliases, IP addresses, friends' logins"`
	IgnorList           []string `default:"" usage:"List of lines that will be excluded from the final log"`
	ListenAddr          string   `default:":3031" usage:"Listen address for HTTP-server"`
	LogLevel            string   `default:"info" usage:"Log level: panic, fatal, error, warn, info, debug, trace"`
	GomtcSshHost        string   `default:"http://127.0.0.1:3034" usage:"The address of the Mikrotik router, from which the data on the comparison of the MAC address and IP address is taken"`
	ParseDelay          string   `default:"10m" usage:"Delay parsing logs"`
	BlockGroup          string   `default:"block" usage:"The name of the address list in MikrotiK with which access is blocked to users who have exceeded the quota."`
	ManualGroup         string   `default:"manual" usage:"Name of the address list in MikrotiK, in the presence of which the device is manually controlled."`
	DefaultQuotaHourly  uint64   `default:"0" usage:"Default hourly traffic consumption quota"`
	DefaultQuotaDaily   uint64   `default:"0" usage:"Default daily traffic consumption quota"`
	DefaultQuotaMonthly uint64   `default:"0" usage:"Default monthly traffic consumption quota"`
	SizeOneKilobyte     uint64   `default:"1024" usage:"The number of bytes in one megabyte"`
	NoControl           bool     `default:"true" usage:"No need to control the Mikrotik, just read."`
	NoMT                bool     `default:"true" usage:"Disables all communication with Mikrotik"`
	ParseAllFiles       bool     `default:"false" usage:"Scans all files in the folder where access.lоg is located once, deleting all data from the database"`
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
		},
		FileDecoders: map[string]aconfig.FileDecoder{
			// from `aconfigyaml` submodule
			// see submodules in repo for more formats
			".yaml": aconfigyaml.New(),
		},
		Args: args[1:], // [1:] важно, см. доку к FlagSet.Parse
	})

	if err := loader.Load(); err != nil {
		log.Error(err)
	}

	if cfg.DSN == "" {
		cfg.DSN = path.Join(cfg.ConfigPath, "sqlite.db")
	}
	lvl, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Errorf("Error parse the level of logs (%v). Installed by default = Info", cfg.LogLevel)
		lvl, _ = log.ParseLevel("info")
	}
	log.SetLevel(lvl)

	Location, err = time.LoadLocation(cfg.Loc)
	if err != nil {
		log.Errorf("Error loading Location(%v):%v", cfg.Loc, err)
		Location = time.UTC
	}
	log.Debugf("Config %#v:", cfg)

	return &cfg
}
