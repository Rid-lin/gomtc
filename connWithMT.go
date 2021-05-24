package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// func (t *Transport) loopGetDataFromMT() {
// 	t.updateDevices()
// 	for {
// 		<-t.timerMT.C
// 		t.updateDevices()
// 	}
// }

// func (t *Transport) getParseCred() parseType {
// 	p := parseType{}
// 	t.RLock()
// 	p.SSHCredentials = t.sshCredentials
// 	p.BlockAddressList = t.BlockAddressList
// 	p.QuotaType = t.QuotaType
// 	p.Location = t.Location
// 	p.DevicesRetryDelay = t.DevicesRetryDelay
// 	t.RUnlock()
// 	return p
// }

func (t *Transport) setTimerMT(IntervalStr string) {
	interval, err := time.ParseDuration(IntervalStr)
	if err != nil {
		t.timerMT = time.NewTimer(15 * time.Minute)
	} else {
		t.timerMT = time.NewTimer(interval)
	}

}

func (t *Transport) updateDevices() {
	t.Lock()
	t.devices = parseInfoFromMTToSlice2(parseType{
		SSHCredentials:   t.sshCredentials,
		QuotaType:        t.QuotaType,
		BlockAddressList: t.BlockAddressList,
		Location:         t.Location,
	})
	t.lastUpdatedMT = time.Now()
	t.setTimerMT(t.DevicesRetryDelay)
	t.Unlock()
}

// func (data *Transport) updateQuotas(p parseType) {
// 	data.Lock()
// 	for key, value := range data.dataCashe {
// 		infoD, err := data.devices.findInfoDByAlias(value.Alias, p.QuotaType)
// 		if err != nil {
// 			continue
// 		}
// 		value.InfoOfDeviceType = infoD
// 		data.dataCashe[key] = value
// 		// value.InfoOfDeviceType = data.devices.getInfoD(value.Alias, p.QuotaType)
// 	}
// 	data.Unlock()
// }

func (t *Transport) updateQuotas(p parseType) {
	t.Lock()
	for key, value := range t.dataCashe {
		if value.Alias == "" {
			value.Alias = key.Alias
		}
		if value.DateStr == "" {
			value.DateStr = key.DateStr
		}
		// if value.Alias == "E8:D8:D1:47:55:93" {
		// 	runtime.Breakpoint()
		// }
		value.AliasType = t.devices.findDeviceToConvertInfoD(value.Alias, p.QuotaType)
		t.dataCashe[key] = value
	}
	t.Unlock()
}

// func (data *Transport) loopGetDataFromMTOverAPI() {
// 	for {
// 		data.updateInfoOfDevicesFromMT()
// 		if err := data.getStatusallDevices(); err == nil {
// 			// if err := transport.getStatusDevices(cfg); err == nil {
// 			data.checkQuota()
// 			// transport.updateStatusDevicesToMT(cfg)
// 		}
// 		interval, err := time.ParseDuration(cfg.Interval)
// 		if err != nil {
// 			interval = 1 * time.Minute
// 		}
// 		time.Sleep(interval)

// 	}
// }

func parseParamertToStr(inpuStr string) string {
	inpuStr = strings.Trim(inpuStr, "=")
	inpuStr = strings.ReplaceAll(inpuStr, "==", "=")
	arr := strings.Split(inpuStr, "=")
	if len(arr) > 1 {
		return arr[1]
	} else {
		log.Warnf("Parameter error. The parameter is specified incorrectly or not specified at all.(%v)", inpuStr)
	}
	return ""
}

