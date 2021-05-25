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

// func parseInfoFromMTToSlice(p parseType) []DeviceType {
// 	devices := DevicesType{}
// 	var b bytes.Buffer
// 	var i int = 1

// 	for b.Len() == 0 {
// 		if (i - p.MaxSSHRetries) == 0 {
// 			os.Exit(2)
// 		}
// 		fmt.Printf("\rTrying connect to MT(%d) ", i)
// 		b = getResponseOverSSHfMT(p.SSHCredentials, "/ip dhcp-server lease print detail without-paging")
// 		i++
// 	}
// 	fmt.Println("Connection successful.Get data.")
// 	devices.parseLeasePrint(b)
// 	return devices
// }

// func (ds *DevicesType) parseLeasePrint(b bytes.Buffer) {
// 	var d, dTemp DeviceType
// 	var disabled, radius, dynamic, blocked, n string
// 	var indexN int
// 	inputStr := b.String()
// 	inputStr = strings.ReplaceAll(inputStr, "\n", "")
// 	inputStr = strings.ReplaceAll(inputStr, "\t", "")
// 	inputStr = strings.ReplaceAll(inputStr, "\r", "")
// 	for i := 0; i >= 0; i = strings.Index(inputStr, "  ") {
// 		inputStr = strings.ReplaceAll(inputStr, "  ", " ")
// 	}
// 	arr := strings.Split(inputStr, " ")
// 	_ = saveArrToFile(arr)
// 	for index := 0; index < len(arr)-1; index++ {
// 		// if index == 3290 {
// 		// 	runtime.Breakpoint()
// 		// }
// 		switch {
// 		case arr[index] == "Flags:" || arr[index] == "-":
// 			continue
// 		case arr[index] == "disabled,":
// 			disabled = arr[index-2]
// 		case arr[index] == "radius,":
// 			radius = arr[index-2]
// 		case arr[index] == "dynamic,":
// 			dynamic = arr[index-2]
// 		case arr[index] == "blocked":
// 			blocked = arr[index-2]
// 		case isNumDot(arr[index]):
// 			n = arr[index]
// 			indexN = index
// 			dTemp = d
// 		case isParametr(arr[index], "address"):
// 			d = DeviceType{}
// 			d.address = parseParamertToStr(arr[index])
// 			if index-indexN > 2 {
// 				d.comment = strings.Join(arr[indexN+2:index], " ")
// 				// fmt.Printf("indexN(%v), arr[indexN:index](%v), d.comment(%v)\n", indexN, arr[indexN:index], d.comment)
// 			}
// 			dTemp.Id = n
// 			*(ds) = append(*(ds), dTemp)
// 		case isParametr(arr[index], "address-lists"):
// 			d.addressLists = parseParamertToStr(arr[index])
// 		case isParametr(arr[index], "server"):
// 			d.server = parseParamertToStr(arr[index])
// 		case isParametr(arr[index], "dhcp-option"):
// 			d.dhcpOption = parseParamertToStr(arr[index])
// 		case isParametr(arr[index], "status"):
// 			d.status = parseParamertToStr(arr[index])
// 		case isParametr(arr[index], "expires-after"):
// 			d.expiresAfter = parseParamertToStr(arr[index])
// 		case isParametr(arr[index], "last-seen"):
// 			d.lastSeen = parseParamertToStr(arr[index])
// 		case isParametr(arr[index], "active-address"):
// 			d.activeAddress = parseParamertToStr(arr[index])
// 		case isParametr(arr[index], "active-mac-address"):
// 			d.activeMacAddress = parseParamertToStr(arr[index])
// 		case isParametr(arr[index], "mac-address"):
// 			d.macAddress = parseParamertToStr(arr[index])
// 			// if d.macAddress == "E8:D8:D1:47:55:93" {
// 			// 	fmt.Printf("arr[index-20:index+20](%v)", arr[index-20:index+20])
// 			// 	runtime.Breakpoint()
// 			// }
// 		case isParametr(arr[index], "mac-address"):
// 			d.activeMacAddress = parseParamertToStr(arr[index])
// 		case isParametr(arr[index], "active-client-id"):
// 			d.activeClientId = parseParamertToStr(arr[index])
// 		case isParametr(arr[index], "client-id"):
// 			d.clientId = parseParamertToStr(arr[index])
// 		case isParametr(arr[index], "active-server"):
// 			d.activeServer = parseParamertToStr(arr[index])
// 		case isParametr(arr[index], "host-name"):
// 			d.hostName = parseParamertToStr(arr[index])
// 		case isParametr(arr[index], "radius"):
// 			d.radius = parseParamertToStr(arr[index])
// 		case arr[index] == disabled:
// 			d.disabledL = "yes"
// 			// fmt.Printf("arr[index-5:index+5](%v)", arr[index-5:index+5])
// 		case arr[index] == blocked:
// 			d.blocked = "yes"
// 			// fmt.Printf("arr[index-5:index+5](%v)", arr[index-5:index+5])
// 		case arr[index] == dynamic:
// 			d.dynamic = "yes"
// 			// fmt.Printf("arr[index-5:index+5](%v)", arr[index-5:index+5])
// 		case arr[index] == radius:
// 			d.radius = "yes"
// 			// fmt.Printf("arr[index-5:index+5](%v)", arr[index-5:index+5])
// 		}
// 	}
// }

