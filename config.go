package main

import (
	"flag"

	"github.com/ilyakaznacheev/cleanenv"
	log "github.com/sirupsen/logrus"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "List of strings"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type Config struct {
	SubNets                arrayFlags `yaml:"SubNets" toml:"subnets" env:"SUBNETS"`
	IgnorList              arrayFlags `yaml:"IgnorList" toml:"ignorlist" env:"IGNORLIST"`
	LogLevel               string     `yaml:"LogLevel" toml:"loglevel" env:"LOG_LEVEL"`
	FlowAddr               string     `yaml:"FlowAddr" toml:"flowaddr" env:"FLOW_ADDR" env-default:"0.0.0.0:2055"`
	NameFileToLog          string     `yaml:"FileToLog" toml:"log" env:"FLOW_LOG"`
	BindAddr               string     `yaml:"BindAddr" toml:"bindaddr" env:"ADDR_M4M" envdefault:":3030"`
	MTAddr                 string     `yaml:"MTAddr" toml:"mtaddr" env:"ADDR_MT"`
	MTUser                 string     `yaml:"MTUser" toml:"mtuser" env:"USER_MT"`
	MTPass                 string     `yaml:"MTPass" toml:"mtpass" env:"PASS_MT"`
	GMT                    string     `yaml:"GMT" toml:"gmt" env:"GMT"`
	Interval               string
	SQLArddr               string `yaml:"SQLArddr" toml:"sqlarddr" env:"SQL_ADDR"`
	receiveBufferSizeBytes int    `yaml:"receiveBufferSizeBytes" toml:"receiveBufferSizeBytes" env:"GONFLUX_BUFSIZE"`
	useTLS                 bool   `yaml:"tls" toml:"tls" env:"TLS"`
}

var (
	cfg Config
)

func newConfig(configFilename string) *Config {
	/* Parse command-line arguments */
	flag.IntVar(&cfg.receiveBufferSizeBytes, "buffer", 212992, "Size of RxQueue, i.e. value for SO_RCVBUF in bytes")
	flag.StringVar(&cfg.FlowAddr, "addr", "0.0.0.0:2055", "Address and port to listen NetFlow packets")
	flag.StringVar(&cfg.LogLevel, "loglevel", "info", "Log level")
	flag.Var(&cfg.SubNets, "subnet", "List of subnets traffic between which will not be counted")
	flag.Var(&cfg.IgnorList, "ignorlist", "List of lines that will be excluded from the final log")
	flag.StringVar(&cfg.NameFileToLog, "log", "", "The file where logs will be written in the format of squid logs")
	flag.StringVar(&cfg.GMT, "gmt", "+0500", "GMT offset time")
	flag.StringVar(&cfg.MTAddr, "mtaddr", "", "The address of the Mikrotik router, from which the data on the comparison of the MAC address and IP address is taken")
	flag.StringVar(&cfg.MTUser, "u", "", "User of the Mikrotik router, from which the data on the comparison of the MAC address and IP address is taken")
	flag.StringVar(&cfg.MTPass, "p", "", "The password of the user of the Mikrotik router, from which the data on the comparison of the mac-address and IP-address is taken")
	flag.StringVar(&cfg.BindAddr, "m4maddr", ":3030", "Listen address for HTTP-server")
	flag.StringVar(&cfg.SQLArddr, "sqladdr", "", "string to connect DB (e.g. username:password@protocol(address)/dbname?param=value) More details in https://github.com/go-sql-driver/mysql#dsn-data-source-name")
	flag.StringVar(&cfg.Interval, "interval", "10m", "Interval to getting info from Mikrotik")
	flag.BoolVar(&cfg.useTLS, "tls", false, "Using TLS to connect to a router")

	flag.Parse()

	var config_source string
	err := cleanenv.ReadConfig(configFilename, &cfg)
	if err != nil {
		log.Warningf("No config file(%v) found: %v", configFilename, err)
		config_source = "ENV/CFG"
	} else {
		config_source = "CLI"
	}

	lvl, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Errorf("Error in determining the level of logs (%v). Installed by default = Info", cfg.LogLevel)
		lvl, _ = log.ParseLevel("info")
	}
	log.SetLevel(lvl)

	log.Debugf("Config read from %s: %#v",
		config_source,
		cfg)

	return &cfg
}
