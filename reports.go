package main

import (
	"math"
	"sort"
	"time"
)

type ReportDataType []LineOfDisplay

func (t *Transport) reportDailyHourlyByMac(rq RequestForm, showFriends bool) (DisplayDataType, error) {
	start := time.Now()
	t.RLock()
	// data := t.statofYears
	SizeOneKilobyte := t.SizeOneKilobyte
	Quota := t.QuotaType
	Copyright := t.Copyright
	Mail := t.Mail
	// BlockAddressList := t.BlockAddressList
	LastUpdated := t.lastUpdated.Format("2006-01-02 15:04:05.999")
	LastUpdatedMT := t.lastUpdatedMT.Format("2006-01-02 15:04:05.999")
	t.RUnlock()
	day := t.getDay(rq.ToLine())
	ReportData := ReportDataType{}
	line := LineOfDisplay{}
	var totalVolumePerDay uint64
	var totalVolumePerHour [24]uint64
	t.RLock()
	for key, value := range day.devicesStat {

		line.Alias = key.mac
		line.VolumePerDay = value.VolumePerDay
		totalVolumePerDay += value.VolumePerDay
		// TODO подумать над ключом
		line.InfoType.PersonType = t.Aliases[key.mac].PersonType
		line.InfoType.QuotaType = t.Aliases[key.mac].QuotaType
		line.InfoType.DeviceType = t.devices[key]
		for i := range line.PerHour {
			line.PerHour[i] = value.PerHour[i]
			totalVolumePerHour[i] += value.PerHour[i]
		}
		ReportData = add(ReportData, line)
	}
	t.RUnlock()
	line = LineOfDisplay{}
	line.Alias = "Всего"
	line.VolumePerDay = totalVolumePerDay
	line.PerHour = totalVolumePerHour
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
		DateFrom:       rq.dateFrom,
		DateTo:         "",
		LastUpdated:    LastUpdated,
		LastUpdatedMT:  LastUpdatedMT,
		TimeToGenerate: time.Since(start),
		ReferURL:       rq.referURL,
		Path:           rq.path,
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

func add(slice []LineOfDisplay, line LineOfDisplay) []LineOfDisplay {
	for index, item := range slice {
		if line.Alias == item.Alias {
			slice[index].PerHour = line.PerHour
			return slice
		}
	}
	return append(slice, line)
}

func (rq *RequestForm) ToLine() *lineOfLogType {
	l := lineOfLogType{}
	tn, err := time.Parse(DateLayout, rq.dateFrom)
	if err != nil {
		tn = time.Now()
	}
	l.year = tn.Year()
	l.month = tn.Month()
	l.day = tn.Day()
	l.hour = tn.Hour()
	l.minute = tn.Minute()
	return &l
}
