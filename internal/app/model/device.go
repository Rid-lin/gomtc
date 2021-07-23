package model

import (
	"strings"

	"git.vegner.org/vsvegner/gomtc/internal/app/validation"
	"github.com/sirupsen/logrus"
)

type DeviceType struct {
	// From MT
	Id                  string `json:"Id"`
	ActiveAddress       string `json:"ActiveAddress"`  // 192.168.65.85
	ActiveClientId      string `json:"ActiveClientId"` // 1:e8:d8:d1:47:55:93
	AllowDualStackQueue string `json:"AllowDualStackQueue"`
	ActiveMacAddress    string `json:"ActiveMacAddress"` // E8:D8:D1:47:55:93
	ActiveServer        string `json:"ActiveServer"`     // dhcp_lan
	Address             string `json:"Address"`          // pool_admin
	AddressLists        string `json:"AddressLists"`     // inet
	ClientId            string `json:"ClientId"`         // 1:e8:d8:d1:47:55:93
	Comment             string `json:"Comment"`          // nb=Vlad/com=UTTiST/col=Admin/quotahourly=500000000/quotadaily=50000000000
	DhcpOption          string `json:"DhcpOption"`       //
	DisabledL           string `json:"DisabledL"`        // false
	Dynamic             string `json:"Dynamic"`          // false
	ExpiresAfter        string `json:"ExpiresAfter"`     // 6m32s
	HostName            string `json:"HostName"`         // root-hp
	LastSeen            string `json:"LastSeen"`         // 3m28s
	MacAddress          string `json:"MacAddress"`       // E8:D8:D1:47:55:93
	Radius              string `json:"Radius"`           // false
	Server              string `json:"Server"`           // dhcp_lan
	Status              string `json:"Status"`           // bound
	InsertQueueBefore   string `json:"InsertQueueBefore"`
	RateLimit           string `json:"RateLimit"`
	UseSrcMac           string `json:"UseSrcMac"`
	AgentCircuitId      string `json:"AgentCircuitId"`
	BlockAccess         string `json:"BlockAccess"`
	LeaseTime           string `json:"LeaseTime"`
	AgentRemoteId       string `json:"AgentRemoteId"`
	DhcpOptionSet       string `json:"DhcpOptionSet"`
	SrcMacAddress       string `json:"SrcMacAddress"`
	AlwaysBroadcast     string `json:"AlwaysBroadcast"`
	//User Defined
	// timeout         time.Time
	Hardware        bool   `json:"Hardware"`
	Manual          bool   `json:"Manual"`
	Blocked         bool   `json:"Blocked"`
	ShouldBeBlocked bool   `json:"ShouldBeBlocked"`
	TypeD           string `json:"TypeD"`
	ID              int    `json:"id"`
	IP              string `json:"ip"`
	Mac             string `json:"mac"`
	Disabled        bool   `json:"Disabled"`
	TimeoutBlock    string `json:"TimeoutBlock"`
	HourlyQuota     uint64 `json:"HourlyQuota"`
	DailyQuota      uint64 `json:"DailyQuota"`
	MonthlyQuota    uint64 `json:"MonthlyQuota"`
	Name            string `json:"Name"`
	Position        string `json:"Position"`
	Company         string `json:"Company"`
	IDUser          string `json:"IDUser"`
}

func (d *DeviceType) ToQuota() QuotaType {
	var q QuotaType
	commentArray := strings.Split(d.Comment, "/")
	for _, value := range commentArray {
		switch {
		case strings.Contains(value, "quotahourly="):
			q.HourlyQuota = validation.ParseParamertToUint(value)
		case strings.Contains(value, "quotadaily="):
			q.DailyQuota = validation.ParseParamertToUint(value)
		case strings.Contains(value, "quotamonthly="):
			q.MonthlyQuota = validation.ParseParamertToUint(value)
			// case strings.Contains(value, "manual="):
			// 	q.Manual = parseParamertToBool(value)
		}
		q.Blocked = q.Blocked || d.Blocked
		q.Disabled = q.Disabled || validation.ParameterToBool(d.DisabledL)
	}
	return q
}

