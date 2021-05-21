package main

import (
	"reflect"
	"testing"
	"time"
)

var (
	DS DevicesType
)

func Test_parseParamertToStr(t *testing.T) {
	type args struct {
		inpuStr string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "1",
			args: args{
				inpuStr: "test1=nari fiah fsrfarferfbui erf",
			},
			want: "nari fiah fsrfarferfbui erf",
		},
		{
			name: "2",
			args: args{
				inpuStr: "=test1=nari fiah fsrfarferfbui erf",
			},
			want: "nari fiah fsrfarferfbui erf",
		},
		{
			name: "3",
			args: args{
				inpuStr: "test1==nari fiah fsrfarferfbui erf",
			},
			want: "nari fiah fsrfarferfbui erf",
		},
		{
			name: "4",
			args: args{
				inpuStr: "test1=nari fiah fsrfarferfbui erf=",
			},
			want: "nari fiah fsrfarferfbui erf",
		},
		{
			name: "5",
			args: args{
				inpuStr: "=test1==nari fiah fsrfarferfbui erf=",
			},
			want: "nari fiah fsrfarferfbui erf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseParamertToStr(tt.args.inpuStr); got != tt.want {
				t.Errorf("Test(%s) got ='%v', want '%v'", tt.name, got, tt.want)
			}
		})
	}
}

func Test_parseParamertToUint(t *testing.T) {
	type args struct {
		inputValue string
	}
	tests := []struct {
		name      string
		args      args
		wantQuota uint64
	}{
		{
			name: "1",
			args: args{
				inputValue: "q=5000",
			},
			wantQuota: 5000,
		},
		{
			name: "2",
			args: args{
				inputValue: "q=50000000",
			},
			wantQuota: 50000000,
		},
		{
			name: "3",
			args: args{
				inputValue: "q=9000000000000000000",
			},
			wantQuota: 9000000000000000000,
		},
		{
			name: "4",
			args: args{
				inputValue: "q=900000000.0000000000",
			},
			wantQuota: 900000000,
		},
		{
			name: "5",
			args: args{
				inputValue: "q=68945264502789.867308976308973689",
			},
			wantQuota: 68945264502789,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotQuota := parseParamertToUint(tt.args.inputValue); gotQuota != tt.wantQuota {
				t.Errorf("Test(%s) got = %v, want %v", tt.name, gotQuota, tt.wantQuota)
			}
		})
	}
}

func Test_paramertToUint(t *testing.T) {
	type args struct {
		inputValue string
	}
	tests := []struct {
		name      string
		args      args
		wantQuota uint64
	}{
		{
			name: "1",
			args: args{
				inputValue: "5000",
			},
			wantQuota: 5000,
		},
		{
			name: "2",
			args: args{
				inputValue: "50000000",
			},
			wantQuota: 50000000,
		},
		{
			name: "3",
			args: args{
				inputValue: "9000000000000000000",
			},
			wantQuota: 9000000000000000000,
		},
		{
			name: "4",
			args: args{
				inputValue: "900000000.0000000000",
			},
			wantQuota: 900000000,
		},
		{
			name: "5",
			args: args{
				inputValue: "68945264502789.867308976308973689",
			},
			wantQuota: 68945264502789,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotQuota := paramertToUint(tt.args.inputValue); gotQuota != tt.wantQuota {
				t.Errorf("paramertToUint() = %v, want %v", gotQuota, tt.wantQuota)
			}
		})
	}
}

func Test_parseParamertToBool(t *testing.T) {
	type args struct {
		inputValue string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "1",
			args: args{
				inputValue: "q=68945264502789.867308976308973689",
			},
			want: false,
		},
		{
			name: "2",
			args: args{
				inputValue: "hjkhkjh=yes",
			},
			want: true,
		},
		{
			name: "3",
			args: args{
				inputValue: "hjkhkjh=true",
			},
			want: true,
		},
		{
			name: "4",
			args: args{
				inputValue: "hjkhkjh=true1",
			},
			want: false,
		},
		{
			name: "5",
			args: args{
				inputValue: "hjkhkjh=yеs", //Russian 'е'
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseParamertToBool(tt.args.inputValue); got != tt.want {
				t.Errorf("Test(%s) parseParamertToBool() = '%v', want '%v'", tt.name, got, tt.want)
			}
		})
	}
}

