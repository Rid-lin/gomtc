package main

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/go-routeros/routeros"
	log "github.com/sirupsen/logrus"
)

func dial(cfg *Config) (*routeros.Client, error) {
	if cfg.UseTLS {
		return routeros.DialTLS(cfg.MTAddr, cfg.MTUser, cfg.MTPass, nil)
	}
	return routeros.Dial(cfg.MTAddr, cfg.MTUser, cfg.MTPass)
}

func tryingToReconnectToMokrotik(cfg *Config) *routeros.Client {
	c, err := dial(cfg)
	if err != nil {
		if cfg.NumOfTryingConnectToMT == 0 {
			log.Fatalf("Error connect to %v:%v", cfg.MTAddr, err)
		} else if cfg.NumOfTryingConnectToMT < 0 {
			cfg.NumOfTryingConnectToMT = -1
		}
		log.Errorf("Error connect to %v:%v", cfg.MTAddr, err)
		time.Sleep(15 * time.Second)
		c = tryingToReconnectToMokrotik(cfg)
		cfg.NumOfTryingConnectToMT--
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
		response.DeviceType = ipStruct.DeviceType
		response.Comment = ipStruct.Comments
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
	defer func() {
		if e := recover(); e != nil {
			log.Errorf("Error while trying to get data from the router:%v", e)
		}
	}()
	for {
		data.updateDataFromMT()

		interval, err := time.ParseDuration(cfg.Interval)
		if err != nil {
			interval = 1 * time.Minute
		}
		time.Sleep(interval)

	}
}

func (data *Transport) updateDataFromMT() {
	data.Lock()
	data.infoOfDevices = getDataFromMT(data.QuotaType, data.clientROS)
	data.lastUpdatedMT = time.Now()
	log.Tracef("Info Of Devices updated from MT")
	data.Unlock()
}

func getDataFromMT(quota QuotaType, connRos *routeros.Client) map[string]InfoOfDeviceType {

	quotahourly := quota.HourlyQuota
	quotadaily := quota.DailyQuota
	quotamonthly := quota.MonthlyQuota

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
		// lineOfData.timeoutInt = time.Now().Add(1 * time.Minute).Unix()
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
		// lineOfData.timeout = re.Map["expires-after"]
		lineOfData.HostName = re.Map["host-name"]
		lineOfData.Comments = re.Map["comment"]
		lineOfData.HourlyQuota, lineOfData.DailyQuota, lineOfData.MonthlyQuota, lineOfData.Name, lineOfData.Position, lineOfData.Company, lineOfData.TypeD, lineOfData.IDUser = parseComments(lineOfData.Comments)
		lineOfData.HourlyQuota = checkNULLQuota(lineOfData.HourlyQuota, quotahourly)
		lineOfData.DailyQuota = checkNULLQuota(lineOfData.DailyQuota, quotadaily)
		lineOfData.MonthlyQuota = checkNULLQuota(lineOfData.MonthlyQuota, quotamonthly)
		disable := re.Map["disabled"]
		if disable == "yes" {
			lineOfData.Blocked = true
		} else {
			lineOfData.Blocked = false
		}
		lineOfData.Groups = re.Map["address-lists"]
		lineOfData.AddressLists = strings.Split(lineOfData.Groups, ",")

		ipToMac[lineOfData.IP] = lineOfData

	}
	return ipToMac
}

