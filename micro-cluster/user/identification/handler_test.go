package identification

import (
	"context"
	"github.com/pingcap-inc/tiem/message"
	"reflect"
	"testing"
)

func TestProvide(t *testing.T) {
	type args struct {
		ctx     context.Context
		request message.ProvideTokenReq
	}
	tests := []struct {
		name    string
		args    args
		want    message.ProvideTokenResp
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Provide(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Provide() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Provide() got = %v, want %v", got, tt.want)
			}
		})
	}
}