func parseParamertToUint(inputValue string) uint64 {
	// var err error
	var q uint64
	inputValue = strings.Trim(inputValue, "=")
	inputValue = strings.ReplaceAll(inputValue, "==", "=")
	Arr := strings.Split(inputValue, "=")
	if len(Arr) > 1 {
		quotaStr := Arr[1]
		q = paramertToUint(quotaStr)
		// quotaStr = strings.Trim(quotaStr, "\r")
		// quota, err := strconv.ParseUint(quotaStr, 10, 64)
		// if err != nil {
		// 	quotaF, err2 := strconv.ParseFloat(quotaStr, 64)
		// 	if err != nil {
		// 		log.Errorf("Error parse quota from string(%v):(%v)(%v)", quotaStr, err, err2)
		// 		return 0
		// 	}
		// 	quota = uint64(quotaF)
		// }
		return q
	} else {
		log.Warnf("Parameter error. The parameter is specified incorrectly or not specified at all.(%v)", inputValue)
	}
	return q
}

func parseParamertToBool(inputValue string) bool {
	inputValue = strings.Trim(inputValue, "=")
	inputValue = strings.ReplaceAll(inputValue, "==", "=")
	arr := strings.Split(inputValue, "=")
	if len(arr) > 1 {
		value := arr[1]
		if value == "true" || value == "yes" {
			return true
		}
	}
	return false
}

func paramertToUint(inputValue string) uint64 {
	inputValue = strings.Trim(inputValue, "\r")
	quota, err := strconv.ParseUint(inputValue, 10, 64)
	if err != nil {
		quotaF, err2 := strconv.ParseFloat(inputValue, 64)
		if err != nil {
			log.Errorf("Error parse quota from input string(%v):(%v)(%v)", inputValue, err, err2)
			return 0
		}
		quota = uint64(quotaF)
	}
	return quota
}

func paramertToBool(inputValue string) bool {
	if inputValue == "true" || inputValue == "yes" {
		return true
	}
	return false
}

func makeCommentFromIodt(a AliasType, q QuotaType) string {
	comment := "/"

	if a.TypeD != "" && a.Name != "" {
		comment = fmt.Sprintf("%v=%v",
			a.TypeD,
			a.Name)
	} else if a.TypeD == "" && a.Name != "" {
		comment = fmt.Sprintf("name=%v",
			a.Name)
	}
	if a.IDUser != "" {
		comment = fmt.Sprintf("%v/id=%v",
			comment,
			a.IDUser)
	}
	if a.Company != "" {
		comment = fmt.Sprintf("%v/com=%v",
			comment,
			a.Company)
	}
	if a.Position != "" {
		comment = fmt.Sprintf("%v/pos=%v",
			comment,
			a.Position)
	}
	if a.HourlyQuota != 0 && a.HourlyQuota != q.HourlyQuota {
		comment = fmt.Sprintf("%v/quotahourly=%v",
			comment,
			a.HourlyQuota)
	}
	if a.DailyQuota != 0 && a.DailyQuota != q.DailyQuota {
		comment = fmt.Sprintf("%v/quotadaily=%v",
			comment,
			a.DailyQuota)
	}
	if a.MonthlyQuota != 0 && a.MonthlyQuota != q.MonthlyQuota {
		comment = fmt.Sprintf("%v/quotamonthly=%v",
			comment,
			a.MonthlyQuota)
	}
	if a.Manual {
		comment = fmt.Sprintf("%v/manual=%v",
			comment,
			a.Manual)
	}
	if a.Comment != "" {
		comment = fmt.Sprintf("%v/comment=%v",
			comment,
			a.Comment)
	}
	comment = strings.ReplaceAll(comment, "//", "/")
	if comment == "/" {
		comment = ""
	}

	return comment
}

