package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	. "git.vegner.org/vsvegner/gomtc/internal/ssh"

	log "github.com/sirupsen/logrus"
)

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

func (ds *DevicesType) parseLeasePrintAsValue(b bytes.Buffer) {
	var d DeviceType
	var addedTo string
	inputStr := b.String()
	arr := strings.Split(inputStr, ";")
	// For Debug
	_ = saveStrToFile(".config/str.temp", inputStr)
	_ = saveArrToFile(".config/arr.temp", arr)
	for _, lineItem := range arr {
		switch {
		case isParametr(lineItem, ".id"):
			d.Id = parseParamertToStr(lineItem)
		case isParametr(lineItem, "active-address"):
			d.ActiveAddress = parseParamertToStr(lineItem)
		case isParametr(lineItem, "address"):
			d.address = parseParamertToStr(lineItem)
		case isParametr(lineItem, "allow-dual-stack-queue"):
			d.allowDualStackQueue = parseParamertToStr(lineItem)
		case isParametr(lineItem, "client-id"):
			addedTo = "client-id"
			d.clientId = parseParamertToStr(lineItem)
		case isParametr(lineItem, "disabled"):
			d.disabledL = parseParamertToStr(lineItem)
		case isParametr(lineItem, "insert-queue-before"):
			d.insertQueueBefore = parseParamertToStr(lineItem)
		case isParametr(lineItem, "radius"):
			d.radius = parseParamertToStr(lineItem)
		case isParametr(lineItem, "active-client-id"):
			d.activeClientId = parseParamertToStr(lineItem)
		case isParametr(lineItem, "address-lists"):
			addedTo = "address-lists"
			d.AddressLists = parseParamertToStr(lineItem)
		case isParametr(lineItem, "always-broadcast"):
			d.alwaysBroadcast = parseParamertToStr(lineItem)
		case isParametr(lineItem, "dynamic"):
			d.dynamic = parseParamertToStr(lineItem)
		case isParametr(lineItem, "last-seen"):
			d.lastSeen = parseParamertToStr(lineItem)
		case isParametr(lineItem, "rate-limit"):
			d.rateLimit = parseParamertToStr(lineItem)
		case isParametr(lineItem, "use-src-mac"):
			d.useSrcMac = parseParamertToStr(lineItem)
		case isParametr(lineItem, "active-mac-address"):
			d.ActiveMacAddress = parseParamertToStr(lineItem)
		case isParametr(lineItem, "agent-circuit-id"):
			d.agentCircuitId = parseParamertToStr(lineItem)
		case isParametr(lineItem, "block-access"):
			d.blockAccess = parseParamertToStr(lineItem)
		case isParametr(lineItem, "dhcp-option"):
			d.dhcpOption = parseParamertToStr(lineItem)
		case isParametr(lineItem, "expires-after"):
			d.expiresAfter = parseParamertToStr(lineItem)
		case isParametr(lineItem, "lease-time"):
			d.leaseTime = parseParamertToStr(lineItem)
		case isParametr(lineItem, "server"):
			d.server = parseParamertToStr(lineItem)
		case isParametr(lineItem, "active-server"):
			d.activeServer = parseParamertToStr(lineItem)
		case isParametr(lineItem, "agent-remote-id"):
			d.agentRemoteId = parseParamertToStr(lineItem)
		// case isParametr(lineItem, "blocked"):
		// 	d.blocked = parseParamertToStr(lineItem)
		case isParametr(lineItem, "dhcp-option-set"):
			d.dhcpOptionSet = parseParamertToStr(lineItem)
		case isParametr(lineItem, "host-name"):
			d.HostName = parseParamertToStr(lineItem)
		case isParametr(lineItem, "mac-address"):
			d.macAddress = parseParamertToStr(lineItem)
		case isParametr(lineItem, "src-mac-address"):
			d.srcMacAddress = parseParamertToStr(lineItem)
		case isComment(lineItem, "comment"):
			d.Comment = parseParamertToComment(lineItem)
		case isParametr(lineItem, "status"):
			d.status = parseParamertToStr(lineItem)
			*ds = append(*ds, d)
			d = DeviceType{}
		case addedTo == "address-lists":
			d.AddressLists = d.AddressLists + "," + lineItem
		}
	}
}

func ToReportData(as map[string]AliasType, sd map[KeyDevice]StatDeviceType, ds DevicesMapType) ReportDataType {
	var totalVolumePerDay uint64
	var totalVolumePerHour [24]uint64

	ReportData := ReportDataType{}
	for key, value := range sd {
		line := LineOfDisplay{}
		line.Alias = key.mac
		line.VolumePerDay = value.VolumePerDay
		totalVolumePerDay += value.VolumePerDay
		// TODO подумать над ключом
		line.InfoType.PersonType = as[key.mac].PersonType
		line.InfoType.QuotaType = as[key.mac].QuotaType
		line.InfoType.DeviceType = ds[key]
		for i := range line.PerHour {
			line.PerHour[i] = value.PerHour[i]
			totalVolumePerHour[i] += value.PerHour[i]
		}
		ReportData = add(ReportData, line)
	}
	line := LineOfDisplay{}
	line.Alias = "Всего"
	line.VolumePerDay = totalVolumePerDay
	line.PerHour = totalVolumePerHour
	ReportData = add(ReportData, line)
	return ReportData
}

func parseInfoFromMTAsValueToSlice(p parseType) []DeviceType {
	devices := DevicesType{}
	b, err := GetResponseOverSSHfMTWithBuffer(p.SSHHost, p.SSHPort, p.SSHUser, p.SSHPass,
		":put [/ip dhcp-server lease print detail as-value]",
		p.MaxSSHRetries, int(p.SSHRetryDelay))
	if err != nil {
		return devices
	}
	devices.parseLeasePrintAsValue(b)
	return devices
}

func (a *BlockDevices) sendLeaseSet(p parseType) {
	var command string
	firstCommand := "/ip dhcp-server lease set "
	for _, item := range *a {
		if item.Id == "" {
			continue
		}
		itemCommand := fmt.Sprintf("number=%s disabled=%s address-lists=%s\n",
			item.Id, boolToParamert(item.Disabled), item.Groups)
		command = command + firstCommand + itemCommand
	}
	// For Debug
	_ = saveStrToFile("./config/command.temp", command)
	b := GetResponseOverSSHfMT(p.SSHHost, p.SSHPort, p.SSHUser, p.SSHPass, command)
	if b.Len() > 0 {
		log.Errorf("Error save device to Mikrotik(%v) with command:\n%v", b.String(), command)
	}
}

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

func boolToParamert(trigger bool) string {
	if trigger {
		return "yes"
	}
	return "no"
}

func saveArrToFile(nameFile string, arr []string) error {
	f, _ := os.Create(nameFile)
	defer f.Close()
	w := bufio.NewWriter(f)
	for index := 0; index < len(arr)-1; index++ {
		fmt.Fprintln(w, arr[index])
	}
	w.Flush()
	return nil
}

func saveStrToFile(nameFile, str string) error {
	f, _ := os.Create(nameFile)
	defer f.Close()
	_, _ = f.WriteString(str)
	return nil
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
