package main

import (
	"time"

	"github.com/jinzhu/now"
)

func findOutTheCurrentDay(timestamp int64) int64 {
	myConfig := &now.Config{
		WeekStartDay: time.Monday,
		TimeLocation: Location,
	}
	return myConfig.With(time.Unix(timestamp, 0)).BeginningOfDay().Unix()

}

// func setLastdates(lastdate, lastday int64, cfg *Config) {
// 	cfg.LastDate = lastdate
// 	cfg.LastDay = lastday
// 	cfg.LastDayStr = time.Unix(lastday, 0).In(cfg.Location).Format(cfg.dateLayout)
// 	cfg.LastDateStr = time.Unix(lastdate, 0).In(cfg.Location).Format(cfg.dateLayout)
// }
