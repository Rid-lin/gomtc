package main

import (
	"os"
	"time"

	_ "net/http/pprof"

	. "git.vegner.org/vsvegner/gomtc/internal/config"
	. "git.vegner.org/vsvegner/gomtc/internal/pid"
	. "git.vegner.org/vsvegner/gomtc/pkg/gsshutdown"
	log "github.com/sirupsen/logrus"
)

var (
	// cfg                 Config

	fileDestination     *os.File
	csvFiletDestination *os.File
	Location            *time.Location // Global variable
)

const DateLayout = "2006-01-02"
const DateTimeLayout = "2006-01-02 15:04:05"

func Exit(ve interface{}) func(ve interface{}) {
	return func(ve interface{}) {
		cfg, ok := ve.(*Config)
		if ok {
			if !cfg.NoFlow {
				if err := fileDestination.Sync(); err != nil {
					log.Errorf("File(%v) sync error:%v", fileDestination.Name(), err)
				}
				if err := fileDestination.Close(); err != nil {
					log.Errorf("File(%v) close error:%v", fileDestination.Name(), err)
				}
			}
			if err := os.Remove(cfg.Pidfile); err != nil {
				log.Errorf("File (%v) deletion error:%v", cfg.Pidfile, err)
			}
		}
	}
}

func GetSIGHUP(vr interface{}) func(vr interface{}) {
	return func(ve interface{}) {
		var err error
		cfg, ok := ve.(*Config)
		if ok {
			log.Println("Received a signal from logrotate, close the file.")
			if err := fileDestination.Sync(); err != nil {
				log.Errorf("File(%v) sync error:%v", fileDestination.Name(), err)
			}
			if err := fileDestination.Close(); err != nil {
				log.Errorf("File(%v) close error:%v", fileDestination.Name(), err)
			}
			if !cfg.NoFlow {
				fileDestination, err = os.OpenFile(cfg.NameFileToLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					fileDestination.Close()
					log.Fatalf("Error, the '%v' file could not be created (there are not enough premissions or it is busy with another program): %v", cfg.NameFileToLog, err)
				}
			}
		}
	}
}

func main() {
	cfg := NewConfig()
	Location = cfg.Location

	if err := WritePID(cfg.Pidfile); err != nil {
		log.Error(err)
		os.Exit(2)
	}

	// TODO Сделать разделение на 3 части программы со своими параметрами, чтобы например... управление микротом было на запуск с ключом "-control", получение инфы по нетфлоу было с ключом "-flow", а просто статистика с ключом "-statistics"
	// TODO и кодовую базу разнести соответствующе (хотя бы попытаться)

	// TODO после "разделения" на три части сделать общение между частями по JSON или gRPC

	gss := NewGSS(
		Exit(cfg), cfg,
		GetSIGHUP(cfg), cfg,
	)

	t := NewTransport(cfg, gss)

	// Endless file parsing loop
	go func(cfg *Config) {
		t.runOnce(cfg)
		for {
			<-t.timerParse.C
			t.runOnce(cfg)
		}
	}(cfg)
	t.handleRequest(cfg)

}
