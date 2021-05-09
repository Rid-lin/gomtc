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

	timeInt, err := strconv.ParseInt(request.Time, 10, 64)
	if err != nil {
		log.Errorf("Error parsing timeStamp(%v) from request:%v", timeInt, err)
		//With an incorrect time, removes 30 seconds from the current time to be able to identify the IP address
		timeInt = time.Now().Add(-30 * time.Second).Unix()
	}
	request.timeInt = timeInt
	data.RLock()
	ipStruct, ok := data.infoOfDevices[request.IP]
	data.RUnlock()
	if ok && timeInt < ipStruct.timeoutInt {
		log.Tracef("IP:%v to MAC:%v, hostname:%v, comment:%v", ipStruct.IP, ipStruct.Mac, ipStruct.HostName, ipStruct.Comment)
		response.Mac = ipStruct.Mac
		response.IP = ipStruct.IP
		response.Hostname = ipStruct.HostName
		response.Comment = ipStruct.Comment
	} else if ok {
		// TODO remove
		log.Tracef("IP:%v to MAC:%v, hostname:%v, comment:%v", ipStruct.IP, ipStruct.Mac, ipStruct.HostName, ipStruct.Comment)
		response.Mac = ipStruct.Mac
		response.IP = ipStruct.IP
		response.Hostname = ipStruct.HostName
		response.Comment = ipStruct.Comment
	} else if !ok {
		// TODO Make information about the mac-address loaded from the router
		log.Tracef("IP:'%v' not find in table lease of router:'%v'", ipStruct.IP, cfg.MTAddr)
		response.Mac = request.IP
		response.IP = request.IP
	} else {
		log.Tracef("IP:'%v' not find in table lease of router:'%v'", ipStruct.IP, cfg.MTAddr)
		response.Mac = request.IP
		response.IP = request.IP
	}
	if response.Mac == "" {
		response.Mac = request.IP
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

		data.Lock()
		data.infoOfDevices = getDataFromMT(data.QuotaType, data.clientROS)
		data.Unlock()

		// ipToMac := data.getDataFromMT()
		// data.Lock()
		// data.ipToMac = ipToMac
		// data.Unlock()
		// ipToMac = map[string]LineOfData{}

		interval, err := time.ParseDuration(cfg.Interval)
		if err != nil {
			interval = 10 * time.Minute
		}
		time.Sleep(interval)

	}
}

func getDataFromMT(quota QuotaType, connRos *routeros.Client) map[string]LineOfData {

	quotahourly := quota.HourlyQuota
	quotadaily := quota.DailyQuota
	quotamonthly := quota.MonthlyQuota

	lineOfData := LineOfData{}
	ipToMac := map[string]LineOfData{}
	reply, err := connRos.Run("/ip/arp/print")
	if err != nil {
		log.Error(err)
	}
	for _, re := range reply.Re {
		lineOfData.IP = re.Map["address"]
		lineOfData.Mac = re.Map["mac-address"]
		if lineOfData.HourlyQuota == 0 {
			lineOfData.HourlyQuota = quotahourly
		}
		if lineOfData.DailyQuota == 0 {
			lineOfData.DailyQuota = quotadaily
		}
		if lineOfData.MonthlyQuota == 0 {
			lineOfData.MonthlyQuota = quotamonthly
		}
		lineOfData.timeoutInt = time.Now().Add(1 * time.Minute).Unix()
		ipToMac[lineOfData.IP] = lineOfData
	}
	reply2, err2 := connRos.Run("/ip/dhcp-server/lease/print") //, "?status=bound") //, "?disabled=false")
	if err2 != nil {
		log.Error(err2)
	}
	for _, re := range reply2.Re {
		lineOfData.Id = re.Map[".id"]
		lineOfData.IP = re.Map["active-address"]
		lineOfData.Mac = re.Map["active-mac-address"]
		lineOfData.timeout = re.Map["expires-after"]
		lineOfData.HostName = re.Map["host-name"]
		lineOfData.Comment = re.Map["comment"]
		lineOfData.HourlyQuota, lineOfData.DailyQuota, lineOfData.MonthlyQuota, lineOfData.Name, lineOfData.Position, lineOfData.Company, lineOfData.TypeD = parseComments(lineOfData.Comment)
		if lineOfData.HourlyQuota == 0 {
			lineOfData.HourlyQuota = quotahourly
		}
		if lineOfData.DailyQuota == 0 {
			lineOfData.DailyQuota = quotadaily
		}
		if lineOfData.MonthlyQuota == 0 {
			lineOfData.MonthlyQuota = quotamonthly
		}
		lineOfData.disable = re.Map["disabled"]
		addressLists := re.Map["address-lists"]
		lineOfData.addressLists = strings.Split(addressLists, ",")

		//Calculating the time when the lease of the address ends
		timeStr, err := time.ParseDuration(lineOfData.timeout)
		if err != nil {
			timeStr = 10 * time.Second
		}
		// Writes to a variable for further quick comparison
		lineOfData.timeoutInt = time.Now().Add(timeStr).Unix()

		ipToMac[lineOfData.IP] = lineOfData

	}
	return ipToMac
}

