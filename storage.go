package main

import (
	"time"
)

// type VariableToWriteToLogDB struct {
// 	LastDate       int64
// 	ArrayLogsOfJob []LogsOfJob
// }

type LogsOfJob struct {
	StartTime       time.Time
	EndTime         time.Time
	SizeOneKilobyte uint64
	Count
}

// func writeLogDB(path string, LogToDB *VariableToWriteToLogDB) error {
// 	if _, err := os.Stat(path); err != nil {
// 		if os.IsNotExist(err) {
// 			if err := os.Mkdir(path, 0644); err != nil {
// 				return fmt.Errorf("Error to create folder:%v", err)
// 			}
// 		} else {
// 			return fmt.Errorf("Error to read folder:%v", err)
// 		}
// 	}

// 	path = filepath.Join(path, "log.gob")

// 	// Oppen a file
// 	dataFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
// 	if err != nil {
// 		return fmt.Errorf("Error to read log-file:%v", err)
// 	}

// 	// serialize the data
// 	dataEncoder := gob.NewEncoder(dataFile)
// 	if err := dataEncoder.Encode(LogToDB); err != nil {
// 		return fmt.Errorf("Error to write log-file:%v", err)
// 	}
// 	dataFile.Close()
// 	return nil
// }

// func readLogDB(path string) VariableToWriteToLogDB {
// 	var data VariableToWriteToLogDB

// 	path = filepath.Join(path, "log.gob")
// 	// Open data file
// 	dataFile, err := os.Open(path)

// 	if err != nil {
// 		log.Errorf("Error to read log-file(path):%v", err)
// 	}

// 	dataDecoder := gob.NewDecoder(dataFile)
// 	err = dataDecoder.Decode(&data)

// 	if err != nil {
// 		log.Errorf("Error to decode log-file(path):%v", err)
// 	}

// 	dataFile.Close()

// 	return data
// }

// func saveDeviceToCSV(devices []DeviceType) error {
// 	f, e := os.Create("./Device.csv")
// 	if e != nil {
// 		fmt.Println(e)
// 	}

// 	writer := csv.NewWriter(f)

// 	for _, device := range devices {
// 		fields := device.convertToSlice()
// 		err := writer.Write(fields)
// 		if err != nil {
// 			fmt.Println(e)
// 		}
// 	}
// 	return nil
// }

// func (d DeviceType) convertToSlice() []string {
// 	var slice []string
// 	slice = append(slice, d.Id)
// 	slice = append(slice, d.activeAddress)
// 	slice = append(slice, d.activeClientId)
// 	slice = append(slice, d.activeMacAddress)
// 	slice = append(slice, d.activeServer)
// 	slice = append(slice, d.address)
// 	slice = append(slice, d.addressLists)
// 	slice = append(slice, fmt.Sprint(d.blocked))
// 	slice = append(slice, d.clientId)
// 	slice = append(slice, d.comment)
// 	slice = append(slice, d.dhcpOption)
// 	slice = append(slice, d.disabledL)
// 	slice = append(slice, d.dynamic)
// 	slice = append(slice, d.expiresAfter)
// 	slice = append(slice, d.hostName)
// 	slice = append(slice, d.lastSeen)
// 	slice = append(slice, d.macAddress)
// 	slice = append(slice, d.radius)
// 	slice = append(slice, d.server)
// 	slice = append(slice, d.status)
// 	slice = append(slice, fmt.Sprint(d.Manual))
// 	slice = append(slice, fmt.Sprint(d.ShouldBeBlocked))
// 	slice = append(slice, d.timeout.Format("2006-01-02 15:04:05 -0700 MST "))
// 	return slice
// }
