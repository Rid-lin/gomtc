package main

import (
	"bytes"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"gopkg.in/routeros.v2"
)

func dial(MTAddr, MTUser, MTPass string, UseTLS bool) (*routeros.Client, error) {
	if UseTLS {
		return routeros.DialTLS(MTAddr, MTUser, MTPass, nil)
	}
	return routeros.Dial(MTAddr, MTUser, MTPass)
}

func tryingToReconnectToMokrotik(MTAddr, MTUser, MTPass string, UseTLS bool, NumOfTryingConnectToMT int) *routeros.Client {
	c, err := dial(MTAddr, MTUser, MTPass, UseTLS)
	if err != nil {
		if NumOfTryingConnectToMT == 0 {
			log.Fatalf("Error connect to %v:%v", MTAddr, err)
		} else if NumOfTryingConnectToMT < 0 {
			cfg.NumOfTryingConnectToMT = -1
		}
		log.Errorf("Error connect to %v:%v", MTAddr, err)
		time.Sleep(15 * time.Second)
		NumOfTryingConnectToMT--
		c = tryingToReconnectToMokrotik(MTAddr, MTUser, MTPass, UseTLS, NumOfTryingConnectToMT)
	}
	return c
}

func (data *Transport) GetInfo(request *request) ResponseType {
	var response ResponseType

	data.RLock()
	ipStruct, ok := data.infoOfDevices[request.IP]
	data.RUnlock()
	if ok {
		log.Tracef("IP:%v to MAC:%v, hostname:%v, comment:%v", ipStruct.IP, ipStruct.Mac, ipStruct.HostName, ipStruct.Comments)
		response.DeviceOldType = ipStruct.DeviceOldType
		response.Comments = ipStruct.Comments
	} else {
		log.Tracef("IP:'%v' not find in table lease of router:'%v'", ipStruct.IP, cfg.MTAddr)
		response.Mac = request.IP
		response.IP = request.IP
	}
	if response.Mac == "" {
		response.Mac = request.IP
		log.Error("Mac of Device = '' O_o", request)
	}

	return response
}

func (data *Transport) loopGetDataFromMT() {
	// defer func() {
	// 	if e := recover(); e != nil {
	// 		log.Errorf("Error while trying to get data from the router:%v", e)
	// 		data.exitChan <- os.Kill
	// 		fmt.Printf("\n")
	// 		panic(e)
	// 	}
	// }()
	for {
		data.updateDataFromMT()

		if err := data.getStatusallDevices(); err == nil {
			// if err := transport.getStatusDevices(cfg); err == nil {
			data.checkQuota()
			// transport.updateStatusDevicesToMT(cfg)
		}

		interval, err := time.ParseDuration(cfg.Interval)
		if err != nil {
			interval = 1 * time.Minute
		}
		time.Sleep(interval)

	}
}

func (data *Transport) updateDataFromMT() {
	data.Lock()
	BlockAddressList := data.BlockAddressList
	data.infoOfDevices = getDataFromMT(data.QuotaType, data.clientROS, BlockAddressList)
	data.lastUpdatedMT = time.Now()
	data.Unlock()
	log.Tracef("Info Of Devices updated from MT")
}

