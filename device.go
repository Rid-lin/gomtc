package main

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	. "git.vegner.org/vsvegner/gomtc/internal/ssh"

	log "github.com/sirupsen/logrus"
)

func (t *Transport) getDevices() {
	t.Lock()
	devices := parseInfoFromMTAsValueToSlice(parseType{
		SSHCredentials:   t.sshCredentials,
		QuotaType:        t.QuotaType,
		BlockAddressList: t.BlockAddressList,
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

func (t *Transport) SendGroupStatus(NoControl bool) {
	if NoControl {
		return
	}
	t.RLock()
	p := parseType{
		SSHCredentials:   t.sshCredentials,
		QuotaType:        t.QuotaType,
		BlockAddressList: t.BlockAddressList,
	}
	t.RUnlock()
	t.change.sendLeaseSet(p)
	t.Lock()
	t.change = BlockDevices{}
	t.Unlock()
}

func (ds *DevicesType) parseLeasePrintAsValue(b bytes.Buffer) {
	var d DeviceType
	var addedTo string
	inputStr := b.String()
	arr := strings.Split(inputStr, ";")
	// For Debug
	_ = saveStrToFile("./str.temp", inputStr)
	_ = saveArrToFile("./arr.temp", arr)
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

// func (info InfoType) sendByAll(p parseType, qDefault QuotaType) error {
// 	var err error
// 	var idStr string
// 	comments := info.convertToComment(qDefault)

// 	switch {
// 	case info.Id != "":
// 		idStr = info.Id
// 	case info.activeMacAddress != "":
// 		idStr, err = reciveIDByMac(p, info.activeMacAddress)
// 		if err != nil {
// 			return err
// 		}
// 	case info.macAddress != "":
// 		idStr, err = reciveIDByMac(p, info.macAddress)
// 		if err != nil {
// 			return err
// 		}
// 	case info.activeAddress != "":
// 		idStr, err = reciveIDByIP(p, info.activeAddress)
// 		if err != nil {
// 			return err
// 		}
// 	case info.InfoName != "" && isMac(info.InfoName):
// 		idStr, err = reciveIDByMac(p, info.InfoName)
// 		if err != nil {
// 			return err
// 		}
// 	case info.InfoName != "" && isIP(info.InfoName):
// 		idStr, err = reciveIDByIP(p, info.InfoName)
// 		if err != nil {
// 			return err
// 		}
// 	default:
// 		return fmt.Errorf("Mac and IP addres canot be empty")
// 	}
// 	if idStr == "" {
// 		return fmt.Errorf("Device not found")
// 	}
// 	idArr := strings.Split(idStr, ";")
// 	for _, id := range idArr {
// 		command := fmt.Sprintf(`/ip dhcp-server lease set number="%s" address-lists="%s" disabled="%s" comment="%s"`,
// 			id,
// 			info.addressLists,
// 			boolToParamert(info.Disabled),
// 			comments)
// 		b := getResponseOverSSHfMT(p.SSHCredentials, command)
// 		result := b.String()
// 		fmt.Printf("command:'%s',result:'%s'\n", command, result)
// 		if b.Len() > 0 {
// 			return fmt.Errorf(b.String())
// 		}
// 	}
// 	return nil
// }

// func reciveIDByMac(p parseType, mac string) (string, error) {
// 	if mac == "" {
// 		return "", fmt.Errorf("MAC address cannot be empty")
// 	}
// 	return reciveIDBy(p, "mac-address", mac)
// }

// func reciveIDBy(p parseType, entity, value string) (string, error) {
// 	command := fmt.Sprintf(`:put [/ip dhcp-server lease find %s="%s"]`,
// 		entity, value)
// 	b := getResponseOverSSHfMT(p.SSHCredentials, command)
// 	id := b.String()
// 	id = strings.ReplaceAll(id, `"`, "")
// 	id = strings.ReplaceAll(id, "\n", "")
// 	id = strings.ReplaceAll(id, "\r", "")
// 	return id, nil
// }

// func reciveIDByIP(p parseType, ip string) (string, error) {
// 	if ip == "" {
// 		return "", fmt.Errorf("IP address cannot be empty")
// 	}
// 	return reciveIDBy(p, "active-address", ip)
// }

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
	_ = saveStrToFile("./command.temp", command)
	b := GetResponseOverSSHfMT(p.SSHHost, p.SSHPort, p.SSHUser, p.SSHPass, command)
	if b.Len() > 0 {
		log.Errorf("Error save device to Mikrotik(%v) with command:\n%v", b.String(), command)
	}
}
