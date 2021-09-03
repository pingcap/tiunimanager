package framework

import (
	"testing"
)

func TestAssert(t *testing.T) {
	type args struct {
		b bool
	}
	tests := []struct {
		name      string
		args      args
		withPanic bool
	}{
		{"true", args{true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Assert(tt.args.b)
		})
	}
}

func TestAssertNoErr(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name      string
		args      args
		withPanic bool
	}{
		{"true", args{nil}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AssertNoErr(tt.args.err)
		})
	}
}

func TestAssertWithInfo(t *testing.T) {
	type args struct {
		b    bool
		info string
	}
	tests := []struct {
		name      string
		args      args
		withPanic bool
	}{
		{"true", args{true, "sdf"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AssertWithInfo(tt.args.b, tt.args.info)
		})
	}
}
