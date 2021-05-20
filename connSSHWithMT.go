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

func getResponseOverSSHfMT(SSHCred SSHCredetinals, commands []string) bytes.Buffer {
	// Add the last command (exit), if not, to reduce the number of commands passed in parameters.
	if len(commands) > 0 {
		lastCommand := commands[len(commands)-1]
		if lastCommand != "exit" {
			commands = append(commands, "exit")
		}
	} else {
		commands = append(commands, "exit")
	}
	sshConfig := &ssh.ClientConfig{
		User: SSHCred.SSHUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(SSHCred.SSHPass),
		},
		// Non-production only
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	// Established connection
	connection, err := ssh.Dial("tcp", SSHCred.SSHHost+":"+SSHCred.SSHPort, sshConfig)
	if err != nil {
		log.Errorf("Failed to dial: %s", err)
	}
	defer connection.Close()
	// crete ssh session
	session, err := connection.NewSession()
	if err != nil {
		log.Errorf("Failed to ceate session: %s", err)
	}
	defer session.Close()
	// StdinPipe for commands
	stdin, err := session.StdinPipe()
	if err != nil {
		log.Errorf("Failed redirect StdinPipe: %s", err)
	}
	// Uncomment to store output in variable
	var b bytes.Buffer
	session.Stdout = &b
	session.Stderr = &b
	// Enable system stdout
	// Comment these if you uncomment to store in variable
	// session.Stdout = os.Stdout
	// session.Stderr = os.Stderr
	// Start remote shell
	err = session.Shell()
	if err != nil {
		log.Errorf("Failed to ceate Shell: %s", err)
	}
	// send the commands
	for _, cmd := range commands {
		_, err = fmt.Fprintf(stdin, "%s\n", cmd)
		if err != nil {
			log.Errorf("Failed to send command(%s): %s", cmd, err)
		}
	}
	// session.Close()
	// Wait for sess to finish
	err = session.Wait()
	if err != nil {
		log.Errorf("Failed to wait session: %s", err)
	}
	// str := b.String()
	// fmt.Print(str)
	return b
}

func parseInfoFromMTToSlice(p parseType) []DeviceType {

	device, device2 := DeviceType{}, DeviceType{}
	devices := []DeviceType{}
	var b bytes.Buffer
	for b.Len() < 1 {
		b = getResponseOverSSHfMT(p.SSHCredetinals, []string{"/ip dhcp-server lease print detail without-paging"})
	}
	inputStr := b.String()
	inputArr := strings.Split(inputStr, "\n")
	for _, line := range inputArr {
		if device.parseLine(line) != nil {
			device2.timeout = time.Now().In(p.Location)
			devices = append(devices, device2)
			// fmt.Printf("1=%#v\n2=%#v\n", device, device2)
			device = DeviceType{}
		}
		device2 = device
	}
	fmt.Print(saveDeviceToCSV(devices))
	return devices
}

func (d *DeviceType) parseLine(l string) (err error) {
	l = strings.Trim(l, " ")
	l = strings.ReplaceAll(l, "  ", " ")
	arr := strings.Split(l, " ")
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
		case isNumDot(s):
			err = fmt.Errorf("New line")
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

func isNumDot(s string) bool {
	if len(s) == 0 {
		return false
	} else if s == `
` {
		return false
	}
	dotFound := false
	for _, v := range s {
		if v == '.' {
			if dotFound {
				return false
			}
			dotFound = true
		} else if v < '0' || v > '9' {
			return false
		}
	}
	return true
}

func saveDeviceToCSV(devices []DeviceType) error {
	f, e := os.Create("./Device.csv")
	if e != nil {
		fmt.Println(e)
	}

	writer := csv.NewWriter(f)

	for _, value := range devices {
		fields := deviceToSlice(value)
		err := writer.Write(fields)
		if err != nil {
			fmt.Println(e)
		}
	}
	return nil
}

func deviceToSlice(d DeviceType) []string {
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

func (ds *DevicesType) getInfo(alias string, quota QuotaType) InfoOfDeviceType {
	InfoOfDevice := InfoOfDeviceType{}
	var (
		ip, mac, name, position, company, typeD, IDUser, comment string
		hourlyQuota, dailyQuota, monthlyQuota                    uint64
		manual                                                   bool
	)
	for _, d := range *ds {
		if d.activeAddress == alias || d.activeMacAddress == alias || d.address == alias || d.macAddress == alias {
			ip = validateIP(d.activeAddress, d.address)
			mac = validateMac(d.activeMacAddress, d.macAddress, d.clientId, d.activeClientId)
			hourlyQuota, dailyQuota, monthlyQuota, name, position, company, typeD, IDUser, comment, manual = parseComments(d.comment)
			InfoOfDevice = InfoOfDeviceType{
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
				},
			}
		}
	}

	return InfoOfDevice
}

func validateIP(ip, altIp string) string {
	if ip != "" {
		return ip
	} else if isIP(altIp) {
		return altIp
	}
	return ""
}

func validateMac(mac, altMac, hopeMac, lastHopeMac string) string {
	switch {
	case mac != "":
		return mac
	case altMac != "":
		return altMac
	case isMac(hopeMac[2:]):
		return hopeMac[2:]
	case isMac(lastHopeMac[2:]):
		return lastHopeMac[2:]
	}
	return ""
}
