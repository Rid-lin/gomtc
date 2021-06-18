package main

import (
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func parseParamertToStr(inpuStr string) string {
	inpuStr = strings.Trim(inpuStr, "=")
	inpuStr = strings.ReplaceAll(inpuStr, "==", "=")
	arr := strings.Split(inpuStr, "=")
	if len(arr) > 1 {
		return arr[1]
	}
	return ""
}

func parseParamertToComment(inpuStr string) string {
	inpuStr = strings.Trim(inpuStr, "=")
	inpuStr = strings.ReplaceAll(inpuStr, "==", "=")
	arr := strings.Split(inpuStr, "=")
	if len(arr) > 1 {
		return strings.Join(arr[1:], "=")
	}
	return ""
}

func parseParamertToUint(inputValue string) uint64 {
	var q uint64
	inputValue = strings.Trim(inputValue, "=")
	inputValue = strings.ReplaceAll(inputValue, "==", "=")
	Arr := strings.Split(inputValue, "=")
	if len(Arr) > 1 {
		quotaStr := Arr[1]
		q = paramertToUint(quotaStr)
		return q
	}
	return q
}

// func parseParamertToBool(inputValue string) bool {
// 	inputValue = strings.Trim(inputValue, "=")
// 	inputValue = strings.ReplaceAll(inputValue, "==", "=")
// 	arr := strings.Split(inputValue, "=")
// 	if len(arr) > 1 {
// 		value := arr[1]
// 		if value == "true" || value == "yes" {
// 			return true
// 		}
// 	}
// 	return false
// }

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

// func makeCommentFromIodt(a InfoOldType, q QuotaType) string {
// 	comment := "/"

// 	if a.TypeD != "" && a.Name != "" {
// 		comment = fmt.Sprintf("%v=%v",
// 			a.TypeD,
// 			a.Name)
// 	} else if a.TypeD == "" && a.Name != "" {
// 		comment = fmt.Sprintf("name=%v",
// 			a.Name)
// 	}
// 	if a.IDUser != "" {
// 		comment = fmt.Sprintf("%v/id=%v",
// 			comment,
// 			a.IDUser)
// 	}
// 	if a.Company != "" {
// 		comment = fmt.Sprintf("%v/com=%v",
// 			comment,
// 			a.Company)
// 	}
// 	if a.Position != "" {
// 		comment = fmt.Sprintf("%v/pos=%v",
// 			comment,
// 			a.Position)
// 	}
// 	if a.HourlyQuota != 0 && a.HourlyQuota != q.HourlyQuota {
// 		comment = fmt.Sprintf("%v/quotahourly=%v",
// 			comment,
// 			a.HourlyQuota)
// 	}
// 	if a.DailyQuota != 0 && a.DailyQuota != q.DailyQuota {
// 		comment = fmt.Sprintf("%v/quotadaily=%v",
// 			comment,
// 			a.DailyQuota)
// 	}
// 	if a.MonthlyQuota != 0 && a.MonthlyQuota != q.MonthlyQuota {
// 		comment = fmt.Sprintf("%v/quotamonthly=%v",
// 			comment,
// 			a.MonthlyQuota)
// 	}
// 	// if a.Manual {
// 	// 	comment = fmt.Sprintf("%v/manual=%v",
// 	// 		comment,
// 	// 		a.Manual)
// 	// }
// 	if a.Comment != "" {
// 		comment = fmt.Sprintf("%v/comment=%v",
// 			comment,
// 			a.Comment)
// 	}
// 	comment = strings.ReplaceAll(comment, "//", "/")
// 	if comment == "/" {
// 		comment = ""
// 	}

// 	return comment
// }

// func (a *InfoType) convertToComment(q QuotaType) string {
// 	comment := "/"

// 	if a.TypeD != "" && a.Name != "" {
// 		comment = fmt.Sprintf("%v=%v",
// 			a.TypeD,
// 			a.Name)
// 	} else if a.TypeD == "" && a.Name != "" {
// 		comment = fmt.Sprintf("name=%v",
// 			a.Name)
// 	}
// 	if a.IDUser != "" {
// 		comment = fmt.Sprintf("%v/id=%v",
// 			comment,
// 			a.IDUser)
// 	}
// 	if a.Company != "" {
// 		comment = fmt.Sprintf("%v/com=%v",
// 			comment,
// 			a.Company)
// 	}
// 	if a.Position != "" {
// 		comment = fmt.Sprintf("%v/pos=%v",
// 			comment,
// 			a.Position)
// 	}
// 	if a.HourlyQuota != 0 && a.HourlyQuota != q.HourlyQuota {
// 		comment = fmt.Sprintf("%v/quotahourly=%v",
// 			comment,
// 			a.HourlyQuota)
// 	}
// 	if a.DailyQuota != 0 && a.DailyQuota != q.DailyQuota {
// 		comment = fmt.Sprintf("%v/quotadaily=%v",
// 			comment,
// 			a.DailyQuota)
// 	}
// 	if a.MonthlyQuota != 0 && a.MonthlyQuota != q.MonthlyQuota {
// 		comment = fmt.Sprintf("%v/quotamonthly=%v",
// 			comment,
// 			a.MonthlyQuota)
// 	}
// 	// if a.DeviceType.Manual {
// 	// 	comment = fmt.Sprintf("%v/manual=%v",
// 	// 		comment,
// 	// 		a.DeviceType.Manual)
// 	// }
// 	if a.Comment != "" {
// 		comment = fmt.Sprintf("%v/comment=%v",
// 			comment,
// 			a.Comment)
// 	}
// 	comment = strings.ReplaceAll(comment, "//", "/")
// 	if comment == "/" {
// 		comment = ""
// 	}

// 	return comment
// }

// func parseComment(comment string) (
// 	quotahourly, quotadaily, quotamonthly uint64,
// 	name, position, company, typeD, IDUser, Comment string,
// 	manual bool) {
// 	commentArray := strings.Split(comment, "/")
// 	var comments string
// 	for _, value := range commentArray {
// 		switch {
// 		case strings.Contains(value, "tel="):
// 			typeD = "tel"
// 			name = parseParamertToStr(value)
// 		case strings.Contains(value, "nb="):
// 			typeD = "nb"
// 			name = parseParamertToStr(value)
// 		case strings.Contains(value, "ws="):
// 			typeD = "ws"
// 			name = parseParamertToStr(value)
// 		case strings.Contains(value, "srv"):
// 			typeD = "srv"
// 			name = parseParamertToStr(value)
// 		case strings.Contains(value, "prn="):
// 			typeD = "prn"
// 			name = parseParamertToStr(value)
// 		case strings.Contains(value, "ap="):
// 			typeD = "ap"
// 			name = parseParamertToStr(value)
// 		case strings.Contains(value, "name="):
// 			typeD = "other"
// 			name = parseParamertToStr(value)
// 		case strings.Contains(value, "col="):
// 			position = parseParamertToStr(value)
// 		case strings.Contains(value, "pos="):
// 			position = parseParamertToStr(value)
// 		case strings.Contains(value, "com="):
// 			company = parseParamertToStr(value)
// 		case strings.Contains(value, "id="):
// 			IDUser = parseParamertToStr(value)
// 		case strings.Contains(value, "quotahourly="):
// 			quotahourly = parseParamertToUint(value)
// 		case strings.Contains(value, "quotadaily="):
// 			quotadaily = parseParamertToUint(value)
// 		case strings.Contains(value, "quotamonthly="):
// 			quotamonthly = parseParamertToUint(value)
// 		// case strings.Contains(value, "manual="):
// 		// 	manual = parseParamertToBool(value)
// 		case strings.Contains(value, "comment="):
// 			Comment = parseParamertToStr(value)
// 		case value != "":
// 			comments = comments + "/" + value
// 			// default:
// 			// 	comments = comments + "/" + value
// 		}
// 	}
// 	Comment = Comment + comments
// 	return
// }

// func (d DeviceType) convertToInfo(blockGroup, ManualAddresList string) InfoOldType {
// 	var (
// 		ip, mac, name, position, company, typeD, IDUser, comment string
// 		hourlyQuota, dailyQuota, monthlyQuota                    uint64
// 		manual                                                   bool
// 	)
// 	ip = validateIP(d.activeAddress, d.address)
// 	mac = getSwithMac(d.activeMacAddress, d.macAddress, d.clientId, d.activeClientId)
// 	hourlyQuota, dailyQuota, monthlyQuota, name, position, company, typeD, IDUser, comment, _ = parseComment(d.comment)
// 	blocked := inAddressList(d.addressLists, blockGroup)
// 	manual = inAddressList(d.addressLists, ManualAddresList)
// 	infoD := InfoOldType{
// 		DeviceOldType: DeviceOldType{
// 			Id:       d.Id,
// 			IP:       ip,
// 			Mac:      mac,
// 			AMac:     mac,
// 			HostName: d.hostName,
// 			Groups:   d.addressLists,
// 			timeout:  d.timeout,
// 			TypeD:    typeD,
// 			Manual:   manual,
// 		},
// 		QuotaType: QuotaType{
// 			HourlyQuota:  hourlyQuota,
// 			DailyQuota:   dailyQuota,
// 			MonthlyQuota: monthlyQuota,
// 			Blocked:      blocked,
// 			Dynamic:      paramertToBool(d.dynamic),
// 		},
// 		PersonType: PersonType{
// 			IDUser:   IDUser,
// 			Name:     name,
// 			Position: position,
// 			Company:  company,
// 			Comment:  comment,
// 			Comments: d.comment,
// 		},
// 	}
// 	return infoD
// }

// func (a *InfoType) convertToDevice(quotaDef QuotaType) DeviceType {
// 	return DeviceType{
// 		activeAddress:    a.activeAddress,
// 		address:          a.activeAddress,
// 		activeMacAddress: a.activeMacAddress,
// 		addressLists:     a.addressLists,
// 		blocked:          a.Blocked,
// 		comment:          a.convertToComment(quotaDef),
// 		disabledL:        fmt.Sprint(a.Disabled),
// 		hostName:         a.hostName,
// 		macAddress:       a.macAddress,
// 		Manual:           a.DeviceType.Manual,
// 		ShouldBeBlocked:  a.DeviceType.ShouldBeBlocked,
// 		timeout:          a.timeout,
// 	}
// }

// func (d1 *DeviceType) compare(d2 *DeviceType) bool {
// 	switch {
// 	case d1.macAddress == d2.macAddress && d1.macAddress != "" && d2.macAddress != "":
// 		return true
// 	case d1.activeMacAddress == d2.macAddress && d1.activeMacAddress != "" && d2.macAddress != "":
// 		return true
// 	case d1.macAddress == d2.activeMacAddress && d1.macAddress != "" && d2.activeMacAddress != "":
// 		return true
// 	case d1.activeMacAddress == d2.activeMacAddress && d1.activeMacAddress != "" && d2.activeMacAddress != "":
// 		return true
// 	case d1.activeClientId == d2.activeClientId && d1.activeClientId != "" && d2.activeClientId != "":
// 		return true
// 	case d1.activeClientId == d2.clientId && d1.activeClientId != "" && d2.clientId != "":
// 		return true
// 	case d1.clientId == d2.activeClientId && d1.clientId != "" && d2.activeClientId != "":
// 		return true
// 	case d1.clientId == d2.clientId && d1.clientId != "" && d2.clientId != "":
// 		return true
// 	}
// 	return false
// }

func boolToParamert(trigger bool) string {
	if trigger {
		return "yes"
	}
	return "no"
}

func lNow() *lineOfLogType {
	var l lineOfLogType
	tn := time.Now()
	l.year = tn.Year()
	l.month = tn.Month()
	l.day = tn.Day()
	l.hour = tn.Hour()
	l.minute = tn.Minute()
	return &l
}
