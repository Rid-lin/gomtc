package main

import (
	"time"

	_ "net/http/pprof"

	"git.vegner.org/vsvegner/gomtc/internal/config"
)

var (
	// fileDestination     *os.File
	// csvFiletDestination *os.File
	Location *time.Location // Global variable
)

const DateLayout = "2006-01-02"
const DateTimeLayout = "2006-01-02 15:04:05"

func main() {
	var cfg *config.Config
	cfg, Location = config.NewConfig()

	t := NewTransport(cfg)

	t.Start(cfg)
}
