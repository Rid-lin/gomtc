package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

type VariableToWriteToLogDB struct {
	LastDate       int64
	ArrayLogsOfJob []LogsOfJob
}

type LogsOfJob struct {
	StartTime       time.Time
	EndTime         time.Time
	SizeOneKilobyte uint64
	Count
}

func writeLogDB(path string, LogToDB *VariableToWriteToLogDB) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(path, 0644); err != nil {
				return fmt.Errorf("Error to create folder:%v", err)
			}
		} else {
			return fmt.Errorf("Error to read folder:%v", err)
		}
	}

	path = filepath.Join(path, "log.gob")

	// Oppen a file
	dataFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Error to read log-file:%v", err)
	}

	// serialize the data
	dataEncoder := gob.NewEncoder(dataFile)
	if err := dataEncoder.Encode(LogToDB); err != nil {
		return fmt.Errorf("Error to write log-file:%v", err)
	}
	dataFile.Close()
	return nil
}

func readLogDB(path string) VariableToWriteToLogDB {
	var data VariableToWriteToLogDB

	path = filepath.Join(path, "log.gob")
	// Open data file
	dataFile, err := os.Open(path)

	if err != nil {
		log.Errorf("Error to read log-file(path):%v", err)
	}

	dataDecoder := gob.NewDecoder(dataFile)
	err = dataDecoder.Decode(&data)

	if err != nil {
		log.Errorf("Error to decode log-file(path):%v", err)
	}

	dataFile.Close()

	return data
}
