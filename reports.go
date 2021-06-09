package main

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type ReportDataType []LineOfDisplay

// func (t *Transport) reportTrafficHourlyByLogins(request RequestForm, showFriends bool) DisplayDataType {
// 	start := time.Now()
// 	t.RLock()
// 	dataChashe := t.dataCasheOld
// 	SizeOneKilobyte := t.SizeOneKilobyte
// 	Quota := t.QuotaType
// 	Copyright := t.Copyright
// 	Mail := t.Mail
// 	LastUpdated := t.lastUpdated.Format("2006-01-02 15:04:05.999")
// 	LastUpdatedMT := t.lastUpdatedMT.Format("2006-01-02 15:04:05.999")
// 	t.RUnlock()

// 	ReportData := ReportDataType{}
// 	line := LineOfDisplay{}
// 	for key, value := range dataChashe {
// 		if key.DateStr != request.dateFrom {
// 			continue
// 		}
// 		line.Alias = key.Alias
// 		line.InfoType = value.InfoType
// 		line.StatOldType = value.StatOldType
// 		ReportData = add(ReportData, line)
// 	}

// 	sort.Sort(ReportData)
// 	ReportData = ReportData.percentileCalculation(1)
// 	if !showFriends {
// 		ReportData = ReportData.FiltredFriendS(t.friends)
// 	}

// 	return DisplayDataType{
// 		ArrayDisplay:   ReportData,
// 		Logs:           []LogsOfJob{},
// 		Header:         "Отчёт почасовой по трафику пользователей с логинами и IP-адресами",
// 		DateFrom:       request.dateFrom,
// 		DateTo:         "",
// 		LastUpdated:    LastUpdated,
// 		LastUpdatedMT:  LastUpdatedMT,
// 		TimeToGenerate: time.Since(start),
// 		ReferURL:       request.referURL,
// 		Path:           request.path,
// 		SizeOneType: SizeOneType{
// 			SizeOneKilobyte: SizeOneKilobyte,
// 			SizeOneMegabyte: SizeOneKilobyte * SizeOneKilobyte,
// 			SizeOneGigabyte: SizeOneKilobyte * SizeOneKilobyte * SizeOneKilobyte,
// 		},
// 		Author: Author{Copyright: Copyright,
// 			Mail: Mail,
// 		},
// 		QuotaType: Quota,
// 	}

// }

func (t *Transport) reportTrafficHourlyByLoginsNew(request RequestForm, showFriends bool) (DisplayDataType, error) {
	start := time.Now()
	t.RLock()
	data := t.statofYears
	SizeOneKilobyte := t.SizeOneKilobyte
	Quota := t.QuotaType
	Copyright := t.Copyright
	Mail := t.Mail
	// BlockAddressList := t.BlockAddressList
	LastUpdated := t.lastUpdated.Format("2006-01-02 15:04:05.999")
	LastUpdatedMT := t.lastUpdatedMT.Format("2006-01-02 15:04:05.999")
	t.RUnlock()

	tn, err := time.Parse("2006-01-02", request.dateFrom)
	if err != nil {
		tn = time.Now()
	}
	yearStat, ok := data[tn.Year()]
	if !ok {
		return DisplayDataType{}, fmt.Errorf("Year(%d) missing from statistics", tn.Year())
	}
	monthStat, ok := yearStat.monthsStat[tn.Month()]
	if !ok {
		return DisplayDataType{}, fmt.Errorf("Month(%s) missing from statistics", tn.Month().String())
	}
	day, ok := monthStat.daysStat[tn.Day()]
	if !ok {
		return DisplayDataType{}, fmt.Errorf("Day(%d) missing from statistics", tn.Day())
	}

	ReportData := ReportDataType{}
	line := LineOfDisplay{}
	var totalVolumePerDay uint64
	var totalVolumePerHour [24]uint64
	for key, value := range day.devicesStat {

		line.Alias = key.mac
		line.VolumePerDay = value.VolumePerDay
		totalVolumePerDay += value.VolumePerDay
		// TODO подумать над ключом
		line.InfoOldType.PersonType = t.Aliases[key.mac].PersonType
		for i := range line.VolumePerHour {
			line.VolumePerHour[i] = value.StatPerHour[i].Hour
			totalVolumePerHour[i] += value.StatPerHour[i].Hour
		}
		ReportData = add(ReportData, line)
	}
	line = LineOfDisplay{}
	line.Alias = "Всего"
	line.VolumePerDay = totalVolumePerDay
	line.VolumePerHour = totalVolumePerHour
	ReportData = add(ReportData, line)

	sort.Sort(ReportData)
	ReportData = ReportData.percentileCalculation(1)
	if !showFriends {
		ReportData = ReportData.FiltredFriendS(t.friends)
	}

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
	}, nil

}

func (a ReportDataType) Len() int           { return len(a) }
func (a ReportDataType) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ReportDataType) Less(i, j int) bool { return a[i].VolumePerDay > a[j].VolumePerDay }

func (data ReportDataType) percentileCalculation(cub uint8) ReportDataType {
	var maxIndex = 0
	var PrecentilIndex int
	var sum uint64
	if len(data) == 0 {
		return data
	}
	SizeOfPrecentil := uint64(float64(data[maxIndex].VolumePerDay) * 0.9)
	sumTotal := data[maxIndex].VolumePerDay // МАксимальная сумма необходима для расчёта претентиля 90
	// cubf := math.Pow(10, float64(cub)) // Высчитываем степерь округления
	// Если сумма скаченного трафика текущего пользователя и тех кого уже прошли будет больше чем размер прецентиля, то мы отмечает порядковый номер данного пользователя для последующей обработки
	for index := 1; index < len(data)-1; index++ {
		if SizeOfPrecentil < sum {
			PrecentilIndex = index
			break
		} else {
			// ... инвче прибавляем к текущей сумме объём скаченного пользователем
			sum = (data[index-1].VolumeOfPrecentil + data[index].VolumePerDay)
			data[index].VolumeOfPrecentil = sum
			PrecentilIndex = index
		}
	}
	AverageTotal := data[maxIndex].VolumePerDay / uint64(PrecentilIndex)
	data[maxIndex].Average = AverageTotal

	for index := 1; index < PrecentilIndex; index++ {
		// data[index].Average = math.Round(data[index].Size/float64(PrecentilIndex)*cubf) / cubf
		data[index].Precent = math.Round(float64(data[index].VolumePerDay)/float64(sumTotal)*1000) / 10
		if data[index].VolumePerDay > AverageTotal {
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

func add(slice []LineOfDisplay, line LineOfDisplay) []LineOfDisplay {
	for index, item := range slice {
		if line.Alias == item.Alias {
			slice[index].VolumePerHour = line.VolumePerHour
			return slice
		}
	}
	return append(slice, line)
}
