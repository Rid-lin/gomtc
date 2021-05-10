package main

import (
	"reflect"
	"testing"

	"github.com/go-routeros/routeros"
)

func Test_dial(t *testing.T) {
	type args struct {
		cfg *Config
	}
	tests := []struct {
		name    string
		args    args
		want    *routeros.Client
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dial(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("dial() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dial() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tryingToReconnectToMokrotik(t *testing.T) {
	type args struct {
		cfg *Config
	}
	tests := []struct {
		name string
		args args
		want *routeros.Client
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tryingToReconnectToMokrotik(tt.args.cfg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tryingToReconnectToMokrotik() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransport_GetInfo(t *testing.T) {
	type args struct {
		request *request
	}
	tests := []struct {
		name string
		data *Transport
		args args
		want ResponseType
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.data.GetInfo(tt.args.request); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Transport.GetInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransport_loopGetDataFromMT(t *testing.T) {
	tests := []struct {
		name string
		data *Transport
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.data.loopGetDataFromMT()
		})
	}
}

func TestTransport_updateDataFromMT(t *testing.T) {
	tests := []struct {
		name string
		data *Transport
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.data.updateDataFromMT()
		})
	}
}

func Test_getDataFromMT(t *testing.T) {
	type args struct {
		quota   QuotaType
		connRos *routeros.Client
	}
	tests := []struct {
		name string
		args args
		want map[string]InfoOfDeviceType
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDataFromMT(tt.args.quota, tt.args.connRos); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDataFromMT() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseComments(t *testing.T) {
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
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuotahourly, gotQuotadaily, gotQuotamonthly, gotName, gotPosition, gotCompany, gotTypeD, gotIDUser := parseComments(tt.args.comment)
			if gotQuotahourly != tt.wantQuotahourly {
				t.Errorf("parseComments() gotQuotahourly = %v, want %v", gotQuotahourly, tt.wantQuotahourly)
			}
			if gotQuotadaily != tt.wantQuotadaily {
				t.Errorf("parseComments() gotQuotadaily = %v, want %v", gotQuotadaily, tt.wantQuotadaily)
			}
			if gotQuotamonthly != tt.wantQuotamonthly {
				t.Errorf("parseComments() gotQuotamonthly = %v, want %v", gotQuotamonthly, tt.wantQuotamonthly)
			}
			if gotName != tt.wantName {
				t.Errorf("parseComments() gotName = %v, want %v", gotName, tt.wantName)
			}
			if gotPosition != tt.wantPosition {
				t.Errorf("parseComments() gotPosition = %v, want %v", gotPosition, tt.wantPosition)
			}
			if gotCompany != tt.wantCompany {
				t.Errorf("parseComments() gotCompany = %v, want %v", gotCompany, tt.wantCompany)
			}
			if gotTypeD != tt.wantTypeD {
				t.Errorf("parseComments() gotTypeD = %v, want %v", gotTypeD, tt.wantTypeD)
			}
			if gotIDUser != tt.wantIDUser {
				t.Errorf("parseComments() gotIDUser = %v, want %v", gotIDUser, tt.wantIDUser)
			}
		})
	}
}

func Test_parseParamertToStr(t *testing.T) {
	type args struct {
		inpuStr string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseParamertToStr(tt.args.inpuStr); got != tt.want {
				t.Errorf("parseParamertToStr() = %v, want %v", got, tt.want)
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotQuota := parseParamertToUint(tt.args.inputValue); gotQuota != tt.wantQuota {
				t.Errorf("parseParamertToUint() = %v, want %v", gotQuota, tt.wantQuota)
			}
		})
	}
}

func TestTransport_syncStatusDevices(t *testing.T) {
	type args struct {
		inputSync map[string]bool
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
			tt.transport.syncStatusDevices(tt.args.inputSync)
		})
	}
}

