package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func (transport *Transport) getStatusDevices(cfg *Config) error {
	pathToPost := "http://" + cfg.GonsquidAddr + "/getstatusdevices"
	pathToPost = strings.TrimSpace(pathToPost)
	resp, err := http.Get(pathToPost)
	if err != nil {
		log.Errorf("Error in receiving information from the remote device management server(%v):%v", pathToPost, err)
		return err
	}

	result := new(map[string]responseMapType)
	// result := new(map[string]interface{})

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	if err != nil {
		log.Errorf("%v", err)
	}

	resultMap := *result
	defaultLine := resultMap["default"]

	DefaultHourlyQuota := defaultLine.HourlyQuota
	DefaultDailyQuota := defaultLine.DailyQuota
	DefaultMonthlyQuota := defaultLine.MonthlyQuota

	transport.Lock()
	transport.HourlyQuota = uint64(DefaultHourlyQuota)
	transport.DailyQuota = uint64(DefaultDailyQuota)
	transport.MonthlyQuota = uint64(DefaultMonthlyQuota)

	delete(resultMap, "default")

	for key, value := range transport.dataChashe {
		value.HourlyQuota = checkNULLQuota(0, DefaultHourlyQuota)
		value.DailyQuota = checkNULLQuota(0, DefaultDailyQuota)
		value.MonthlyQuota = checkNULLQuota(0, DefaultMonthlyQuota)
		for _, value2 := range resultMap {
			if key.Alias == value2.IP || key.Alias == value2.Mac || key.Alias == value2.HostName {
				// value.HourlyQuota = checkNULLQuota(value2.HourlyQuota, DefaultHourlyQuota)
				// value.DailyQuota = checkNULLQuota(value2.DailyQuota, DefaultDailyQuota)
				// value.MonthlyQuota = checkNULLQuota(value2.MonthlyQuota, DefaultMonthlyQuota)
				value.PersonType = value2.PersonType
				value.DeviceType = value2.DeviceType
				value.QuotaType = value2.QuotaType
				// if value.HourlyQuota == 0 {
				// 	runtime.Breakpoint()
				// }
				// value.Blocked = value2.Blocked
				// value.Comments = value2.Comments
				// value.Group = value2.Group
				// value.HostName = value2.HostName
				// value.IP = value2.IP
				// value.Id = value2.Id
				// value.Mac = value2.Mac
				// value.Name = value2.Name
				// value.Position = value2.Position
				break
			}

		}
		transport.dataChashe[key] = value
	}
	transport.Unlock()
	return nil
}

func checkNULLQuota(setValue, deafultValue uint64) uint64 {
	if setValue == 0 {
		return uint64(deafultValue)
	}
	return uint64(setValue)
}

func (transport *Transport) checkQuota() {
	hour := time.Now().Hour()
	transport.Lock()
	for key, value := range transport.dataChashe {
		if key.Alias == "Всего" {
			continue
		}
		if value.SizeInBytes >= value.DailyQuota {
			value.Blocked = true
			log.Tracef("Login (%v) was disabled due to exceeding the daily quota(%v)", key.Alias, value.DailyQuota)
		} else if value.SizeOfHour[hour] >= value.HourlyQuota {
			value.Blocked = true
			log.Tracef("Login (%v) was disabled due to exceeding the hourly quota(%v)", key.Alias, value.DailyQuota)
		} else if value.Blocked {
			value.Blocked = false
			log.Tracef("Login (%v)has been enabled, the quota(%v) has not been exceeded", key.Alias, value.HourlyQuota)
		}
		transport.data[key] = value

	}
	transport.Unlock()
}

func (transport *Transport) sendStatusDevices(cfg *Config) {
	SliceToBlock := map[string]bool{}
	transport.RLock()
	for key, value := range transport.data {
		SliceToBlock[key.Alias] = value.Blocked
	}
	transport.RUnlock()

	bytesRepresentation, err := json.MarshalIndent(SliceToBlock, "", "	")
	if err != nil {
		log.Error(err)
	}

	log.Tracef("%s", bytesRepresentation)

	// pathToPost := cfg.GonsquidAddr + "/setstatusdevices"

	// resp, err := http.Post(pathToPost, "application/json", bytes.NewBuffer(bytesRepresentation))
	// if err != nil {
	// 	log.Errorf("Error in receiving information from the remote device management server(%v):%v", pathToPost, err)
	// 	return
	// }

	// result := new(map[string]interface{})

	// decoder := json.NewDecoder(resp.Body)
	// err = decoder.Decode(&result)
	// if err != nil {
	// 	log.Errorf("%v", err)
	// }

}