func (d *DeviceType) ToPerson() PersonType {
	var p PersonType
	var comments string
	commentArray := strings.Split(d.Comment, "/")
	for _, value := range commentArray {
		switch {
		case strings.Contains(value, "tel="):
			// p.TypeD = "tel"
			p.Name = validation.ParseParamertToStr(value)
		case strings.Contains(value, "nb="):
			// p.TypeD = "nb"
			p.Name = validation.ParseParamertToStr(value)
		case strings.Contains(value, "ws="):
			// p.TypeD = "ws"
			p.Name = validation.ParseParamertToStr(value)
		case strings.Contains(value, "srv"):
			// p.TypeD = "srv"
			p.Name = validation.ParseParamertToStr(value)
		case strings.Contains(value, "prn="):
			// p.TypeD = "prn"
			p.Name = validation.ParseParamertToStr(value)
		case strings.Contains(value, "ap="):
			// p.TypeD = "ap"
			p.Name = validation.ParseParamertToStr(value)
		case strings.Contains(value, "name="):
			// p.TypeD = "other"
			p.Name = validation.ParseParamertToStr(value)
		case strings.Contains(value, "col="):
			p.Position = validation.ParseParamertToStr(value)
		case strings.Contains(value, "pos="):
			p.Position = validation.ParseParamertToStr(value)
		case strings.Contains(value, "com="):
			p.Company = validation.ParseParamertToStr(value)
		case strings.Contains(value, "id="):
			p.IDUser = validation.ParseParamertToStr(value)
		case strings.Contains(value, "comment="):
			comments = validation.ParseParamertToStr(value)
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
			comments = validation.ParseParamertToStr(value)
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
	logrus.Debugf("Device (%17v;%15v;%17v;%17v) was disabled due to exceeding the quota", key.Mac, d.ActiveAddress, d.ActiveMacAddress, d.MacAddress)
	return d
}

func (d DeviceType) UnBlock(group string, key KeyDevice) DeviceType {
	d.AddressLists = strings.Replace(d.AddressLists, group, "", 1)
	d.AddressLists = strings.ReplaceAll(d.AddressLists, ",,", ",")
	d.AddressLists = strings.Trim(d.AddressLists, ",")
	d.AddressLists = strings.ReplaceAll(d.AddressLists, `"`, "")
	logrus.Debugf("Device (%17v;%15v;%17v;%17v) has been enabled, the quota has not been exceeded", key.Mac, d.ActiveAddress, d.ActiveMacAddress, d.MacAddress)
	d.Blocked = false
	d.ShouldBeBlocked = false
	return d
}

func (d *DeviceType) IsNULL() bool {
	switch {
	case d.ActiveClientId != "":
		return false
	case d.ActiveServer != "":
		return false
	case d.Address != "":
		return false
	case d.AgentCircuitId != "":
		return false
	case d.AgentRemoteId != "":
		return false
	case d.AllowDualStackQueue != "":
		return false
	case d.AlwaysBroadcast != "":
		return false
	case d.BlockAccess != "":
		return false
	case d.ClientId != "":
		return false
	case d.DhcpOption != "":
		return false
	case d.DhcpOptionSet != "":
		return false
	case d.DisabledL != "":
		return false
	case d.Dynamic != "":
		return false
	case d.ExpiresAfter != "":
		return false
	case d.InsertQueueBefore != "":
		return false
	case d.LastSeen != "":
		return false
	case d.LeaseTime != "":
		return false
	case d.MacAddress != "":
		return false
	case d.Radius != "":
		return false
	case d.RateLimit != "":
		return false
	case d.Server != "":
		return false
	case d.SrcMacAddress != "":
		return false
	case d.Status != "":
		return false
	case d.UseSrcMac != "":
		return false
	case d.ActiveAddress != "":
		return false
	case d.ActiveMacAddress != "":
		return false
	case d.AddressLists != "":
		return false
	case d.Comment != "":
		return false
	case d.HostName != "":
		return false
	case d.Id != "":
		return false
	case d.TypeD != "":
		return false
	}
	return true
}
