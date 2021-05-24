package main

import (
	"math"
	"sort"
	"time"
)

type ReportDataType []LineOfDisplay

func (t *Transport) reportTrafficHourlyByLogins(request RequestForm, showFriends bool) DisplayDataType {
	start := time.Now()
	t.RLock()
	dataChashe := t.dataCashe
	SizeOneKilobyte := t.SizeOneKilobyte
	Quota := t.QuotaType
	Copyright := t.Copyright
	Mail := t.Mail
	LastUpdated := t.lastUpdated.Format("2006-01-02 15:04:05.999")
	LastUpdatedMT := t.lastUpdatedMT.Format("2006-01-02 15:04:05.999")
	t.RUnlock()

	ReportData := ReportDataType{}
	line := LineOfDisplay{}
	for key, value := range dataChashe {
		if key.DateStr != request.dateFrom {
			continue
		}
		line.Alias = key.Alias
		line.AliasType = value.AliasType
		line.StatType = value.StatType
		ReportData = add(ReportData, line)
	}

	sort.Sort(ReportData)
	ReportData = ReportData.percentileCalculation(1)
	if !showFriends {
		ReportData = ReportData.FiltredFriendS(t.friends)
	}
	// for _, dl := range ReportData {
	// 	if dl.Alias == "4C:63:71:75:C6:B6" {
	// 		runtime.Breakpoint()
	// 	}
	// }

	return DisplayDataType{
		ArrayDisplay:   ReportData,
		Logs:           []LogsOfJob{},
		Header:         "Отчёт почасовой по трафику пользователей с логинами и IP-адресами",
		DateFrom:       request.dateFrom,
		DateTo:         "",
		LastUpdated:    LastUpdated,
		LastUpdatedMT:  LastUpdatedMT,
		TimeToGenerate: time.Since(start),
		ReferURL:       request.referURL,
		Path:           request.path,
		SizeOneType: SizeOneType{
			SizeOneKilobyte: SizeOneKilobyte,
			SizeOneMegabyte: SizeOneKilobyte * SizeOneKilobyte,
			SizeOneGigabyte: SizeOneKilobyte * SizeOneKilobyte * SizeOneKilobyte,
		},
		Author: Author{Copyright: Copyright,
			Mail: Mail,
		},
		QuotaType: Quota,
	}

}

func (a ReportDataType) Len() int           { return len(a) }
func (a ReportDataType) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ReportDataType) Less(i, j int) bool { return a[i].Size > a[j].Size }

func (data ReportDataType) percentileCalculation(cub uint8) ReportDataType {
	var maxIndex = 0
	var PrecentilIndex int
	var sum uint64
	if len(data) == 0 {
		return data
	}
	SizeOfPrecentil := uint64(float64(data[maxIndex].Size) * 0.9)
	sumTotal := data[maxIndex].Size // МАксимальная сумма необходима для расчёта претентиля 90
	// cubf := math.Pow(10, float64(cub)) // Высчитываем степерь округления
	// Если сумма скаченного трафика текущего пользователя и тех кого уже прошли будет больше чем размер прецентиля, то мы отмечает порядковый номер данного пользователя для последующей обработки
	for index := 1; index < len(data)-1; index++ {
		if SizeOfPrecentil < sum {
			PrecentilIndex = index
			break
		} else {
			// ... инвче прибавляем к текущей сумме объём скаченного пользователем
			sum = (data[index-1].SizeOfPrecentil + data[index].Size)
			data[index].SizeOfPrecentil = sum
			PrecentilIndex = index
		}
	}
	AverageTotal := data[maxIndex].Size / uint64(PrecentilIndex)
	data[maxIndex].Average = AverageTotal

	for index := 1; index < PrecentilIndex; index++ {
		// data[index].Average = math.Round(data[index].Size/float64(PrecentilIndex)*cubf) / cubf
		data[index].Precent = math.Round(float64(data[index].Size)/float64(sumTotal)*1000) / 10
		if data[index].Size > AverageTotal {
			data[maxIndex].Count++
		}
	}
	return data
}

func (rData ReportDataType) FiltredFriendS(friends []string) ReportDataType {
	dataLen := len(rData)
	for index := 0; index < dataLen; index++ {
		for jndex := range friends {
			if rData[index].Login == friends[jndex] || rData[index].Alias == friends[jndex] || rData[index].IP == friends[jndex] {
				rData = append(rData[:index], rData[index+1:]...)
				index--
				dataLen--
			}
		}
	}
	return rData
}

// func (data ReportDataType) Format(cub uint8) ReportDataType {
// 	for index := 1; index < len(data)-1; index++ {
// 		// data[index].AverageStr = fmt.Sprintf("%6.2f", data[index].Average)
// 		data[index].PrecentStr = fmt.Sprintf("%6.2f", data[index].Precent)
// 		HourSize := data[index].SizeOfHourU
// 		HourSizeStr := data[index].SizeOfHourStr
// 		for hourIndex := range HourSize {
// 			HourSizeStr[hourIndex] = fmt.Sprintf("%6.2f", HourSize[hourIndex])
// 		}
// 		data[index].SizeOfHourStr = HourSizeStr
// 	}
// 	return data
// }

func add(slice []LineOfDisplay, line LineOfDisplay) []LineOfDisplay {
	for index, item := range slice {
		if line.Alias == item.Alias {
			slice[index].SizeOfHour = line.SizeOfHour
			return slice
		}
	}
	return append(slice, line)
}

// // roundToMb Function rounds a number to megabyte with cub precision
// func roundToMb(sizeInBytes uint64, SizeOneKilobyte uint64, cub int) float64 {
// 	if SizeOneKilobyte == 0 {
// 		SizeOneKilobyte = 1048576
// 	}
// 	sizeInBytesf := float64(sizeInBytes)
// 	cubf := math.Pow(10, float64(cub))
// 	// return (math.Round((float64(sizeInBytes) / SizeOneKilobyte * math.Pow(10, float64(cub+1))) / math.Pow(10, float64(cub+1))))
// 	return math.Round(sizeInBytesf/float64(SizeOneKilobyte)*cubf) / cubf
// }
