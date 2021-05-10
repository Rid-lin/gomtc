package main

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
)

func (transport *Transport) getStatusDevices(cfg *Config) error {
	transport.Lock()
	resultMap := transport.infoOfDevices

	DefaultHourlyQuota := transport.HourlyQuota
	DefaultDailyQuota := transport.DailyQuota
	DefaultMonthlyQuota := transport.MonthlyQuota

	for key, value := range transport.dataCashe {
		value.HourlyQuota = checkNULLQuota(0, DefaultHourlyQuota)
		value.DailyQuota = checkNULLQuota(0, DefaultDailyQuota)
		value.MonthlyQuota = checkNULLQuota(0, DefaultMonthlyQuota)
		for _, value2 := range resultMap {
			if key.Alias == value2.IP || key.Alias == value2.Mac || key.Alias == value2.HostName {
				value.PersonType = value2.PersonType
				value.DeviceType = value2.DeviceType
				value.QuotaType = value2.QuotaType
				break
			}

		}
		transport.dataCashe[key] = value
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
	for key, value := range transport.dataCashe {
		if key.Alias == "Всего" {
			continue
		}
		if value.Size >= value.DailyQuota {
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
