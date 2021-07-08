package main

import (
	"math"
	"sort"
	"time"

	"git.vegner.org/vsvegner/gomtc/internal/app/model"
)

type ReportDataType []model.LineOfDisplay

func (t *Transport) reportDailyHourlyByMac(rq model.RequestForm, showFriends bool) (model.DisplayDataType, error) {
	start := time.Now()
	devicesStat := GetDayStat(rq.DateFrom, rq.DateTo, t.DSN)
	ReportData := ToReportData(t.Aliases, devicesStat, t.devices)
	sort.Sort(ReportData)
	ReportData = ReportData.percentileCalculation(1)
	if !showFriends {
		ReportData = ReportData.FiltredFriendS(t.friends)
	}
	t.RLock()
	defer t.RUnlock()
	return model.DisplayDataType{
		ArrayDisplay:   ReportData,
		Header:         "Отчёт почасовой по трафику пользователей с логинами",
		DateFrom:       rq.DateFrom,
		DateTo:         rq.DateTo,
		LastUpdated:    t.lastUpdated.Format("2006-01-02 15:04:05.999"),
		LastUpdatedMT:  t.lastUpdatedMT.Format("2006-01-02 15:04:05.999"),
		TimeToGenerate: time.Since(start),
		ReferURL:       rq.ReferURL,
		Path:           rq.Path,
		SizeOneType: model.SizeOneType{
			SizeOneKilobyte: t.SizeOneKilobyte,
			SizeOneMegabyte: t.SizeOneKilobyte * t.SizeOneKilobyte,
			SizeOneGigabyte: t.SizeOneKilobyte * t.SizeOneKilobyte * t.SizeOneKilobyte,
		},
		Author: model.Author{Copyright: t.Copyright,
			Mail: t.Mail,
		},
		QuotaType: t.QuotaType,
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
			sum = (uint64(data[index-1].Precent) + data[index].VolumePerDay)
			data[index].Precent = float64(sum)
			PrecentilIndex = index
		}
	}
	if PrecentilIndex == 0 {
		PrecentilIndex = 1
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
			if rData[index].Login == friends[jndex] || rData[index].Alias == friends[jndex] || rData[index].ActiveAddress == friends[jndex] {
				rData = append(rData[:index], rData[index+1:]...)
				index--
				dataLen--
			}
		}
	}
	return rData
}

func add(slice []model.LineOfDisplay, line model.LineOfDisplay) []model.LineOfDisplay {
	for index, item := range slice {
		if line.Alias == item.Alias {
			slice[index].PerHour = line.PerHour
			return slice
		}
	}
	return append(slice, line)
}

func ToReportData(as map[string]model.AliasType, sd map[model.KeyDevice]model.StatDeviceType, ds DevicesMapType) ReportDataType {
	var totalVolumePerDay uint64
	var totalVolumePerHour [24]uint64

	ReportData := ReportDataType{}
	for key, value := range sd {
		line := model.LineOfDisplay{}
		line.Alias = key.Mac
		line.VolumePerDay = value.VolumePerDay
		totalVolumePerDay += value.VolumePerDay
		// TODO подумать над ключом
		line.InfoType.PersonType = as[key.Mac].PersonType
		line.InfoType.QuotaType = as[key.Mac].QuotaType
		line.InfoType.DeviceType = ds[key]
		for i := range line.PerHour {
			line.PerHour[i] = value.PerHour[i]
			totalVolumePerHour[i] += value.PerHour[i]
		}
		ReportData = add(ReportData, line)
	}
	line := model.LineOfDisplay{}
	line.Alias = "Всего"
	line.VolumePerDay = totalVolumePerDay
	line.PerHour = totalVolumePerHour
	ReportData = add(ReportData, line)
	return ReportData
}
