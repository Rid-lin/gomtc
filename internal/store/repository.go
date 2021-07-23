package store

import "git.vegner.org/vsvegner/gomtc/internal/app/model"

type DeviceStatRepository interface {
	AddLine(*model.StatDevice)
	GetDayStat(model.KeyStat) []*model.StatDevice
}

type LogRepository interface {
	AddLine(*model.LogLine)
	Flush() error
	GetAll() []*model.LogLine
}
