package main

// func (t *Transport) addLineOutToMapOfReportsOld(value *lineOfLogType, cfg *Config) {
// 	tm := time.Unix(value.timestamp, value.nsec)
// 	value.alias = determiningAlias(*value)
// 	key := KeyMapOfReports{
// 		DateStr: tm.Format(cfg.dateLayout),
// 		Alias:   value.alias,
// 	}
// 	_, ok := t.dataOld[key]
// 	if !ok {
// 		t.dataOld[key] = AliasOld{}
// 	}
// 	// Подсчёт трафика для пользователя и в определенный час
// 	t.trafficСountingOld(key, value)
// }

// func (t *Transport) AddLineToMapData(key KeyMapOfReports, value lineOfLogType) {
// 	var SizeOfHour [24]uint64
// 	t.Lock()
// 	SizeOfHour[value.hour] = value.sizeInBytes
// 	valueMapOfReports := AliasOld{
// 		Hits: 1,
// 		StatOldType: StatOldType{
// 			VolumePerHour: SizeOfHour,
// 			VolumePerDay:  value.sizeInBytes,
// 		},
// 	}
// 	t.dataOld[key] = valueMapOfReports
// 	t.Unlock()
// }

// func (t *Transport) trafficСountingOld(key KeyMapOfReports, value *lineOfLogType) {
// 	t.RLock()
// 	// Приваеваем данные в карте временной переменной для того чтобы предыдущие значения не потерялись
// 	valueMapOfReports := t.dataOld[key]
// 	t.RUnlock()
// 	// Расчет суммы трафика для дальшейшего отображения
// 	valueMapOfReports.VolumePerDay = valueMapOfReports.VolumePerDay + value.sizeInBytes
// 	valueMapOfReports.Hits++
// 	valueMapOfReports.HostName = value.hostname
// 	valueMapOfReports.Comments = value.comments
// 	VolumePerHour := valueMapOfReports.VolumePerHour
// 	VolumePerHour[value.hour] = VolumePerHour[value.hour] + value.sizeInBytes
// 	// Подсчёт окончен
// 	// Обработанные данные из временных переменных помещаем в карту....
// 	valueMapOfReports.VolumePerHour = VolumePerHour
// 	// .... блокируя её для записи во избежании коллизий
// 	valueMapOfReports.Alias = key.Alias
// 	valueMapOfReports.DateStr = key.DateStr
// 	t.Lock()
// 	t.dataOld[key] = valueMapOfReports
// 	t.Unlock()
// }

// func (t *Transport) updateAliases(p parseType) {
// 	t.Lock()
// 	for key, aliases := range t.AliasesOld {
// 		for index, alias := range aliases {
// 			if alias.InfoName == "" {
// 				alias.InfoName = key
// 			}
// 			infoD := t.devices.findDeviceToConvertInfoD(alias.InfoName, p.BlockAddressList, p.QuotaType)
// 			alias.DeviceType = infoD.convertToDevice(p.QuotaType)
// 			alias.PersonType = infoD.PersonType
// 			alias.QuotaType = infoD.QuotaType
// 			alias.QuotaType = checkNULLQuotas(alias.QuotaType, p.QuotaType)
// 			aliases[index] = alias
// 		}
// 		t.AliasesOld[key] = aliases
// 	}
// 	t.Unlock()
// }
