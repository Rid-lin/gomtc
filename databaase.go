package main

import (
	"time"

	"git.vegner.org/vsvegner/gomtc/store"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

const (
	insertSQL = `
INSERT INTO stat (
	year, month, day, hour, size, login, ipaddress
) VALUES (
	?,?,?,?,?,?,?
)
`

	schemaSQL = `
	CREATE TABLE IF NOT EXISTS "stat" (
		"id"	INTEGER NOT NULL UNIQUE,
		"year"	INTEGER NOT NULL,
		"month"	INTEGER NOT NULL,
		"day"	INTEGER NOT NULL,
		"hour"	INTEGER NOT NULL,
		"size"	INTEGER NOT NULL,
		"login"	TEXT NOT NULL,
		"ipaddress"	TEXT NOT NULL,
		PRIMARY KEY("id" AUTOINCREMENT)
		);
`
)

func (t *Transport) SaveStatisticswithBuffer(fileName string, bufSize int) {
	// fmt.Print("Start SaveStatisticswithBuffer ")
	startTime := time.Now()
	lineAdded := 0
	db, err := store.NewDBStat(fileName, bufSize)
	if err != nil {
		log.Error(err)
		return
	}
	defer db.CloseStat()
	t.RLock()
	for _, year := range t.statofYears {
		for _, month := range year.monthsStat {
			for _, day := range month.daysStat {
				for _, d := range day.devicesStat {
					for i := range d.StatType.PerHour {
						if d.PerHour[i] == 0 {
							continue
						}
						err := db.AddStat(store.Stat{
							Year:      year.year,
							Month:     int(month.month),
							Day:       day.day,
							Hour:      i,
							Size:      d.PerHour[i],
							Login:     d.mac,
							Ipaddress: d.ip,
						})
						if err != nil {
							if err.Error() == "buffer is full" {
								if err2 := db.FlushStat(); err2 != nil {
									log.Error(err2)
									log.Error(err)
									continue
								}
							}
							log.Error(err)
						}
						lineAdded++
					}
				}
			}
		}
	}
	t.RUnlock()
	deltaTime := time.Since(startTime)
	// fmt.Print("SaveStatisticswithBuffer ended ")
	// fmt.Printf("Statistics save Execution time:%v speed:%v\n", deltaTime.Seconds(), float64(lineAdded)/deltaTime.Seconds())
	log.Debugf("Statistics save Execution time:%v speed:%v\n", deltaTime.Seconds(), float64(lineAdded)/deltaTime.Seconds())
}