func getDataFromMT(quotaDefault QuotaType, connRos *routeros.Client, BlockAddressList string) map[string]InfoOfDeviceType {

	quotahourly := quotaDefault.HourlyQuota
	quotadaily := quotaDefault.DailyQuota
	quotamonthly := quotaDefault.MonthlyQuota

	lineOfData := InfoOfDeviceType{}
	ipToMac := map[string]InfoOfDeviceType{}
	reply, err := connRos.Run("/ip/arp/print")
	if err != nil {
		log.Error(err)
	}
	for _, re := range reply.Re {
		lineOfData.IP = re.Map["address"]
		lineOfData.Mac = re.Map["mac-address"]
		lineOfData.HourlyQuota = checkNULLQuota(lineOfData.HourlyQuota, quotahourly)
		lineOfData.DailyQuota = checkNULLQuota(lineOfData.DailyQuota, quotadaily)
		lineOfData.MonthlyQuota = checkNULLQuota(lineOfData.MonthlyQuota, quotamonthly)
		lineOfData.timeout = time.Now()
		ipToMac[lineOfData.IP] = lineOfData
	}

	reply2, err2 := connRos.Run("/ip/dhcp-server/lease/print") //, "?status=bound") //, "?disabled=false")
	if err2 != nil {
		log.Error(err2)
	}
	for _, re := range reply2.Re {
		lineOfData.Id = re.Map[".id"]
		lineOfData.IP = re.Map["active-address"]
		lineOfData.Mac = re.Map["mac-address"]
		lineOfData.AMac = re.Map["active-mac-address"]
		lineOfData.HostName = re.Map["host-name"]
		lineOfData.Comments = re.Map["comment"]
		lineOfData.HourlyQuota, lineOfData.DailyQuota, lineOfData.MonthlyQuota, lineOfData.Name, lineOfData.Position, lineOfData.Company, lineOfData.TypeD, lineOfData.IDUser, lineOfData.Comment, lineOfData.Manual = parseComments(lineOfData.Comments)
		lineOfData.HourlyQuota = checkNULLQuota(lineOfData.HourlyQuota, quotahourly)
		lineOfData.DailyQuota = checkNULLQuota(lineOfData.DailyQuota, quotadaily)
		lineOfData.MonthlyQuota = checkNULLQuota(lineOfData.MonthlyQuota, quotamonthly)
		lineOfData.Disabled = paramertToBool(re.Map["disabled"])
		lineOfData.Groups = re.Map["address-lists"]
		if BlockAddressList != "" {
			lineOfData.Blocked = strings.Contains(lineOfData.Groups, BlockAddressList)
		}
		lineOfData.timeout = time.Now()

		// lineOfData.AddressLists = strings.Split(lineOfData.Groups, ",")

		ipToMac[lineOfData.IP] = lineOfData

	}
	return ipToMac
}

func getLeasesOverSSHfMT(SSHCred SSHCredetinals) []DeviceType {
	// device := DeviceType{}
	devices := []DeviceType{}

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

	command := "/ip/arp/print"
	// тут происходит запуск uname -a на удалённом сервере
	b, err := session.CombinedOutput(command)
	if err != nil {
		panic(err)
	}
	// выводим результат
	fmt.Print(string(b))

	// for _, re := range reply.Re {
	// 	device.IP = re.Map["address"]
	// 	device.Mac = re.Map["mac-address"]
	// 	device.timeout = time.Now()
	// 	devices = append(devices, device)
	// }

	// reply2, err2 := connRos.Run("/ip/dhcp-server/lease/print") //, "?status=bound") //, "?disabled=false")
	// if err2 != nil {
	// 	log.Error(err2)
	// }
	// for _, re := range reply2.Re {
	// 	device.Id = re.Map[".id"]
	// 	device.IP = re.Map["active-address"]
	// 	device.Mac = re.Map["mac-address"]
	// 	device.AMac = re.Map["active-mac-address"]
	// 	device.HostName = re.Map["host-name"]
	// 	device.Comments = re.Map["comment"]
	// 	device.Disabled = paramertToBool(re.Map["disabled"])
	// 	device.Groups = re.Map["address-lists"]
	// 	device.timeout = time.Now()

	// 	devices = append(devices, device)
	// }
	return devices
}

// func getInfofMT(quotaDefault QuotaType, SSHCred SSHCredetinals, BlockAddressList string) map[string]InfoOfDeviceType {

// 	quotahourly := quotaDefault.HourlyQuota
// 	quotadaily := quotaDefault.DailyQuota
// 	quotamonthly := quotaDefault.MonthlyQuota

// 	lineOfData := InfoOfDeviceType{}
// 	ipToMac := map[string]InfoOfDeviceType{}
// 	reply, err := connRos.Run("/ip/arp/print")
// 	if err != nil {
// 		log.Error(err)
// 	}
// 	for _, re := range reply.Re {
// 		lineOfData.IP = re.Map["address"]
// 		lineOfData.Mac = re.Map["mac-address"]
// 		lineOfData.HourlyQuota = checkNULLQuota(lineOfData.HourlyQuota, quotahourly)
// 		lineOfData.DailyQuota = checkNULLQuota(lineOfData.DailyQuota, quotadaily)
// 		lineOfData.MonthlyQuota = checkNULLQuota(lineOfData.MonthlyQuota, quotamonthly)
// 		lineOfData.timeout = time.Now()
// 		ipToMac[lineOfData.IP] = lineOfData
// 	}

