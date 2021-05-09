package main

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"time"
)

type ReportDataType []LineOfDisplay

func (t *Transport) reportTrafficHourlyByLogins(request RequestForm, showFriends bool) DisplayDataType {
	start := time.Now()
	ReportData := ReportDataType{}
	line := LineOfDisplay{}
	// t.Lock()
	for key, value := range t.dataChashe {
		if key.DateStr != request.dateFrom {
			continue
		}
		line.Alias = key.Alias
		if key.Alias == "" {
			runtime.Breakpoint()
		}
		if line.Alias == "" {
			runtime.Breakpoint()
		}
		line.PersonType = value.PersonType
		line.QuotaType = value.QuotaType
		line.DeviceType = value.DeviceType
		// line.HostName = value.HostName
		// line.Comments = value.Comments
		line.Size = roundToMb(value.SizeInBytes, t.SizeOneMegabyte, 1)
		for index := range line.HourSize {
			line.HourSize[index] = roundToMb(value.SizeOfHour[index], t.SizeOneMegabyte, 1)
		}
		// if value.Blocked {
		// 	runtime.Breakpoint()
		// }
		// line.Blocked = value.Blocked
		ReportData = add(ReportData, line)
	}
	// t.Unlock()

	sort.Sort(ReportData)
	ReportData = ReportData.percentileCalculation(1)
	// TODO Доделать фильтрацию друзей
	if !showFriends {
		ReportData = ReportData.FiltredFriendS(t.friends)
	}
	ReportData = ReportData.Format(1)

	return DisplayDataType{
		ArrayDisplay: ReportData,
		Logs:         []LogsOfJob{},
		QuotaType: QuotaType{
			HourlyQuota:  t.HourlyQuota / uint64(t.SizeOneMegabyte),
			DailyQuota:   t.DailyQuota / uint64(t.SizeOneMegabyte),
			MonthlyQuota: t.MonthlyQuota / uint64(t.SizeOneMegabyte),
		},
		Header:         "Отчёт почасовой по трафику пользователей с логинами и IP-адресами",
		DateFrom:       request.dateFrom,
		DateTo:         "",
		LastUpdated:    t.lastUpdated.Format("2006-01-02 15:04:05.999"),
		TimeToGenerate: time.Since(start),
		Author: Author{Copyright: t.Copyright,
			Mail: t.Mail,
		},
	}

}

// func (t *Transport) fillDisplayData(request RequestForm, header string) DisplayDataType {
// 	start := time.Now()
// 	ReportData := ReportDataType{}
// 	line := LineOfDisplay{}
// 	t.RLock()
// 	for key, value := range t.data {
// 		if key.DateStr != request.dateFrom {
// 			continue
// 		}
// 		line.Alias = key.Alias
// 		line.Hostname = value.Hostname
// 		line.Comments = value.Comments
// 		line.Size = roundToMb(value.SizeInBytes, t.SizeOneMegabyte, 1)
// 		for index := range line.HourSize {
// 			line.HourSize[index] = roundToMb(value.SizeOfHour[index], t.SizeOneMegabyte, 1)
// 		}
// 		ReportData = add(ReportData, line)
// 	}
// 	t.RUnlock()

// 	sort.Sort(ReportData)
// 	ReportData = ReportData.percentileCalculation(1)
// 	// TODO Доделать фильтрацию друзей
// 	// ReportData = ReportData.FiltredFriend(t.cfg)
// 	ReportData = ReportData.Format(1)

// 	return DisplayDataType{
// 		ArrayDisplay:   ReportData,
// 		Header:         header,
// 		DateFrom:       request.dateFrom,
// 		TimeToGenerate: time.Since(start),
// 		Copyright:      t.Copyright,
// 		Mail:           t.Mail,
// 	}

// }

func (a ReportDataType) Len() int           { return len(a) }
func (a ReportDataType) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ReportDataType) Less(i, j int) bool { return a[i].Size > a[j].Size }

