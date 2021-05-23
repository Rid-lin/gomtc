package main

import (
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func (t *Transport) GetInfo(request *request) ResponseType {
	var response ResponseType

	t.RLock()
	ipStruct, ok := t.infoOfDevices[request.IP]
	t.RUnlock()
	if ok {
		log.Tracef("IP:%v to MAC:%v, hostname:%v, comment:%v", ipStruct.IP, ipStruct.Mac, ipStruct.HostName, ipStruct.Comments)
		response.DeviceOldType = ipStruct.DeviceOldType
		response.Comments = ipStruct.Comments
	} else {
		log.Tracef("IP:'%v' not find in table lease of router:'%v'", ipStruct.IP, cfg.MTAddr)
		response.Mac = request.IP
		response.IP = request.IP
	}
	if response.Mac == "" {
		response.Mac = request.IP
		log.Error("Mac of Device = '' O_o", request)
	}

	return response
}

func checkNULLQuota(setValue, deafultValue uint64) uint64 {
	if setValue == 0 {
		return uint64(deafultValue)
	}
	return uint64(setValue)
}

func checkNULLQuotas(setValue, deafultValue QuotaType) QuotaType {
	quotaReturned := setValue
	if setValue.DailyQuota == 0 {
		quotaReturned.DailyQuota = deafultValue.DailyQuota
	}
	if setValue.HourlyQuota == 0 {
		quotaReturned.HourlyQuota = deafultValue.HourlyQuota
	}
	if setValue.MonthlyQuota == 0 {
		quotaReturned.MonthlyQuota = deafultValue.MonthlyQuota
	}
	return quotaReturned
}

func (t *Transport) checkQuotas() {
	hour := time.Now().Hour()
	t.Lock()
	for key, value := range t.dataCashe {
		// if key.Alias == "E8:D8:D1:47:55:93" {
		// 	runtime.Breakpoint()
		// }
		if key.Alias == "Всего" {
			continue
		}
		if value.Manual {
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
		t.data[key] = value

	}
	t.Unlock()
}

func (t *Transport) updateStatusDevicesToMT(cfg *Config) {

	t.RLock()
	BlockGroup := t.BlockAddressList
	data := t.dataCashe
	Quota := t.QuotaType
	t.RUnlock()

	for _, dInfo := range data {
		if dInfo.Manual {
			continue
		}
		if dInfo.ShouldBeBlocked && !dInfo.Blocked {
			if dInfo.QuotaType == Quota {
				dInfo.QuotaType = QuotaType{}
			}
			dInfo.Groups = dInfo.Groups + "," + BlockGroup
			dInfo.Groups = strings.Trim(dInfo.Groups, ",")
			if err := dInfo.convertToDevice().send(); err != nil {
				log.Errorf(`An error occurred while saving the device:%v`, err.Error())
			}
		} else if !dInfo.ShouldBeBlocked && dInfo.Blocked {
			if dInfo.QuotaType == Quota {
				dInfo.QuotaType = QuotaType{}
			}
			dInfo.Groups = strings.Replace(dInfo.Groups, BlockGroup, "", 1)
			dInfo.Groups = strings.ReplaceAll(dInfo.Groups, ",,", ",")
			dInfo.Groups = strings.Trim(dInfo.Groups, ",")
			if err := dInfo.convertToDevice().send(); err != nil {
				log.Errorf(`An error occurred while saving the device:%v`, err.Error())
			}
		}
	}
	t.updateDevices()
}
