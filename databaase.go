package main

import (
	"database/sql"
	"fmt"
	"path"
	"time"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

func PrepareDB(fileName string) {
	db, err := sql.Open("sqlite3", path.Join(fileName))
	if err != nil {
		log.Error(err)
	}
	defer db.Close()
	PrepareSQL := `CREATE TABLE IF NOT EXIST "stat" (
	"id"	INTEGER NOT NULL UNIQUE,
	"year"	INTEGER NOT NULL,
	"month"	INTEGER NOT NULL,
	"day"	INTEGER NOT NULL,
	"minute"	INTEGER NOT NULL,
	"size"	INTEGER NOT NULL DEFAULT 0,
	"login"	TEXT NOT NULL,
	"ipaddress"	TEXT NOT NULL,
	"comment"	TEXT NOT NULL,
	PRIMARY KEY("id" AUTOINCREMENT)
	);`
	_, err = db.Exec(PrepareSQL)
	if err != nil {
		log.Error(err)
	}

}

func (t *Transport) SaveReport(fileName string) {
	db, err := sql.Open("sqlite3", path.Join(fileName))
	if err != nil {
		log.Error(err)
	}
	defer db.Close()
	t.RLock()
	for _, year := range t.statofYears {
		for _, month := range year.monthsStat {
			for _, day := range month.daysStat {
				tn := time.Now()
				if year.year == tn.In(cfg.Location).Year() && month.month == tn.In(cfg.Location).Month() && day.day == tn.In(cfg.Location).Day() {
					continue
				}
				for _, d := range day.devicesStat {
					if err := d.Write(year.year, int(month.month), day.day, db); err != nil {
						log.Errorf("Error write device's statistics to DB:%v", err)
					}
				}
			}
		}
	}
	t.RUnlock()

}

func (sd *StatDeviceType) Write(y, m, d int, db *sql.DB) error {

	result, err := db.Exec("INSERT INTO stat (year, month, day, minute, size, login, ipaddress, comment) values ('iPhone X', $1, $2)",
		"Apple", 72000)
	if err != nil {
		log.Error(err)
	}
	fmt.Println(result.LastInsertId()) // id последнего добавленного объекта
	fmt.Println(result.RowsAffected()) // количество добавленных строк
}
