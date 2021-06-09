package main

import (
	log "github.com/sirupsen/logrus"
)

func (t *Transport) GetInfo(request *request) ResponseType {
	var response ResponseType
	t.RLock()
	for _, device := range t.devices {
		switch {
		case request.IP == device.activeAddress && device.activeMacAddress != "":
			response.Mac = device.activeMacAddress
			response.Comments = device.comment
		case request.IP == device.activeAddress && device.activeMacAddress == "":
			response.Mac = request.IP
			response.Comments = device.comment
		}
	}
	t.RUnlock()
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