func TestTransport_setStatusDevice(t *testing.T) {
	type args struct {
		number string
		status bool
	}
	tests := []struct {
		name    string
		data    *Transport
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.data.setStatusDevice(tt.args.number, tt.args.status); (err != nil) != tt.wantErr {
				t.Errorf("Transport.setStatusDevice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTransport_getInfoOfDeviceFromMT(t *testing.T) {
	type args struct {
		alias string
	}
	cfg := Config{
		NameFileToLog:       "./logs/access.log",
		MTAddr:              "192.168.65.1:8728",
		MTUser:              "getmac",
		MTPass:              "getmac_password",
		DefaultQuotaHourly:  50000000,
		DefaultQuotaDaily:   300000000,
		DefaultQuotaMonthly: 9000000000,
	}
	data := NewTransport(&cfg)
	data.DailyQuota = cfg.DefaultQuotaDaily
	data.HourlyQuota = cfg.DefaultQuotaHourly
	data.MonthlyQuota = cfg.DefaultQuotaMonthly
	defer data.fileDestination.Close()
	tests := []struct {
		name string
		data *Transport
		args args
		want InfoOfDeviceType
	}{
		{name: "1",
			data: data,
			args: args{alias: "E8:D8:D1:47:55:93"},
			want: InfoOfDeviceType{
				DeviceType: DeviceType{
					Id:           "*E6FF8",
					IP:           "192.168.65.85",
					TypeD:        "nb",
					Mac:          "E8:D8:D1:47:55:93",
					AMac:         "E8:D8:D1:47:55:93",
					HostName:     "root-hp",
					Groups:       "inet_over_vpn",
					AddressLists: []string{"inet_over_vpn"}},
				PersonType: PersonType{
					Comments: "nb=Admin/quotahourly=500000000/quotadaily=50000000000",
					Name:     "Admin",
					Position: "",
					Company:  "",
					IDUser:   ""},
				QuotaType: QuotaType{
					HourlyQuota:  0x1dcd6500,
					DailyQuota:   0xba43b7400,
					MonthlyQuota: 0x218711a00,
					Blocked:      false},
			},
		},
		{name: "2",
			data: data,
			args: args{alias: "192.168.65.85"},
			want: InfoOfDeviceType{
				DeviceType: DeviceType{
					Id:           "*E6FF8",
					IP:           "192.168.65.85",
					TypeD:        "nb",
					Mac:          "E8:D8:D1:47:55:93",
					AMac:         "E8:D8:D1:47:55:93",
					HostName:     "root-hp",
					Groups:       "inet_over_vpn",
					AddressLists: []string{"inet_over_vpn"}},
				PersonType: PersonType{
					Comments: "nb=Admin/quotahourly=500000000/quotadaily=50000000000",
					Name:     "Admin",
					Position: "",
					Company:  "",
					IDUser:   ""},
				QuotaType: QuotaType{
					HourlyQuota:  0x1dcd6500,
					DailyQuota:   0xba43b7400,
					MonthlyQuota: 0x218711a00,
					Blocked:      false},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.data.getInfoOfDeviceFromMT(tt.args.alias); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Transport.getInfoOfDeviceFromMT() = %#v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransport_aliasToDevice(t *testing.T) {
	type args struct {
		alias string
	}
	tests := []struct {
		name string
		data *Transport
		args args
		want DeviceType
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.data.aliasToDevice(tt.args.alias); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Transport.aliasToDevice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isMac(t *testing.T) {
	type args struct {
		inputStr string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMac(tt.args.inputStr); got != tt.want {
				t.Errorf("isMac() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isIP(t *testing.T) {
	type args struct {
		inputStr string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIP(tt.args.inputStr); got != tt.want {
				t.Errorf("isIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransport_findInfoOfDevice(t *testing.T) {
	type args struct {
		alias string
	}
	tests := []struct {
		name      string
		transport *Transport
		args      args
		want      DeviceType
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.transport.findInfoOfDevice(tt.args.alias)
			if (err != nil) != tt.wantErr {
				t.Errorf("Transport.findInfoOfDevice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Transport.findInfoOfDevice() = %v, want %v", got, tt.want)
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