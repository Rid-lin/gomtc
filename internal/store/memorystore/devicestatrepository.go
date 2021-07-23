package memorystore

import (
	"sync"

	"git.vegner.org/vsvegner/gomtc/internal/app/model"
)

//DeviceRepository ..
type DeviceStatRepository struct {
	store *Store
	// stats map[model.KeyStat]*model.StatDay
	stats map[model.KeyStat][]*model.StatDevice
	sync.RWMutex
}

// Create ...
func (r *DeviceStatRepository) AddLine(sd *model.StatDevice) {
	key := model.KeyStat{
		Year:  sd.Year,
		Month: sd.Month,
		Day:   sd.Day,
	}
	stat := r.GetDayStat(key)
	r.Lock()
	for index := range stat {
		if sd.Ip == stat[index].Ip || sd.Mac == stat[index].Mac {
			stat[index].Count++
			stat[index].VolumePerDay += sd.VolumePerDay
			stat[index].VolumePerCheck += sd.VolumePerCheck
			for hour := range stat[index].PerHour {
				stat[index].PerHour[hour] += sd.PerHour[hour]
			}
			r.stats[key] = stat
			r.Unlock()
			return
		}
	}
	r.stats[key] = append(r.stats[key], sd)
	r.Unlock()
}

//FindByMac ..
func (r *DeviceStatRepository) GetDayStat(key model.KeyStat) []*model.StatDevice {
	r.Lock()
	defer r.Unlock()
	if r.stats[key] != nil {
		return r.stats[key]
	}
	statDay := []*model.StatDevice{}
	r.stats[key] = statDay
	return r.stats[key]
}
