package domain

import (
	"testing"
)

func TestCommonStatusFromStatus(t *testing.T) {
	type args struct {
		status int32
	}
	tests := []struct {
		name string
		args args
		want CommonStatus
	}{
		{"valid", args{0}, Valid},
		{"Invalid", args{1}, Invalid},
		{"Deleted", args{2}, Deleted},
		{"UnrecognizedStatus", args{99}, UnrecognizedStatus},
		{"UnrecognizedStatus", args{-1}, UnrecognizedStatus},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CommonStatusFromStatus(tt.args.status); got != tt.want {
				t.Errorf("CommonStatusFromStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommonStatus_IsValid(t *testing.T) {
	tests := []struct {
		name string
		s    CommonStatus
		want bool
	}{
		{"normal", Valid, true},
		{"invalid", Invalid, false},
		{"Deleted", Deleted, false},
		{"UnrecognizedStatus", UnrecognizedStatus, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
