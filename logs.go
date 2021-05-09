package main

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func (t *Transport) readLog(cfg *Config) {
	LogToDB := readLogDB(cfg.ConfigPath)
	ArrayLogsOfJob := LogToDB.ArrayLogsOfJob

	cfg.LastDate = LogToDB.LastDate
	log.Tracef("From DB LastDate=%v(%v),LastDay=%v(%v), ArrayLogsOfJob=%v",
		LogToDB.LastDate,
		time.Unix(LogToDB.LastDate, 0).Format(cfg.dateTimeLayout),
		findOutTheCurrentDay(LogToDB.LastDate, cfg.Location),
		time.Unix(findOutTheCurrentDay(LogToDB.LastDate, cfg.Location), 0).Format(cfg.dateTimeLayout),
		ArrayLogsOfJob)
	t.Lock()
	t.logs = ArrayLogsOfJob
	t.Unlock()
}

func (t *Transport) writeLog(cfg *Config) {
	logsOfJob := LogsOfJob{
		StartTime:       cfg.startTime,
		EndTime:         cfg.endTime,
		SizeOneMegabyte: cfg.SizeOneMegabyte,
		Count: Count{
			LineParsed: cfg.totalLineParsed,
			LineSkiped: cfg.totalLineSkiped,
			LineAdded:  cfg.totalLineAdded,
			LineRead:   cfg.totalLineRead,
		},
	}
	t.Lock()
	t.logs = append(t.logs, logsOfJob)
	t.Unlock()
	t.RLock()
	LogToDB := VariableToWriteToLogDB{
		LastDate:       cfg.LastDate,
		ArrayLogsOfJob: t.logs,
	}
	t.RUnlock()
	// We add information about the written and read lines to be able to control from the web interface
	if err := writeLogDB(cfg.ConfigPath, &LogToDB); err != nil {
		log.Error(err)
	}
}
