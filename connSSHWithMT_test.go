package main

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
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
				SSHCred: SSHCredentials{
					SSHHost: "192.168.65.1",
					SSHPort: "22",
					SSHUser: "getmac",
					SSHPass: "getmac_password",
				},
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

// func Test_parseInfoFromMTToSlice(t *testing.T) {
// 	type args struct {
// 		p parseType
// 	}

// 	Location, err := time.LoadLocation("Asia/Yekaterinburg")
// 	if err != nil {
// 		log.Errorf("Error loading Location(%v):%v", "Asia/Yekaterinburg", err)
// 		Location = time.UTC
// 	}

// 	tests := []struct {
// 		name string
// 		args args
// 		want []DeviceType
// 	}{
// 		{
// 			name: "1",
// 			args: args{
// 				p: parseType{
// 					QuotaType:        QuotaType{},
// 					BlockAddressList: "Block",
// 					SSHCredentials: SSHCredentials{
// 						SSHHost: "192.168.65.1",
// 						SSHPort: "22",
// 						SSHUser: "getmac",
// 						SSHPass: "getmac_password",
// 					},
// 					Location: Location,
// 				}},
// 			want: []DeviceType{},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := parseInfoFromMTToSlice(tt.args.p); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("parseInfoFromMTToSlice() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestDeviceType_parseLine(t *testing.T) {
// 	type args struct {
// 		l string
// 	}
// 	tests := []struct {
// 		name    string
// 		d       *DeviceType
// 		args    args
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.d.parseLine(tt.args.l); (err != nil) != tt.wantErr {
// 				t.Errorf("DeviceType.parseLine() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

func Test_parseInfoFromMTToSlice2(t *testing.T) {
	type args struct {
		p parseType
	}

	Location, err := time.LoadLocation("Asia/Yekaterinburg")
	if err != nil {
		log.Errorf("Error loading Location(%v):%v", "Asia/Yekaterinburg", err)
		Location = time.UTC
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
					QuotaType:        QuotaType{},
					BlockAddressList: "Block",
					SSHCredentials: SSHCredentials{
						SSHHost:       "192.168.65.1",
						SSHPort:       "22",
						SSHUser:       "getmac",
						SSHPass:       "getmac_password",
						MaxSSHRetries: 60,
						SSHRetryDelay: 3,
					},
					Location: Location,
				}},
			want: []DeviceType{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseInfoFromMTToSlice2(tt.args.p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseInfoFromMTToSlice2() = %v, want %v", got, tt.want)
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
