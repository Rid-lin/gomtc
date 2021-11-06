package main

import (
	"git.vegner.org/vsvegner/gomtc/internal/app/model"
)

func GetDevicesFromRemote(p model.ParseType) []model.DeviceType {
	return GetDataOverApi(p)
}

func (a *BlockDevices) SendToBlockDevices(p model.ParseType) {
	BlockOverAPI(a, p)
}

// func parseInfoFromMTAsValueToSlice(p parseType) []DeviceType {
// 	devices := DevicesType{}
// 	b, err := GetResponseOverSSHfMTWithBuffer(p.SSHHost, p.SSHPort, p.SSHUser, p.SSHPass,
// 		":put [/ip dhcp-server lease print detail as-value]",
// 		p.MaxSSHRetries, int(p.SSHRetryDelay))
// 	if err != nil {
// 		return devices
// 	}
// 	devices.parseLeasePrintAsValue(b)
// 	return devices
// }

// func (a *BlockDevices) sendLeaseSet(p parseType) {
// 	var command string
// 	firstCommand := "/ip dhcp-server lease set "
// 	for _, item := range *a {
// 		if item.Id == "" {
// 			continue
// 		}
// 		itemCommand := fmt.Sprintf("number=%s disabled=%s address-lists=%s\n",
// 			item.Id, boolToParamert(item.Disabled), item.Groups)
// 		command = command + firstCommand + itemCommand
// 	}
// 	// For Debug
// 	_ = saveStrToFile("./config/command.temp", command)
// 	b := GetResponseOverSSHfMT(p.SSHHost, p.SSHPort, p.SSHUser, p.SSHPass, command)
// 	if b.Len() > 0 {
// 		log.Errorf("Error save device to Mikrotik(%v) with command:\n%v", b.String(), command)
// 	}
// }
