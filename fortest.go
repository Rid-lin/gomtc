package main

// func isIPnetaddr(inputStr string) bool {
// 	_, err := netaddr.ParseIP(inputStr)
// 	return err == nil
// }

// func isMacNetaddr(inputStr string) bool {
// 	ip := netaddr.ma
// 	return !(ip == nil)
// }

// func validateMac(inputStr string) bool {
// 	arr := strings.Split(inputStr, ":")
// 	for i := range arr {
// 		if !isHexColon(arr[i]) {
// 			return false
// 		}
// 	}
// 	return len(arr) == 6
// }

// func is_ipv4(host string) bool {
// 	parts := strings.Split(host, ".")
// 	if len(parts) < 4 {
// 		return false
// 	}
// 	for _, x := range parts {
// 		if i, err := strconv.Atoi(x); err == nil {
// 			if i < 0 || i > 255 {
// 				return false
// 			}
// 		} else {
// 			return false
// 		}
// 	}
// 	return true
// }

// func validIP4(ipAddress string) bool {
// 	ipAddress = strings.Trim(ipAddress, " ")

// 	re, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
// 	return re.MatchString(ipAddress)
// }

// // bug in the function
// func isIP(inputStr string) bool {
// 	arr := strings.Split(inputStr, ".")
// 	for _, item := range arr {
// 		if item < "0" || item > "254" || len(item) > 3 {
// 			return false
// 		}
// 	}
// 	return len(arr) == 4
// }
