package main

import (
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
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
		device.Manual = inAddressList(device.AddressLists, t.ManualAddresList)
		device.Blocked = inAddressList(device.AddressLists, t.BlockAddressList)
		t.devices[KeyDevice{
			// ip: device.ActiveAddress,
			mac: device.ActiveMacAddress}] = device
	}
	t.lastUpdatedMT = time.Now()
	t.Unlock()
}

func (d *DeviceType) ToQuota() QuotaType {
	var q QuotaType
	commentArray := strings.Split(d.Comment, "/")
	for _, value := range commentArray {
		switch {
		case strings.Contains(value, "quotahourly="):
			q.HourlyQuota = parseParamertToUint(value)
		case strings.Contains(value, "quotadaily="):
			q.DailyQuota = parseParamertToUint(value)
		case strings.Contains(value, "quotamonthly="):
			q.MonthlyQuota = parseParamertToUint(value)
			// case strings.Contains(value, "manual="):
			// 	q.Manual = parseParamertToBool(value)
		}
		q.Blocked = q.Blocked || d.Blocked
		q.Disabled = q.Disabled || paramertToBool(d.disabledL)
	}
	return q
}

func (d *DeviceType) ToPerson() PersonType {
	var p PersonType
	var comments string
	commentArray := strings.Split(d.Comment, "/")
	for _, value := range commentArray {
		switch {
		// case strings.Contains(value, "tel="):
		// 	p.TypeD = "tel"
		// 	p.Name = parseParamertToStr(value)
		// case strings.Contains(value, "nb="):
		// 	p.TypeD = "nb"
		// 	p.Name = parseParamertToStr(value)
		// case strings.Contains(value, "ws="):
		// 	p.TypeD = "ws"
		// 	p.Name = parseParamertToStr(value)
		// case strings.Contains(value, "srv"):
		// 	p.TypeD = "srv"
		// 	p.Name = parseParamertToStr(value)
		// case strings.Contains(value, "prn="):
		// 	p.TypeD = "prn"
		// 	p.Name = parseParamertToStr(value)
		// case strings.Contains(value, "ap="):
		// 	p.TypeD = "ap"
		// 	p.Name = parseParamertToStr(value)
		// case strings.Contains(value, "name="):
		// 	p.TypeD = "other"
		// p.Name = parseParamertToStr(value)
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

func (d *DeviceType) ParseComment() {
	var comments string
	commentArray := strings.Split(d.Comment, "/")
	for _, value := range commentArray {
		switch {
		case strings.Contains(value, "tel="):
			d.TypeD = "tel"
		case strings.Contains(value, "nb="):
			d.TypeD = "nb"
		case strings.Contains(value, "ws="):
			d.TypeD = "ws"
		case strings.Contains(value, "srv"):
			d.TypeD = "srv"
		case strings.Contains(value, "prn="):
			d.TypeD = "prn"
		case strings.Contains(value, "ap="):
			d.TypeD = "ap"
		case strings.Contains(value, "name="):
			d.TypeD = "other"
		case strings.Contains(value, "comment="):
			comments = parseParamertToStr(value)
		case value != "":
			comments = comments + "/" + value
		}
	}
}

func (d DeviceType) Block(group string, key KeyDevice) DeviceType {
	d.AddressLists = d.AddressLists + "," + group
	d.AddressLists = strings.Trim(d.AddressLists, ",")
	d.AddressLists = strings.ReplaceAll(d.AddressLists, `"`, "")
	d.Blocked = true
	d.ShouldBeBlocked = true
	log.Debugf("Device (%17v;%15v;%17v;%17v) was disabled due to exceeding the quota", key.mac, d.ActiveAddress, d.ActiveMacAddress, d.macAddress)
	return d
}

func (d DeviceType) UnBlock(group string, key KeyDevice) DeviceType {
	d.AddressLists = strings.Replace(d.AddressLists, group, "", 1)
	d.AddressLists = strings.ReplaceAll(d.AddressLists, ",,", ",")
	d.AddressLists = strings.Trim(d.AddressLists, ",")
	d.AddressLists = strings.ReplaceAll(d.AddressLists, `"`, "")
	log.Debugf("Device (%17v;%15v;%17v;%17v) has been enabled, the quota has not been exceeded", key.mac, d.ActiveAddress, d.ActiveMacAddress, d.macAddress)
	d.Blocked = false
	d.ShouldBeBlocked = false
	return d
}

func (t *Transport) SendGroupStatus() {
	t.RLock()
	p := parseType{
		SSHCredentials:   t.sshCredentials,
		QuotaType:        t.QuotaType,
		BlockAddressList: t.BlockAddressList,
		Location:         t.Location,
	}
	t.RUnlock()
	t.change.sendLeaseSet(p)
	t.Lock()
	t.change = BlockDevices{}
	t.Unlock()
}
