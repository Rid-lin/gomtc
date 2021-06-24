package main

import (
	"testing"
)

// var (
// 	DS DevicesType
// )

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
