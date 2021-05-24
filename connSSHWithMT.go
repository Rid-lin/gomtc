package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

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

// func parseInfoFromMTToSlice(p parseType) []DeviceTyp {

// 	deviceTemp, device := DeviceType{}, DeviceTyp{}
// 	devices := []DeviceTyp{}
// 	// var b bytes.Bufer
// 	b := getResponseOverSSHfMT(p.SSHCredentials, "/ip dhcp-server lease print detail without-pagin")
// 	inputStr := b.Strin()
// 	inputArr := strings.Split(inputStr, "\")
// 	for _, line := range inputAr {
// 		// Если возникает ошибка, то это новое устройсво
// 		if deviceTemp.parseLine(line) != ni {

// 			// Поэтому мы его дополняем предыдущее значение устройства дополняем..
// 			device.timeout = time.Now().In(p.Locatin)
// 			// ... и записываем в массив устройтв
// 			devices = append(devices, devie)
// 			// Обнуляем временное хранилище информации об устройстве для дальнейшей кокатенации из распарсенных стрк.
// 			deviceTemp = DeviceTyp{}
// 			_ = deviceTemp.parseLine(lie)
// 	}
// 		device = deviceTmp
//	}
// 	return devies
// / }

// func (d *DeviceType) parseLine(l string) (err error {
// 	l = strings.Trim(l, "")
// 	l = strings.ReplaceAll(l, "  ", "")
// 	arr := strings.Split(l, "")
// 	if len(arr) > 1 && isNumDot(arr[0] {
// 		err = fmt.Errorf("New lin")
//	}
// 	for index, s := range ar {
// 		switc {
// 		case s == "Flags":
// 			return il
// 		case strings.Contains(s, "address-lists=):
// 			d.addressLists = parseParamertToStrs)
// 		case strings.Contains(s, "server=):
// 			d.server = parseParamertToStrs)
// 		case strings.Contains(s, "dhcp-option=):
// 			d.dhcpOption = parseParamertToStrs)
// 		case strings.Contains(s, "status=):
// 			d.status = parseParamertToStrs)
// 		case strings.Contains(s, "expires-after=):
// 			d.expiresAfter = parseParamertToStrs)
// 		case strings.Contains(s, "last-seen=):
// 			d.lastSeen = parseParamertToStrs)
// 		case strings.Contains(s, "active-address=):
// 			d.activeAddress = parseParamertToStrs)
// 		case strings.Contains(s, "active-mac-address=):
// 			d.activeMacAddress = parseParamertToStrs)
// 		case strings.Contains(s, "mac-address=):
// 			d.macAddress = parseParamertToStrs)
// 		case strings.Contains(s, "address=):
// 			d.address = parseParamertToStrs)
// 		case strings.Contains(s, "mac-address=):
// 			d.activeMacAddress = parseParamertToStrs)
// 		case strings.Contains(s, "active-client-id=):
// 			d.activeClientId = parseParamertToStrs)
// 		case strings.Contains(s, "client-id=):
// 			d.clientId = parseParamertToStrs)
// 		case strings.Contains(s, "active-server=):
// 			d.activeServer = parseParamertToStrs)
// 		case strings.Contains(s, "host-name=):
// 			d.hostName = parseParamertToStrs)
// 		case strings.Contains(s, "radius=):
// 			d.radius = parseParamertToStrs)
// 		// case isNumDot():
// 		// 	err = fmt.Errorf("New lin")
// 		case l == "\":
// 			return fmt.Errorf("New lin")
// 			// fallthrogh
// 		// case strings.Contains(s, ";;;):
// 		case s == "":
// 			d.disabled = "ys"
// 		case s == "":
// 			d.blocked = "ys"
// 		case s == "":
// 			d.dynamic = "ys"
// 		case s == ";;":
// 			d.comment = strings.Join(arr[index+1:], "")
// 			return rr
// 	}
//	}
// 	return rr
// / }

func parseInfoFromMTToSlice2(p parseType) []DeviceType {
	devices := DevicesType{}
	var b bytes.Buffer
	var i int = 1

	for b.Len() == 0 {
		if (i - p.MaxSSHRetries) == 0 {
			os.Exit(2)
		}
		fmt.Printf("\rTrying connect to MT(%d) ", i)
		b = getResponseOverSSHfMT(p.SSHCredentials, "/ip dhcp-server lease print detail without-paging")
		i++
	}
	fmt.Println("Connection successful.Get data.")
	devices.parseBuffer(b)
	return devices
}

