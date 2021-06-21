package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

func (t *Transport) getAliasesArr(cfg *Config) {
	aliases := make(map[string][]string)
	path := path.Join(cfg.ConfigPath, "realname.cfg")
	buf, err := os.Open(path)
	if err != nil {
		log.Error(err)
		return
	}
	defer func() {
		if err = buf.Close(); err != nil {
			log.Error(err)
		}
	}()
	snl := bufio.NewScanner(buf)
	for snl.Scan() {
		line := snl.Text()
		lineArray := strings.Split(line, " ")
		if len(lineArray) <= 1 {
			continue
		}
		key := lineArray[0]
		value := lineArray[1:]
		if test, ok := aliases[key]; ok {
			aliases[key] = append(test, fmt.Sprint(value))
		} else {
			aliases[key] = []string{fmt.Sprint(value)}
		}
	}
	err = snl.Err()
	if err != nil {
		log.Error(err)
		return
	}
	t.Lock()
	t.AliasesStrArr = aliases
	t.Unlock()
	log.Trace(aliases)
}

func (t *Transport) updateAliases(p parseType) {
	t.Lock()
	var contains bool
	for _, year := range t.statofYears {
		for _, month := range year.monthsStat {
			for _, day := range month.daysStat {
				for key, deviceStat := range day.devicesStat {
					device := t.devices[key]
					for _, alias := range t.Aliases {
						switch {
						case alias.DeviceInAlias(key):
							alias.UpdateQuota(device.ToQuota())
							alias.UpdatePerson(device.ToPerson())
							goto gotAlias
						// case alias.IPOnlyInAlias(key) && day.day == time.Now().Day():
						// 	alias.UpdateQuota(device.ToQuota())
						// 	alias.UpdatePerson(device.ToPerson())
						// 	goto gotAlias
						case alias.MacInAlias(key):
							alias.UpdateQuota(device.ToQuota())
							alias.UpdatePerson(device.ToPerson())
							goto gotAlias
						}
					}
					if !contains {
						device := t.devices[key]
						t.Aliases[deviceStat.mac] = AliasType{
							AliasName: deviceStat.mac,
							KeyArr: []KeyDevice{{
								// ip: deviceStat.ip,
								mac: deviceStat.mac}},
							QuotaType:  checkNULLQuotas(device.ToQuota(), p.QuotaType),
							PersonType: device.ToPerson(),
						}
					}
				gotAlias:
				}
			}
		}
	}
	t.Unlock()
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
		if item.mac == key.mac {
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
		a.Disabled = paramertToBool(params["disabled"][0])
	} else {
		a.Disabled = false
	}
	if len(params["quotahourly"]) > 0 {
		a.HourlyQuota = paramertToUint(params["quotahourly"][0])
	} else {
		a.HourlyQuota = 0
	}
	if len(params["quotadaily"]) > 0 {
		a.DailyQuota = paramertToUint(params["quotadaily"][0])
	} else {
		a.DailyQuota = 0
	}
	if len(params["quotamonthly"]) > 0 {
		a.MonthlyQuota = paramertToUint(params["quotamonthly"][0])
	} else {
		a.MonthlyQuota = 0
	}
}

func (t *Transport) BlockAlias(a AliasType, group string) {
	t.Lock()
	for _, key := range a.KeyArr {
		device := t.devices[key]
		if device.IsNULL() {
			delete(t.devices, key)
			continue
		}
		if device.Manual || device.Blocked {
			continue
		}
		device = device.Block(group, key)
		t.change[key] = DeviceToBlock{
			Id:       device.Id,
			Groups:   device.AddressLists,
			Disabled: paramertToBool(device.disabledL),
		}
	}
	t.Unlock()
}

func (t *Transport) UnBlockAlias(a AliasType, group string) {
	t.Lock()
	for _, key := range a.KeyArr {
		device := t.devices[key]
		if device.Manual || !device.Blocked {
			continue
		}
		device = device.UnBlock(group, key)
	}
	t.Unlock()
}

func (d *DeviceType) IsNULL() bool {
	switch {
	case d.activeClientId != "":
		return false
	case d.activeServer != "":
		return false
	case d.address != "":
		return false
	case d.agentCircuitId != "":
		return false
	case d.agentRemoteId != "":
		return false
	case d.allowDualStackQueue != "":
		return false
	case d.alwaysBroadcast != "":
		return false
	case d.blockAccess != "":
		return false
	case d.clientId != "":
		return false
	case d.dhcpOption != "":
		return false
	case d.dhcpOptionSet != "":
		return false
	case d.disabledL != "":
		return false
	case d.dynamic != "":
		return false
	case d.expiresAfter != "":
		return false
	case d.insertQueueBefore != "":
		return false
	case d.lastSeen != "":
		return false
	case d.leaseTime != "":
		return false
	case d.macAddress != "":
		return false
	case d.radius != "":
		return false
	case d.rateLimit != "":
		return false
	case d.server != "":
		return false
	case d.srcMacAddress != "":
		return false
	case d.status != "":
		return false
	case d.useSrcMac != "":
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
