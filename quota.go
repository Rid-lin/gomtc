package main

import (
	"fmt"
	"time"

	"github.com/bmuller/arrow"

	log "github.com/sirupsen/logrus"
)

func (t *Transport) GetInfo(request *request) ResponseType {
	var response ResponseType

	t.RLock()
	ipStruct, ok := t.Infos[request.IP]
	t.RUnlock()
	if ok {
		log.Tracef("IP:%v to MAC:%v, hostname:%v, comment:%v", ipStruct.IP, ipStruct.Mac, ipStruct.HostName, ipStruct.Comments)
		response.DeviceOldType = ipStruct.DeviceOldType
		response.Comments = ipStruct.Comments
	} else {
		log.Tracef("IP:'%v' not find in table lease of router", ipStruct.IP)
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
	t.RLock()
	quota := t.QuotaType
	p := parseType{
		SSHCredentials:   t.sshCredentials,
		QuotaType:        t.QuotaType,
		BlockAddressList: t.BlockAddressList,
		Location:         t.Location,
	}
	t.RUnlock()
	tn := time.Now().Format("2006-01-02")
	hour := time.Now().Hour()
	t.Lock()
	for key, alias := range t.dataCasheOld {
		switch {
		case key.Alias == "Всего":
			continue
		case alias.Manual:
			continue
		case (alias.VolumePerDay >= alias.DailyQuota || alias.VolumePerHour[hour] >= alias.HourlyQuota) && alias.DateStr == tn && alias.Blocked:
			continue
		case alias.VolumePerDay >= alias.DailyQuota && alias.DateStr == tn && !alias.Blocked:
			alias.ShouldBeBlocked = true
			// alias.TimeoutBlock = setDailyTimeout()
			alias.addBlockGroup(p.BlockAddressList)
			t.change = append(t.change, alias)
			log.Debugf("Login (%17s) was disabled due to exceeding the daily quota(%v)", alias.Alias, alias.DailyQuota)
		case alias.VolumePerHour[hour] >= alias.HourlyQuota && alias.DateStr == tn && !alias.Blocked:
			alias.ShouldBeBlocked = true
			// alias.TimeoutBlock = setHourlyTimeout()
			alias.addBlockGroup(p.BlockAddressList)
			t.change = append(t.change, alias)
			log.Debugf("Login (%17s) was disabled due to exceeding the hourly quota(%v)", alias.Alias, alias.DailyQuota)
		case alias.Blocked && alias.DateStr == tn:
			alias.ShouldBeBlocked = false
			alias.delBlockGroup(p.BlockAddressList)
			t.change = append(t.change, alias)
			log.Debugf("Login (%17s) has been enabled, the quota(%v) has not been exceeded(Blocked:%v)",
				alias.Alias, alias.HourlyQuota, alias.Blocked)
		}
		t.dataOld[key] = alias

	}
	for _, device := range t.devices {
		for _, blockedDevice := range t.change {
			if device.macAddress == blockedDevice.Mac || device.macAddress == blockedDevice.Alias || device.activeMacAddress == blockedDevice.Mac {
				break
			}
		}
		if paramertToBool(device.blocked) {
			device.ShouldBeBlocked = false
			alias := AliasOld{
				Alias:   device.macAddress,
				DateStr: time.Now().In(t.Location).Format("2006-01-02"),
			}
			alias.delBlockGroup(p.BlockAddressList)
			t.change = append(t.change, alias)
		}
	}
	t.Unlock()
	t.change.sendLeaseSet(p, quota)
	t.Lock()
	t.change = AliasesOldType{}
	t.Unlock()
}

func setHourlyTimeout() string {
	endOfHour := arrow.Now().AddHours(1).AtBeginningOfHour()
	delta := endOfHour.Sub(arrow.Now()).Seconds()
	if delta-31 < 0 {
		delta = 0
	} else {
		delta = delta - 31
	}
	return fmt.Sprintf("00:00:%.0f", delta)
}

func setDailyTimeout() string {
	endOfDay := arrow.Now().AddDays(1).AtBeginningOfDay()
	delta := endOfDay.Sub(arrow.Now()).Seconds()
	if delta-31 < 0 {
		delta = 0
	} else {
		delta = delta - 31
	}
	return fmt.Sprintf("00:00:%.0f", delta)
}
