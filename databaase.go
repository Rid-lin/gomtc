package main

import (
	"database/sql"
	"fmt"
	"path"
	"strings"
	"time"

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

func PrepareDB(fileName string) {
	log.Debug("Save to DB")
	db, err := sql.Open("sqlite3", path.Join(fileName))
	if err != nil {
		log.Error(err)
	}
	defer db.Close()

	_, err = db.Exec(schemaSQL)
	if err != nil {
		log.Error(err)
	}
}

func (t *Transport) SaveReport(fileName string, cfg *Config) {
	fmt.Print("Start save to DB ")
	db, err := sql.Open("sqlite3", path.Join(fileName))
	if err != nil {
		log.Error(err)
	}
	defer db.Close()
	lineAdded := uint64(0)
	startTime := time.Now()
	t.RLock()
	for _, year := range t.statofYears {
		for _, month := range year.monthsStat {
			for _, day := range month.daysStat {
				for _, d := range day.devicesStat {
					if err := d.Write(year.year, int(month.month), day.day, db, &lineAdded); err != nil {
						log.Errorf("Error write device's statistics to DB:%v", err)
					}
					// fmt.Printf("\r year:%v, month:%v, day:%v, lineAdded:%v", year.year, int(month.month), day.day, lineAdded)
				}
			}
		}
	}
	t.RUnlock()
	deltaTime := time.Since(startTime)
	fmt.Printf("Execution time:%v speed:%v ", deltaTime.Seconds(), float64(lineAdded)/deltaTime.Seconds())
	fmt.Print("Save ended\n")
	// log.Debug("Save ended")
}

func (sd *StatDeviceType) Write(y, m, d int, db *sql.DB, lineAdded *uint64) error {
	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for i := range sd.StatType.PerHour {
		if sd.PerHour[i] == 0 {
			continue
		}
		_, err := tx.Stmt(stmt).Exec(
			y, m, d, i,
			sd.PerHour[i],
			sd.mac,
			sd.ip)
		if err != nil {
			if err2 := tx.Rollback(); err != nil {
				return fmt.Errorf("Error exec(%v), error RollBack(%v)", err, err2)
			}
			return err
		}
		*lineAdded++
	}
	return tx.Commit()
}

func (t *Transport) SaveReport3(fileName string, cfg *Config) {
	fmt.Print("Start save to DB ")
	db, err := sql.Open("sqlite3", path.Join(fileName))
	if err != nil {
		log.Error(err)
	}
	defer db.Close()
	lineAdded := uint64(0)
	startTime := time.Now()
	t.RLock()
	for _, year := range t.statofYears {
		for _, month := range year.monthsStat {
			for _, day := range month.daysStat {
				for _, d := range day.devicesStat {
					if err := d.Write3(year.year, int(month.month), day.day, db, &lineAdded); err != nil {
						log.Errorf("Error write device's statistics to DB:%v", err)
					}
					// fmt.Printf("\r year:%v, month:%v, day:%v, lineAdded:%v", year.year, int(month.month), day.day, lineAdded)
				}
			}
		}
	}
	t.RUnlock()
	deltaTime := time.Since(startTime)
	fmt.Printf("Execution time:%v speed:%v ", deltaTime.Seconds(), float64(lineAdded)/deltaTime.Seconds())
	fmt.Print("Save3 ended\n")
	// log.Debug("Save ended")
}

func (sd *StatDeviceType) Write3(y, m, d int, db *sql.DB, lineAdded *uint64) error {
	SQL := `INSERT INTO stat (
		year, month, day, hour, size, login, ipaddress
	) VALUES `
	sqlAdd := ""
	for i := range sd.StatType.PerHour {
		if sd.PerHour[i] == 0 {
			continue
		}
		sqlAdd += fmt.Sprintf("('%v','%v','%v','%v','%v','%v','%v'),",
			y, m, d, i,
			sd.PerHour[i],
			sd.mac,
			sd.ip)
		*lineAdded++
	}
	for strings.Contains(sqlAdd, ",,") {
		sqlAdd = strings.ReplaceAll(sqlAdd, ",,", ",")
	}
	sqlAdd = strings.Trim(sqlAdd, ",")
	SQL += sqlAdd
	// print(SQL)
	_, err := db.Exec(SQL)
	if err != nil {
		return fmt.Errorf("SQL REquest(%v), error:%v", sqlAdd, err)
	}
	return nil
}

func (t *Transport) SaveReport2(fileName string, cfg *Config) {
	fmt.Print("Start save to DB ")
	db, err := sql.Open("sqlite3", path.Join(fileName))
	if err != nil {
		log.Error(err)
	}
	defer db.Close()
	lineAdded := uint64(0)
	startTime := time.Now()
	SQL := `INSERT INTO stat (
		year, month, day, hour, size, login, ipaddress
	) VALUES `
	sqlAdd := ""

	t.RLock()
	for _, year := range t.statofYears {
		for _, month := range year.monthsStat {
			for _, day := range month.daysStat {
				for _, d := range day.devicesStat {
					for i := range d.StatType.PerHour {
						if d.PerHour[i] == 0 {
							continue
						}
						sqlAdd += fmt.Sprintf("('%v','%v','%v','%v','%v','%v','%v'),",
							year.year, int(month.month), day.day, i,
							d.PerHour[i],
							d.mac,
							d.ip)
						lineAdded++
					}
				}
			}
		}
	}
	t.RUnlock()
	for strings.Contains(sqlAdd, ",,") {
		sqlAdd = strings.ReplaceAll(sqlAdd, ",,", ",")
	}
	sqlAdd = strings.Trim(sqlAdd, ",")
	SQL += sqlAdd
	// print(SQL)
	_, err2 := db.Exec(SQL)
	if err2 != nil {
		log.Errorf("SQL REquest(%v), error:%v", sqlAdd, err2)
	}
	deltaTime := time.Since(startTime)
	fmt.Printf("Execution time:%v speed:%v ", deltaTime.Seconds(), float64(lineAdded)/deltaTime.Seconds())
	fmt.Print("Save3 ended\n")
}
