package main

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func (data *Transport) loopGetDataFromMT() {
	p := parseType{}
	for {
		data.RLock()
		p.SSHCredetinals = data.sshCredetinals
		p.BlockAddressList = data.BlockAddressList
		p.QuotaType = data.QuotaType
		p.Location = data.Location
		data.RUnlock()
		data.Lock()
		data.devices = parseInfoFromMTToSlice(p)
		data.Unlock()
		data.checkQuota()
		// data.updateStatusDevicesToMT(&cfg)
		interval, err := time.ParseDuration(cfg.Interval)
		if err != nil {
			interval = 1 * time.Minute
		}
		time.Sleep(interval)

	}
}

// func (data *Transport) loopGetDataFromMTOverAPI() {
// 	for {
// 		data.updateInfoOfDevicesFromMT()
// 		if err := data.getStatusallDevices(); err == nil {
// 			// if err := transport.getStatusDevices(cfg); err == nil {
// 			data.checkQuota()
// 			// transport.updateStatusDevicesToMT(cfg)
// 		}
// 		interval, err := time.ParseDuration(cfg.Interval)
// 		if err != nil {
// 			interval = 1 * time.Minute
// 		}
// 		time.Sleep(interval)

// 	}
// }

func parseParamertToStr(inpuStr string) string {
	inpuStr = strings.Trim(inpuStr, "=")
	inpuStr = strings.ReplaceAll(inpuStr, "==", "=")
	arr := strings.Split(inpuStr, "=")
	if len(arr) > 1 {
		return arr[1]
	} else {
		log.Errorf("Parameter error. The parameter is specified incorrectly or not specified at all.(%v)", inpuStr)
	}
	return ""
}

func parseParamertToUint(inputValue string) uint64 {
	var err error
	var quota uint64
	inputValue = strings.Trim(inputValue, "=")
	inputValue = strings.ReplaceAll(inputValue, "==", "=")
	Arr := strings.Split(inputValue, "=")
	if len(Arr) > 1 {
		quotaStr := Arr[1]
		quota, err = strconv.ParseUint(quotaStr, 10, 64)
		if err != nil {
			quotaF, err := strconv.ParseFloat(quotaStr, 64)
			if err != nil {
				log.Errorf("Error parse quota from:(%v) with:(%v)", quotaStr, err)
				return 0
			}
			quota = uint64(quotaF)
		}
		return quota
	} else {
		log.Errorf("Parameter error. The parameter is specified incorrectly or not specified at all.(%v)", inputValue)
	}
	return quota
}

func parseParamertToBool(inputValue string) bool {
	inputValue = strings.Trim(inputValue, "=")
	inputValue = strings.ReplaceAll(inputValue, "==", "=")
	arr := strings.Split(inputValue, "=")
	if len(arr) > 1 {
		value := arr[1]
		if value == "true" || value == "yes" {
			return true
		}
	}
	return false
}

func paramertToUint(inputValue string) uint64 {
	quota, err := strconv.ParseUint(inputValue, 10, 64)
	if err != nil {
		quotaF, err := strconv.ParseFloat(inputValue, 64)
		if err != nil {
			log.Errorf("Error parse quota from:(%v) with:(%v)", inputValue, err)
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

func makeCommentFromIodt(d InfoOfDeviceType, quota QuotaType) string {
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
	if d.HourlyQuota != 0 && d.HourlyQuota != quota.HourlyQuota {
		comment = fmt.Sprintf("%v/quotahourly=%v",
			comment,
			d.HourlyQuota)
	}
	if d.DailyQuota != 0 && d.DailyQuota != quota.DailyQuota {
		comment = fmt.Sprintf("%v/quotadaily=%v",
			comment,
			d.DailyQuota)
	}
	if d.MonthlyQuota != 0 && d.MonthlyQuota != quota.MonthlyQuota {
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

	return comment
}

func parseComment(comment string) (
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

func (d DeviceType) convertToInfo() InfoOfDeviceType {
	var (
		ip, mac, name, position, company, typeD, IDUser, comment string
		hourlyQuota, dailyQuota, monthlyQuota                    uint64
		manual                                                   bool
	)
	ip = validateIP(d.activeAddress, d.address)
	mac = validateMac(d.activeMacAddress, d.macAddress, d.clientId, d.activeClientId)
	hourlyQuota, dailyQuota, monthlyQuota, name, position, company, typeD, IDUser, comment, manual = parseComment(d.comment)
	infoD := InfoOfDeviceType{
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
	return infoD
}

func (dInfo *InfoOfDeviceType) convertToDevice() DeviceType {
	return DeviceType{
		activeAddress:    dInfo.IP,
		address:          dInfo.IP,
		activeMacAddress: dInfo.AMac,
		addressLists:     dInfo.Groups,
		blocked:          fmt.Sprint(dInfo.Blocked),
		comment:          makeCommentFromIodt(*dInfo, dInfo.QuotaType),
		disabled:         fmt.Sprint(dInfo.Disabled),
		hostName:         dInfo.HostName,
		macAddress:       dInfo.Mac,
		Manual:           dInfo.Manual,
		ShouldBeBlocked:  dInfo.ShouldBeBlocked,
		timeout:          dInfo.timeout,
	}
}

func (d1 *DeviceType) compare(d2 *DeviceType) bool {
	switch {
	case d1.macAddress == d2.macAddress:
		return true
	case d1.activeMacAddress == d2.macAddress:
		return true
	case d1.macAddress == d2.activeMacAddress:
		return true
	case d1.activeMacAddress == d2.activeMacAddress:
		return true
	case d1.activeClientId == d2.activeClientId:
		return true
	case d1.activeClientId == d2.clientId:
		return true
	case d1.clientId == d2.activeClientId:
		return true
	case d1.clientId == d2.clientId:
		return true
	}
	return false
}

func (ds *DevicesType) find(d *DeviceType) int {
	for index, device := range *ds {
		if d.compare(&device) {
			return index
		}
	}
	return -1
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
