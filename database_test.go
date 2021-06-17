package main

import (
	"path"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestGetDayStat(t *testing.T) {
	cfg := newConfig()
	type args struct {
		y        int
		m        int
		d        int
		fileName string
	}
	tests := []struct {
		name string
		args args
		want map[KeyDevice]StatDeviceType
	}{
		{
			name: "1",
			args: args{2021, 6, 16, path.Join(cfg.ConfigPath, "sqlite.db")},
			want: map[KeyDevice]StatDeviceType{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDayStat(tt.args.y, tt.args.m, tt.args.d, tt.args.fileName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDayStat() = %v, want %v", got, tt.want)
			}
		})
	}
}