// func (data *Transport) getDataFromMT() map[string]LineOfData {

// 	quotahourly := data.HourlyQuota
// 	quotadaily := data.DailyQuota
// 	quotamonthly := data.MonthlyQuota

// 	lineOfData := LineOfData{}
// 	ipToMac := map[string]LineOfData{}
// 	reply, err := data.clientROS.Run("/ip/arp/print")
// 	if err != nil {
// 		log.Error(err)
// 	}
// 	for _, re := range reply.Re {
// 		lineOfData.IP = re.Map["address"]
// 		lineOfData.Mac = re.Map["mac-address"]
// 		if lineOfData.HourlyQuota == 0 {
// 			lineOfData.HourlyQuota = quotahourly
// 		}
// 		if lineOfData.DailyQuota == 0 {
// 			lineOfData.DailyQuota = quotadaily
// 		}
// 		if lineOfData.MonthlyQuota == 0 {
// 			lineOfData.MonthlyQuota = quotamonthly
// 		}
// 		lineOfData.timeoutInt = time.Now().Add(1 * time.Minute).Unix()
// 		ipToMac[lineOfData.IP] = lineOfData
// 	}
// 	reply2, err2 := data.clientROS.Run("/ip/dhcp-server/lease/print") //, "?status=bound") //, "?disabled=false")
// 	if err2 != nil {
// 		log.Error(err2)
// 	}
// 	for _, re := range reply2.Re {
// 		lineOfData.Id = re.Map[".id"]
// 		lineOfData.IP = re.Map["active-address"]
// 		lineOfData.Mac = re.Map["active-mac-address"]
// 		lineOfData.timeout = re.Map["expires-after"]
// 		lineOfData.HostName = re.Map["host-name"]
// 		lineOfData.Comment = re.Map["comment"]
// 		lineOfData.HourlyQuota, lineOfData.DailyQuota, lineOfData.MonthlyQuota, lineOfData.Name, lineOfData.Position, lineOfData.Company, lineOfData.TypeD = parseComments(lineOfData.Comment)
// 		if lineOfData.HourlyQuota == 0 {
// 			lineOfData.HourlyQuota = quotahourly
// 		}
// 		if lineOfData.DailyQuota == 0 {
// 			lineOfData.DailyQuota = quotadaily
// 		}
// 		if lineOfData.MonthlyQuota == 0 {
// 			lineOfData.MonthlyQuota = quotamonthly
// 		}
// 		lineOfData.disable = re.Map["disabled"]
// 		addressLists := re.Map["address-lists"]
// 		lineOfData.addressLists = strings.Split(addressLists, ",")

// 		//Calculating the time when the lease of the address ends
// 		timeStr, err := time.ParseDuration(lineOfData.timeout)
// 		if err != nil {
// 			timeStr = 10 * time.Second
// 		}
// 		// Writes to a variable for further quick comparison
// 		lineOfData.timeoutInt = time.Now().Add(timeStr).Unix()

// 		ipToMac[lineOfData.IP] = lineOfData

// 	}
// 	return ipToMac
// }

func parseComments(comment string) (
	quotahourly, quotadaily, quotamonthly uint64,
	name, position, company, typeD string) {
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
	ipToMac := getDataFromMT(transport.QuotaType, transport.clientROS)
	transport.Lock()
	transport.infoOfDevices = ipToMac
	transport.Unlock()

	for _, value := range ipToMac {
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

func (transport *Transport) readsStreamFromMT(cfg *Config) {
	addr, err := net.ResolveUDPAddr("udp", cfg.FlowAddr)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

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