// 	reply2, err2 := connRos.Run("/ip/dhcp-server/lease/print") //, "?status=bound") //, "?disabled=false")
// 	if err2 != nil {
// 		log.Error(err2)
// 	}
// 	for _, re := range reply2.Re {
// 		lineOfData.Id = re.Map[".id"]
// 		lineOfData.IP = re.Map["active-address"]
// 		lineOfData.Mac = re.Map["mac-address"]
// 		lineOfData.AMac = re.Map["active-mac-address"]
// 		lineOfData.HostName = re.Map["host-name"]
// 		lineOfData.Comments = re.Map["comment"]
// 		lineOfData.HourlyQuota, lineOfData.DailyQuota, lineOfData.MonthlyQuota, lineOfData.Name, lineOfData.Position, lineOfData.Company, lineOfData.TypeD, lineOfData.IDUser, lineOfData.Comment, lineOfData.Manual = parseComments(lineOfData.Comments)
// 		lineOfData.HourlyQuota = checkNULLQuota(lineOfData.HourlyQuota, quotahourly)
// 		lineOfData.DailyQuota = checkNULLQuota(lineOfData.DailyQuota, quotadaily)
// 		lineOfData.MonthlyQuota = checkNULLQuota(lineOfData.MonthlyQuota, quotamonthly)
// 		lineOfData.Disabled = paramertToBool(re.Map["disabled"])
// 		lineOfData.Groups = re.Map["address-lists"]
// 		if BlockAddressList != "" {
// 			lineOfData.Blocked = strings.Contains(lineOfData.Groups, BlockAddressList)
// 		}
// 		lineOfData.timeout = time.Now()

// 		// lineOfData.AddressLists = strings.Split(lineOfData.Groups, ",")

// 		ipToMac[lineOfData.IP] = lineOfData

// 	}
// 	return ipToMac
// }

