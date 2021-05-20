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
	for {
		data.updateInfoOfDevicesFromMT()
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

func isMac(inputStr string) bool {
	arr := strings.Split(inputStr, ":")
	for i := range arr {
		if !isHexColon(arr[i]) {
			return false
		}
	}
	return len(arr) == 6
}

func isIP(inputStr string) bool {
	arr := strings.Split(inputStr, ".")
	for i := range arr {
		if !isNumDot(arr[i]) {
			return false
		}
	}
	return len(arr) == 4
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

func isHexColon(s string) bool {
	if len(s) != 2 {
		return false
	} else if s == `
` {
		return false
	}
	colonFound := 5
	for _, v := range s {
		if v == ':' {
			if colonFound < 0 {
				return false
			}
			colonFound = colonFound - 1
		} else if (v < '0' || v > '9') && (v < 'a' || v > 'f') && (v < 'A' || v > 'F') {
			return false
		}
	}
	return true
}

// validateIP Returns the IP address if the first is an IP address,
// otherwise it checks if the second parameter is an IP address.
// Otherwise, it returns an empty string.
func validateIP(ip, altIp string) string {
	if isIP(ip) {
		return ip
	} else if isIP(altIp) {
		return altIp
	}
	return ""
}

func validateMac(mac, altMac, hopeMac, lastHopeMac string) string {
	var hopeMacR, lastHopeMacR string
	if len(hopeMac) > 2 {
		hopeMacR = hopeMac[2:]
	}
	if len(lastHopeMac) > 2 {
		lastHopeMacR = lastHopeMac[2:]
	}
	switch {
	case mac != "":
		return mac
	case altMac != "":
		return altMac
	case isMac(hopeMacR):
		return hopeMacR
	case isMac(lastHopeMacR):
		return lastHopeMacR
	}
	return ""
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
