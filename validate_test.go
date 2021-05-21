package main

import (
	"testing"
)

func Test_validateMac(t *testing.T) {
	type args struct {
		mac         string
		altMac      string
		hopeMac     string
		lastHopeMac string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "1 The first one is correct",
			args: args{
				mac:         "F6:5E:B9:A2:D4:51",
				altMac:      "F6:5E:B9:A2:D4:51",
				hopeMac:     "F6:5E:B9:A2:D4:51",
				lastHopeMac: "F6:5E:B9:A2:D4:521",
			},
			want: "F6:5E:B9:A2:D4:51",
		},
		{
			name: "2",
			args: args{
				mac:         "",
				altMac:      "F6:5E:B9:A2:D4:51",
				hopeMac:     "F6:5E:B9:A2:D4:51",
				lastHopeMac: "F6:5E:B9:A2:D4:51",
			},
			want: "F6:5E:B9:A2:D4:51",
		},
		{
			name: "3",
			args: args{
				mac:         "",
				altMac:      "",
				hopeMac:     "1:F6:5E:B9:A2:D4:53",
				lastHopeMac: "1:F6:5E:B9:A2:D4:53",
			},
			want: "F6:5E:B9:A2:D4:53",
		},
		{
			name: "4",
			args: args{
				mac:         "",
				altMac:      "",
				hopeMac:     "",
				lastHopeMac: "1:F6:5E:B9:A2:D4:54",
			},
			want: "F6:5E:B9:A2:D4:54",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateMac(tt.args.mac, tt.args.altMac, tt.args.hopeMac, tt.args.lastHopeMac); got != tt.want {
				t.Errorf("validateMac() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateIP(t *testing.T) {
	type args struct {
		ip    string
		altIp string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "The first one is correct",
			args: args{
				ip:    "192.168.65.40",
				altIp: "192.168.65.41",
			},
			want: "192.168.65.40",
		},
		{
			name: "The second is correct, the first is empty.",
			args: args{
				ip:    "",
				altIp: "192.168.65.41",
			},
			want: "192.168.65.41",
		},
		{
			name: "The first one is empty. The second is not an IP",
			args: args{
				ip:    "",
				altIp: "192.168.65.41ю",
			},
			want: "",
		},
		{
			name: "The first one is empty. The second is not an IP",
			args: args{
				ip:    "",
				altIp: "192.168.65.41.",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateIP(tt.args.ip, tt.args.altIp); got != tt.want {
				t.Errorf("validateIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isHexColon(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "1",
			args: args{
				s: "00",
			},
			want: true,
		},
		{
			name: "2",
			args: args{
				s: "a0",
			},
			want: true,
		},
		{
			name: "3",
			args: args{
				s: "A0",
			},
			want: true,
		},
		{
			name: "4",
			args: args{
				s: "09",
			},
			want: true,
		},
		{
			name: "5",
			args: args{
				s: "0F",
			},
			want: true,
		},
		{
			name: "6",
			args: args{
				s: "0f",
			},
			want: true,
		},
		{
			name: "7",
			args: args{
				s: "ff",
			},
			want: true,
		},
		{
			name: "8",
			args: args{
				s: "FF",
			},
			want: true,
		},
		{
			name: "9",
			args: args{
				s: "FF0",
			},
			want: false,
		},
		{
			name: "10",
			args: args{
				s: "g0",
			},
			want: false,
		},
		{
			name: "11",
			args: args{
				s: "0",
			},
			want: false,
		},
		{
			name: "12",
			args: args{
				s: "0g",
			},
			want: false,
		},
		{
			name: "7",
			args: args{
				s: "gg",
			},
			want: false,
		},
		{
			name: "7",
			args: args{
				s: "zz",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isHexColon(tt.args.s); got != tt.want {
				t.Errorf("isHexColon() = %v, want %v", got, tt.want)
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
		{
			name: "1-true",
			args: args{
				inputStr: "192.168.0.1",
			},
			want: true,
		},
		{
			name: "2-true",
			args: args{
				inputStr: "0.0.0.0",
			},
			want: true,
		},
		{
			name: "3-true",
			args: args{
				inputStr: "10.0.0.1",
			},
			want: true,
		},
		{
			name: "1-false",
			args: args{
				inputStr: "10.0.0.1000",
			},
			want: false,
		},
		{
			name: "2-false",
			args: args{
				inputStr: "10.0.0.1.",
			},
			want: false,
		},
		{
			name: "3-false",
			args: args{
				inputStr: "10.0.0.",
			},
			want: false,
		},
		{
			name: "4-false",
			args: args{
				inputStr: "10.0.0.255",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIP(tt.args.inputStr); got != tt.want {
				t.Errorf("isIP() = %v, want %v(s:%s)", got, tt.want, tt.args.inputStr)
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
		{
			name: "1",
			args: args{
				inputStr: "00:00:00:00:00:00",
			},
			want: true,
		},
		{
			name: "2",
			args: args{
				inputStr: "00:00:00:00:00:00:",
			},
			want: false,
		},
		{
			name: "3",
			args: args{
				inputStr: "000:00:00:00:00:00",
			},
			want: false,
		},
		{
			name: "4",
			args: args{
				inputStr: "0G:00:00:00:00:00",
			},
			want: false,
		},
		{
			name: "5",
			args: args{
				inputStr: "0:00:00:00:00:00",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMac(tt.args.inputStr); got != tt.want {
				t.Errorf("isMac() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isNumDot(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "1",
			args: args{
				s: "300",
			},
			want: true,
		},
		{
			name: "2",
			args: args{
				s: "300.0",
			},
			want: true,
		},
		{
			name: "3",
			args: args{
				s: "1111300.0",
			},
			want: true,
		},
		{
			name: "4",
			args: args{
				s: "1111300.0.",
			},
			want: false,
		},
		{
			name: "4",
			args: args{
				s: "a1111300.0",
			},
			want: false,
		},
		{
			name: "5",
			args: args{
				s: "kp",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNumDot(tt.args.s); got != tt.want {
				t.Errorf("isNumDot() = %v, want %v", got, tt.want)
			}
		})
	}
}
