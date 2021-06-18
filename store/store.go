package store

// https://github.com/ardanlabs/python-go/blob/master/sqlite/trades/trades.go
// Package trades provides an SQLite based trades database.
// package trades

// Your main or test packages require this import so
// the sql package is properly initialized.
// _ "github.com/mattn/go-sqlite3"

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

const (
	insertSQL = `
INSERT INTO stat (
	date_str, year, month, day, hour, size, login, ipaddress
) VALUES (
	?,?,?,?,?,?,?,?
)
`
	deleteSQL = `
INSERT INTO stat (
	date_str, year, month, day, hour, size, login, ipaddress
) VALUES (
	?,?,?,?,?,?,?,?
)
`

	schemaSQL = `
CREATE TABLE IF NOT EXISTS "stat" (
	"id"	INTEGER NOT NULL UNIQUE,
	"date_str"	TEXT NOT NULL DEFAULT '1970-01-01',
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

// Stat is a buy/sell trade for symbol.
type Stat struct {
	Date                   string
	Year, Month, Day, Hour int
	Size                   uint64
	Login, Ipaddress       string
}

// DB is a database of stock trades.
type DB struct {
	sql    *sql.DB
	stmt   *sql.Stmt
	buffer []Stat
}

// NewDBStat constructs a Trades value for managing stock trades in a
// SQLite database. This API is not thread safe.
func NewDBStat(dbFile string, bufSize int) (*DB, error) {
	sqlDB, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}

	if _, err = sqlDB.Exec(schemaSQL); err != nil {
		return nil, err
	}

	stmt, err := sqlDB.Prepare(insertSQL)
	if err != nil {
		return nil, err
	}

	if bufSize <= 1024 {
		bufSize = 1024
	}
	db := DB{
		sql:    sqlDB,
		stmt:   stmt,
		buffer: make([]Stat, 0, bufSize),
	}
	return &db, nil
}

// NewDB constructs a Trades value for managing stock trades in a
// SQLite database. This API is not thread safe.
func OpenDB(dbFile string) (*sql.DB, error) {
	sqlDB, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return sqlDB, nil
}

// AddStat stores a stat into the buffer. Once the buffer is full, the
// stats are flushed to the database.
func (db *DB) AddStat(stat Stat) error {
	if len(db.buffer) == cap(db.buffer) {
		return errors.New("buffer is full")
	}

	db.buffer = append(db.buffer, stat)
	if len(db.buffer) == cap(db.buffer) {
		if err := db.FlushStat(); err != nil {
			return fmt.Errorf("unable to flush: %w", err)
		}
	}

	return nil
}

// FlushStat inserts pending trades into the database.
func (db *DB) FlushStat() error {
	tx, err := db.sql.Begin()
	if err != nil {
		return err
	}

	for _, stat := range db.buffer {
		_, err := tx.Stmt(db.stmt).Exec(stat.Date, stat.Year, stat.Month, stat.Day, stat.Hour, stat.Size, stat.Login, stat.Ipaddress)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	db.buffer = db.buffer[:0]
	return tx.Commit()
}

// CloseStat flushes all trades to the database and prevents any future trading.
func (db *DB) CloseStat() error {
	defer func() {
		db.stmt.Close()
		db.sql.Close()
	}()

	if err := db.FlushStat(); err != nil {
		return err
	}

	return nil
}
