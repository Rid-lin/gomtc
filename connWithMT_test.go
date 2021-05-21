package main

import (
	"reflect"
	"testing"
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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuotahourly, gotQuotadaily, gotQuotamonthly, gotName, gotPosition, gotCompany, gotTypeD, gotIDUser, gotComment, gotManual := parseComment(tt.args.comment)
			if gotQuotahourly != tt.wantQuotahourly {
				t.Errorf("parseComment() gotQuotahourly = %v, want %v", gotQuotahourly, tt.wantQuotahourly)
			}
			if gotQuotadaily != tt.wantQuotadaily {
				t.Errorf("parseComment() gotQuotadaily = %v, want %v", gotQuotadaily, tt.wantQuotadaily)
			}
			if gotQuotamonthly != tt.wantQuotamonthly {
				t.Errorf("parseComment() gotQuotamonthly = %v, want %v", gotQuotamonthly, tt.wantQuotamonthly)
			}
			if gotName != tt.wantName {
				t.Errorf("parseComment() gotName = %v, want %v", gotName, tt.wantName)
			}
			if gotPosition != tt.wantPosition {
				t.Errorf("parseComment() gotPosition = %v, want %v", gotPosition, tt.wantPosition)
			}
			if gotCompany != tt.wantCompany {
				t.Errorf("parseComment() gotCompany = %v, want %v", gotCompany, tt.wantCompany)
			}
			if gotTypeD != tt.wantTypeD {
				t.Errorf("parseComment() gotTypeD = %v, want %v", gotTypeD, tt.wantTypeD)
			}
			if gotIDUser != tt.wantIDUser {
				t.Errorf("parseComment() gotIDUser = %v, want %v", gotIDUser, tt.wantIDUser)
			}
			if gotComment != tt.wantComment {
				t.Errorf("parseComment() gotComment = %v, want %v", gotComment, tt.wantComment)
			}
			if gotManual != tt.wantManual {
				t.Errorf("parseComment() gotManual = %v, want %v", gotManual, tt.wantManual)
			}
		})
	}
}

func TestDeviceType_convertToInfo(t *testing.T) {
	tests := []struct {
		name string
		d    DeviceType
		want InfoOfDeviceType
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.convertToInfo(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeviceType.convertToInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInfoOfDeviceType_convertToDevice(t *testing.T) {
	tests := []struct {
		name  string
		dInfo *InfoOfDeviceType
		want  DeviceType
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dInfo.convertToDevice(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InfoOfDeviceType.convertToDevice() = %v, want %v", got, tt.want)
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d1.compare(tt.args.d2); got != tt.want {
				t.Errorf("DeviceType.compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDevicesType_find(t *testing.T) {
	type args struct {
		d *DeviceType
	}
	tests := []struct {
		name string
		ds   *DevicesType
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ds.find(tt.args.d); got != tt.want {
				t.Errorf("DevicesType.find() = %v, want %v", got, tt.want)
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
