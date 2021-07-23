package model

import "time"

type NewContType struct {
	Count
	StartTime   time.Time
	EndTime     time.Time
	LastUpdated time.Time
	LastDate    int64
	LastDayStr  string
}

type StatOfYearType struct {
	MonthsStat map[time.Month]StatOfMonthType // Statistics of all devices that were seen that year
	Year       int
}

type StatOfMonthType struct {
	DaysStat map[int]StatOfDayType // Statistics of all devices that were seen that month
	Month    time.Month
}

type StatOfDayType struct {
	DevicesStat map[KeyDevice]StatDeviceType // Statistics of all devices that were seen that day
	Day         int
	StatType    // General statistics of the day, to speed up access
}

type StatDeviceType struct {
	Mac string
	Ip  string
	StatType
}
