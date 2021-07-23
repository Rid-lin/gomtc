package memorystore

import (
	"database/sql"

	"git.vegner.org/vsvegner/gomtc/internal/app/model"
	"git.vegner.org/vsvegner/gomtc/internal/store"

	_ "github.com/mattn/go-sqlite3" //..
)

//Store ..
type Store struct {
	db                    *sql.DB
	devicesStatRepository *DeviceStatRepository
	logRepository         *LogRepository
}

//New ..
func New(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

//DeviceStat ..
func (s *Store) DeviceStat() store.DeviceStatRepository {
	if s.devicesStatRepository != nil {
		return s.devicesStatRepository
	}

	s.devicesStatRepository = &DeviceStatRepository{
		store: s,
		stats: make(map[model.KeyStat][]*model.StatDevice),
	}
	return s.devicesStatRepository
}

func (s *Store) Log() store.LogRepository {
	if s.logRepository != nil {
		return s.logRepository
	}

	s.logRepository = &LogRepository{
		store: s,
	}
	return s.logRepository
}