func (a *AliasType) convetrToString(q QuotaType) string {
	comment := "/"

	if a.TypeD != "" && a.Name != "" {
		comment = fmt.Sprintf("%v=%v",
			a.TypeD,
			a.Name)
	} else if a.TypeD == "" && a.Name != "" {
		comment = fmt.Sprintf("name=%v",
			a.Name)
	}
	if a.IDUser != "" {
		comment = fmt.Sprintf("%v/id=%v",
			comment,
			a.IDUser)
	}
	if a.Company != "" {
		comment = fmt.Sprintf("%v/com=%v",
			comment,
			a.Company)
	}
	if a.Position != "" {
		comment = fmt.Sprintf("%v/pos=%v",
			comment,
			a.Position)
	}
	if a.HourlyQuota != 0 && a.HourlyQuota != q.HourlyQuota {
		comment = fmt.Sprintf("%v/quotahourly=%v",
			comment,
			a.HourlyQuota)
	}
	if a.DailyQuota != 0 && a.DailyQuota != q.DailyQuota {
		comment = fmt.Sprintf("%v/quotadaily=%v",
			comment,
			a.DailyQuota)
	}
	if a.MonthlyQuota != 0 && a.MonthlyQuota != q.MonthlyQuota {
		comment = fmt.Sprintf("%v/quotamonthly=%v",
			comment,
			a.MonthlyQuota)
	}
	if a.Manual {
		comment = fmt.Sprintf("%v/manual=%v",
			comment,
			a.Manual)
	}
	if a.Comment != "" {
		comment = fmt.Sprintf("%v/comment=%v",
			comment,
			a.Comment)
	}
	comment = strings.ReplaceAll(comment, "//", "/")
	if comment == "/" {
		comment = ""
	}

	return comment
}

func parseComment(comment string) (
	quotahourly, quotadaily, quotamonthly uint64,
	name, position, company, typeD, IDUser, Comment string,
	manual bool) {
	commentArray := strings.Split(comment, "/")
	var comments string
	for _, value := range commentArray {
		switch {
		case strings.Contains(value, "tel="):
			typeD = "tel"
			name = parseParamertToStr(value)
		case strings.Contains(value, "nb="):
			typeD = "nb"
			name = parseParamertToStr(value)
		case strings.Contains(value, "ws="):
			typeD = "ws"
			name = parseParamertToStr(value)
		case strings.Contains(value, "srv"):
			typeD = "srv"
			name = parseParamertToStr(value)
		case strings.Contains(value, "prn="):
			typeD = "prn"
			name = parseParamertToStr(value)
		case strings.Contains(value, "ap="):
			typeD = "ap"
			name = parseParamertToStr(value)
		case strings.Contains(value, "name="):
			typeD = "other"
			name = parseParamertToStr(value)
		case strings.Contains(value, "col="):
			position = parseParamertToStr(value)
		case strings.Contains(value, "pos="):
			position = parseParamertToStr(value)
		case strings.Contains(value, "com="):
			company = parseParamertToStr(value)
		case strings.Contains(value, "id="):
			IDUser = parseParamertToStr(value)
		case strings.Contains(value, "quotahourly="):
			quotahourly = parseParamertToUint(value)
		case strings.Contains(value, "quotadaily="):
			quotadaily = parseParamertToUint(value)
		case strings.Contains(value, "quotamonthly="):
			quotamonthly = parseParamertToUint(value)
		case strings.Contains(value, "manual="):
			manual = parseParamertToBool(value)
		case strings.Contains(value, "comment="):
			Comment = parseParamertToStr(value)
		default:
			comments = comments + "/" + value
		}
	}
	Comment = Comment + comments
	return
}

func (d DeviceType) convertToInfo() AliasType {
	var (
		ip, mac, name, position, company, typeD, IDUser, comment string
		hourlyQuota, dailyQuota, monthlyQuota                    uint64
		manual                                                   bool
	)
	ip = validateIP(d.activeAddress, d.address)
	mac = getSwithMac(d.activeMacAddress, d.macAddress, d.clientId, d.activeClientId)
	hourlyQuota, dailyQuota, monthlyQuota, name, position, company, typeD, IDUser, comment, manual = parseComment(d.comment)
	infoD := AliasType{
		DeviceOldType: DeviceOldType{
			IP:       ip,
			Mac:      mac,
			AMac:     mac,
			HostName: d.hostName,
			Groups:   d.addressLists,
			timeout:  d.timeout,
		},
		QuotaType: QuotaType{
			HourlyQuota:  hourlyQuota,
			DailyQuota:   dailyQuota,
			MonthlyQuota: monthlyQuota,
			Manual:       manual,
		},
		PersonType: PersonType{
			IDUser:   IDUser,
			Name:     name,
			Position: position,
			Company:  company,
			TypeD:    typeD,
			Comment:  comment,
			Comments: d.comment,
		},
	}
	return infoD
}

