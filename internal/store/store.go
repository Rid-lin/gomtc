package store

//Store ..
type Store interface {
	DeviceStat() DeviceStatRepository
	Log() LogRepository
}
