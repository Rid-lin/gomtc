package main

import (
	"path"
	"time"
)

func (t *Transport) runOnce(cfg *Config) {
	p := parseType{}
	t.RLock()
	p.SSHCredentials = t.sshCredentials
	p.BlockAddressList = t.BlockAddressList
	p.QuotaType = t.QuotaType
	p.Location = t.Location
	t.RUnlock()

	t.readLog(cfg)

	t.getDevices()
	t.delOldData(t.newCount.LastDateNew, t.Location)
	t.parseAllFilesAndCountingTraffic(cfg)
	t.updateAliases(p)
	t.checkQuotas()
	t.BlockDevices()
	t.SendGroupStatus(cfg.NoControl)
	t.getDevices()

	t.SaveStatisticswithBuffer(path.Join(cfg.ConfigPath, "sqlite.db"), 1024*64)

	t.writeLog(cfg)
	t.newCount.Count = Count{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	t.setTimerParse(cfg.ParseDelay)
}

func (t *Transport) setTimerParse(IntervalStr string) {
	interval, err := time.ParseDuration(IntervalStr)
	if err != nil {
		t.timerParse = time.NewTimer(15 * time.Minute)
	} else {
		t.timerParse = time.NewTimer(interval)
	}
}
