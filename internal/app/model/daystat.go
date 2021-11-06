package model

import "time"

// type StatDay struct {
// 	devicesStat []StatDevice // Statistics of all devices that were seen that day
// 	// 	day         int
// }

type KeyStat struct {
	Year  int
	Month time.Month
	Day   int
}

type StatDevice struct {
	PerHour        [24]uint64
	Mac            string
	Ip             string
	VolumePerDay   uint64
	VolumePerCheck uint64
	Count          uint32
	Year           int
	Month          time.Month
	Day            int
	Hour           int
}
