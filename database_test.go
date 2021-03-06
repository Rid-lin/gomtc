package main

import (
	"reflect"
	"testing"

	"git.vegner.org/vsvegner/gomtc/internal/app/model"
	. "git.vegner.org/vsvegner/gomtc/internal/config"

	_ "github.com/mattn/go-sqlite3"
)

func TestGetDayStat(t *testing.T) {
	cfg, _ := NewConfig()
	type args struct {
		from     string
		to       string
		fileName string
	}
	tests := []struct {
		name string
		args args
		want map[model.KeyDevice]model.StatDeviceType
	}{
		{
			name: "1",
			args: args{"2021-06-16", "2021-06-16", cfg.DSN},
			want: map[model.KeyDevice]model.StatDeviceType{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDayStat(tt.args.from, tt.args.to, tt.args.fileName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDayStat() = %v, want %v", got, tt.want)
			}
		})
	}
}
