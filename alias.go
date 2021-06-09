package main

import (
	"bufio"
	"fmt"
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
					for _, alias := range t.Aliases {
						switch {
						case alias.DeviceInAlias(key):
							device := t.devices[key]
							alias.UpdateQuota(device.ToQuota())
							alias.UpdatePerson(device.ToPerson())
							contains = true
							break
						case alias.IPOnlyInAlias(key):
						}
					}
					if !contains {
						device := t.devices[key]
						t.Aliases[deviceStat.mac] = AliasType{
							AliasName:  deviceStat.mac,
							KeyArr:     []KeyDevice{KeyDevice{ip: deviceStat.ip, mac: deviceStat.mac}},
							QuotaType:  checkNULLQuotas(device.ToQuota(), p.QuotaType),
							PersonType: device.ToPerson(),
						}
					}
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

func (a *AliasType) IPOnlyInAlias(key KeyDevice) bool {
	for _, item := range a.KeyArr {
		if item.ip == key.ip && item.ip == "" {
			return true
		}
	}
	return false
}

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
	if p.TypeD != "" {
		a.PersonType.TypeD = p.TypeD
	}
	if p.Comment != "" {
		a.PersonType.Comment = p.Comment
	}
	if p.IDUser != "" {
		a.PersonType.IDUser = p.IDUser
	}
}