func (ds *DevicesType) parseBuffer(b bytes.Buffer) {
	var d, dTemp DeviceType
	var disabled, radius, dynamic, blocked, n string
	var indexN int
	inputStr := b.String()
	inputStr = strings.ReplaceAll(inputStr, "\n", "")
	inputStr = strings.ReplaceAll(inputStr, "\t", "")
	inputStr = strings.ReplaceAll(inputStr, "\r", "")
	for i := 0; i >= 0; i = strings.Index(inputStr, "  ") {
		inputStr = strings.ReplaceAll(inputStr, "  ", " ")
	}
	arr := strings.Split(inputStr, " ")
	_ = saveStrToFile(arr)
	for index := 0; index < len(arr)-1; index++ {
		// if index == 3290 {
		// 	runtime.Breakpoint()
		// }
		switch {
		case arr[index] == "Flags:" || arr[index] == "-":
			continue
		case arr[index] == "disabled,":
			disabled = arr[index-2]
		case arr[index] == "radius,":
			radius = arr[index-2]
		case arr[index] == "dynamic,":
			dynamic = arr[index-2]
		case arr[index] == "blocked":
			blocked = arr[index-2]
		case isNumDot(arr[index]):
			n = arr[index]
			indexN = index
			dTemp = d
		case isParametr(arr[index], "address"):
			d = DeviceType{}
			d.address = parseParamertToStr(arr[index])
			if index-indexN > 2 {
				d.comment = strings.Join(arr[indexN+2:index], " ")
				// fmt.Printf("indexN(%v), arr[indexN:index](%v), d.comment(%v)\n", indexN, arr[indexN:index], d.comment)
			}
			dTemp.Id = n
			*(ds) = append(*(ds), dTemp)
		case isParametr(arr[index], "address-lists"):
			d.addressLists = parseParamertToStr(arr[index])
		case isParametr(arr[index], "server"):
			d.server = parseParamertToStr(arr[index])
		case isParametr(arr[index], "dhcp-option"):
			d.dhcpOption = parseParamertToStr(arr[index])
		case isParametr(arr[index], "status"):
			d.status = parseParamertToStr(arr[index])
		case isParametr(arr[index], "expires-after"):
			d.expiresAfter = parseParamertToStr(arr[index])
		case isParametr(arr[index], "last-seen"):
			d.lastSeen = parseParamertToStr(arr[index])
		case isParametr(arr[index], "active-address"):
			d.activeAddress = parseParamertToStr(arr[index])
		case isParametr(arr[index], "active-mac-address"):
			d.activeMacAddress = parseParamertToStr(arr[index])
		case isParametr(arr[index], "mac-address"):
			d.macAddress = parseParamertToStr(arr[index])
			// if d.macAddress == "E8:D8:D1:47:55:93" {
			// 	fmt.Printf("arr[index-20:index+20](%v)", arr[index-20:index+20])
			// 	runtime.Breakpoint()
			// }
		case isParametr(arr[index], "mac-address"):
			d.activeMacAddress = parseParamertToStr(arr[index])
		case isParametr(arr[index], "active-client-id"):
			d.activeClientId = parseParamertToStr(arr[index])
		case isParametr(arr[index], "client-id"):
			d.clientId = parseParamertToStr(arr[index])
		case isParametr(arr[index], "active-server"):
			d.activeServer = parseParamertToStr(arr[index])
		case isParametr(arr[index], "host-name"):
			d.hostName = parseParamertToStr(arr[index])
		case isParametr(arr[index], "radius"):
			d.radius = parseParamertToStr(arr[index])
		case arr[index] == disabled:
			d.disabled = "yes"
			// fmt.Printf("arr[index-5:index+5](%v)", arr[index-5:index+5])
		case arr[index] == blocked:
			d.blocked = "yes"
			// fmt.Printf("arr[index-5:index+5](%v)", arr[index-5:index+5])
		case arr[index] == dynamic:
			d.dynamic = "yes"
			// fmt.Printf("arr[index-5:index+5](%v)", arr[index-5:index+5])
		case arr[index] == radius:
			d.radius = "yes"
			// fmt.Printf("arr[index-5:index+5](%v)", arr[index-5:index+5])
		}
	}
}

func saveDeviceToCSV(devices []DeviceType) error {
	f, e := os.Create("./Device.csv")
	if e != nil {
		fmt.Println(e)
	}

	writer := csv.NewWriter(f)

	for _, device := range devices {
		fields := device.convertToSlice()
		err := writer.Write(fields)
		if err != nil {
			fmt.Println(e)
		}
	}
	return nil
}

func (d DeviceType) convertToSlice() []string {
	var slice []string
	slice = append(slice, d.Id)
	slice = append(slice, d.activeAddress)
	slice = append(slice, d.activeClientId)
	slice = append(slice, d.activeMacAddress)
	slice = append(slice, d.activeServer)
	slice = append(slice, d.address)
	slice = append(slice, d.addressLists)
	slice = append(slice, d.blocked)
	slice = append(slice, d.clientId)
	slice = append(slice, d.comment)
	slice = append(slice, d.dhcpOption)
	slice = append(slice, d.disabled)
	slice = append(slice, d.dynamic)
	slice = append(slice, d.expiresAfter)
	slice = append(slice, d.hostName)
	slice = append(slice, d.lastSeen)
	slice = append(slice, d.macAddress)
	slice = append(slice, d.radius)
	slice = append(slice, d.server)
	slice = append(slice, d.status)
	slice = append(slice, fmt.Sprint(d.Manual))
	slice = append(slice, fmt.Sprint(d.ShouldBeBlocked))
	slice = append(slice, d.timeout.Format("2006-01-02 15:04:05 -0700 MST "))
	return slice
}

func (t *Transport) getAliasS(alias string) AliasOld {
	key := KeyMapOfReports{
		Alias:   alias,
		DateStr: time.Now().In(t.Location).Format("2006-01-02"),
	}
	InfoD, ok := t.dataCasheOld[key]
	if !ok {
		return AliasOld{}
	}
	return InfoD
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

// func (a AliasType) sendByMacAndIP(p parseType, qDefault QuotaType) error {
// 	var err2 error
// 	comments := a.convertToComment(qDefault)

// 	idStr, err := reciveIDByMac(p, a.Mac)
// 	if err != nil {
// 		return err
// 	}
// 	if idStr == "" {
// 		// return fmt.Errorf("Device not found")
// 		idStr, err2 = reciveIDByIP(p, a.IP)
// 		if err2 != nil {
// 			return err2
// 		}
// 		if idStr == "" {
// 			return fmt.Errorf("Device not found")
// 		}
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
