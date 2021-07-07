package main

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"git.vegner.org/vsvegner/gomtc/internal/config"

	log "github.com/sirupsen/logrus"
)

func (t *Transport) parseLog(cfg *config.Config) {
	if cfg.ParseAllFiles {
		// Getting the current time to calculate the running time
		t.StartTime = time.Now()
		fmt.Printf("Parsing has started.\r")
		tn := time.Unix(0, 0)
		t.DeletingDateData(tn.Format(DateLayout), path.Join(cfg.ConfigPath, "sqlite.db"))
		t.Lock()
		t.LastDate = tn.Unix()
		t.Unlock()
		t.parseAllFilesAndCountingTraffic(cfg)
		t.timeCalculationAndPrinting()
		cfg.ParseAllFiles = false
	} else {
		// Getting the current time to calculate the running time
		t.StartTime = time.Now()
		fmt.Printf("Parsing has started.\r")
		tn := time.Now().In(Location)
		t.DeletingDateData(tn.Format(DateLayout), path.Join(cfg.ConfigPath, "sqlite.db"))
		t.Lock()
		td := time.Date(tn.Year(), tn.Month(), tn.Day(), 0, 0, 1, 0, Location)
		t.LastDate = td.Unix()
		t.StartTime = time.Now()
		t.Unlock()
		if err := t.parseOneFilesAndCountingTraffic(path.Join(cfg.LogPath, cfg.FnStartsWith), cfg); err != nil {
			log.Error(err)
		}
		t.SumAndReset()
		t.timeCalculationAndPrinting()
	}
}

func replaceQuotes(lineOld string) string {
	lineNew := strings.ReplaceAll(lineOld, "'", "&quot")
	line := strings.ReplaceAll(lineNew, `"`, "&quot")
	return line
}

func SortFileByModTime(files []os.FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Unix() < files[j].ModTime().Unix()
	})
}

func filtredMessage(message string, IgnorList []string) string {
	for _, ignorStr := range IgnorList {
		if strings.Contains(message, ignorStr) {
			log.Tracef("Line of log :%v contains ignorstr:%v, skipping...", message, ignorStr)
			return ""
		}
	}
	return message
}
