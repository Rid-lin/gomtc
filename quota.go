package main

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func (transport *Transport) getStatusDevices(cfg *Config) error {

	// transport.mergeMap()

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

// func (transport *Transport) mergeMap() {
// 	var key KeyMapOfReports
// 	DateStr := time.Now().In(transport.Location).Format("2006-01-02")
// 	transport.Lock()
// 	resultMap := transport.infoOfDevices
// 	for _, value := range resultMap {
// 		key = KeyMapOfReports{
// 			Alias:   value.Mac,
// 			DateStr: DateStr,
// 		}
// 		ValueMapOfReports := ValueMapOfReportsType{}
// 		ValueMapOfReports.Alias = value.Mac
// 		ValueMapOfReports.DateStr = DateStr
// 		ValueMapOfReports.InfoOfDeviceType = value
// 		transport.dataCashe[key] = ValueMapOfReports
// 	}
// 	transport.Unlock()
// }

func (transport *Transport) GetData(key KeyMapOfReports) (ValueMapOfReportsType, error) {
	transport.RLock()
	if data, ok := transport.dataCashe[key]; ok {
		transport.RUnlock()
		return data, nil
	}
	transport.RUnlock()
	return ValueMapOfReportsType{}, fmt.Errorf("Map element with key(%v) not found", key)
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
			value.ShouldBeBlocked = true
			log.Tracef("Login (%v) was disabled due to exceeding the daily quota(%v)", key.Alias, value.DailyQuota)
		} else if value.SizeOfHour[hour] >= value.HourlyQuota {
			value.ShouldBeBlocked = true
			log.Tracef("Login (%v) was disabled due to exceeding the hourly quota(%v)", key.Alias, value.DailyQuota)
		} else if value.ShouldBeBlocked {
			value.ShouldBeBlocked = false
			log.Tracef("Login (%v)has been enabled, the quota(%v) has not been exceeded", key.Alias, value.HourlyQuota)
		}
		transport.data[key] = value

	}
	transport.Unlock()
}

func (transport *Transport) updateStatusDevicesToMT(cfg *Config) {

	transport.RLock()
	BlockGroup := transport.BlockAddressList
	data := transport.dataCashe
	Quota := transport.QuotaType
	transport.RUnlock()

	for _, device := range data {
		if device.Manual {
			continue
		}
		if device.ShouldBeBlocked && !device.Blocked {
			if device.QuotaType == Quota {
				device.QuotaType = QuotaType{}
			}
			device.Groups = device.Groups + "," + BlockGroup
			device.Groups = strings.Trim(device.Groups, ",")
			if err := transport.setGroupOfDeviceToMT(device.InfoOfDeviceType); err != nil {
				log.Errorf(`An error occurred while saving the device(%v):%v`, device, err.Error())
			}
		} else if !device.ShouldBeBlocked && device.Blocked {
			if device.QuotaType == Quota {
				device.QuotaType = QuotaType{}
			}
			device.Groups = strings.Replace(device.Groups, BlockGroup, "", 1)
			device.Groups = strings.ReplaceAll(device.Groups, ",,", ",")
			device.Groups = strings.Trim(device.Groups, ",")
			if err := transport.setGroupOfDeviceToMT(device.InfoOfDeviceType); err != nil {
				log.Errorf(`An error occurred while saving the device(%v):%v`, device, err.Error())
			}
		}
	}

	transport.updateDataFromMT()

}
