package main

import (
	"fmt"
	"time"

	"git.vegner.org/vsvegner/gomtc/store"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

const (
	insertSQL = `
INSERT INTO stat (
	date_str, year, month, day, hour, size, login, ipaddress
) VALUES (
	?,?,?,?,?,?,?,?
)
`

	schemaSQL = `
	CREATE TABLE IF NOT EXISTS "stat" (
		"id"	INTEGER NOT NULL UNIQUE,
		"date"	TEXT NOT NULL DEFAULT '1970-01-01',
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
	t.RLock()
	for _, year := range t.statofYears {
		for _, month := range year.monthsStat {
			for _, day := range month.daysStat {
				for _, d := range day.devicesStat {
					for i := range d.StatType.PerHour {
						var m, dayStr string
						if d.PerHour[i] == 0 {
							continue
						}
						if int(month.month) < 10 {
							m = "0"
						}
						if day.day < 10 {
							dayStr = "0"
						}
						m += fmt.Sprint(int(month.month))
						dayStr += fmt.Sprint(day.day)
						err := db.AddStat(store.Stat{
							Date:      fmt.Sprintf("%v-%v-%v", year.year, m, dayStr),
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
	if err := db.CloseStat(); err != nil {
		log.Errorf("unable to flush: %w", err)
	}
	deltaTime := time.Since(startTime)
	// fmt.Print("SaveStatisticswithBuffer ended ")
	// fmt.Printf("Statistics save Execution time:%v speed:%v\n", deltaTime.Seconds(), float64(lineAdded)/deltaTime.Seconds())
	log.Debugf("Statistics save Execution time:%v speed:%v\n", deltaTime.Seconds(), float64(lineAdded)/deltaTime.Seconds())
}

func GetDayStat(from, to string, fileName string) map[KeyDevice]StatDeviceType {
	devStats := map[KeyDevice]StatDeviceType{}
	db, err := store.OpenDB(fileName)
	if err != nil {
		return devStats
	}
	SQL := fmt.Sprintf(`SELECT ipaddress, login, sum(size), hour
	FROM stat
	WHERE date(date_str) BETWEEN date('%s') AND date('%s')
	GROUP BY login, hour
	ORDER BY sum(size) DESC;`, from, to)
	rows, err := db.Query(SQL)
	if err != nil {
		return devStats
	}
	for rows.Next() {
		var hour int
		var size uint64
		var ip, mac string
		var pHour [24]uint64
		err := rows.Scan(&ip, &mac, &size, &hour)
		if err != nil {
			log.Error(err)
			continue
		}
		stat, ok := devStats[KeyDevice{mac: mac}]
		if !ok {
			pHour[hour] = size
			stat = StatDeviceType{
				ip:  ip,
				mac: mac,
				StatType: StatType{
					PerHour:      pHour,
					VolumePerDay: size,
				},
			}
		} else {
			stat.PerHour[hour] += size
			stat.VolumePerDay += size
		}

		devStats[KeyDevice{mac: mac}] = stat
	}
	return devStats
}

func (t *Transport) DeletingDateData(date, fileName string) {
	db, err := store.OpenDB(fileName)
	if err != nil {
		return
	}
	defer db.Close()
	SQL := fmt.Sprintf("delete from stat where date_str = '%s'", date)
	result, err := db.Exec(SQL)
	if err != nil {
		log.Error("error to delete from table stat:%v", err)
		return
	}
	row, err := result.RowsAffected()
	log.Trace("result to delete from table stat:%v,%v", row, err)
}
