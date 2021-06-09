package main

import (
	"testing"
)

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

// func TestTransport_checkQuota(t *testing.T) {
// 	tests := []struct {
// 		name      string
// 		transport *Transport
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.transport.checkQuotas()
// 		})
// 	}
// }

// func TestTransport_updateStatusDevicesToMT(t *testing.T) {
// 	type args struct {
// 		cfg *Config
// 	}
// 	tests := []struct {
// 		name      string
// 		transport *Transport
// 		args      args
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.transport.updateStatusDevicesToMT(tt.args.cfg)
// 		})
// 	}
// }

func Test_setDailyTimeout(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "1",
			want: "00:00:150",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := setDailyTimeout(); got != tt.want {
				t.Errorf("setDailyTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setHourlyTimeout(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "1",
			want: "00:00:150",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := setHourlyTimeout(); got != tt.want {
				t.Errorf("setHourlyTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}
