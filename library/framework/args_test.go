package framework

import (
	"github.com/micro/cli/v2"
	"testing"
)

func TestAllFlags(t *testing.T) {
	type args struct {
		receiver *ClientArgs
	}
	tests := []struct {
		name string
		args args
		asserts []func(receiver *ClientArgs, fs []cli.Flag) bool
	}{
		{"normal", args{&ClientArgs{}}, []func(receiver *ClientArgs, fs []cli.Flag) bool{
			func(receiver *ClientArgs, fs []cli.Flag) bool {return len(fs) > 4},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AllFlags(tt.args.receiver)
			for i,v := range tt.asserts {
				if !v(tt.args.receiver, got) {
					t.Errorf("AllFlags assert false, assert index = %v, got = %v", i, got)
				}
			}

		})
	}
}