func (data ReportDataType) percentileCalculation(cub uint8) ReportDataType {
	var maxIndex = 0
	var PrecentilIndex int
	var sum float64
	if len(data) == 0 {
		return data
	}
	SizeOfPrecentil := data[maxIndex].Size * 0.9
	sumTotal := data[maxIndex].Size    // МАксимальная сумма необходима для расчёта претентиля 90
	cubf := math.Pow(10, float64(cub)) // Высчитываем степерь округления
	// Если сумма скаченного трафика текущего пользователя и тех кого уже прошли будет больше чем размер прецентиля, то мы отмечает порядковый номер данного пользователя для последующей обработки
	for index := 1; index < len(data)-1; index++ {
		if SizeOfPrecentil < sum {
			PrecentilIndex = index
			break
		} else {
			// ... инвче прибавляем к текущей сумме объём скаченного пользователем
			sum = math.Round((data[index-1].SizeOfPrecentil+data[index].Size)*cubf) / cubf
			data[index].SizeOfPrecentil = sum
			PrecentilIndex = index
		}
	}
	AverageTotal := math.Round(data[maxIndex].Size/float64(PrecentilIndex)*cubf) / cubf
	data[maxIndex].Average = AverageTotal

	for index := 1; index < PrecentilIndex; index++ {
		// data[index].Average = math.Round(data[index].Size/float64(PrecentilIndex)*cubf) / cubf
		data[index].Precent = math.Round(data[index].Size/sumTotal*1000) / 10
		if data[index].Size > AverageTotal {
			data[maxIndex].Count++
		}
	}
	return data
}

// TODO Сделать форматирование вывода как то вместо 0.9 (Мб) использовать 935 Кб
// TODO вместо 4384 Мб использовать 4,38 Гб или 4.4 Гб
// TODO Вывод сделать чисто в строковом формате

func (data ReportDataType) FiltredFriendS(friends []string) ReportDataType {
	// newData := ReportDataType
	dataLen := len(data)
	for index := 0; index < dataLen; index++ {
		for jndex := range friends {
			if data[index].Login == friends[jndex] || data[index].Alias == friends[jndex] || data[index].IP == friends[jndex] {
				data = append(data[:index], data[index+1:]...)
				index--
				dataLen--
			}
		}
	}
	return data
}

// func (data ReportDataType) FiltredFriendS(friends []string) ReportDataType {
// 	newData := ReportDataType{}
// 	for index := range data {
// 	nextDevice:
// 		for jndex := range friends {
// 			if data[index].Login == friends[jndex] || data[index].Alias == friends[jndex] || data[index].IP == friends[jndex] {
// 				continue nextDevice
// 			}
// 		}
// 		newData = append(newData, data[index])
// 	}
// 	return data
// }

func (data ReportDataType) Format(cub uint8) ReportDataType {
	for index := 1; index < len(data)-1; index++ {
		// data[index].AverageStr = fmt.Sprintf("%6.2f", data[index].Average)
		data[index].PrecentStr = fmt.Sprintf("%6.2f", data[index].Precent)
		HourSize := data[index].HourSize
		HourSizeStr := data[index].HourSizeStr
		for jndex := range HourSize {
			HourSizeStr[jndex] = fmt.Sprintf("%6.2f", HourSize[jndex])
		}
		data[index].HourSizeStr = HourSizeStr
	}
	return data
}

func add(slice []LineOfDisplay, line LineOfDisplay) []LineOfDisplay {
	for index, item := range slice {
		if line.Alias == item.Alias {
			slice[index].HourSize = line.HourSize
			return slice
		}
	}
	return append(slice, line)
}

// roundToMb Function rounds a number to megabyte with cub precision
func roundToMb(sizeInBytes uint64, sizeOneMegabyte uint64, cub int) float64 {
	if sizeOneMegabyte == 0 {
		sizeOneMegabyte = 1048576
	}
	sizeInBytesf := float64(sizeInBytes)
	cubf := math.Pow(10, float64(cub))
	// return (math.Round((float64(sizeInBytes) / sizeOneMegabyte * math.Pow(10, float64(cub+1))) / math.Pow(10, float64(cub+1))))
	return math.Round(sizeInBytesf/float64(sizeOneMegabyte)*cubf) / cubf
}