func (a *AliasType) convertToDevice(quotaDef QuotaType) DeviceType {
	return DeviceType{
		activeAddress:    a.IP,
		address:          a.IP,
		activeMacAddress: a.AMac,
		addressLists:     a.Groups,
		blocked:          fmt.Sprint(a.Blocked),
		comment:          a.convetrToString(quotaDef),
		disabled:         fmt.Sprint(a.Disabled),
		hostName:         a.HostName,
		macAddress:       a.Mac,
		Manual:           a.Manual,
		ShouldBeBlocked:  a.ShouldBeBlocked,
		timeout:          a.timeout,
	}
}

func (d1 *DeviceType) compare(d2 *DeviceType) bool {
	switch {
	case d1.macAddress == d2.macAddress && d1.macAddress != "" && d2.macAddress != "":
		return true
	case d1.activeMacAddress == d2.macAddress && d1.activeMacAddress != "" && d2.macAddress != "":
		return true
	case d1.macAddress == d2.activeMacAddress && d1.macAddress != "" && d2.activeMacAddress != "":
		return true
	case d1.activeMacAddress == d2.activeMacAddress && d1.activeMacAddress != "" && d2.activeMacAddress != "":
		return true
	case d1.activeClientId == d2.activeClientId && d1.activeClientId != "" && d2.activeClientId != "":
		return true
	case d1.activeClientId == d2.clientId && d1.activeClientId != "" && d2.clientId != "":
		return true
	case d1.clientId == d2.activeClientId && d1.clientId != "" && d2.activeClientId != "":
		return true
	case d1.clientId == d2.clientId && d1.clientId != "" && d2.clientId != "":
		return true
	}
	return false
}

func (ds *DevicesType) findIndexOfDevice(d *DeviceType) int {
	for index, device := range *ds {
		if d.compare(&device) {
			return index
		}
	}
	return -1
}

func (t *Transport) updateStatusDevicesToMT(cfg *Config) {

	t.RLock()
	blockGroup := t.BlockAddressList
	data := t.dataCashe
	quota := t.QuotaType
	p := parseType{
		SSHCredentials:   t.sshCredentials,
		QuotaType:        t.QuotaType,
		BlockAddressList: t.BlockAddressList,
		Location:         t.Location,
	}
	t.RUnlock()

	for _, dInfo := range data {
		if dInfo.Manual {
			continue
		}
		if dInfo.ShouldBeBlocked && !dInfo.Blocked {
			if dInfo.QuotaType == quota {
				dInfo.QuotaType = QuotaType{}
			}
			dInfo.Groups = dInfo.Groups + "," + blockGroup
			dInfo.Groups = strings.Trim(dInfo.Groups, ",")
			if err := dInfo.send(p, quota); err != nil {
				log.Errorf(`An error occurred while saving the device:%v`, err.Error())
			}
		} else if !dInfo.ShouldBeBlocked && dInfo.Blocked {
			if dInfo.QuotaType == quota {
				dInfo.QuotaType = QuotaType{}
			}
			dInfo.Groups = strings.Replace(dInfo.Groups, blockGroup, "", 1)
			dInfo.Groups = strings.ReplaceAll(dInfo.Groups, ",,", ",")
			dInfo.Groups = strings.Trim(dInfo.Groups, ",")
			if err := dInfo.send(p, quota); err != nil {
				log.Errorf(`An error occurred while saving the device:%v`, err.Error())
			}
		}
	}
	t.updateDevices()
}

func (ds *DevicesType) updateInfo(deviceNew DeviceType) error {
	index := ds.findIndexOfDevice(&deviceNew)
	if index == -1 {
		*ds = append(*ds, deviceNew)
	}
	(*ds)[index] = deviceNew
	return nil
}
