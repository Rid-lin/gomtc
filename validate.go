package main

import "strings"

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
	for _, item := range arr {
		if item < "0" || item > "254" || len(item) > 3 {
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
