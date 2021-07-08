package memorystore

import (
	"git.vegner.org/vsvegner/gomtc/internal/app/model"
	"git.vegner.org/vsvegner/gomtc/internal/store"
)

//Store ..
type Store struct {
	devicesStatRepository *DeviceStatRepository
}

//New ..
func New() *Store {
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
