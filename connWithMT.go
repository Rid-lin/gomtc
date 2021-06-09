package main

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

func parseParamertToStr(inpuStr string) string {
	inpuStr = strings.Trim(inpuStr, "=")
	inpuStr = strings.ReplaceAll(inpuStr, "==", "=")
	arr := strings.Split(inpuStr, "=")
	if len(arr) > 1 {
		return arr[1]
		// } else {
		// 	log.Warnf("Parameter error. The parameter is specified incorrectly or not specified at all.(%v)", inpuStr)
	}
	return ""
}

func parseParamertToComment(inpuStr string) string {
	inpuStr = strings.Trim(inpuStr, "=")
	inpuStr = strings.ReplaceAll(inpuStr, "==", "=")
	arr := strings.Split(inpuStr, "=")
	if len(arr) > 1 {
		return strings.Join(arr[1:], "=")
		// } else {
		// 	log.Warnf("Parameter error. The parameter is specified incorrectly or not specified at all.(%v)", inpuStr)
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
		return q
		// } else {
		// 	log.Warnf("Parameter error. The parameter is specified incorrectly or not specified at all.(%v)", inputValue)
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

func makeCommentFromIodt(a InfoOldType, q QuotaType) string {
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

func (a *InfoType) convertToComment(q QuotaType) string {
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
	if a.DeviceType.Manual {
		comment = fmt.Sprintf("%v/manual=%v",
			comment,
			a.DeviceType.Manual)
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
		case value != "":
			comments = comments + "/" + value
			// default:
			// 	comments = comments + "/" + value
		}
	}
	Comment = Comment + comments
	return
}

func (d DeviceType) convertToInfo(blockGroup string) InfoOldType {
	var (
		ip, mac, name, position, company, typeD, IDUser, comment string
		hourlyQuota, dailyQuota, monthlyQuota                    uint64
		manual                                                   bool
	)
	ip = validateIP(d.activeAddress, d.address)
	mac = getSwithMac(d.activeMacAddress, d.macAddress, d.clientId, d.activeClientId)
	hourlyQuota, dailyQuota, monthlyQuota, name, position, company, typeD, IDUser, comment, manual = parseComment(d.comment)
	blocked := isBlocked(d.addressLists, blockGroup)
	// blocked := strings.Contains(d.addressLists, blockGroup)

	infoD := InfoOldType{
		DeviceOldType: DeviceOldType{
			Id:       d.Id,
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
			Blocked:      blocked,
			Dynamic:      paramertToBool(d.dynamic),
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

func (a *InfoType) convertToDevice(quotaDef QuotaType) DeviceType {
	return DeviceType{
		activeAddress:    a.activeAddress,
		address:          a.activeAddress,
		activeMacAddress: a.activeMacAddress,
		addressLists:     a.addressLists,
		blocked:          fmt.Sprint(a.Blocked),
		comment:          a.convertToComment(quotaDef),
		disabledL:        fmt.Sprint(a.Disabled),
		hostName:         a.hostName,
		macAddress:       a.macAddress,
		Manual:           a.DeviceType.Manual,
		ShouldBeBlocked:  a.DeviceType.ShouldBeBlocked,
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

// func (ds *DevicesType) findIndexOfDevice(d *DeviceType) int {
// 	for index, device := range *ds {
// 		if d.compare(&device) {
// 			return index
// 		}
// 	}
// 	return -1
// }

// func (t *Transport) updateStatusDevicesToMT(cfg *Config) {

// 	t.RLock()
// 	blockGroup := t.BlockAddressList
// 	dataCashe := t.dataCasheOld
// 	quota := t.QuotaType
// 	p := parseType{
// 		SSHCredentials:   t.sshCredentials,
// 		QuotaType:        t.QuotaType,
// 		BlockAddressList: t.BlockAddressList,
// 		Location:         t.Location,
// 	}
// 	t.RUnlock()
// 	tn := time.Now().Format("2006-01-02")

// 	for key, alias := range dataCashe {
// 		if alias.Manual {
// 			continue
// 		}
// 		if key.DateStr != tn || alias.Dynamic {
// 			continue
// 		}
// 		// key := KeyMapOfReports{
// 		// 	Alias:   alias,
// 		// 	DateStr: time.Now().In(t.Location).Format("2006-01-02"),
// 		// }
// 		if alias.ShouldBeBlocked && !alias.Blocked {
// 			alias.removeDefaultQuotas(quota)
// 			alias.addBlockGroup(blockGroup)
// 			if err := alias.sendByAll(p, quota); err != nil {
// 				log.Errorf(`An error occurred while saving the device(%v::%v::%v):%v`,
// 					alias.Alias, alias.Mac, alias.AMac, err.Error())
// 			}
// 		} else if !alias.ShouldBeBlocked && alias.Blocked {
// 			alias.removeDefaultQuotas(quota)
// 			alias.delBlockGroup(blockGroup)
// 			if err := alias.sendByAll(p, quota); err != nil {
// 				log.Errorf(`An error occurred while saving the device(%v::%v::%v):%v`,
// 					alias.Alias, alias.Mac, alias.AMac, err.Error())
// 			}
// 		}
// 	}
// 	t.getDevices()
// }

// func (ds *DevicesType) updateInfo(deviceNew DeviceType) error {
// 	index := ds.findIndexOfDevice(&deviceNew)
// 	if index == -1 {
// 		*ds = append(*ds, deviceNew)
// 	} else {
// 		(*ds)[index] = deviceNew
// 	}
// 	return nil
// }

// func (dms *DevicesMapType) updateInfo(deviceNew DeviceType) error {

// 	index := dms.findIndexOfDevice(&deviceNew)
// 	if index == -1 {
// 		*dms = append(*dms, deviceNew)
// 	} else {
// 		(*dms)[index] = deviceNew
// 	}
// 	return nil
// }

func boolToParamert(trigger bool) string {
	if trigger {
		return "yes"
	}
	return "no"
}

func removeDefaultQuota(setValue, deafultValue uint64) uint64 {
	if setValue == deafultValue {
		return uint64(0)
	}
	return uint64(setValue)
}

func (a *InfoOldType) removeDefaultQuotas(qDef QuotaType) {
	a.HourlyQuota = removeDefaultQuota(a.HourlyQuota, qDef.MonthlyQuota)
	a.DailyQuota = removeDefaultQuota(a.DailyQuota, qDef.DailyQuota)
	a.MonthlyQuota = removeDefaultQuota(a.MonthlyQuota, qDef.MonthlyQuota)
}

func (a *InfoOldType) addBlockGroup(group string) {
	a.Groups = a.Groups + "," + group
	a.Groups = strings.Trim(a.Groups, ",")
	a.Groups = strings.ReplaceAll(a.Groups, `"`, "")
}

func (a *InfoOldType) delBlockGroup(group string) {
	a.Groups = strings.Replace(a.Groups, group, "", 1)
	a.Groups = strings.ReplaceAll(a.Groups, ",,", ",")
	a.Groups = strings.Trim(a.Groups, ",")
	a.Groups = strings.ReplaceAll(a.Groups, `"`, "")
}

// func (t *Transport) getAliasS(alias string) InfoType {
// 	key := KeyMapOfReports{
// 		Alias:   alias,
// 		DateStr: time.Now().In(t.Location).Format("2006-01-02"),
// 	}
// 	InfoD, ok := t.dataCasheOld[key]
// 	if !ok {
// 		return InfoType{}
// 	}
// 	return InfoD
// }
