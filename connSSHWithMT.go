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
	var i uint16 = 1
	// crete ssh session
	for b.Len() == 0 {
		if sshCred.MaxSSHRetries == 1 {
			os.Exit(2)
		}
		fmt.Printf("\rTrying connect to MT(%d) ", i)

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
		i++
	}
	fmt.Println("Connection successful.Get data.")
	return b
}

func parseInfoFromMTToSlice(p parseType) []DeviceType {

	deviceTemp, device := DeviceType{}, DeviceType{}
	devices := []DeviceType{}
	// var b bytes.Buffer
	b := getResponseOverSSHfMT(p.SSHCredentials, "/ip dhcp-server lease print detail without-paging")
	inputStr := b.String()
	inputArr := strings.Split(inputStr, "\n")
	for _, line := range inputArr {
		// Если возникает ошибка, то это новое устройство
		if deviceTemp.parseLine(line) != nil {

			// Поэтому мы его дополняем предыдущее значение устройства дополняем...
			device.timeout = time.Now().In(p.Location)
			// ... и записываем в массив устройств
			devices = append(devices, device)
			// Обнуляем временное хранилище информации об устройстве для дальнейшей кокатенации из распарсенных строк.
			deviceTemp = DeviceType{}
			_ = deviceTemp.parseLine(line)
		}
		device = deviceTemp
	}
	return devices
}

func (d *DeviceType) parseLine(l string) (err error) {
	l = strings.Trim(l, " ")
	l = strings.ReplaceAll(l, "  ", " ")
	arr := strings.Split(l, " ")
	if len(arr) > 1 && isNumDot(arr[0]) {
		err = fmt.Errorf("New line")
	}
	for index, s := range arr {
		switch {
		case s == "Flags:":
			return nil
		case strings.Contains(s, "address-lists="):
			d.addressLists = parseParamertToStr(s)
		case strings.Contains(s, "server="):
			d.server = parseParamertToStr(s)
		case strings.Contains(s, "dhcp-option="):
			d.dhcpOption = parseParamertToStr(s)
		case strings.Contains(s, "status="):
			d.status = parseParamertToStr(s)
		case strings.Contains(s, "expires-after="):
			d.expiresAfter = parseParamertToStr(s)
		case strings.Contains(s, "last-seen="):
			d.lastSeen = parseParamertToStr(s)
		case strings.Contains(s, "active-address="):
			d.activeAddress = parseParamertToStr(s)
		case strings.Contains(s, "active-mac-address="):
			d.activeMacAddress = parseParamertToStr(s)
		case strings.Contains(s, "mac-address="):
			d.macAddress = parseParamertToStr(s)
		case strings.Contains(s, "address="):
			d.address = parseParamertToStr(s)
		case strings.Contains(s, "mac-address="):
			d.activeMacAddress = parseParamertToStr(s)
		case strings.Contains(s, "active-client-id="):
			d.activeClientId = parseParamertToStr(s)
		case strings.Contains(s, "client-id="):
			d.clientId = parseParamertToStr(s)
		case strings.Contains(s, "active-server="):
			d.activeServer = parseParamertToStr(s)
		case strings.Contains(s, "host-name="):
			d.hostName = parseParamertToStr(s)
		case strings.Contains(s, "radius="):
			d.radius = parseParamertToStr(s)
		// case isNumDot(s):
		// 	err = fmt.Errorf("New line")
		case l == "\n":
			return fmt.Errorf("New line")
			// fallthrough
		// case strings.Contains(s, ";;;"):
		case s == "X":
			d.disabled = "yes"
		case s == "B":
			d.blocked = "yes"
		case s == "D":
			d.dynamic = "yes"
		case s == ";;;":
			d.comment = strings.Join(arr[index+1:], " ")
			return err
		}
	}
	return err
}

