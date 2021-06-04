package main

import (
	"bufio"
	"fmt"
	"os"
)

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

func saveArrToFile(nameFile string, arr []string) error {
	f, _ := os.Create(nameFile)
	defer f.Close()
	w := bufio.NewWriter(f)
	for index := 0; index < len(arr)-1; index++ {
		fmt.Fprintln(w, arr[index])
	}
	w.Flush()
	return nil
}

func saveStrToFile(nameFile, str string) error {
	f, _ := os.Create(nameFile)
	defer f.Close()
	_, _ = f.WriteString(str)
	return nil
}

func (t *Transport) addLineOutToMapOfReports(value *lineOfLogType) {
	value.alias = determiningAlias(*value)
	t.trafficСounting(value)
}

func (t *Transport) trafficСounting(l *lineOfLogType) {
	// Идея такая.
	// посчитать статистику для каждого отдельного случая, когда:
	// есть и мак и айпи, есть только айпи, есть только мак
	// записать это в слайс и привязать к отдельному оборудованию.
	// при чём по привязка только айпи адресу идёт только в течении сегодняшнего дня, потом не учитывается
	// привязка по маку и мак+айпи идёт всегда, т.к. устройство опознано.
	t.Lock()
	statForDate, iStatDay, _ := t.getStatForDate(l) // статистка за день по всем устройствам и общая
	// Присваеваем данные в массиве временной переменной для того чтобы предыдущие значения не потерялись
	devStat, iStatDev, _ := statForDate.findStat(l)
	// Расчет суммы трафика для устройства для дальшейшего отображения
	devStat.VolumePerDay = devStat.VolumePerDay + l.sizeInBytes
	devStat.VolumePerCheck = devStat.VolumePerCheck + l.sizeInBytes
	devStat.StatPerHour[l.hour].Hour = devStat.StatPerHour[l.hour].Hour + l.sizeInBytes
	devStat.StatPerHour[l.hour].Minute[l.minute] = devStat.StatPerHour[l.hour].Minute[l.minute] + l.sizeInBytes
	// Расчет суммы трафика для дня для дальшейшего отображения
	statForDate.VolumePerDay = statForDate.VolumePerDay + l.sizeInBytes
	statForDate.VolumePerCheck = statForDate.VolumePerCheck + l.sizeInBytes
	statForDate.StatPerHour[l.hour].Hour = statForDate.StatPerHour[l.hour].Hour + l.sizeInBytes
	statForDate.StatPerHour[l.hour].Minute[l.minute] = statForDate.StatPerHour[l.hour].Minute[l.minute] + l.sizeInBytes
	// Возвращаем данные обратно
	statForDate.devicesStat[iStatDev] = devStat
	t.stats[iStatDay] = statForDate

	t.Unlock()
}

func (t *Transport) getStatForDate(l *lineOfLogType) (StatDayType, int, error) {
	date := fmt.Sprintf("%d-%s-%d", l.year, l.month.String(), l.day)
	for index, statForDay := range t.stats {
		if statForDay.date == date {
			return statForDay, index, nil
		}
	}
	t.stats = append(t.stats, StatDayType{})
	return StatDayType{}, len(t.stats) - 1, fmt.Errorf("Not found statistic of Day")
}

func (ss *StatDayType) findStat(l *lineOfLogType) (StatDeviceType, int, error) {
	for index, s := range ss.devicesStat {
		if s.Mac == l.login || s.IP == l.ipaddress {
			return s, index, nil
		}
	}
	var mac, ip string
	if isMac(l.login) {
		mac = l.login
	}
	if isIP(l.ipaddress) {
		ip = l.ipaddress
	}
	ss.devicesStat = append(ss.devicesStat, StatDeviceType{Mac: mac, IP: ip})
	return ss.devicesStat[len(ss.devicesStat)-1], len(ss.devicesStat) - 1, nil
}
