package main

import (
	"strings"
	"time"
)

func (t *Transport) getDevices() {
	t.Lock()
	devices := parseInfoFromMTAsValueToSlice(parseType{
		SSHCredentials:   t.sshCredentials,
		QuotaType:        t.QuotaType,
		BlockAddressList: t.BlockAddressList,
		Location:         t.Location,
	})
	for _, device := range devices {
		t.devices[KeyDevice{ip: device.activeAddress, mac: device.activeMacAddress}] = device
	}
	t.lastUpdatedMT = time.Now()
	t.Unlock()
}

func (d *DeviceType) Update(p parseType) {
	// TODO
	// Сделать отправку информации об устройстве отбратно в микротик
	// Обратить внимание, на:
	// квоту, если она равна той которая по-умолчанию, то не писать её
	// Блокировку
	// Ручное управление
	// Полное время
	return nil
}

func (d *DeviceType) ToQuota() QuotaType {
	var q QuotaType
	commentArray := strings.Split(d.comment, "/")
	for _, value := range commentArray {
		switch {
		case strings.Contains(value, "quotahourly="):
			q.HourlyQuota = parseParamertToUint(value)
		case strings.Contains(value, "quotadaily="):
			q.DailyQuota = parseParamertToUint(value)
		case strings.Contains(value, "quotamonthly="):
			q.MonthlyQuota = parseParamertToUint(value)
		case strings.Contains(value, "manual="):
			q.Manual = parseParamertToBool(value)
		}
	}
	return q
}

func (d *DeviceType) ToPerson() PersonType {
	var p PersonType
	var comments string
	commentArray := strings.Split(d.comment, "/")
	for _, value := range commentArray {
		switch {
		case strings.Contains(value, "tel="):
			p.TypeD = "tel"
			p.Name = parseParamertToStr(value)
		case strings.Contains(value, "nb="):
			p.TypeD = "nb"
			p.Name = parseParamertToStr(value)
		case strings.Contains(value, "ws="):
			p.TypeD = "ws"
			p.Name = parseParamertToStr(value)
		case strings.Contains(value, "srv"):
			p.TypeD = "srv"
			p.Name = parseParamertToStr(value)
		case strings.Contains(value, "prn="):
			p.TypeD = "prn"
			p.Name = parseParamertToStr(value)
		case strings.Contains(value, "ap="):
			p.TypeD = "ap"
			p.Name = parseParamertToStr(value)
		case strings.Contains(value, "name="):
			p.TypeD = "other"
			p.Name = parseParamertToStr(value)
		case strings.Contains(value, "col="):
			p.Position = parseParamertToStr(value)
		case strings.Contains(value, "pos="):
			p.Position = parseParamertToStr(value)
		case strings.Contains(value, "com="):
			p.Company = parseParamertToStr(value)
		case strings.Contains(value, "id="):
			p.IDUser = parseParamertToStr(value)
		case strings.Contains(value, "comment="):
			comments = parseParamertToStr(value)
		case value != "":
			comments = comments + "/" + value
		}
	}
	return p
}