func parseComments(comment string) (
	quotahourly, quotadaily, quotamonthly uint64,
	name, position, company, typeD, IDUser string) {
	commentArray := strings.Split(comment, "/")
	for _, value := range commentArray {
		switch {
		case strings.Contains(value, "tel"):
			typeD = "tel"
			name = parseParamertToStr(value)
		case strings.Contains(value, "nb"):
			typeD = "nb"
			name = parseParamertToStr(value)
		case strings.Contains(value, "ws"):
			typeD = "ws"
			name = parseParamertToStr(value)
		case strings.Contains(value, "srv"):
			typeD = "srv"
			name = parseParamertToStr(value)
		case strings.Contains(value, "prn"):
			typeD = "prn"
			name = parseParamertToStr(value)
		case strings.Contains(value, "name"):
			typeD = "other"
			name = parseParamertToStr(value)
		case strings.Contains(value, "col"):
			position = parseParamertToStr(value)
		case strings.Contains(value, "com"):
			company = parseParamertToStr(value)
		case strings.Contains(value, "id"):
			IDUser = parseParamertToStr(value)
		case strings.Contains(value, "quotahourly"):
			quotahourly = parseParamertToUint(value)
		case strings.Contains(value, "quotadaily"):
			quotadaily = parseParamertToUint(value)
		case strings.Contains(value, "quotamonthly"):
			quotamonthly = parseParamertToUint(value)
		}
	}
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

func (transport *Transport) syncStatusDevices(inputSync map[string]bool) {
	result := map[string]bool{}
	infoOfDevices := getDataFromMT(transport.QuotaType, transport.clientROS)
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
func (data *Transport) getInfoOfDeviceFromMT(alias string) InfoOfDeviceType {
	var device InfoOfDeviceType
	var entity string

	data.RLock()
	quotahourly := data.HourlyQuota
	quotadaily := data.DailyQuota
	quotamonthly := data.MonthlyQuota
	data.RUnlock()

	if isMac(alias) {
		entity = "mac-address"
	} else if isIP(alias) {
		entity = "active-address"
	}

	// /ip dhcp-server lease get [/ip dhcp-server lease find mac-address="E8:D8:D1:47:55:93"]

	reply, err := data.clientROS.Run("/ip/dhcp-server/lease/print", "?"+entity+"="+alias)
	if err != nil {
		log.Errorf("Error to get Info from MT( alias:'%v'):%v", alias, err)
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
		device.HourlyQuota, device.DailyQuota, device.MonthlyQuota, device.Name, device.Position, device.Company, device.TypeD, device.IDUser = parseComments(device.Comments)
		device.HourlyQuota = checkNULLQuota(device.HourlyQuota, quotahourly)
		device.DailyQuota = checkNULLQuota(device.DailyQuota, quotadaily)
		device.MonthlyQuota = checkNULLQuota(device.MonthlyQuota, quotamonthly)
		disable := re.Map["disabled"]
		if disable == "yes" {
			device.Blocked = true
		} else {
			device.Blocked = false
		}
		device.Groups = re.Map["address-lists"]
		device.AddressLists = strings.Split(device.Groups, ",")
	}
	log.Tracef("Response from Mikrotik(numbers):%v(%v::%v)", reply, device, reply)

	return device
}

func (data *Transport) aliasToDevice(alias string) InfoOfDeviceType {
	device, err := data.findInfoOfDevice(alias)
	if err != nil {
		device = data.getInfoOfDeviceFromMT(alias)
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
		DateStr: time.Now().In(cfg.Location).Format(cfg.dateLayout),
	}
	device := InfoOfDeviceType{}
	var err error

	transport.RLock()
	defer transport.RUnlock()

	if isMac(alias) {
		lineOfReportCashe, ok := transport.dataCashe[key]
		if ok {
			device = InfoOfDeviceType{
				DeviceType: lineOfReportCashe.DeviceType,
				QuotaType:  lineOfReportCashe.QuotaType,
				PersonType: lineOfReportCashe.PersonType,
			}
		} else {
			lineOfReport, ok := transport.data[key]
			if ok {
				device = InfoOfDeviceType{
					DeviceType: lineOfReport.DeviceType,
					QuotaType:  lineOfReport.QuotaType,
					PersonType: lineOfReport.PersonType,
				}
			} else {
				for _, value := range transport.dataCashe {
					if value.Mac != alias {
						continue
					}
					device = InfoOfDeviceType{
						DeviceType: value.DeviceType,
						QuotaType:  value.QuotaType,
						PersonType: value.PersonType,
					}
				}
				for _, value := range transport.data {
					if value.Mac != alias {
						continue
					}
					device = InfoOfDeviceType{
						DeviceType: value.DeviceType,
						QuotaType:  value.QuotaType,
						PersonType: value.PersonType,
					}
				}
				for _, value := range transport.infoOfDevices {
					if value.Mac != alias {
						continue
					}
					device = InfoOfDeviceType{
						DeviceType: value.DeviceType,
						QuotaType:  value.QuotaType,
						PersonType: value.PersonType,
					}
				}

			}

		}
	} else if isIP(alias) {
		lineOfInfo, ok := transport.infoOfDevices[alias]
		if ok {
			device = InfoOfDeviceType{
				DeviceType: lineOfInfo.DeviceType,
				QuotaType:  lineOfInfo.QuotaType,
				PersonType: lineOfInfo.PersonType,
			}
		} else {
			for _, value := range transport.infoOfDevices {
				if value.IP != alias {
					continue
				}
				device = InfoOfDeviceType{
					DeviceType: value.DeviceType,
					QuotaType:  value.QuotaType,
					PersonType: value.PersonType,
				}
			}
			for _, value := range transport.data {
				if value.IP != alias {
					continue
				}
				device = InfoOfDeviceType{
					DeviceType: value.DeviceType,
					QuotaType:  value.QuotaType,
					PersonType: value.PersonType,
				}
			}
			for _, value := range transport.infoOfDevices {
				if value.IP != alias {
					continue
				}
				device = InfoOfDeviceType{
					DeviceType: value.DeviceType,
					QuotaType:  value.QuotaType,
					PersonType: value.PersonType,
				}
			}

		}
	} else {
		err = fmt.Errorf("Alias '%v' not found", alias)
		log.Debugf("Alias '%v' not found", alias)
	}

	return device, err
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
					buf := make([]byte, 4096)
					rlen, remote, err := transport.conn.ReadFromUDP(buf)

					if err != nil {
						log.Errorf("Error: %v\n", err)
					} else {

						stream := bytes.NewBuffer(buf[:rlen])

						go handlePacket(stream, remote, transport.outputChannel, cfg)
					}
				}
			}
		}

	}

}
