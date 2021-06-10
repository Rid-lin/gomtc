package main

import (
	"bytes"
	"reflect"
	"testing"
)

var (
	cfgTest *Config        = newConfig()
	sshCred SSHCredentials = SSHCredentials{
		SSHHost:       cfgTest.MTAddr,
		SSHPort:       cfgTest.SSHPort,
		SSHUser:       cfgTest.MTUser,
		SSHPass:       cfgTest.MTPass,
		MaxSSHRetries: cfgTest.MaxSSHRetries,
		SSHRetryDelay: cfgTest.SSHRetryDelay,
	}
	qDef QuotaType = QuotaType{
		HourlyQuota:  cfgTest.DefaultQuotaHourly,
		DailyQuota:   cfgTest.DefaultQuotaDaily,
		MonthlyQuota: cfgTest.DefaultQuotaMonthly,
	}
	// block string = cfg.BlockGroup
)

func Test_getResponseOverSSHfMT(t *testing.T) {
	type args struct {
		SSHCred SSHCredentials
		command string
	}

	tests := []struct {
		name string
		args args
		want bytes.Buffer
	}{
		{
			name: "1 with exit",
			args: args{
				SSHCred: sshCred,
				command: "/ip dhcp-server lease print detail without-paging",
			},
			want: bytes.Buffer{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getResponseOverSSHfMT(tt.args.SSHCred, tt.args.command); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getResponseOverSSHfMT() = %v, want %v", got.String(), tt.want.String())
			}
		})
	}
}

func Test_saveDeviceToCSV(t *testing.T) {
	type args struct {
		devices []DeviceType
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := saveDeviceToCSV(tt.args.devices); (err != nil) != tt.wantErr {
				t.Errorf("saveDeviceToCSV() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_deviceToSlice(t *testing.T) {
	type args struct {
		d DeviceType
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.d.convertToSlice(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("deviceToSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestAliasType_send(t *testing.T) {
// 	type args struct {
// 		p        parseType
// 		qDefault QuotaType
// 	}
// 	tn := time.Now()
// 	tests := []struct {
// 		name    string
// 		a       InfoType
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name: "1",
// 			a: InfoType{
// 				InfoName: "E8:D8:D1:47:55:93",
// 				DeviceType: DeviceType{
// 					activeAddress:    "192.168.65.85",
// 					activeMacAddress: "E8:D8:D1:47:55:93",
// 					hostName:         "root-hp",
// 					addressLists:     "inet,Block",
// 					timeout:          tn,
// 				},
// 				PersonType: PersonType{
// 					Name:     "Vlad",
// 					Position: "Admin",
// 					Company:  "UTTiST",
// 					Comment:  "",
// 				},
// 				QuotaType: QuotaType{
// 					HourlyQuota:     600000000,
// 					DailyQuota:      6000000000,
// 					MonthlyQuota:    60000000000,
// 					Manual:          false,
// 					Blocked:         true,
// 					Disabled:        false,
// 					ShouldBeBlocked: true,
// 				},
// 			},
// 			args: args{
// 				p: parseType{
// 					SSHCredentials:   sshCred,
// 					QuotaType:        qDef,
// 					BlockAddressList: block,
// 					Location:         cfgTest.Location,
// 				},
// 				qDefault: qDef,
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.a.sendByAll(tt.args.p, tt.args.qDefault); (err != nil) != tt.wantErr {
// 				t.Errorf("AliasType.send() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

func Test_parseInfoFromMTAsValueToSlice(t *testing.T) {
	type args struct {
		p parseType
	}
	tests := []struct {
		name string
		args args
		want []DeviceType
	}{
		{
			name: "1",
			args: args{
				p: parseType{
					QuotaType:        qDef,
					BlockAddressList: "Block",
					SSHCredentials:   sshCred,
					Location:         cfg.Location,
				}},
			want: []DeviceType{},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseInfoFromMTAsValueToSlice(tt.args.p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseInfoFromMTAsValueToSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