func parseInfoFromMTAsValueToSlice(p parseType) []DeviceType {
	devices := DevicesType{}
	var b bytes.Buffer
	var i int = 1

	for b.Len() == 0 {
		if (i - p.MaxSSHRetries) == 0 {
			os.Exit(2)
		}
		fmt.Printf("\rTrying connect to MT(%d) ", i)
		b = getResponseOverSSHfMT(p.SSHCredentials, ":put [/ip dhcp-server lease print as-value]")
		i++
	}
	fmt.Println("Connection successful.Get data.")
	devices.parseLeasePrintAsValue(b)
	return devices
}

func (ds *DevicesType) parseLeasePrintAsValue(b bytes.Buffer) {
	var d DeviceType
	inputStr := b.String()
	// inputStr = strings.ReplaceAll(inputStr, "\n", "")
	_ = saveStrToFile(inputStr)
	arr := strings.Split(inputStr, ";")
	_ = saveArrToFile(arr)
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
			// case d.address == "192.168.65.126":
			// 	fmt.Print("q")
		}

	}
}

func (ds *DevicesType) findDeviceToConvertInfoD(alias, blockGroup string, q QuotaType) InfoType {
	for _, d := range *ds {
		if d.activeAddress == alias || d.activeMacAddress == alias || d.address == alias || d.macAddress == alias {
			infoD := d.convertToInfo(blockGroup)
			infoD.QuotaType = checkNULLQuotas(infoD.QuotaType, q)

			return infoD
		}
	}
	return InfoType{}
}

func (a AliasOld) sendByAll(p parseType, qDefault QuotaType) error {
	var err error
	var idStr string
	comments := a.convertToComment(qDefault)

	switch {
	case a.Id != "":
		idStr = a.Id
	case a.Mac != "":
		idStr, err = reciveIDByMac(p, a.Mac)
		if err != nil {
			return err
		}
	case a.AMac != "":
		idStr, err = reciveIDByMac(p, a.AMac)
		if err != nil {
			return err
		}
	case a.IP != "":
		idStr, err = reciveIDByIP(p, a.IP)
		if err != nil {
			return err
		}
	case a.Alias != "" && isMac(a.Alias):
		idStr, err = reciveIDByMac(p, a.Alias)
		if err != nil {
			return err
		}
	case a.Alias != "" && isIP(a.Alias):
		idStr, err = reciveIDByIP(p, a.Alias)
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
			a.Groups,
			boolToParamert(a.Disabled),
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

// func (a AliasOld) sendByAll(p parseType, qDefault QuotaType) error {
// 	var err error
// 	var idStr string
// 	comments := a.convertToComment(qDefault)
// 	switch {
// 	case a.Mac != "":
// 		idStr, err = reciveIDByMac(p, a.Mac)
// 		if err != nil {
// 			return err
// 		}
// 	case a.AMac != "":
// 		idStr, err = reciveIDByMac(p, a.AMac)
// 		if err != nil {
// 			return err
// 		}
// 	case a.IP != "":
// 		idStr, err = reciveIDByIP(p, a.IP)
// 		if err != nil {
// 			return err
// 		}
// 	case a.Alias != "" && isMac(a.Alias):
// 		idStr, err = reciveIDByMac(p, a.Alias)
// 		if err != nil {
// 			return err
// 		}
// 	case a.Alias != "" && isIP(a.Alias):
// 		idStr, err = reciveIDByIP(p, a.Alias)
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
// 			a.Groups,
// 			boolToParamert(a.Disabled),
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
