package main

import (
	"reflect"
	"testing"
	"time"
)

func TestTransport_getStatusDevices(t *testing.T) {
	type args struct {
		cfg *Config
	}
	tests := []struct {
		name      string
		transport *Transport
		args      args
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.transport.getStatusDevices(tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("Transport.getStatusDevices() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTransport_GetData(t *testing.T) {
	transport := NewTransport(&Config{
		MTAddr: "192.168.65.1:8728",
		MTUser: "getmac",
		MTPass: "getmac_password",
		NoFlow: true,
	})

	data := MapOfReports{}

	key := KeyMapOfReports{
		Alias:   "E8:D8:D1:47:55:93",
		DateStr: time.Now().In(transport.Location).Format("2006-01-02"),
	}
	transport.dataCashe[key] = ValueMapOfReportsType{
		Alias:   "E8:D8:D1:47:55:93",
		DateStr: time.Now().In(transport.Location).Format("2006-01-02"),
		Hits:    3,
	}

	type args struct {
		key KeyMapOfReports
		// alias string
	}
	tests := []struct {
		name      string
		transport *Transport
		args      args
		want      ValueMapOfReportsType
		wantErr   bool
	}{
		{
			name:      "1",
			transport: transport,
			args: args{
				key: KeyMapOfReports{
					Alias: "E8:D8:D1:47:55:93",
				},
			},
			want:    data[key],
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.transport.GetData(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Transport.GetData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Transport.GetData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkNULLQuota(t *testing.T) {
	type args struct {
		setValue     uint64
		deafultValue uint64
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkNULLQuota(tt.args.setValue, tt.args.deafultValue); got != tt.want {
				t.Errorf("checkNULLQuota() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransport_checkQuota(t *testing.T) {
	tests := []struct {
		name      string
		transport *Transport
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.transport.checkQuotas()
		})
	}
}

func TestTransport_updateStatusDevicesToMT(t *testing.T) {
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
			tt.transport.updateStatusDevicesToMT(tt.args.cfg)
		})
	}
}
