package old

import (
	"database/sql"
	"fmt"
	"path"
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

func (t *Transport) SaveStatistics(fileName string) {
	fmt.Print("Start save4 to DB ")
	db, err := sql.Open("sqlite3", path.Join(fileName))
	if err != nil {
		log.Error(err)
	}
	defer db.Close()
	lineAdded := uint64(0)
	startTime := time.Now()
	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		log.Error(err)
	}
	defer stmt.Close()
	tx, err := db.Begin()
	if err != nil {
		log.Error(err)
	}

	t.RLock()
	for _, year := range t.statofYears {
		for _, month := range year.monthsStat {
			for _, day := range month.daysStat {
				for _, d := range day.devicesStat {
					for i := range d.StatType.PerHour {
						if d.PerHour[i] == 0 {
							continue
						}
						_, err := tx.Stmt(stmt).Exec(
							year.year, int(month.month), day.day, i,
							d.PerHour[i],
							d.mac,
							d.ip)
						if err != nil {
							if err2 := tx.Rollback(); err != nil {
								log.Errorf("Error exec(%v), error RollBack(%v)", err, err2)
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
	if err := tx.Commit(); err != nil {
		log.Error(err)
	}
	deltaTime := time.Since(startTime)
	fmt.Printf("Execution time:%v speed:%v ", deltaTime.Seconds(), float64(lineAdded)/deltaTime.Seconds())
	fmt.Print("Save4 ended\n")
}