func parseInfoFromMTToSlice2(p parseType) []DeviceType {

	devices := DevicesType{}
	var b bytes.Buffer
	for i := p.MaxSSHRetries; b.Len() < 1; i-- {
		if i == 0 {
			os.Exit(1)
		}
		fmt.Printf("Trying connect to MT(%d)\r", i)
		b = getResponseOverSSHfMT(p.SSHCredentials, "/ip dhcp-server lease print detail without-paging")
		time.Sleep(time.Second * time.Duration(p.SSHRetryDelay))
	}
	fmt.Println("Connection successful.Get data.\r")
	devices.parseBuffer(b)
	return devices
}

func (ds *DevicesType) parseBuffer(b bytes.Buffer) {
	inputStr := b.String()
	inputStr = strings.ReplaceAll(inputStr, "\n", "")
	inputStr = strings.ReplaceAll(inputStr, "\t", "")
	inputStr = strings.ReplaceAll(inputStr, "\r", "")
	for i := 1; i >= 0; i = strings.Index(inputStr, "  ") {
		inputStr = strings.ReplaceAll(inputStr, "  ", " ")
	}
	arr := strings.Split(inputStr, " ")
	// saveStrToFile(arr)
	d := DeviceType{}
	for index := 0; index < len(arr)-1; index++ {
		switch {
		case arr[index] == "Flags:":
			continue
		case strings.Contains(arr[index], "address-lists="):
			d.addressLists = parseParamertToStr(arr[index])
		case strings.Contains(arr[index], "server="):
			d.server = parseParamertToStr(arr[index])
		case strings.Contains(arr[index], "dhcp-option="):
			d.dhcpOption = parseParamertToStr(arr[index])
		case strings.Contains(arr[index], "status="):
			d.status = parseParamertToStr(arr[index])
		case strings.Contains(arr[index], "expires-after="):
			d.expiresAfter = parseParamertToStr(arr[index])
		case strings.Contains(arr[index], "last-seen="):
			d.lastSeen = parseParamertToStr(arr[index])
		case strings.Contains(arr[index], "active-address="):
			d.activeAddress = parseParamertToStr(arr[index])
		case strings.Contains(arr[index], "active-mac-address="):
			d.activeMacAddress = parseParamertToStr(arr[index])
		case strings.Contains(arr[index], "mac-address="):
			d.macAddress = parseParamertToStr(arr[index])
		case strings.Contains(arr[index], "address="):
			d.address = parseParamertToStr(arr[index])
		case strings.Contains(arr[index], "mac-address="):
			d.activeMacAddress = parseParamertToStr(arr[index])
		case strings.Contains(arr[index], "active-client-id="):
			d.activeClientId = parseParamertToStr(arr[index])
		case strings.Contains(arr[index], "client-id="):
			d.clientId = parseParamertToStr(arr[index])
		case strings.Contains(arr[index], "active-server="):
			d.activeServer = parseParamertToStr(arr[index])
		case strings.Contains(arr[index], "host-name="):
			d.hostName = parseParamertToStr(arr[index])
		case strings.Contains(arr[index], "radius="):
			d.radius = parseParamertToStr(arr[index])
		case arr[index] == "X":
			d.disabled = "yes"
		case arr[index] == "B":
			d.blocked = "yes"
		case arr[index] == "D":
			d.dynamic = "yes"
		case arr[index] == ";;;":
			d.comment = strings.Join(arr[index+1:], " ")
			return
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

func (ds *DevicesType) getInfoD(alias string, quota QuotaType) InfoOfDeviceType {
	for _, d := range *ds {
		if d.activeAddress == alias || d.activeMacAddress == alias || d.address == alias || d.macAddress == alias {
			ifoD := d.convertToInfo()
			ifoD.QuotaType = checkNULLQuotas(ifoD.QuotaType, quota)
			return ifoD
		}
	}
	return InfoOfDeviceType{}
}

func (ds *DevicesType) updateInfo(deviceNew DeviceType) error {
	index := ds.findIndexOfDevice(&deviceNew)
	if index == -1 {
		// p := *ds
		// p = append(p, deviceNew)
		// ds = &p
		*ds = append(*ds, deviceNew)
	}
	// p := *ds
	// p[index] = deviceNew
	(*ds)[index] = deviceNew
	return nil
}

func (d DeviceType) send() error {
	// TODO доделать функцию
	return nil
}
