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