func Test_paramertToBool(t *testing.T) {
	type args struct {
		inputValue string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "1",
			args: args{
				inputValue: "68945264502789.867308976308973689",
			},
			want: false,
		},
		{
			name: "2",
			args: args{
				inputValue: "yes",
			},
			want: true,
		},
		{
			name: "3",
			args: args{
				inputValue: "true",
			},
			want: true,
		},
		{
			name: "4",
			args: args{
				inputValue: "true1",
			},
			want: false,
		},
		{
			name: "5",
			args: args{
				inputValue: "yеs", //Russian 'е'
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := paramertToBool(tt.args.inputValue); got != tt.want {
				t.Errorf("Test(%s) paramertToBool() = '%v', want '%v'", tt.name, got, tt.want)
			}
		})
	}
}

func Test_makeCommentFromIodt(t *testing.T) {
	type args struct {
		d     InfoOfDeviceType
		quota QuotaType
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "vlad",
			args: args{
				d: InfoOfDeviceType{
					QuotaType: QuotaType{
						HourlyQuota:  500000000,
						DailyQuota:   50000000000,
						MonthlyQuota: 0,
						Manual:       true,
					},
					PersonType: PersonType{
						Name:     "Vlad",
						Position: "Admin",
						Company:  "Home",
						TypeD:    "nb",
						IDUser:   "33785",
						Comment:  "interesnaya fignya",
					},
				},
			},
			want: "nb=Vlad/id=33785/com=Home/pos=Admin/quotahourly=500000000/quotadaily=50000000000/manual=true/comment=interesnaya fignya",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeCommentFromIodt(tt.args.d, tt.args.quota); got != tt.want {
				t.Errorf("makeCommentFromIodt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseComment(t *testing.T) {
	type args struct {
		comment string
	}
	tests := []struct {
		name             string
		args             args
		wantQuotahourly  uint64
		wantQuotadaily   uint64
		wantQuotamonthly uint64
		wantName         string
		wantPosition     string
		wantCompany      string
		wantTypeD        string
		wantIDUser       string
		wantComment      string
		wantManual       bool
	}{
		{
			name: "vlad",
			args: args{
				comment: "nb=Vlad/id=33785/com=Home/pos=Admin/quotahourly=500000000/quotadaily=50000000000/manual=true/comment=interesnaya fignya",
			},
			wantQuotahourly:  500000000,
			wantQuotadaily:   50000000000,
			wantQuotamonthly: 0,
			wantName:         "Vlad",
			wantPosition:     "Admin",
			wantCompany:      "Home",
			wantTypeD:        "nb",
			wantIDUser:       "33785",
			wantComment:      "interesnaya fignya",
			wantManual:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuotahourly, gotQuotadaily, gotQuotamonthly, gotName, gotPosition, gotCompany, gotTypeD, gotIDUser, gotComment, gotManual := parseComment(tt.args.comment)
			if gotQuotahourly != tt.wantQuotahourly {
				t.Errorf("Test(%s) parseComment() gotQuotahourly = %v, want %v", tt.name, gotQuotahourly, tt.wantQuotahourly)
			}
			if gotQuotadaily != tt.wantQuotadaily {
				t.Errorf("Test(%s) parseComment() gotQuotadaily = %v, want %v", tt.name, gotQuotadaily, tt.wantQuotadaily)
			}
			if gotQuotamonthly != tt.wantQuotamonthly {
				t.Errorf("Test(%s) parseComment() gotQuotamonthly = %v, want %v", tt.name, gotQuotamonthly, tt.wantQuotamonthly)
			}
			if gotName != tt.wantName {
				t.Errorf("Test(%s) parseComment() gotName = %v, want %v", tt.name, gotName, tt.wantName)
			}
			if gotPosition != tt.wantPosition {
				t.Errorf("Test(%s) parseComment() gotPosition = %v, want %v", tt.name, gotPosition, tt.wantPosition)
			}
			if gotCompany != tt.wantCompany {
				t.Errorf("Test(%s) parseComment() gotCompany = %v, want %v", tt.name, gotCompany, tt.wantCompany)
			}
			if gotTypeD != tt.wantTypeD {
				t.Errorf("Test(%s) parseComment() gotTypeD = %v, want %v", tt.name, gotTypeD, tt.wantTypeD)
			}
			if gotIDUser != tt.wantIDUser {
				t.Errorf("Test(%s) parseComment() gotIDUser = %v, want %v", tt.name, gotIDUser, tt.wantIDUser)
			}
			if gotComment != tt.wantComment {
				t.Errorf("Test(%s) parseComment() gotComment = %v, want %v", tt.name, gotComment, tt.wantComment)
			}
			if gotManual != tt.wantManual {
				t.Errorf("Test(%s) parseComment() gotManual = %v, want %v", tt.name, gotManual, tt.wantManual)
			}
		})
	}
}

func TestDeviceType_convertToInfo(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		d    DeviceType
		want InfoOfDeviceType
	}{
		{
			name: "vlad",
			d: DeviceType{
				activeAddress:    "192.168.65.85",
				activeClientId:   "1:e8:d8:d1:47:55:93",
				activeMacAddress: "E8:D8:D1:47:55:93",
				activeServer:     "dhcp_lan",
				address:          "pool_admin",
				addressLists:     "inet",
				blocked:          "false",
				clientId:         "1:e8:d8:d1:47:55:93",
				comment:          "nb=Vlad/id=33785/com=Home/pos=Admin/quotahourly=500000000/quotadaily=50000000000/manual=true/comment=interesnaya fignya",
				dhcpOption:       "",
				disabled:         "false",
				dynamic:          "false",
				expiresAfter:     "6m32s",
				hostName:         "root-hp",
				lastSeen:         "3m28s",
				macAddress:       "E8:D8:D1:47:55:93",
				radius:           "false",
				server:           "dhcp_lan",
				status:           "bound",
				Manual:           true,
				ShouldBeBlocked:  false,
				timeout:          now,
			},
			want: InfoOfDeviceType{
				QuotaType: QuotaType{
					HourlyQuota:     500000000,
					DailyQuota:      50000000000,
					MonthlyQuota:    0,
					Disabled:        false,
					Blocked:         false,
					Manual:          true,
					ShouldBeBlocked: false,
				},
				PersonType: PersonType{
					Comment:  "interesnaya fignya",
					Comments: "nb=Vlad/id=33785/com=Home/pos=Admin/quotahourly=500000000/quotadaily=50000000000/manual=true/comment=interesnaya fignya",
					Name:     "Vlad",
					Position: "Admin",
					Company:  "Home",
					IDUser:   "33785",
					TypeD:    "nb",
				},
				DeviceOldType: DeviceOldType{
					Id:       "",
					IP:       "192.168.65.85",
					Mac:      "E8:D8:D1:47:55:93",
					AMac:     "E8:D8:D1:47:55:93",
					HostName: "root-hp",
					Groups:   "inet",
					timeout:  now,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.convertToInfo(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Test(%s) DeviceType.convertToInfo() = \n%#v, \nwant \n%#v", tt.name, got, tt.want)
			}
		})
	}
}

func TestInfoOfDeviceType_convertToDevice(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name  string
		dInfo *InfoOfDeviceType
		want  DeviceType
	}{
		{
			name: "vlad",
			want: DeviceType{
				activeAddress:    "192.168.65.85",
				activeClientId:   "1:e8:d8:d1:47:55:93",
				activeMacAddress: "E8:D8:D1:47:55:93",
				activeServer:     "dhcp_lan",
				address:          "pool_admin",
				addressLists:     "inet",
				blocked:          "false",
				clientId:         "1:e8:d8:d1:47:55:93",
				comment:          "nb=Vlad/id=33785/com=Home/pos=Admin/quotahourly=500000000/quotadaily=50000000000/manual=true/comment=interesnaya fignya",
				dhcpOption:       "",
				disabled:         "false",
				dynamic:          "false",
				expiresAfter:     "6m32s",
				hostName:         "root-hp",
				lastSeen:         "3m28s",
				macAddress:       "E8:D8:D1:47:55:93",
				radius:           "false",
				server:           "dhcp_lan",
				status:           "bound",
				Manual:           true,
				ShouldBeBlocked:  false,
				timeout:          now,
			},
			dInfo: &InfoOfDeviceType{
				QuotaType: QuotaType{
					HourlyQuota:     500000000,
					DailyQuota:      50000000000,
					MonthlyQuota:    0,
					Disabled:        false,
					Blocked:         false,
					Manual:          true,
					ShouldBeBlocked: false,
				},
				PersonType: PersonType{
					Comment:  "interesnaya fignya",
					Comments: "nb=Vlad/id=33785/com=Home/pos=Admin/quotahourly=500000000/quotadaily=50000000000/manual=true/comment=interesnaya fignya",
					Name:     "Vlad",
					Position: "Admin",
					Company:  "Home",
					IDUser:   "33785",
					TypeD:    "nb",
				},
				DeviceOldType: DeviceOldType{
					Id:       "",
					IP:       "192.168.65.85",
					Mac:      "E8:D8:D1:47:55:93",
					AMac:     "E8:D8:D1:47:55:93",
					HostName: "root-hp",
					Groups:   "inet",
					timeout:  now,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dInfo.convertToDevice(); !tt.want.compare(&got) {
				t.Errorf("Test(%s) InfoOfDeviceType.convertToDevice() = \n%#v, \nwant \n%#v", tt.name, got, tt.want)
			}
		})
	}
}

func TestDeviceType_compare(t *testing.T) {
	type args struct {
		d2 *DeviceType
	}
	tests := []struct {
		name string
		d1   *DeviceType
		args args
		want bool
	}{
		{
			name: "1",
			d1: &DeviceType{
				activeClientId:   "1:e8:d8:d1:47:55:93",
				activeMacAddress: "E8:D8:D1:47:55:93",
				clientId:         "1:e8:d8:d1:47:55:93",
				macAddress:       "E8:D8:D1:47:55:93",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "1:e8:d8:d1:47:55:93",
					activeMacAddress: "E8:D8:D1:47:55:93",
					clientId:         "1:e8:d8:d1:47:55:93",
					macAddress:       "E8:D8:D1:47:55:93",
				},
			},
			want: true,
		},
		{
			name: "2",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "",
				clientId:         "",
				macAddress:       "E8:D8:D1:47:55:93",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "",
					clientId:         "",
					macAddress:       "E8:D8:D1:47:55:93",
				},
			},
			want: true,
		},
		{
			name: "3",
			d1: &DeviceType{
				activeClientId:   "1:E8:D8:D1:47:55:93",
				activeMacAddress: "",
				clientId:         "",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "1:E8:D8:D1:47:55:93",
					activeMacAddress: "",
					clientId:         "",
					macAddress:       "",
				},
			},
			want: true,
		},
		{
			name: "4",
			d1: &DeviceType{
				activeClientId:   "1:E8:D8:D1:47:55:93",
				activeMacAddress: "",
				clientId:         "",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "E8:D8:D1:47:55:93",
					clientId:         "",
					macAddress:       "",
				},
			},
			want: false,
		},
		{
			name: "5",
			d1: &DeviceType{
				activeClientId:   "1:E8:D8:D1:47:55:93",
				activeMacAddress: "",
				clientId:         "",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "",
					clientId:         "1:E8:D8:D1:47:55:93",
					macAddress:       "",
				},
			},
			want: true,
		},
		{
			name: "6",
			d1: &DeviceType{
				activeClientId:   "1:E8:D8:D1:47:55:93",
				activeMacAddress: "",
				clientId:         "",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "",
					clientId:         "",
					macAddress:       "E8:D8:D1:47:55:93",
				},
			},
			want: false,
		},
		{
			name: "7",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "E8:D8:D1:47:55:93",
				clientId:         "",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "1:E8:D8:D1:47:55:93",
					activeMacAddress: "",
					clientId:         "",
					macAddress:       "",
				},
			},
			want: false,
		},
		{
			name: "8",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "E8:D8:D1:47:55:93",
				clientId:         "",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "E8:D8:D1:47:55:93",
					clientId:         "",
					macAddress:       "",
				},
			},
			want: true,
		},
		{
			name: "9",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "E8:D8:D1:47:55:93",
				clientId:         "",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "",
					clientId:         "1:E8:D8:D1:47:55:93",
					macAddress:       "",
				},
			},
			want: false,
		},
		{
			name: "10",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "E8:D8:D1:47:55:93",
				clientId:         "",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "",
					clientId:         "",
					macAddress:       "E8:D8:D1:47:55:93",
				},
			},
			want: true,
		},
		{
			name: "11",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "",
				clientId:         "1:E8:D8:D1:47:55:93",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "1:E8:D8:D1:47:55:93",
					activeMacAddress: "",
					clientId:         "",
					macAddress:       "",
				},
			},
			want: true,
		},
		{
			name: "12",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "",
				clientId:         "1:E8:D8:D1:47:55:93",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "E8:D8:D1:47:55:93",
					clientId:         "",
					macAddress:       "",
				},
			},
			want: false,
		},
		{
			name: "13",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "",
				clientId:         "1:E8:D8:D1:47:55:93",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "",
					clientId:         "1:E8:D8:D1:47:55:93",
					macAddress:       "",
				},
			},
			want: true,
		},
		{
			name: "14",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "",
				clientId:         "1:E8:D8:D1:47:55:93",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "",
					clientId:         "",
					macAddress:       "E8:D8:D1:47:55:93",
				},
			},
			want: false,
		},
		{
			name: "15",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "",
				clientId:         "",
				macAddress:       "E8:D8:D1:47:55:93",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "1:E8:D8:D1:47:55:93",
					activeMacAddress: "",
					clientId:         "",
					macAddress:       "",
				},
			},
			want: false,
		},
		{
			name: "16",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "",
				clientId:         "",
				macAddress:       "E8:D8:D1:47:55:93",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "E8:D8:D1:47:55:93",
					clientId:         "",
					macAddress:       "",
				},
			},
			want: true,
		},
		{
			name: "17",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "",
				clientId:         "",
				macAddress:       "E8:D8:D1:47:55:93",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "",
					clientId:         "1:E8:D8:D1:47:55:93",
					macAddress:       "",
				},
			},
			want: false,
		},
		{
			name: "18",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "",
				clientId:         "",
				macAddress:       "E8:D8:D1:47:55:93",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "",
					clientId:         "",
					macAddress:       "E8:D8:D1:47:55:93",
				},
			},
			want: true,
		},
		{
			name: "1f",
			d1: &DeviceType{
				activeClientId:   "1:e8:d8:d1:47:55:93",
				activeMacAddress: "E8:D8:D1:47:55:93",
				clientId:         "1:e8:d8:d1:47:55:93",
				macAddress:       "E8:D8:D1:47:55:93",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "1:e8:d8:d1:47:55:94",
					activeMacAddress: "E8:D8:D1:47:55:94",
					clientId:         "1:e8:d8:d1:47:55:94",
					macAddress:       "E8:D8:D1:47:55:94",
				},
			},
			want: false,
		},
		{
			name: "2f",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "",
				clientId:         "",
				macAddress:       "E8:D8:D1:47:55:93",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "",
					clientId:         "",
					macAddress:       "E8:D8:D1:47:55:94",
				},
			},
			want: false,
		},
		{
			name: "3f",
			d1: &DeviceType{
				activeClientId:   "1:E8:D8:D1:47:55:93",
				activeMacAddress: "",
				clientId:         "",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "1:E8:D8:D1:47:55:9f",
					activeMacAddress: "",
					clientId:         "",
					macAddress:       "",
				},
			},
			want: false,
		},
		{
			name: "4f",
			d1: &DeviceType{
				activeClientId:   "1:E8:D8:D1:47:55:93",
				activeMacAddress: "",
				clientId:         "",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "E8:D8:D1:47:55:9f",
					clientId:         "",
					macAddress:       "",
				},
			},
			want: false,
		},
		{
			name: "5f",
			d1: &DeviceType{
				activeClientId:   "1:E8:D8:D1:47:55:9f",
				activeMacAddress: "",
				clientId:         "",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "",
					clientId:         "1:E8:D8:D1:47:55:93",
					macAddress:       "",
				},
			},
			want: false,
		},
		{
			name: "6f",
			d1: &DeviceType{
				activeClientId:   "1:E8:D8:D1:47:55:9f",
				activeMacAddress: "",
				clientId:         "",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "",
					clientId:         "",
					macAddress:       "E8:D8:D1:47:55:93",
				},
			},
			want: false,
		},
		{
			name: "7f",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "E8:D8:D1:47:55:9f",
				clientId:         "",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "1:E8:D8:D1:47:55:93",
					activeMacAddress: "",
					clientId:         "",
					macAddress:       "",
				},
			},
			want: false,
		},
		{
			name: "8f",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "E8:D8:D1:47:55:9f",
				clientId:         "",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "E8:D8:D1:47:55:93",
					clientId:         "",
					macAddress:       "",
				},
			},
			want: false,
		},
		{
			name: "9f",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "E8:D8:D1:47:55:9f",
				clientId:         "",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "",
					clientId:         "1:E8:D8:D1:47:55:93",
					macAddress:       "",
				},
			},
			want: false,
		},
		{
			name: "10f",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "E8:D8:D1:47:55:9f",
				clientId:         "",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "",
					clientId:         "",
					macAddress:       "E8:D8:D1:47:55:93",
				},
			},
			want: false,
		},
		{
			name: "11f",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "",
				clientId:         "1:E8:D8:D1:47:55:9f",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "1:E8:D8:D1:47:55:93",
					activeMacAddress: "",
					clientId:         "",
					macAddress:       "",
				},
			},
			want: false,
		},
		{
			name: "12f",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "",
				clientId:         "1:E8:D8:D1:47:55:9f",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "E8:D8:D1:47:55:93",
					clientId:         "",
					macAddress:       "",
				},
			},
			want: false,
		},
		{
			name: "13f",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "",
				clientId:         "1:E8:D8:D1:47:55:9f",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "",
					clientId:         "1:E8:D8:D1:47:55:93",
					macAddress:       "",
				},
			},
			want: false,
		},
		{
			name: "14f",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "",
				clientId:         "1:E8:D8:D1:47:55:9f",
				macAddress:       "",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "",
					clientId:         "",
					macAddress:       "E8:D8:D1:47:55:93",
				},
			},
			want: false,
		},
		{
			name: "15f",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "",
				clientId:         "",
				macAddress:       "E8:D8:D1:47:55:9f",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "1:E8:D8:D1:47:55:93",
					activeMacAddress: "",
					clientId:         "",
					macAddress:       "",
				},
			},
			want: false,
		},
		{
			name: "16f",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "",
				clientId:         "",
				macAddress:       "E8:D8:D1:47:55:9f",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "E8:D8:D1:47:55:93",
					clientId:         "",
					macAddress:       "",
				},
			},
			want: false,
		},
		{
			name: "17f",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "",
				clientId:         "",
				macAddress:       "E8:D8:D1:47:55:9f",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "",
					clientId:         "1:E8:D8:D1:47:55:93",
					macAddress:       "",
				},
			},
			want: false,
		},
		{
			name: "18f",
			d1: &DeviceType{
				activeClientId:   "",
				activeMacAddress: "",
				clientId:         "",
				macAddress:       "E8:D8:D1:47:55:9f",
			},
			args: args{
				d2: &DeviceType{
					activeClientId:   "",
					activeMacAddress: "",
					clientId:         "",
					macAddress:       "E8:D8:D1:47:55:93",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d1.compare(tt.args.d2); got != tt.want {
				t.Errorf("test(%v) DeviceType.compare() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestDevicesType_find(t *testing.T) {
	now := time.Now()
	DS = DevicesType{
		{
			activeAddress:    "192.168.65.86",
			activeClientId:   "1:e8:d8:d1:47:55:96",
			activeMacAddress: "E8:D8:D1:47:55:96",
			activeServer:     "dhcp_lan",
			address:          "pool_admin",
			addressLists:     "inet",
			blocked:          "false",
			clientId:         "1:e8:d8:d1:47:55:96",
			comment:          "nb=Vlad/id=33786/com=Home/pos=Admin/quotahourly=500000000/quotadaily=50000000000/manual=true/comment=interesnaya fignya",
			dhcpOption:       "",
			disabled:         "false",
			dynamic:          "false",
			expiresAfter:     "6m32s",
			hostName:         "root-hp6",
			lastSeen:         "3m28s",
			macAddress:       "E8:D8:D1:47:55:96",
			radius:           "false",
			server:           "dhcp_lan",
			status:           "bound",
			Manual:           true,
			ShouldBeBlocked:  false,
			timeout:          now,
		},
		{
			activeAddress:    "192.168.65.87",
			activeClientId:   "1:e8:d8:d1:47:55:97",
			activeMacAddress: "E8:D8:D1:47:55:97",
			activeServer:     "dhcp_lan",
			address:          "pool_admin",
			addressLists:     "inet",
			blocked:          "false",
			clientId:         "1:e8:d8:d1:47:55:97",
			comment:          "nb=Vlad/id=33787/com=Home/pos=Admin/quotahourly=500000000/quotadaily=50000000000/manual=true/comment=interesnaya fignya",
			dhcpOption:       "",
			disabled:         "false",
			dynamic:          "false",
			expiresAfter:     "6m32s",
			hostName:         "root-hp7",
			lastSeen:         "3m28s",
			macAddress:       "E8:D8:D1:47:55:97",
			radius:           "false",
			server:           "dhcp_lan",
			status:           "bound",
			Manual:           true,
			ShouldBeBlocked:  false,
			timeout:          now,
		},
		{
			activeAddress:    "192.168.65.85",
			activeClientId:   "1:e8:d8:d1:47:55:93",
			activeMacAddress: "E8:D8:D1:47:55:93",
			activeServer:     "dhcp_lan",
			address:          "pool_admin",
			addressLists:     "inet",
			blocked:          "false",
			clientId:         "1:e8:d8:d1:47:55:93",
			comment:          "nb=Vlad/id=33785/com=Home/pos=Admin/quotahourly=500000000/quotadaily=50000000000/manual=true/comment=interesnaya fignya",
			dhcpOption:       "",
			disabled:         "false",
			dynamic:          "false",
			expiresAfter:     "6m32s",
			hostName:         "root-hp",
			lastSeen:         "3m28s",
			macAddress:       "E8:D8:D1:47:55:93",
			radius:           "false",
			server:           "dhcp_lan",
			status:           "bound",
			Manual:           true,
			ShouldBeBlocked:  false,
			timeout:          now,
		},
	}
	type args struct {
		d *DeviceType
	}
	tests := []struct {
		name string
		ds   *DevicesType
		args args
		want int
	}{
		{
			name: "1",
			ds:   &DS,
			args: args{
				d: &DeviceType{
					activeAddress:    "192.168.65.86",
					activeClientId:   "1:e8:d8:d1:47:55:96",
					activeMacAddress: "E8:D8:D1:47:55:96",
					activeServer:     "dhcp_lan",
					address:          "pool_admin",
					addressLists:     "inet",
					blocked:          "false",
					clientId:         "1:e8:d8:d1:47:55:96",
					comment:          "nb=Vlad/id=33786/com=Home/pos=Admin/quotahourly=500000000/quotadaily=50000000000/manual=true/comment=interesnaya fignya",
					dhcpOption:       "",
					disabled:         "false",
					dynamic:          "false",
					expiresAfter:     "6m32s",
					hostName:         "root-hp6",
					lastSeen:         "3m28s",
					macAddress:       "E8:D8:D1:47:55:96",
					radius:           "false",
					server:           "dhcp_lan",
					status:           "bound",
					Manual:           true,
					ShouldBeBlocked:  false,
					timeout:          now,
				},
			},
			want: 0,
		},
		{
			name: "2",
			ds:   &DS,
			args: args{d: &DeviceType{
				activeAddress:    "192.168.65.87",
				activeClientId:   "1:e8:d8:d1:47:55:97",
				activeMacAddress: "E8:D8:D1:47:55:97",
				activeServer:     "dhcp_lan",
				address:          "pool_admin",
				addressLists:     "inet",
				blocked:          "false",
				clientId:         "1:e8:d8:d1:47:55:97",
				comment:          "nb=Vlad/id=33787/com=Home/pos=Admin/quotahourly=500000000/quotadaily=50000000000/manual=true/comment=interesnaya fignya",
				dhcpOption:       "",
				disabled:         "false",
				dynamic:          "false",
				expiresAfter:     "6m32s",
				hostName:         "root-hp7",
				lastSeen:         "3m28s",
				macAddress:       "E8:D8:D1:47:55:97",
				radius:           "false",
				server:           "dhcp_lan",
				status:           "bound",
				Manual:           true,
				ShouldBeBlocked:  false,
				timeout:          now,
			}},
			want: 1,
		},
		{
			name: "3",
			ds:   &DS,
			args: args{d: &DeviceType{
				activeAddress:    "192.168.65.85",
				activeClientId:   "1:e8:d8:d1:47:55:93",
				activeMacAddress: "E8:D8:D1:47:55:93",
				activeServer:     "dhcp_lan",
				address:          "pool_admin",
				addressLists:     "inet",
				blocked:          "false",
				clientId:         "1:e8:d8:d1:47:55:93",
				comment:          "nb=Vlad/id=33785/com=Home/pos=Admin/quotahourly=500000000/quotadaily=50000000000/manual=true/comment=interesnaya fignya",
				dhcpOption:       "",
				disabled:         "false",
				dynamic:          "false",
				expiresAfter:     "6m32s",
				hostName:         "root-hp",
				lastSeen:         "3m28s",
				macAddress:       "E8:D8:D1:47:55:93",
				radius:           "false",
				server:           "dhcp_lan",
				status:           "bound",
				Manual:           true,
				ShouldBeBlocked:  false,
				timeout:          now,
			}},
			want: 2,
		},
		{
			name: "3",
			ds:   &DS,
			args: args{d: &DeviceType{
				activeMacAddress: "E8:D8:D1:47:55:93",
				activeServer:     "dhcp_lan",
				address:          "pool_admin",
				addressLists:     "inet",
				blocked:          "false",
				clientId:         "1:e8:d8:d1:47:55:93",
				comment:          "nb=Vlad/id=33785/com=Home/pos=Admin/quotahourly=500000000/quotadaily=50000000000/manual=true/comment=interesnaya fignya",
				dhcpOption:       "",
				disabled:         "false",
				dynamic:          "false",
				expiresAfter:     "6m32s",
				hostName:         "root-hp",
				lastSeen:         "3m28s",
				macAddress:       "E8:D8:D1:47:55:93",
				radius:           "false",
				server:           "dhcp_lan",
				status:           "bound",
				Manual:           true,
				ShouldBeBlocked:  false,
				timeout:          now,
			}},
			want: 2,
		},
		{
			name: "3",
			ds:   &DS,
			args: args{d: &DeviceType{
				clientId:        "1:e8:d8:d1:47:55:93",
				comment:         "nb=Vlad/id=33785/com=Home/pos=Admin/quotahourly=500000000/quotadaily=50000000000/manual=true/comment=interesnaya fignya",
				dhcpOption:      "",
				disabled:        "false",
				dynamic:         "false",
				expiresAfter:    "6m32s",
				hostName:        "root-hp",
				lastSeen:        "3m28s",
				macAddress:      "E8:D8:D1:47:55:93",
				radius:          "false",
				server:          "dhcp_lan",
				status:          "bound",
				Manual:          true,
				ShouldBeBlocked: false,
				timeout:         now,
			}},
			want: 2,
		},
		{
			name: "3",
			ds:   &DS,
			args: args{d: &DeviceType{
				macAddress:      "E8:D8:D1:47:55:93",
				radius:          "false",
				server:          "dhcp_lan",
				status:          "bound",
				Manual:          true,
				ShouldBeBlocked: false,
				timeout:         now,
			}},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ds.findIndexOfDevice(tt.args.d); got != tt.want {
				t.Errorf("Test(%s)DevicesType.find() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestTransport_readsStreamFromMT(t *testing.T) {
	type args struct {
		cfg *Config
	}
	tests := []struct {
		name      string
		transport *Transport
		args      args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.transport.readsStreamFromMT(tt.args.cfg)
		})
	}
}