func parseComments(comment string) (
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
		case strings.Contains(value, "name="):
			typeD = "other"
			name = parseParamertToStr(value)
		case strings.Contains(value, "col="):
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

func parseParamertToStr(inpuStr string) string {
	Arr := strings.Split(inpuStr, "=")
	if len(Arr) > 1 {
		return Arr[1]
	} else {
		log.Errorf("Parameter error. The parameter is specified incorrectly or not specified at all.(%v)", inpuStr)
	}
	return ""
}

func parseParamertToUint(inputValue string) (quota uint64) {
	var err error
	Arr := strings.Split(inputValue, "=")
	if len(Arr) > 1 {
		quotaStr := Arr[1]
		quota, err = strconv.ParseUint(quotaStr, 10, 64)
		if err != nil {
			log.Errorf("Error parse quota from:(%v) with:(%v)", quotaStr, err)
		}
		return
	} else {
		log.Errorf("Parameter error. The parameter is specified incorrectly or not specified at all.(%v)", inputValue)
	}
	return
}

func parseParamertToBool(inputValue string) bool {
	if strings.Contains(inputValue, "true") || strings.Contains(inputValue, "yes") {
		return true
	}
	return false
}

func paramertToUint(inputValue string) (quota uint64) {
	quota, err := strconv.ParseUint(inputValue, 10, 64)
	if err != nil {
		log.Errorf("Error parse quota from:(%v) with:(%v)", inputValue, err)
	}
	return
}

func paramertToBool(inputValue string) bool {
	if inputValue == "true" || inputValue == "yes" {
		return true
	}
	return false
}

func (transport *Transport) syncStatusDevices(inputSync map[string]bool) {
	result := map[string]bool{}
	transport.RLock()
	BlockAddressList := transport.BlockAddressList
	transport.RUnlock()
	infoOfDevices := getDataFromMT(transport.QuotaType, transport.clientROS, BlockAddressList)
	transport.Lock()
	transport.infoOfDevices = infoOfDevices
	transport.Unlock()

	for _, value := range infoOfDevices {
		for keySync := range inputSync {
			if keySync == value.IP || keySync == value.Mac || keySync == value.HostName {
				result[value.Id] = inputSync[keySync]
			}
		}
	}

	for key := range result {
		err := transport.setStatusDevice(key, result[key])
		if err != nil {
			log.Error(err)
		}
	}
}

func (data *Transport) setDevice(d InfoOfDeviceType) error {

	data.RLock()
	Quota := data.QuotaType
	data.RUnlock()
	var disabled string
	if d.Disabled {
		disabled = "yes"
	} else {
		disabled = "no"
	}
	// .id=*e6ff8;active-address=192.168.65.85;active-client-id=1:e8:d8:d1:47:55:93;active-mac-address=E8:D8:D1:47:55:93;active-server=dhcp_lan;address=pool_admin;address-lists=inet_over_vpn;blocked=false;client-id=1:e8:d8:d1:47:55:93;comment=nb=Admin/quotahourly=500000000/quotadaily=50000000000;dhcp-option=;disabled=false;dynamic=false;expires-after=00:07:53;host-name=root-hp;last-seen=2m7s;mac-address=E8:D8:D1:47:55:93;radius=false;server=dhcp_lan;status=boun

	comment := "/"

	if d.TypeD != "" && d.Name != "" {
		comment = fmt.Sprintf("%v=%v",
			d.TypeD,
			d.Name)
	} else if d.TypeD == "" && d.Name != "" {
		comment = fmt.Sprintf("name=%v",
			d.Name)
	}
	if d.Company != "" {
		comment = fmt.Sprintf("%v/com=%v",
			comment,
			d.Company)
	}
	if d.Position != "" {
		comment = fmt.Sprintf("%v/col=%v",
			comment,
			d.Position)
	}
	if d.HourlyQuota != 0 && d.HourlyQuota != Quota.HourlyQuota {
		comment = fmt.Sprintf("%v/quotahourly=%v",
			comment,
			d.HourlyQuota)
	}
	if d.DailyQuota != 0 && d.DailyQuota != Quota.DailyQuota {
		comment = fmt.Sprintf("%v/quotadaily=%v",
			comment,
			d.DailyQuota)
	}
	if d.MonthlyQuota != 0 && d.MonthlyQuota != Quota.MonthlyQuota {
		comment = fmt.Sprintf("%v/quotamonthly=%v",
			comment,
			d.MonthlyQuota)
	}
	if d.Manual {
		comment = fmt.Sprintf("%v/manual=%v",
			comment,
			d.Manual)
	}
	if d.Comment != "" {
		comment = fmt.Sprintf("%v/comment=%v",
			comment,
			d.Comment)
	}
	comment = strings.ReplaceAll(comment, "//", "/")
	if comment == "/" {
		comment = ""
	}

	reply, err := data.clientROS.RunArgs([]string{
		"/ip/dhcp-server/lease/set",
		"=disabled=" + disabled,
		"=numbers=" + d.Id,
		"=address-lists=" + d.Groups,
		"=comment=" + comment})
	// reply, err := data.clientROS.Run("/ip/dhcp-server/lease/set", "=disabled="+disabled, "=numbers="+d.Id, "=comment="+comment)
	if err != nil {
		return err
	} else if reply.Done.Word != "!done" {
		return fmt.Errorf("%v", reply.Done.Word)
	}
	log.Tracef("Response from Mikrotik(numbers):%v(%v)", reply, d.Id)
	return nil
}

func (data *Transport) setGroupOfDeviceToMT(d InfoOfDeviceType) error {

	if d.Id == "" {
		return fmt.Errorf("Device ID is empty")
	}

	argsForMT := []string{
		"/ip/dhcp-server/lease/set",
		"=numbers=" + d.Id,
		"=address-lists=" + d.Groups,
	}
	log.Debug(argsForMT)

	reply, err := data.clientROS.RunArgs(argsForMT)
	// reply, err := data.clientROS.RunArgs([]string{
	// 	"/ip/dhcp-server/lease/set",
	// 	"=numbers=" + d.Id,
	// 	"=address-lists=" + d.Groups,
	// })
	if err != nil {
		return err
	} else if reply.Done.Word != "!done" {
		return fmt.Errorf("%v", reply.Done.Word)
	}
	log.Tracef("Response from Mikrotik(numbers):%v(%v)", reply, d.Id)
	return nil
}

func (data *Transport) setStatusDevice(number string, status bool) error {

	var statusMtT string
	if status {
		statusMtT = "yes"
	} else {
		statusMtT = "no"
	}

	reply, err := data.clientROS.Run("/ip/dhcp-server/lease/set", "=disabled="+statusMtT, "=numbers="+number)
	if err != nil {
		return err
	} else if reply.Done.Word != "!done" {
		return fmt.Errorf("%v", reply.Done.Word)
	}
	log.Tracef("Response from Mikrotik(numbers):%v(%v)", reply, number)
	return nil
}

func (data *Transport) obtainingInformationFromMTboutOneDevice(alias string) InfoOfDeviceType {
	defer func() {
		if e := recover(); e != nil {
			runtime.Breakpoint()
		}
	}()

	var device InfoOfDeviceType
	var entity string

	data.RLock()
	BlockAddressList := data.BlockAddressList
	DefaultHourlyQuota := data.HourlyQuota
	DefaultDailyQuota := data.DailyQuota
	DefaultMonthlyQuota := data.MonthlyQuota

	data.RUnlock()

	if isMac(alias) {
		entity = "mac-address"
	} else if isIP(alias) {
		entity = "active-address"
	}
	alias = strings.Trim(alias, "]")
	alias = strings.Trim(alias, "[")
	runArgs := []string{
		"/ip/dhcp-server/lease/print",
		"?" + entity + "=" + alias,
	}
	reply, err := data.clientROS.RunArgs(runArgs)
	if err != nil {
		log.Errorf("Error to get Info from MT(alias:'%v', runArgs:'%v'):%v", alias, runArgs, err)
	} else if reply.Done.Word != "!done" {
		log.Errorf("%v", reply.Done.Word)
	}
	if len(reply.Re) > 0 {
		re := reply.Re[0]
		device.Id = re.Map[".id"]
		device.IP = re.Map["active-address"]
		device.Mac = re.Map["mac-address"]
		device.AMac = re.Map["active-mac-address"]
		device.HostName = re.Map["host-name"]
		device.Comments = re.Map["comment"]
		device.HourlyQuota, device.DailyQuota, device.MonthlyQuota, device.Name, device.Position, device.Company, device.TypeD, device.IDUser, device.Comment, device.Manual = parseComments(device.Comments)
		device.ShouldBeBlocked = paramertToBool(re.Map["disabled"])
		device.Groups = re.Map["address-lists"]
		if BlockAddressList != "" {
			device.Blocked = strings.Contains(device.Groups, BlockAddressList)
		}
		device.HourlyQuota = checkNULLQuota(device.HourlyQuota, DefaultHourlyQuota)
		device.DailyQuota = checkNULLQuota(device.DailyQuota, DefaultDailyQuota)
		device.MonthlyQuota = checkNULLQuota(device.MonthlyQuota, DefaultMonthlyQuota)
	}
	log.Tracef("Response from Mikrotik(numbers)(%v) %v", reply, device)

	return device
}

func (data *Transport) aliasToDevice(alias string) InfoOfDeviceType {
	device, err := data.findInfoOfDevice(alias)
	if err != nil {
		device = data.obtainingInformationFromMTboutOneDevice(alias)
	}

	return device
}

func isMac(inputStr string) bool {
	arr := strings.Split(inputStr, ":")
	return len(arr) == 6
}

func isIP(inputStr string) bool {
	arr := strings.Split(inputStr, ".")
	return len(arr) == 4
}

func (transport *Transport) findInfoOfDevice(alias string) (InfoOfDeviceType, error) {
	key := KeyMapOfReports{
		Alias:   alias,
		DateStr: time.Now().In(transport.Location).Format("2006-01-02"),
	}
	device := InfoOfDeviceType{}
	var err error

	transport.RLock()
	defer transport.RUnlock()

	if isMac(alias) {
		lineOfReportCashe, ok := transport.dataCashe[key]
		if ok {
			device = InfoOfDeviceType{
				DeviceOldType: lineOfReportCashe.DeviceOldType,
				QuotaType:     lineOfReportCashe.QuotaType,
				PersonType:    lineOfReportCashe.PersonType,
			}
			return device, nil
		} else {
			lineOfReport, ok := transport.data[key]
			if ok {
				device = InfoOfDeviceType{
					DeviceOldType: lineOfReport.DeviceOldType,
					QuotaType:     lineOfReport.QuotaType,
					PersonType:    lineOfReport.PersonType,
				}
				return device, nil
			} else {
				for _, value := range transport.dataCashe {
					if value.Mac == alias {
						device = InfoOfDeviceType{
							DeviceOldType: value.DeviceOldType,
							QuotaType:     value.QuotaType,
							PersonType:    value.PersonType,
						}
						return device, nil
					}
				}
				for _, value := range transport.data {
					if value.Mac == alias {
						device = InfoOfDeviceType{
							DeviceOldType: value.DeviceOldType,
							QuotaType:     value.QuotaType,
							PersonType:    value.PersonType,
						}
						return device, nil
					}
				}
				for _, value := range transport.infoOfDevices {
					if value.Mac != alias {
						device = InfoOfDeviceType{
							DeviceOldType: value.DeviceOldType,
							QuotaType:     value.QuotaType,
							PersonType:    value.PersonType,
						}
						return device, nil
					}
				}
			}

		}
	} else if isIP(alias) {
		lineOfInfo, ok := transport.infoOfDevices[alias]
		if ok {
			device = InfoOfDeviceType{
				DeviceOldType: lineOfInfo.DeviceOldType,
				QuotaType:     lineOfInfo.QuotaType,
				PersonType:    lineOfInfo.PersonType,
			}
			return device, nil
		} else {
			for _, value := range transport.infoOfDevices {
				if value.IP == alias {
					device = InfoOfDeviceType{
						DeviceOldType: value.DeviceOldType,
						QuotaType:     value.QuotaType,
						PersonType:    value.PersonType,
					}
					return device, nil
				}
			}
			for _, value := range transport.data {
				if value.IP == alias {
					device = InfoOfDeviceType{
						DeviceOldType: value.DeviceOldType,
						QuotaType:     value.QuotaType,
						PersonType:    value.PersonType,
					}
					return device, nil
				}
			}
			for _, value := range transport.infoOfDevices {
				if value.IP != alias {
					device = InfoOfDeviceType{
						DeviceOldType: value.DeviceOldType,
						QuotaType:     value.QuotaType,
						PersonType:    value.PersonType,
					}
					return device, nil
				}
			}

		}
	}
	err = fmt.Errorf("Alias '%v' not found", alias)
	log.Debugf("Alias '%v' not found", alias)

	return device, err
}

func (transport *Transport) updateInfoOfDeviceFromMT(alias string) {
	device := transport.obtainingInformationFromMTboutOneDevice(alias)
	key := KeyMapOfReports{
		Alias:   alias,
		DateStr: time.Now().In(transport.Location).Format("2006-01-02"),
	}
	transport.Lock()
	LineOfReports := transport.data[key]
	LineOfReports.QuotaType = device.QuotaType
	LineOfReports.PersonType = device.PersonType
	transport.data[key] = LineOfReports
	LineOfReports = transport.dataCashe[key]
	LineOfReports.QuotaType = device.QuotaType
	LineOfReports.PersonType = device.PersonType
	transport.dataCashe[key] = LineOfReports
	infoOfDevice := transport.infoOfDevices[device.IP]
	infoOfDevice.QuotaType = device.QuotaType
	infoOfDevice.PersonType = device.PersonType
	transport.Unlock()
}

func (transport *Transport) readsStreamFromMT(cfg *Config) {

	addr, err := net.ResolveUDPAddr("udp", cfg.FlowAddr)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	log.Infof("gomtc listen NetFlow on:'%v'", cfg.FlowAddr)
	for {
		transport.conn, err = net.ListenUDP("udp", addr)
		if err != nil {
			log.Errorln(err)
		} else {
			err = transport.conn.SetReadBuffer(cfg.ReceiveBufferSizeBytes)
			if err != nil {
				log.Errorln(err)
			} else {
				/* Infinite-loop for reading packets */
				for {
					select {
					case <-transport.stopReadFromUDP:
						time.Sleep(5 * time.Second)
						return
					default:
						bufUDP := make([]byte, 4096)
						rlen, remote, err := transport.conn.ReadFromUDP(bufUDP)

						if err != nil {
							log.Errorf("Error read from UDP: %v\n", err)
						} else {

							stream := bytes.NewBuffer(bufUDP[:rlen])

							go handlePacket(stream, remote, transport.outputChannel, cfg)
						}
					}
				}
			}
		}

	}

}
