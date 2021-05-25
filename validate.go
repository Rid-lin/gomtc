package main

import (
	"net"
	"strings"
)

func isMac(inputStr string) bool {
	_, err := net.ParseMAC(inputStr)
	return err == nil
}

func isIP(inputStr string) bool {
	ip := net.ParseIP(inputStr)
	return !(ip == nil)
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
	// colonFound := 2
	for _, v := range s {
		// if v == ':' {
		// 	if colonFound < 0 {
		// 		return false
		// 	}
		// 	colonFound = colonFound - 1
		// } else
		if (v < '0' || v > '9') && (v < 'a' || v > 'f') && (v < 'A' || v > 'F') {
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

func getSwithMac(mac, altMac, hopeMac, lastHopeMac string) string {
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

func isParametr(inputStr, parametr string) bool {
	arrStr := strings.Split(inputStr, "=")
	if len(arrStr) != 2 {
		return false
	}
	if arrStr[0] != parametr {
		return false
	}
	return true
}

func isComment(inputStr, parametr string) bool {
	arrStr := strings.Split(inputStr, "=")
	if len(arrStr) < 2 {
		return false
	}
	if arrStr[0] != parametr {
		return false
	}
	return true
}
