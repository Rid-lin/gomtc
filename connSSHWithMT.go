package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func getResponseOverSSHfMT(sshCred SSHCredentials, command string) bytes.Buffer {
	sshConfig := &ssh.ClientConfig{
		User: sshCred.SSHUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshCred.SSHPass),
		},
		// Non-production only
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	// Established connection
	client, err := ssh.Dial("tcp", sshCred.SSHHost+":"+sshCred.SSHPort, sshConfig)
	if err != nil {
		log.Errorf("Failed to dial: %s", err)
	}
	defer client.Close()
	var b bytes.Buffer
	// crete ssh session

	session, err := client.NewSession()
	if err != nil {
		log.Errorf("Failed to ceate session: %s", err)
	}
	defer session.Close()
	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	session.Stdout = &b
	if err := session.Run(command); err != nil {
		log.Error("Failed to run: " + err.Error())
	}
	session.Close()
	return b
}

func parseInfoFromMTAsValueToSlice(p parseType) []DeviceType {
	devices := DevicesType{}
	var b bytes.Buffer
	var i int = 1

	for b.Len() == 0 {
		if (i - p.MaxSSHRetries) == 0 {
			os.Exit(2)
		}
		fmt.Printf("\rTrying connect to MT(%d) ", i)
		b = getResponseOverSSHfMT(p.SSHCredentials, ":put [/ip dhcp-server lease print detail as-value]")
		i++
	}
	fmt.Println("Connection successful.Get data.")
	devices.parseLeasePrintAsValue(b)
	return devices
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
			d.activeAddress = parseParamertToStr(lineItem)
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
			d.addressLists = parseParamertToStr(lineItem)
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
			d.activeMacAddress = parseParamertToStr(lineItem)
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
		case isParametr(lineItem, "blocked"):
			d.blocked = parseParamertToStr(lineItem)
		case isParametr(lineItem, "dhcp-option-set"):
			d.dhcpOptionSet = parseParamertToStr(lineItem)
		case isParametr(lineItem, "host-name"):
			d.hostName = parseParamertToStr(lineItem)
		case isParametr(lineItem, "mac-address"):
			d.macAddress = parseParamertToStr(lineItem)
		case isParametr(lineItem, "src-mac-address"):
			d.srcMacAddress = parseParamertToStr(lineItem)
		case isComment(lineItem, "comment"):
			d.comment = parseParamertToComment(lineItem)
		case isParametr(lineItem, "status"):
			d.status = parseParamertToStr(lineItem)
			*ds = append(*ds, d)
			d = DeviceType{}
		case addedTo == "address-lists":
			d.addressLists = d.addressLists + "," + lineItem
		}

	}
}

func (info InfoType) sendByAll(p parseType, qDefault QuotaType) error {
	var err error
	var idStr string
	comments := info.convertToComment(qDefault)

	switch {
	case info.Id != "":
		idStr = info.Id
	case info.activeMacAddress != "":
		idStr, err = reciveIDByMac(p, info.activeMacAddress)
		if err != nil {
			return err
		}
	case info.macAddress != "":
		idStr, err = reciveIDByMac(p, info.macAddress)
		if err != nil {
			return err
		}
	case info.activeAddress != "":
		idStr, err = reciveIDByIP(p, info.activeAddress)
		if err != nil {
			return err
		}
	case info.InfoName != "" && isMac(info.InfoName):
		idStr, err = reciveIDByMac(p, info.InfoName)
		if err != nil {
			return err
		}
	case info.InfoName != "" && isIP(info.InfoName):
		idStr, err = reciveIDByIP(p, info.InfoName)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Mac and IP addres canot be empty")
	}
	if idStr == "" {
		return fmt.Errorf("Device not found")
	}
	idArr := strings.Split(idStr, ";")
	for _, id := range idArr {
		command := fmt.Sprintf(`/ip dhcp-server lease set number="%s" address-lists="%s" disabled="%s" comment="%s"`,
			id,
			info.addressLists,
			boolToParamert(info.Disabled),
			comments)
		b := getResponseOverSSHfMT(p.SSHCredentials, command)
		result := b.String()
		fmt.Printf("command:'%s',result:'%s'\n", command, result)
		if b.Len() > 0 {
			return fmt.Errorf(b.String())
		}
	}
	return nil
}

func reciveIDByMac(p parseType, mac string) (string, error) {
	if mac == "" {
		return "", fmt.Errorf("MAC address cannot be empty")
	}
	return reciveIDBy(p, "mac-address", mac)
}

func reciveIDBy(p parseType, entity, value string) (string, error) {
	command := fmt.Sprintf(`:put [/ip dhcp-server lease find %s="%s"]`,
		entity, value)
	b := getResponseOverSSHfMT(p.SSHCredentials, command)
	id := b.String()
	id = strings.ReplaceAll(id, `"`, "")
	id = strings.ReplaceAll(id, "\n", "")
	id = strings.ReplaceAll(id, "\r", "")
	return id, nil
}

func reciveIDByIP(p parseType, ip string) (string, error) {
	if ip == "" {
		return "", fmt.Errorf("IP address cannot be empty")
	}
	return reciveIDBy(p, "active-address", ip)
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
	_ = saveStrToFile("./command.temp", command)
	b := getResponseOverSSHfMT(p.SSHCredentials, command)
	if b.Len() > 0 {
		log.Errorf("Error save device to Mikrotik(%v) with command:\n%v", b.String(), command)
	}
}
