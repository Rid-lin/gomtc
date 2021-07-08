package model

import (
	"net/url"

	"git.vegner.org/vsvegner/gomtc/internal/app/validation"
)

type AliasType struct {
	AliasName string
	KeyArr    []KeyDevice
	QuotaType
	PersonType
}

func (a *AliasType) DeviceInAlias(key KeyDevice) bool {
	for _, item := range a.KeyArr {
		if item == key {
			return true
		}
	}
	return false
}

// func (a *AliasType) IPOnlyInAlias(key KeyDevice) bool {
// 	for _, item := range a.KeyArr {
// 		// Если ip алиаса совпадают с маком устройства или мак пустой, то указан только ip
// 		if item.ip == key.ip && item.mac == "" || item.ip == key.ip && item.mac == key.ip {
// 			return true
// 		}
// 	}
// 	return false
// }

func (a *AliasType) MacInAlias(key KeyDevice) bool {
	for _, item := range a.KeyArr {
		if item.Mac == key.Mac {
			return true
		}
	}
	return false
}

func (a *AliasType) UpdateQuota(q QuotaType) {
	if a.DailyQuota < q.DailyQuota {
		a.DailyQuota = q.DailyQuota
	}
	if a.HourlyQuota < q.HourlyQuota {
		a.HourlyQuota = q.HourlyQuota
	}
	if a.MonthlyQuota < q.MonthlyQuota {
		a.MonthlyQuota = q.MonthlyQuota
	}
}

func (a *AliasType) UpdatePerson(p PersonType) {
	if p.Name != "" {
		a.PersonType.Name = p.Name
	}
	if p.Position != "" {
		a.PersonType.Position = p.Position
	}
	if p.Company != "" {
		a.PersonType.Company = p.Company
	}
	if p.Comment != "" {
		a.PersonType.Comment = p.Comment
	}
	if p.IDUser != "" {
		a.PersonType.IDUser = p.IDUser
	}
}

func (a *AliasType) UpdateFromForm(params url.Values) {
	if len(params["name"]) > 0 {
		a.Name = params["name"][0]
	} else {
		a.Name = ""
	}
	if len(params["col"]) > 0 {
		a.Position = params["col"][0]
	} else {
		a.Position = ""
	}
	if len(params["com"]) > 0 {
		a.Company = params["com"][0]
	} else {
		a.Company = ""
	}
	if len(params["comment"]) > 0 {
		a.Comment = params["comment"][0]
	} else {
		a.Comment = ""
	}
	if len(params["disabled"]) > 0 {
		a.Disabled = validation.ParameterToBool(params["disabled"][0])
	} else {
		a.Disabled = false
	}
	if len(params["quotahourly"]) > 0 {
		a.HourlyQuota = validation.ParamertToUint(params["quotahourly"][0])
	} else {
		a.HourlyQuota = 0
	}
	if len(params["quotadaily"]) > 0 {
		a.DailyQuota = validation.ParamertToUint(params["quotadaily"][0])
	} else {
		a.DailyQuota = 0
	}
	if len(params["quotamonthly"]) > 0 {
		a.MonthlyQuota = validation.ParamertToUint(params["quotamonthly"][0])
	} else {
		a.MonthlyQuota = 0
	}
}
