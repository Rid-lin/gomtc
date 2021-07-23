package memorystore

import (
	"database/sql"

	"git.vegner.org/vsvegner/gomtc/internal/app/model"
	"git.vegner.org/vsvegner/gomtc/internal/store"

	_ "github.com/mattn/go-sqlite3" //..
)

//Store ..
type Store struct {
	devicesStatRepository *DeviceStatRepository
}

//New ..
func New(db *sql.DB) *Store {
	return &Store{}
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
